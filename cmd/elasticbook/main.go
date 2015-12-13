package main

import (
	"math/rand"
	"os"
	"time"

	"github.com/codegangsta/cli"
	"github.com/go-martini/martini"
	"github.com/zeroed/elasticbook"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	app := cli.NewApp()
	app.Name = "ElasticBook"
	app.Usage = "Elasticsearch for your bookmarks"

	var command string
	var verbose bool
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name: "command",
			// Aliases:     []string{"c"},
			Value:       "start",
			Usage:       "",
			Destination: &command,
		},
		cli.BoolFlag{
			Name: "verbose",
			// Aliases:     []string{"v"},
			Usage:       "I wanna read useless stuff",
			Destination: &verbose,
		},
	}

	app.Action = func(cc *cli.Context) {
		if command == "start" || command == "web" {
			m := martini.Classic()
			m.Get("/", func() string {
				return "Hello world!"
			})
			m.Run()
		} else {
			elasticbook.Sample()
		}
	}

	app.Run(os.Args)
}
