package main

import (
	"fmt"
)

var FullTextFields = []string{"content", "subject", "comment"}

type Tag struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func (tag Tag) String() string {
	return fmt.Sprintf("%s$%s", tag.Key, tag.Value)
}

type Document struct {
	Id           uint64   `json:"id"`
	Reference    *string  `json:"reference"`
	DocumentType *string  `json:"documentType"`
	Date         *string  `json:"date"`
	Decision     *string  `json:"decision"`
	AuthorType   *string  `json:"authorType"`
	Author       *string  `json:"author"`
	Area         *string  `json:"area"`
	Subject      *string  `json:"subject"`
	Keywords     []string `json:"keywords"`
	Comments     []string `json:"comments"`
	Content      string   `json:"-"`
	Path         string   `json:"-"`
}

func (doc Document) Tags() []Tag {
	tags := []Tag{}
	if doc.DocumentType != nil {
		tags = append(tags, Tag{"Entscheidungsform", *doc.DocumentType})
	}
	if doc.Decision != nil {
		tags = append(tags, Tag{"Entscheidung", *doc.Decision})
	}
	if doc.AuthorType != nil {
		tags = append(tags, Tag{"Kammer/ERi", *doc.AuthorType})
	}
	if doc.Author != nil {
		tags = append(tags, Tag{"BE/ERi", *doc.Author})
	}
	if doc.Subject != nil {
		tags = append(tags, Tag{"Sachgebiet", *doc.Area})
	}
	for _, keyword := range doc.Keywords {
		tags = append(tags, Tag{"Schlagwort", keyword})
	}
	return tags
}

func (doc Document) FullTextFields() map[string][]string {
	subject := []string{}
	if doc.Subject != nil {
		subject = []string{*doc.Subject}
	}
	return map[string][]string{
		"content": []string{doc.Content},
		"subject": subject,
		"comment": doc.Comments,
	}
}
