package main

import (
	"html/template"
	"io"

	"github.com/jadenrose/go-note/cmd/routes"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Templates struct {
	Templates *template.Template
}

func (t *Templates) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.Templates.ExecuteTemplate(w, name, data)
}

func newTemplate() *Templates {
	return &Templates{
		Templates: template.Must(template.ParseGlob("html/*.html")),
	}
}

func main() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Renderer = newTemplate()

	e.Static("/css", "css")
	e.Static("/img", "img")
	e.Static("/js", "js")

	e.GET("/", routes.Index)

	e.GET("/preview-links", routes.GetPreviewLinks)
	e.GET("/more-options/show", routes.ShowMoreOptions)
	e.GET("/more-options/hide", routes.HideMoreOptions)
	e.DELETE("/notes/:note_id", routes.DeleteNote)

	e.GET("/notes/new", routes.GetNewNote)
	e.POST("/notes", routes.PostNote)
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

	e.GET("/archive", routes.GetArchiveList)
	e.GET("/archive/:archived_note_id", routes.GetArchivedNote)
	e.POST("/archive/:archived_note_id", routes.RestoreArchivedNote)
	e.DELETE("/archive/all", routes.ClearArchive)

	e.POST("/search", routes.QuickSearch)

	e.Logger.Fatal(e.Start(":1337"))
}
