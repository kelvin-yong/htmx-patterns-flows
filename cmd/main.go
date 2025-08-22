package main

import (
	"fmt"
	"html/template"
	"io"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Templates struct {
	templates *template.Template
}

func (t *Templates) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func NewTemplates() *Templates {
	return &Templates{
		templates: template.Must(template.ParseGlob("views/*.html")),
	}
}

func main() {
	e := echo.New()
	e.Static("/static", "static")
	e.Use(middleware.Logger())

	e.Renderer = NewTemplates()

	e.GET("/", func(c echo.Context) error {
		return c.Render(200, "demo01", nil)
	})

	e.GET("/test", func(c echo.Context) error {
		return c.Render(200, "test", nil)
	})

	for i := range 20 {
		demoRoute := fmt.Sprintf("demo%02d", i+1)
		e.GET("/"+demoRoute, func(c echo.Context) error {
			return c.Render(200, demoRoute, nil)
		})
	}

	e.POST("demo01-search", demo1Search)
	e.POST("demo02-search", demo2Search)

	e.Logger.Fatal(e.Start(":9000"))
}

func demo1Search(c echo.Context) error {
	name := c.FormValue("name")

	if name == "Error1" {
		c.Render(200, "demo1-search-form", nil)

		return c.Render(200, "page-message", map[string]any{
			"type": "danger", "message": "The user's account cannot be found. Form is replaced."})
	}

	if name == "Error2" {
		c.Response().Header().Set("HX-Reswap", "none")
		return c.Render(200, "page-message", map[string]any{
			"type": "danger", "message": "The user's account cannot be found. Form replacing skipped with HX-Reswap header"})
	}

	data := map[string]any{
		"name": name,
	}
	c.Render(200, "page-message", nil)
	return c.Render(200, "demo1-search-result", data)
}

func demo2Search(c echo.Context) error {
	name := c.FormValue("name")
	if name == "" {
		return c.Render(200, "page-message", map[string]any{
			"type": "danger", "message": "The user's account cannot be found."})
	}

	// simulate long running event
	time.Sleep(3 * time.Second)
	c.Render(200, "page-message", nil)
	return c.Render(200, "demo2-search-result", map[string]any{"name": name})
}
