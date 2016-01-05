package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/boltdb/bolt"
	"github.com/codegangsta/cli"
	"github.com/fatih/color"
	"github.com/go-martini/martini"
	"github.com/zeroed/elasticbook"
	"github.com/zeroed/elasticbook/utils"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	app := cli.NewApp()
	app.Name = "ElasticBook"
	app.Usage = "Elasticsearch for your bookmarks"
	app.Version = "0.0.1"

	var command string
	var term string
	var verbose bool
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "command, c",
			Usage:       "-c [alias|aliases|unalias|default|indices|index|count|health|parse|delete|web|persist]",
			Destination: &command,
		},
		cli.StringFlag{
			Name:        "search, s",
			Usage:       "-s [term]",
			Destination: &term,
		},
		cli.BoolFlag{
			Name:        "verbose, V",
			Usage:       "I wanna read useless stuff",
			Destination: &verbose,
		},
	}

	app.Action = func(cc *cli.Context) {
		if command != "" && term != "" {
			fmt.Fprintf(os.Stderr, "You cannot set a command AND make a search\n\n")
			os.Exit(1)
		}

		if command == "" && term == "" {
			fmt.Fprintf(os.Stdout, "You shuld set a command OR make a search\n\n")
			cli.ShowAppHelp(cc)
			os.Exit(1)
		}

		if command == "alias" {
			alias()
		} else if command == "aliases" {
			aliases()
		} else if command == "unalias" {
			unalias()
		} else if command == "count" {
			count()
		} else if command == "default" {
			defaultAlias()
		} else if command == "delete" {
			deleteIndex()
		} else if command == "indices" {
			indices()
		} else if command == "index" {
			index()
		} else if command == "health" {
			health()
		} else if command == "parse" {
			parse()
		} else if command == "persist" {
			persist()
		} else if command == "version" {
			version()
		} else if command == "web" {
			web()
		} else {
			if term == "" {
				fmt.Fprintf(os.Stderr, "Command not supported\n\n")

				cli.ShowAppHelp(cc)
			}
		}

		if term != "" {
			searchTerm(term, verbose)
		}
	}

	app.Run(os.Args)
}

func alias() {
	c, _ := aliases()

	ics, err := c.IndexNames()
	if len(ics) == 0 {
		fmt.Fprintf(os.Stderr, "There are no indexes\n")
		os.Exit(1)
	}

	fmt.Fprintf(os.Stdout, "Index ")
	i := askForIndex(len(ics) - 1)
	indexName := ics[i]
	var aliasName string

	fmt.Fprintf(os.Stdout, "Alias name: ")
	_, err = fmt.Scanln(&aliasName)
	if err != nil && err.Error() == "unexpected newline" {
		alias()
	}

	if aliasName == elasticbook.DefaultAliasName {
		fmt.Fprintf(
			os.Stderr,
			"%s is the default alias name. Use `-c default` to assign the default index\n",
			aliasName)
		alias()
	}

	ack, err := c.Alias(indexName, aliasName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}

	if ack {
		aliases()
	} else {
		fmt.Fprintf(os.Stderr, "Cannot create your alias. Maybe is already there...\n")
		os.Exit(1)
	}
}

func aliases() (*elasticbook.Client, []string) {
	c, err := elasticbook.ClientRemote()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}
	var ics []string
	ics, err = c.Aliases()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}
	cyan := color.New(color.FgCyan).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	for i, x := range ics {
		index := fmt.Sprintf("%02d", i)
		xs := strings.Split(x, ":")
		fmt.Fprintf(os.Stdout, "%s] - %s%s\n",
			cyan(index), green(xs[0]), yellow(xs[1]))
	}

	return c, ics
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

	if utils.ContainsString(okayResponses, response) {
		return true
	} else if utils.ContainsString(nokayResponses, response) {
		return false
	} else {
		fmt.Fprintf(os.Stdout, "Please type yes|no and then press enter: ")
	}
	return askForConfirmation()
}

func askForIndex(length int) int {
	msg := ""
	for {
		fmt.Fprintf(os.Stdout, "[0-%02d]: %s ", length, msg)
		var i int
		_, err := fmt.Scanf("%d", &i)
		if err == nil && i >= 0 && i <= length {
			return i
		}
		msg = "(nope)"
	}
}

func chooseCollection(cns []string) int {
	for i, cn := range cns {
		fmt.Fprintf(os.Stdout, "[%d] %s\n", i, cn)
	}
	fmt.Fprintf(os.Stdout, "----\n")
	return askForIndex(len(cns))
}

func count() {
	// TODO: also check local if you want
	fmt.Fprintf(os.Stdout, "Working on %s\n", utils.BookmarksFilePath())
	c, err := elasticbook.ClientRemote()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}
	b := utils.BookmarksFile()
	r, err := c.Parse(b)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Your Bookmarks DB cannot be parsed, sorry\n\n")
		os.Exit(1)
	}

	n := r.Count()
	fmt.Fprintf(os.Stdout, "%+v", n)
}

