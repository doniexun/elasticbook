package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/user"
	"path/filepath"
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
			Usage:       "parse|index|count|delete|web|persist",
			Destination: &command,
		},
		cli.BoolFlag{
			Name:        "verbose, V",
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
			fmt.Fprintf(os.Stdout, "Working on %s\n", bookmarksFile())

			b := file()
			r := elasticbook.Parse(b)
			c := r.Count()
			fmt.Fprintf(os.Stdout, "%+v", c)

		} else if command == "health" {
			// TODO: also check local if you want
			c := elasticbook.ClientRemote()
			h := elasticbook.Health(c)
			fmt.Fprintf(os.Stdout, "%+v\n\n", h)

		} else if command == "version" {
			// TODO: also check local if you want
			// c := elasticbook.ClientRemote()
			// h := elasticbook.Version(c)
			// fmt.Fprintf(os.Stdout, "%+v", h)

		} else if command == "index" {
			b := file()
			r := elasticbook.Parse(b)
			elasticbook.Index(r)
			c := r.Count()
			fmt.Fprintf(os.Stdout, "%+v", c)

		} else if command == "delete" {
			fmt.Fprintf(os.Stdout, "Want to delete the existing index? [y/N]: ")
			if askForConfirmation() {
				elasticbook.Delete()
			} else {
				fmt.Fprintf(os.Stdout, "Whatever\n\n")
			}

		} else {
			fmt.Fprintf(os.Stdout, "unsupported command\n")
		}
	}

	app.Run(os.Args)
}

// askForConfirmation uses Scanln to parse user input. A user must type
// in "yes" or "no" and then press enter. It has fuzzy matching, so "y",
// "Y", "yes", "YES", and "Yes" all count as confirmations. If the input
// is not recognized, it will ask again. The function does not return
// until it gets a valid response from the user. Typically, you should
// use fmt to print out a question before calling askForConfirmation.
// E.g. fmt.Println("WARNING: Are you sure? (yes/no)")
func askForConfirmation() bool {
	const dflt string = "no"
	var response string

	_, err := fmt.Scanln(&response)
	if err != nil && err.Error() == "unexpected newline" {
		response = dflt
	}

	nokayResponses := []string{"n", "N", "no", "No", "NO"}
	okayResponses := []string{"y", "Y", "yes", "Yes", "YES"}

	if containsString(okayResponses, response) {
		return true
	} else if containsString(nokayResponses, response) {
		return false
	} else {
		fmt.Fprintf(os.Stdout, "Please type yes|no and then press enter: ")
	}
	return askForConfirmation()
}

// bookmarksFile want to guess which is the local bookmarks DB from the
// Chrome installation.
// This one is from my OSX, brew-installed, Chrome.
// "/Users/edoardo/Library/Application Support/Google/Chrome/Default/Bookmarks"
func bookmarksFile() string {
	user, err := user.Current()
	if err != nil {
		fmt.Fprintf(os.Stderr, "OS usupported? %s\n", err.Error())
	}

	return filepath.Join(
		user.HomeDir, "Library", "Application Support",
		"Google", "Chrome", "Default", "Bookmarks")
}

// containsString returns true iff slice contains element
func containsString(slice []string, element string) bool {
	return !(posString(slice, element) == -1)
}

func file() []byte {
	b, err := ioutil.ReadFile(bookmarksFile())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to load file (%s)", err.Error())
		os.Exit(1)
	}
	return b
}

// posString returns the first index of element in slice.
// If slice does not contain element, returns -1.
func posString(slice []string, element string) int {
	for i, e := range slice {
		if e == element {
			return i
		}
	}
	return -1
}
