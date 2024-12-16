package main

import (
	"html/template"
	"io"
	"log"

	"github.com/jadenrose/go-note/cmd/notes"
	"github.com/jadenrose/go-note/cmd/routes"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "modernc.org/sqlite"
)

type Templates struct {
	Templates *template.Template
}

func (t *Templates) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.Templates.ExecuteTemplate(w, name, data)
}

func newTemplate() *Templates {
	return &Templates{
		Templates: template.Must(template.ParseGlob("views/*.html")),
	}
}

func main() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Renderer = newTemplate()

	e.Static("/css", "css")
	e.Static("/img", "img")
	e.Static("/js", "js")

	e.GET("/", func(c echo.Context) error {
		notes, err := notes.GetAllNotes()
		if err != nil {
			log.Panic(err)
			return c.NoContent(500)
		}
		return c.Render(200, "index", notes)
	})

	e.GET("/preview-links", routes.GetPreviewLinks)
	e.GET("/notes/:note_id", routes.GetNoteContent)
	e.GET("/notes/:note_id/edit", routes.GetTitleEditor)
	e.PUT("/notes/:note_id", routes.PutTitle)

	e.GET("/blocks/new", routes.GetNewBlock)
	e.GET("/blocks/:block_id/edit", routes.GetBlockEditor)
	e.GET("/blocks/:block_id/move", routes.GetBlockMover)
	e.GET("/blocks/:block_id/move/cancel", routes.CancelBlockMover)
	e.POST("/blocks", routes.PostBlock)
	e.PUT("/blocks/:block_id", routes.PutBlock)
	e.PUT("/blocks/:block_id/move", routes.MoveBlock)
	e.DELETE("/blocks/:block_id", routes.DeleteBlock)

	// e.PUT("/blocks/:block_id/move/:direction", func(c echo.Context) error {
	// 	id, err := strconv.Atoi(c.Param("block_id"))
	// 	if err != nil {
	// 		return c.String(400, "Missing or invalid param :block_id")
	// 	}
	// 	block, err := notes.GetBlock(id)
	// 	if err != nil {
	// 		log.Panic(err)
	// 		return c.NoContent(404)
	// 	}
	// 	direction := c.Param("direction")
	// 	switch direction {
	// 	default:
	// 		return c.String(422, "Direction must be one of: up, down, top, bottom")
	// 	case "up":
	// 		err = notes.MoveBlock(block, block.SortOrder-1)
	// 	case "down":
	// 		err = notes.MoveBlock(block, block.SortOrder+1)
	// 	case "top":
	// 		err = notes.MoveBlock(block, 0)
	// 	case "bottom":
	// 		last_sort_order, err := notes.GetLastSortOrder(block)
	// 		if err != nil {
	// 			log.Panic(err)
	// 			return c.NoContent(500)
	// 		}
	// 		err = notes.MoveBlock(block, last_sort_order)
	// 		if err != nil {
	// 			log.Panic(err)
	// 			return c.NoContent(500)
	// 		}
	// 	}
	// 	if err != nil {
	// 		log.Panic(err)
	// 		return c.NoContent(500)
	// 	}
	// 	note, err := notes.GetNote(block.NoteID)
	// 	if err != nil {
	// 		log.Panic(err)
	// 		return c.NoContent(500)
	// 	}
	// 	return c.Render(200, "note-content", note)
	// })

	e.Logger.Fatal(e.Start(":1337"))
}
