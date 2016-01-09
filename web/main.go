package web

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/go-martini/martini"
	"github.com/martini-contrib/binding"
	"github.com/martini-contrib/render"
	"github.com/zeroed/elasticbook"
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

// Result is the result of a search
type Result struct {
	Index     int
	URL       string
	Title     string
	DateAdded string
	Score     float64
}

// IndexAlias contains an index name and its aliases
type IndexAlias struct {
	Index int
	Name  string
}

func (a *App) aliases(cl *elasticbook.Client, r render.Render, log *log.Logger) {
	sr, err := cl.Aliases()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}

	list := make([]IndexAlias, len(sr))
	for i, ia := range sr {
		list[i] = IndexAlias{
			Index: i,
			Name:  ia,
		}
	}

	nmap := map[string]interface{}{"show": true, "results": list}
	r.HTML(200, "aliases", nmap)
	return
}

func (a *App) search(cl *elasticbook.Client, s Search, r render.Render, log *log.Logger) {
	sr, err := cl.Search(s.Term)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}

	nmap := map[string]interface{}{"show": false, "results": nil}
	if sr.Hits != nil {
		log.Printf("Found a total of %d bookmarks\n", sr.Hits.TotalHits)

		list := make([]Result, len(sr.Hits.Hits))
		for i, hit := range sr.Hits.Hits {
			var t elasticbook.Bookmark
			err := json.Unmarshal(*hit.Source, &t)
			if err != nil {
				continue
			}

			list[i] = Result{
				Index:     i,
				Title:     t.Name,
				URL:       t.URL,
				DateAdded: t.DateAdded,
				Score:     *hit.Score}
		}
		nmap = map[string]interface{}{"show": true, "results": list}

	}
	r.HTML(200, "list", nmap)
	return
}

// Start open a local server
func (a *App) Start() {
	cl, err := elasticbook.ClientRemote()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}

	m := martini.Classic()
	m.Map(cl)
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

	m.Get("/", func(r render.Render) {
		r.Redirect("/elasticbook")
		return
	})

	m.Group("/elasticbook", func(r martini.Router) {

		m.Get("/", func(r render.Render) {
			r.HTML(200, "list", nil)
		})

		r.Get("/aliases", a.aliases)
		r.Post("/search", binding.Bind(Search{}), a.search)
	})

	m.Run()
}
