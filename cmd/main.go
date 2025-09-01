package main

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strconv"
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

var allContacts = getContacts()

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
		case 4:
			data := map[string]any{"curStep": 0, "nextStep": 1, "prevStep": -1, "direction": "none"}

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
	e.GET("errorpage", displayErrorPageHandler)
	e.GET("demo04-error", simulateErrorHandler)
	e.POST("demo05/next", demo05NextHandler)
	e.POST("demo05/prev", demo05BackHandler)

	e.GET("list-items", func(c echo.Context) error {
		return c.Render(200, "contact-list", map[string]any{"ItemList": allContacts})
	})

	e.DELETE("item-delete", func(c echo.Context) error {
		itemId, _ := strconv.Atoi(c.QueryParam("id"))

		if itemId != 5 && itemId != 8 {
			allContacts[itemId].Deleted = true
			return c.Render(200, "page-message", map[string]any{
				"type": "info", "message": fmt.Sprintf("\"%s\" removed", allContacts[itemId].Name)})
		}

		c.Response().Header().Set("HX-Reswap", "none")
		return c.Render(422, "page-message", map[string]any{
			"type": "danger", "message": fmt.Sprintf("\"%s\" cannot be delete", allContacts[itemId].Name)})
	})

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

func displayErrorPageHandler(c echo.Context) error {
	return c.Render(200, "errorpage", map[string]any{"CreatedOn": time.Now()})
}

func simulateErrorHandler(c echo.Context) error {
	var message string

	switch id := c.QueryParam("id"); id {
	case "1":
		message = "404: Not found"
	case "2":
		message = "403: Not authorized"
	case "3":
		message = "Some processing error"
	case "4":
		message = "Forcefully logged out"
		c.Response().Header().Set("HX-Trigger-After-Settle", `{"logoutEvent":{"level" : "Critical", "details" : "No further details"}}`)
	}

	c.Response().Header().Set("HX-Retarget", "body")
	c.Response().Header().Set("HX-Push-Url", "errorpage")
	return c.Render(200, "errorpage", map[string]any{"CreatedOn": time.Now(), "message": message})
}

func demo05NextHandler(c echo.Context) error {
	step, _ := strconv.Atoi(c.FormValue("step"))
	step += 1
	food := c.FormValue("food")
	month := c.FormValue("month")
	colour := c.FormValue("colour")
	if step == 4 {
		return c.HTML(200, "<h4>Thank you for completing the survey</h4>")
	}

	return c.Render(200, "multi-form", map[string]any{
		"curStep": step, "nextStep": step + 1, "prevStep": step - 1, "direction": "next",
		"food": food, "month": month, "colour": colour,
	})
}

func demo05BackHandler(c echo.Context) error {
	step, _ := strconv.Atoi(c.FormValue("step"))
	step -= 1
	food := c.FormValue("food")
	month := c.FormValue("month")
	colour := c.FormValue("colour")
	return c.Render(200, "multi-form", map[string]any{
		"curStep": step, "nextStep": step + 1, "prevStep": step - 1, "direction": "prev",
		"food": food, "month": month, "colour": colour,
	})
}

type Contact struct {
	Name    string
	Id      int
	Deleted bool
}

type Contacts []Contact

func getContacts() Contacts {
	var c Contacts
	for i := range 16 {
		c = append(c, Contact{Name: fmt.Sprintf("Item %d", i), Id: i})
	}
	c[0].Deleted = true
	c[5].Name = "Can't make me move"
	c[8].Name = "Here to stay"
	return c
}
