package routes

import (
	"log"

	"github.com/labstack/echo/v4"
)

type QuickSearchResults struct {
	Results    []QuickSearchResult
	SearchTerm string
}

type QuickSearchResult struct {
	NoteID  int
	Title   string
	Content string
}

func QuickSearch(c echo.Context) error {
	var err error

	search_term := c.FormValue("search-term")
	if len(search_term) == 0 {
		return c.NoContent(200)
	}

	if agent == nil {
		agent = NewDBAgent()
	}
	handleError := func() error {
		log.Panic(err)
		agent.Rollback()
		return c.NoContent(500)
	}

	if err = agent.Open(); err != nil {
		return handleError()
	}
	defer agent.Close()
	rows, err := agent.Query(
		`
        SELECT note_id, title, content
        FROM quick_search($1);
        `,
		search_term,
	)
	if err != nil {
		return handleError()
	}
	results := []QuickSearchResult{}
	for rows.Next() {
		result := QuickSearchResult{}
		if err = rows.Scan(
			&result.NoteID,
			&result.Title,
			&result.Content,
		); err != nil {
			return handleError()
		}
		results = append(results, result)
	}

	if len(results) == 0 {
		return c.Render(200, "no-search-results", nil)
	}

	return c.Render(
		200,
		"search-results-list",
		QuickSearchResults{
			SearchTerm: search_term,
			Results:    results,
		},
	)
}
