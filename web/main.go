package web

import (
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
)

func newSearch() {
}

// Start open a local server
func Start() {
	m := martini.Classic()
	m.Use(martini.Static("public"))
	m.Use(render.Renderer(render.Options{
		Directory:       "templates",                // Specify what path to load the templates from.
		Layout:          "layout",                   // Specify a layout template. Layouts can call {{ yield }} to render the current template.
		Extensions:      []string{".tmpl", ".html"}, // Specify extensions to load for templates.
		Charset:         "UTF-8",                    // Sets encoding for json and html content-types. Default is "UTF-8".
		IndentJSON:      true,                       // Output human readable JSON
		IndentXML:       true,                       // Output human readable XML
		HTMLContentType: "text/html",
	}))

	m.Get("/", func() string { return "Hello world!" })

	m.Group("/elasticbook", func(r martini.Router) {
		m.Get("/", func(r render.Render) {
			r.HTML(200, "list", nil)
		})
		r.Post("/search", newSearch)
	})
	m.Run()
}
