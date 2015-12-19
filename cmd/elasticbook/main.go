package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/boltdb/bolt"
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
			Name:        "command, c",
			Value:       "start",
			Usage:       "",
			Destination: &command,
		},
		cli.BoolFlag{
			Name:        "verbose", //, v",
			Usage:       "I wanna read useless stuff",
			Destination: &verbose,
		},
	}

	app.Action = func(cc *cli.Context) {
		if command == "web" {
			m := martini.Classic()
			m.Get("/", func() string {
				return "Hello world!"
			})
			m.Run()

		} else if command == "persist" {
			db, err := bolt.Open("db/my.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
			if err != nil {
				log.Fatal(err)
			}
			defer db.Close()

			db.Update(func(tx *bolt.Tx) error {
				b := tx.Bucket([]byte("MyBucket"))
				err := b.Put([]byte("answer"), []byte("42"))
				return err
			})

			db.View(func(tx *bolt.Tx) error {
				b := tx.Bucket([]byte("MyBucket"))
				v := b.Get([]byte("answer"))
				fmt.Printf("The answer is: %s\n", v)
				return nil
			})
		} else if command == "parse" {
			b := file()
			elasticbook.Parse(b)

		} else if command == "count" {
			b := file()
			r := elasticbook.Parse(b)
			c := r.Count()
			fmt.Fprintf(os.Stdout, "%+v", c)

		} else if command == "index" {
			b := file()
			x := elasticbook.Parse(b)
			elasticbook.Index(x)

		} else if command == "elastic" {
			elasticbook.Elastic()

		} else {
			fmt.Fprintf(os.Stdout, "unsupported command\n")
		}
	}

	app.Run(os.Args)
}

func file() []byte {
	b, err := ioutil.ReadFile("/Users/edoardo/Downloads/bookmarks_20151215.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to load file (%s)", err.Error())
		os.Exit(1)
	}
	return b
}
