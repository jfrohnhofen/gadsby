package main

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"strings"
	"time"
)

type coreProperties struct {
	Title         string `xml:"title"`
	Subject       string `xml:"subject"`
	Creator       string `xml:"creator"`
	Keywords      string `xml:"keywords"`
	Description   string `xml:"description"`
	Category      string `xml:"category"`
	ContentStatus string `xml:"contentStatus"`
}

type property struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"lpwstr"`
}

type properties struct {
	Properties []property `xml:"property"`
}

func ParseDocument(path string) (Document, []Tag, string, error) {
	log.Println("Parsing document", path)

	zipFile, err := zip.OpenReader(path)
	if err != nil {
		return Document{}, nil, "", fmt.Errorf("reading zip file: %w", err)
	}
	defer zipFile.Close()

	coreXml, err := zipFile.Open("docProps/core.xml")
	if err != nil {
		return Document{}, nil, "", fmt.Errorf("reading core.xml: %w", err)
	}
	defer coreXml.Close()

	coreProps := coreProperties{}
	coreDecoder := xml.NewDecoder(coreXml)
	coreDecoder.Decode(&coreProps)

	customXml, err := zipFile.Open("docProps/custom.xml")
	if err != nil {
		return Document{}, nil, "", fmt.Errorf("reading custom.xml: %w", err)
	}
	defer customXml.Close()

	customDecoder := xml.NewDecoder(customXml)
	props := properties{}
	customDecoder.Decode(&props)

	customProps := map[string]string{}
	for _, prop := range props.Properties {
		value := prop.Value
		customProps[prop.Name] = value
	}

	if _, err := time.Parse("02.01.2006", customProps["Datum"]); err != nil {
		return Document{}, nil, "", fmt.Errorf("parsing data: %w", err)
	}

	areaParts := []string{}
	for _, part := range strings.FieldsFunc(coreProps.Subject, func(r rune) bool { return r == ',' || r == ';' }) {
		part = strings.TrimSpace(part)
		if part != "" {
			areaParts = append(areaParts, part)
		}
	}

	keywords := []string{}
	for _, keyword := range strings.FieldsFunc(coreProps.Keywords, func(r rune) bool { return r == ',' || r == ';' }) {
		keyword = strings.TrimSpace(keyword)
		if keyword != "" {
			keywords = append(keywords, keyword)
		}
	}

	doc := Document{
		Reference:    customProps["Aktenzeichen"],
		DocumentType: customProps["DokumententypVisJustiz"],
		Date:         customProps["Datum"],
		Decision:     coreProps.ContentStatus,
		AuthorType:   coreProps.Category,
		Author:       coreProps.Creator,
		Area:         strings.Join(areaParts, " &#x25b8; "),
		Subject:      coreProps.Title,
		Keywords:     keywords,
		Comments:     strings.Split(coreProps.Description, "\n"),
		Path:         path,
	}

	tags := []Tag{}
	if doc.Reference != "" {
		tags = append(tags, Tag{"Aktenzeichen", doc.Reference})
	}
	if doc.DocumentType != "" {
		tags = append(tags, Tag{"Entscheidungsform", doc.DocumentType})
	}
	if doc.Decision != "" {
		tags = append(tags, Tag{"Entscheidung", doc.Decision})
	}
	if doc.AuthorType != "" && doc.Author != "" {
		tags = append(tags, Tag{doc.AuthorType, doc.Author})
	}
	for _, keyword := range doc.Keywords {
		tags = append(tags, Tag{"Schlagwort", keyword})
	}
	for i := range areaParts {
		tags = append(tags, Tag{"Sachgebiet", strings.Join(areaParts[:i+1], " &#x25B8; ")})
	}

	documentXml, err := zipFile.Open("word/document.xml")
	if err != nil {
		return Document{}, nil, "", fmt.Errorf("reading document.xml: %w", err)
	}
	defer documentXml.Close()

	text, err := extractText(xml.NewDecoder(documentXml))
	if err != nil {
		return Document{}, nil, "", fmt.Errorf("extracting text: %w", err)
	}

	content := strings.Join(append(doc.Comments, doc.Subject, text), "\n")

	return doc, tags, content, nil
}

func extractText(decoder *xml.Decoder) (string, error) {
	text := ""
	expectText := false

	token, err := decoder.Token()
	for ; err == nil; token, err = decoder.Token() {
		if expectText {
			data, ok := token.(xml.CharData)
			if !ok {
				return "", fmt.Errorf("expected character data")
			}
			text = text + string(data) + " "
			expectText = false
		} else {
			elem, ok := token.(xml.StartElement)
			expectText = ok && elem.Name.Local == "t"
		}
	}
	if err != io.EOF {
		return "", err
	}
	return text, nil
}