func defaultAlias() {
	aliases()

	c, err := elasticbook.ClientRemote()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}

	fmt.Fprintf(os.Stdout, "Index name: ")
	icNames, _ := c.IndexNames()
	i := askForIndex(len(icNames) - 1)
	indexName := icNames[i]

	ack, err := c.Default(indexName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}

	if ack {
		aliases()
	} else {
		fmt.Fprintf(os.Stderr, "Cannot switch default alias")
		os.Exit(1)
	}
}

func deleteIndex() {
	c, ics := indices()
	if len(ics) == 0 {
		fmt.Fprintf(os.Stderr, "There are no indexes\n")
		os.Exit(1)
	}

	icNames, _ := c.IndexNames()
	i := askForIndex(len(icNames) - 1)
	indexName := icNames[i]
	fmt.Fprintf(os.Stdout, "Want to delete the %s index? [y/N]: ", indexName)
	if askForConfirmation() {
		c.Delete(indexName)
	} else {
		fmt.Fprintf(os.Stdout, "Whatever\n\n")
	}
}

func health() {
	// TODO: also check local if you want
	c, err := elasticbook.ClientRemote()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}
	h, err := c.Health()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}
	fmt.Fprintf(os.Stdout, "%+v\n\n", h)
}

func indices() (*elasticbook.Client, []string) {
	c, err := elasticbook.ClientRemote()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}
	var ics []string
	ics, err = c.Indices()
	cyan := color.New(color.FgCyan).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()

	for i, x := range ics {
		index := fmt.Sprintf("%02d", i)
		fmt.Fprintf(os.Stdout, "%s] - %s\n",
			cyan(index), green(x))
	}
	return c, ics
}

func index() {
	// TODO: also check local if you want
	c, err := elasticbook.ClientRemote()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}
	b := utils.BookmarksFile()
	r, err := c.Parse(b)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Your Bookmarks DB cannot be parsed, sorry\n\n")
		os.Exit(1)
	}
	_, err = c.Index(r)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	} else {
		fmt.Fprintf(os.Stdout, "Index created \n")
	}
	count := r.Count()
	fmt.Fprintf(os.Stdout, "%+v", count)
}

func parse() {
	// TODO: also check local if you want
	c, err := elasticbook.ClientRemote()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}

	b := utils.BookmarksFile()
	cr, err := c.Parse(b)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Your Bookmarks DB cannot be parsed, sorry\n\n")
	} else {
		fmt.Fprintf(
			os.Stdout,
			"Your Bookmarks DB seems healthy: %d bookmarks found\n\n",
			cr.Count().Total())
	}
}

func persist() {
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
}

func unalias() {
	// TODO: maybe avoid multiple connections?
	aliases()

	c, err := elasticbook.ClientRemote()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}

	var aliasName string

	fmt.Fprintf(os.Stdout, "Delete alias name: ")
	_, err = fmt.Scanln(&aliasName)
	if err != nil && err.Error() == "unexpected newline" {
		unalias()
	}

	if aliasName == elasticbook.DefaultAliasName {
		fmt.Fprintf(
			os.Stderr,
			"%s is the default alias name. Do not delete this, please\n",
			aliasName)
		unalias()
	}

	ack, err := c.Unalias(aliasName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}

	if ack {
		aliases()
	} else {
		fmt.Fprintf(os.Stderr, "Cannot delete your alias")
		os.Exit(1)
	}
}

func searchTerm(term string, verbose bool) {
	// TODO: also check local if you want
	c, err := elasticbook.ClientRemote()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}
	sr, err := c.Search(term)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}

	fmt.Fprintf(os.Stdout, "Query took %d milliseconds\n", sr.TookInMillis)

	if sr.Hits != nil {
		fmt.Printf("Found a total of %d bookmarks\n", sr.Hits.TotalHits)

		// blue := color.New(color.FgBlue).SprintFunc()
		red := color.New(color.FgRed).SprintFunc()
		cyan := color.New(color.FgCyan).SprintFunc()
		green := color.New(color.FgGreen).SprintFunc()
		yellow := color.New(color.FgYellow).SprintFunc()

		for i, hit := range sr.Hits.Hits {
			var t elasticbook.Bookmark
			err := json.Unmarshal(*hit.Source, &t)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", err.Error())
			}

			index := fmt.Sprintf("%02d", i)
			fmt.Fprintf(os.Stdout, "%s] - %s [%s] (%s) {%s}\n",
				cyan(index), green(t.Name), yellow(t.URL), t.DateAdded,
				red(fmt.Sprintf("%f", *hit.Score)),
				// red(strconv.FormatFloat(hit.Score, 'f', 6, 64)),
			)
			if verbose {
				fmt.Fprintf(os.Stdout, "%v\n", hit.Explanation)
			}
		}
	} else {
		// No hits
		fmt.Print("Found no Bookmarks\n")
	}
}

func version() {
	// TODO: also check local if you want
	c, err := elasticbook.ClientRemote()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}
	h := c.Version()
	fmt.Fprintf(os.Stdout, "Elasticsearch version %+v (%s)\n\n", h, c.URL())
}

func web() {
	m := martini.Classic()
	m.Get("/", func() string {
		return "Hello world!"
	})
	m.Run()
}
