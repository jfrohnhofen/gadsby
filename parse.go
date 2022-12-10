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
	Title         *string `xml:"title"`
	Subject       *string `xml:"subject"`
	Creator       *string `xml:"creator"`
	Keywords      string  `xml:"keywords"`
	Description   *string `xml:"description"`
	Category      *string `xml:"category"`
	ContentStatus *string `xml:"contentStatus"`
}

type property struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"lpwstr"`
}

type properties struct {
	Properties []property `xml:"property"`
}

func ParseFile(path string) (Document, error) {
	log.Println("Parsing file", path)

	zipFile, err := zip.OpenReader(path)
	if err != nil {
		return Document{}, err
	}
	defer zipFile.Close()

	coreXml, err := zipFile.Open("docProps/core.xml")
	if err != nil {
		return Document{}, err
	}
	defer coreXml.Close()

	coreProps := coreProperties{}
	coreDecoder := xml.NewDecoder(coreXml)
	coreDecoder.Decode(&coreProps)

	customXml, err := zipFile.Open("docProps/custom.xml")
	if err != nil {
		return Document{}, err
	}
	defer customXml.Close()

	customDecoder := xml.NewDecoder(customXml)
	props := properties{}
	customDecoder.Decode(&props)

	customProps := map[string]*string{}
	for _, prop := range props.Properties {
		value := prop.Value
		customProps[prop.Name] = &value
	}

	keywords := []string{}
	for _, keyword := range strings.FieldsFunc(coreProps.Keywords, func(r rune) bool { return r == ',' || r == ';' }) {
		keyword = strings.TrimSpace(keyword)
		if keyword != "" {
			keywords = append(keywords, keyword)
		}
	}

	if customProps["Datum"] != nil {
		if _, err := time.Parse("02.01.2006", *customProps["Datum"]); err != nil {
			return Document{}, err
		}
	}

	documentXml, err := zipFile.Open("word/document.xml")
	if err != nil {
		return Document{}, err
	}
	defer documentXml.Close()
	content, err := extractText(xml.NewDecoder(documentXml))
	if err != nil {
		return Document{}, err
	}

	comments := []string{}
	if coreProps.Description != nil {
		comments = strings.Split(*coreProps.Description, "\n")
	}

	return Document{
		Reference:    customProps["Aktenzeichen"],
		DocumentType: customProps["DokumententypVisJustiz"],
		Date:         customProps["Datum"],
		Decision:     coreProps.ContentStatus,
		AuthorType:   coreProps.Category,
		Author:       coreProps.Creator,
		Area:         coreProps.Subject,
		Subject:      coreProps.Title,
		Keywords:     keywords,
		Comments:     comments,
		Content:      content,
		Path:         path,
	}, nil
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
