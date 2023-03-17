package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/blugelabs/bluge"
)

type Document struct {
	Id           int      `json:"id"`
	Reference    string   `json:"reference"`
	DocumentType string   `json:"documentType"`
	Date         string   `json:"date"`
	Decision     string   `json:"decision"`
	AuthorType   string   `json:"authorType"`
	Author       string   `json:"author"`
	Area         string   `json:"area"`
	Subject      string   `json:"subject"`
	Keywords     []string `json:"keywords"`
	Comments     []string `json:"comments"`
	Score        float64  `json:"score"`
	Path         string   `json:"-"`
}

type Index struct {
	reader    *bluge.Reader
	tags      map[Tag][]int
	documents []Document
}

func BuildIndex(dataPath string) (Index, error) {
	log.Println("Building index...")

	index := Index{tags: map[Tag][]int{}}

	writer, err := bluge.OpenWriter(bluge.InMemoryOnlyConfig())
	if err != nil {
		return Index{}, fmt.Errorf("create writer: %w", err)
	}

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
		doc, tags, content, err := ParseDocument(filePath)
		if err != nil {
			log.Printf("Failed to parse document %s: %s\n", info.Name(), err)
			return nil
		}

		doc.Id = len(index.documents)
		index.documents = append(index.documents, doc)
		for _, tag := range tags {
			index.tags[tag] = append(index.tags[tag], doc.Id)
		}

		blugeDoc := bluge.Document{
			bluge.NewStoredOnlyField("id", []byte(strconv.Itoa(doc.Id))),
			bluge.NewTextField("content", content),
		}
		return writer.Insert(blugeDoc)
	}); err != nil {
		return index, fmt.Errorf("index document: %w", err)
	}

	index.reader, err = writer.Reader()
	if err != nil {
		return index, fmt.Errorf("create reader: %w", err)
	}

	return index, nil
}

func (index Index) GetTags() []Tag {
	tags := []Tag{}
	for tag := range index.tags {
		tags = append(tags, tag)
	}
	return tags
}

func (index Index) GetDocumentPath(id int) (string, error) {
	if id < len(index.documents) {
		return index.documents[id].Path, nil
	}
	return "", fmt.Errorf("out of range")
}

func (index *Index) Query(query string, tags []Tag) ([]Document, error) {
	matches := map[int]int{}
	for _, tag := range tags {
		for _, id := range index.tags[tag] {
			matches[id]++
		}
	}

	var request bluge.SearchRequest
	if query == "" {
		request = bluge.NewAllMatches(bluge.NewMatchAllQuery())
	} else {
		blugeQuery := bluge.NewBooleanQuery()
		for _, term := range strings.Split(query, " ") {
			term = strings.ToLower(term)
			blugeQuery.AddShould(bluge.NewPrefixQuery(term).SetField("content"))
		}
		request = bluge.NewAllMatches(blugeQuery)
	}

	result, err := index.reader.Search(context.Background(), request)
	if err != nil {
		return nil, fmt.Errorf("bluge search: %w", err)
	}

	for match, err := result.Next(); match != nil || err != nil; match, err = result.Next() {
		if err != nil {
			return nil, fmt.Errorf("fetching result: %w", err)
		}

		match.VisitStoredFields(func(field string, value []byte) bool {
			if field == "id" {
				id, err := strconv.Atoi(string(value))
				if err == nil {
					matches[id]++
					index.documents[id].Score = match.Score
				} else {
					log.Printf("Failed to parse document id.\n")
				}
			}
			return true
		})
	}

	results := []Document{}
	for id, count := range matches {
		if count == len(tags)+1 {
			results = append(results, index.documents[id])
		}
	}
	return results, nil
}
