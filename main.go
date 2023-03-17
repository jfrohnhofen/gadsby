package main

import (
	"embed"
	_ "embed"
	"fmt"
	"log"
	"net/http"
	"path"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
)

const port = 9000

//go:embed static/*
var static embed.FS

type Tag struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Request struct {
	Query string `json:"query"`
	Tags  []Tag  `json:"tags"`
}

func main() {
	index, err := BuildIndex(".")
	if err != nil {
		log.Fatalf("build index: %s", err)
	}

	app := fiber.New()

	app.Use("/", filesystem.New(filesystem.Config{
		Root:       http.FS(static),
		PathPrefix: "static",
	}))

	app.Get("/tags", func(ctx *fiber.Ctx) error {
		ctx.Status(200)
		return ctx.JSON(index.GetTags())
	})

	app.Post("/search", func(ctx *fiber.Ctx) error {
		req := Request{}
		if err := ctx.BodyParser(&req); err != nil {
			return fmt.Errorf("decode request: %w", err)
		}
		results, err := index.Query(req.Query, req.Tags)
		if err != nil {
			return fmt.Errorf("query index: %w", err)
		}
		ctx.Status(200)
		return ctx.JSON(results)
	})

	app.Get("/download/:id", func(ctx *fiber.Ctx) error {
		id, err := ctx.ParamsInt("id")
		if err != nil {
			return fmt.Errorf("parse id: %s", err)
		}
		docPath, err := index.GetDocumentPath(id)
		if err != nil {
			return fmt.Errorf("get document: %w", err)
		}
		ctx.Status(200)
		ctx.Set(fiber.HeaderContentDisposition, fmt.Sprintf(`attachment; filename="%s"`, path.Base(docPath)))
		return ctx.SendFile(docPath)
	})

	log.Fatal(app.Listen(fmt.Sprintf(":%d", port)))
}
