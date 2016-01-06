package web

import (
	"github.com/go-martini/martini"
	"github.com/martini-contrib/binding"
	"github.com/martini-contrib/render"
)

const (
	// DefaultTemplateDir decides where look for templates files
	DefaultTemplateDir = "templates"

	// DefaultVerbose decides if you wanna be bored by some noisy logs
	DefaultVerbose = false
)

// AppOptionFunc is a function that configures a App.
// It is used in NewApp.
type AppOptionFunc func(*App) error

// App is the ElasticBook Web App (Martini powered)
type App struct {
	templates string
	verbose   bool
}

// NewApp Set up the default application
func NewApp(options ...AppOptionFunc) (*App, error) {
	c := &App{
		templates: DefaultTemplateDir,
		verbose:   DefaultVerbose,
	}
	for _, option := range options {
		if err := option(c); err != nil {
			return nil, err
		}
	}
	return c, nil
}

// SetVerbose define the verbose logging
func SetVerbose(vvv bool) AppOptionFunc {
	return func(a *App) error {
		a.verbose = vvv
		return nil
	}
}

// SetTemplateDir define the current TemplateDir used
func SetTemplateDir(t string) AppOptionFunc {
	return func(a *App) error {
		if t != "" {
			a.templates = t
		} else {
			a.templates = DefaultTemplateDir
		}
		return nil
	}
}

// Search represent a search request made from the WebInterface
type Search struct {
	ID      int64 `db:"id"`
	Created int64
	Term    string `form:"term" binding:"required"`
}

func (a *App) newSearch(s Search, r render.Render) {
}

// Start open a local server
func (a *App) Start() {
	m := martini.Classic()
	m.Use(martini.Static("public"))
	m.Use(render.Renderer(render.Options{
		Directory:       a.templates,                // Specify what path to load the templates from.
		Layout:          "layout",                   // Specify a layout template. Layouts can call {{ yield }} to render the current template.
		Extensions:      []string{".tmpl", ".html"}, // Specify extensions to load for templates.
		Charset:         "UTF-8",                    // Sets encoding for json and html content-types. Default is "UTF-8".
		IndentJSON:      true,                       // Output human readable JSON
		IndentXML:       true,                       // Output human readable XML
		HTMLContentType: render.ContentHTML,
	}))

	m.Get("/", func() string { return "Hello world!" })

	m.Group("/elasticbook", func(r martini.Router) {
		m.Get("/", func(r render.Render) {
			r.HTML(200, "list", nil)
		})
		r.Post("/search", binding.Bind(Search{}), a.newSearch)
	})
	m.Run()
}
