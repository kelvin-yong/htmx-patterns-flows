package main

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
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
	fmt.Println(RouteDemo3Step1)
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
		switch i {
		case 2:
			data := struct {
				Step        int
				ItemName    string
				NextPostUrl string
				NextHistory string
			}{
				1, "Food", "demo03-add-01", "demo03-month",
			}
			e.GET("/"+demoRoute, func(c echo.Context) error {
				return c.Render(200, demoRoute, data)
			})
		default:
			e.GET("/"+demoRoute, func(c echo.Context) error {
				return c.Render(200, demoRoute, nil)
			})
		}
	}

	e.POST("demo01-search", demo1Search)
	e.POST("demo02-search", demo2Search)
	e.POST("demo03-add-01", generateDemo3Step("Month", "2/3", "demo03-add-02", "demo03-fav-colour"))
	e.POST("demo03-add-02", generateDemo3Step("Colour", "3/3", "demo03-add-03", "demo03-thankyou"))
	e.POST("demo03-add-03", generateDemo3Step("", "Done", "", ""))

	registerFallbackRoutes(e)

	e.Logger.Fatal(e.Start(":9000"))
}

func registerFallbackRoutes(e *echo.Echo) {
	// may not be the best solution. Instead of registering multiple handlers
	// consider have a catch all route. then check if the path is in one of
	// the virtual routes that has a fallback
	routes := []route{RouteDemo3Step1, RouteDemo3Step2, RouteDemo3Step3}
	for _, r := range routes {
		e.GET(r.Path, func(c echo.Context) error {
			return c.Redirect(http.StatusFound, r.FallbackRoute.Path)
		})
	}
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

func generateDemo3Step(item, nextTitle, nextPost, nextHistory string) func(c echo.Context) error {
	data := struct {
		ItemName    string
		NextPostUrl string
		NextHistory string
	}{
		item, nextPost, nextHistory,
	}
	return func(c echo.Context) error {
		if nextPost != "" {
			c.Render(200, "demo3-add-form", data)
		} else {
			c.HTML(200, "<h4>Thank you for completing the survey</h4>")
		}
		name := c.FormValue("name")
		submittedItem := c.FormValue("item")
		html :=
			`<div hx-swap-oob="beforeend: #friend-list"><p>` + submittedItem + ": <i>" + name + `</i></p></div>`
		c.HTML(200, html)

		return c.Render(200, "page-title", map[string]any{"title": "Multi-step Demo: " + nextTitle})
	}
}
