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

type Request struct {
	Query string `json:"query"`
	Tags  []Tag  `json:"tags"`
}

func main() {
	index, err := NewIndex(".")
	if err != nil {
		log.Fatalf("building index: %s", err)
	}

	app := fiber.New()

	app.Use("/", filesystem.New(filesystem.Config{
		Root: http.Dir("./static/"),
		//Root:       http.FS(static),
		//PathPrefix: "static",
	}))

	app.Get("/tags", func(ctx *fiber.Ctx) error {
		ctx.Status(200)
		return ctx.JSON(index.Tags())
	})

	app.Post("/search", func(ctx *fiber.Ctx) error {
		req := Request{}
		if err := ctx.BodyParser(&req); err != nil {
			return fmt.Errorf("decode request: %w", err)
		}
		results, err := index.Query(req.Query, req.Tags)
		if err != nil {
			return err
		}
		ctx.Status(200)
		return ctx.JSON(results)
	})

	app.Get("/download/:id", func(ctx *fiber.Ctx) error {
		id, err := ctx.ParamsInt("id")
		if err != nil {
			return fmt.Errorf("get id: %s", err)
		}
		documentPath := index.Document(uint64(id)).Path
		ctx.Status(200)
		ctx.Set(fiber.HeaderContentDisposition, fmt.Sprintf(`attachment; filename="%s"`, path.Base(documentPath)))
		return ctx.SendFile(documentPath)
	})

	log.Fatal(app.Listen(fmt.Sprintf(":%d", port)))
}

/*
func renderSnippet(content string, locs []search.FieldTermLocation) string {
	snippets := []string{}
	if len(locs) > 10 {
		locs = locs[:10]
	}
	for _, loc := range locs {
		if loc.Field != "content" {
			continue
		}

		start := loc.Location.Start
		for ; content[start] != '\n' && content[start] != '.' && start > 0; start-- {
		}

		end := loc.Location.End
		for ; content[end] != '\n' && content[end] != '.' && end < len(content)-1; end++ {
		}

		snippets = append(snippets, fmt.Sprintf(
			"%s<b>%s</b>%s",
			content[start:loc.Location.Start],
			content[loc.Location.Start:loc.Location.End],
			content[loc.Location.End:end]))
	}
	return strings.Join(snippets, " ... ")
}
*/
