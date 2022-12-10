package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/blugelabs/bluge"
	"github.com/blugelabs/bluge/search"
)

type Index struct {
	reader    *bluge.Reader
	tags      map[Tag][]uint64
	documents []Document
}

type Result struct {
	Score    float64  `json:"score"`
	Document Document `json:"document"`
	Loctions []uint64 `json:"locations"`
}

func NewIndex(dataPath string) (Index, error) {
	log.Println("Building index...")

	writer, err := bluge.OpenWriter(bluge.InMemoryOnlyConfig())
	if err != nil {
		return Index{}, fmt.Errorf("creating index writer: %w", err)
	}

	documents := []Document{}
	tags := map[Tag][]uint64{}

	if err := filepath.Walk(dataPath, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if path.Ext(info.Name()) != ".docx" {
			return nil
		}
		document, err := ParseFile(filePath)
		if err != nil {
			return err
		}

		document.Id = uint64(len(documents))
		documents = append(documents, document)
		for _, tag := range document.Tags() {
			tags[tag] = append(tags[tag], document.Id)
		}

		blugeDoc := bluge.NewDocument(document.Path)
		for field, values := range document.FullTextFields() {
			for _, value := range values {
				blugeDoc.AddField(bluge.NewTextField(field, value).HighlightMatches())
			}
		}
		return writer.Insert(blugeDoc)
	}); err != nil {
		return Index{}, err
	}

	reader, err := writer.Reader()
	if err != nil {
		return Index{}, fmt.Errorf("creating index reader: %w", err)
	}

	return Index{
		reader:    reader,
		tags:      tags,
		documents: documents,
	}, nil
}

func (index Index) Tags() []Tag {
	tags := []Tag{}
	for tag := range index.tags {
		tags = append(tags, tag)
	}
	return tags
}

func (index Index) Document(id uint64) Document {
	return index.documents[id]
}

func (index *Index) Query(query string, tags []Tag) ([]Result, error) {
	posLists := [][]uint64{}
	for _, tag := range tags {
		posLists = append(posLists, index.tags[tag])
	}

	scores := map[uint64]float64{}
	locations := map[uint64][]search.FieldTermLocation{}

	if query != "" {
		blugeQuery := bluge.NewBooleanQuery()
		for _, term := range strings.Split(query, " ") {
			term = strings.ToLower(term)
			for _, field := range FullTextFields {
				blugeQuery.AddShould(bluge.NewPrefixQuery(term).SetField(field))
				blugeQuery.AddShould(bluge.NewPrefixQuery(term).SetField(field))
				blugeQuery.AddShould(bluge.NewPrefixQuery(term).SetField(field))
			}
		}
		result, err := index.reader.Search(context.Background(), bluge.NewAllMatches(blugeQuery).IncludeLocations())
		if err != nil {
			return nil, fmt.Errorf("bluge search: %w", err)
		}

		posList := []uint64{}
		for match, err := result.Next(); match != nil || err != nil; match, err = result.Next() {
			if err != nil {
				return nil, fmt.Errorf("get match: %w", err)
			}
			posList = append(posList, match.Number)
			scores[match.Number] = match.Score
			locations[match.Number] = match.FieldTermLocations
		}
		posLists = append(posLists, posList)
	}

	mergedPosLists := map[uint64]int{}
	for _, posList := range posLists {
		for _, pos := range posList {
			mergedPosLists[pos]++
		}
	}

	results := []Result{}
	for pos, count := range mergedPosLists {
		if count == len(posLists) {
			results = append(results, Result{
				Document: index.documents[pos],
				Score:    scores[pos],
			})
		}
	}

	return results, nil
}
