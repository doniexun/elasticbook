package elasticbook

// TODO: parse date
// TODO: add mapping to date
// TODO: add fulltext searc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"reflect"
	"strconv"
	"time"
)

// https://godoc.org/gopkg.in/olivere/elastic.v3
import "gopkg.in/olivere/elastic.v3"

// IndexName is the Elasticsearch index
const IndexName = "elasticbook"

// Root is the root of the Bookmarks tree
type Root struct {
	Checksum string `json:"checksum"`
	Version  int    `json:"version"`
	Roots    Roots  `json:"roots"`
}

// Roots is the container of the 4 main bookmark structure (high level)
type Roots struct {
	BookmarkBar            Base   `json:"bookmark_bar"`
	Other                  Base   `json:"other"`
	SyncTransactionVersion string `json:"sync_transaction_version"`
	Synced                 Base   `json:"synced"`
}

// Base is a "folder-like" container of Bookmarks
type Base struct {
	Children     []Bookmark `json:"children"`
	DateAdded    string     `json:"date_added"`
	DataModified string     `json:"date_modified"`
	ID           string     `json:"id"`
	Name         string     `json:"name"`
	NodeType     string     `json:"type"`
}

func (b *Base) String() string {
	return fmt.Sprintf("%s (%d)", b.Name, len(b.Children))
}

// Bookmark is a bookmark entry
type Bookmark struct {
	DateAdded              string `json:"date_added"`
	OriginalID             string `json:"id"`
	MetaInfo               Meta   `json:"meta_info,omitempty"`
	Name                   string `json:"name"`
	SyncTransactionVersion string `json:"sync_transaction_version"`
	Type                   string `json:"type"`
	URL                    string `json:"url"`
}

func (b *Bookmark) toSafe() (bs *BookmarkSafe) {
	bs = new(BookmarkSafe)
	bs.DateAdded = b.DateAdded
	bs.OriginalID = b.OriginalID
	mis := b.MetaInfo.toSafe()
	bs.MetaInfo = *mis
	bs.Name = b.Name
	bs.SyncTransactionVersion = b.SyncTransactionVersion
	bs.Type = b.Type
	bs.URL = b.URL
	return
}

// BookmarkSafe is a bookmark entry with a sanitised MetaInfo
type BookmarkSafe struct {
	DateAdded              string   `json:"date_added"`
	OriginalID             string   `json:"id"`
	MetaInfo               MetaSafe `json:"meta_info,omitempty"`
	Name                   string   `json:"name"`
	SyncTransactionVersion string   `json:"sync_transaction_version"`
	Type                   string   `json:"type"`
	URL                    string   `json:"url"`
}

// CountResult contains the bookmarks counter
type CountResult struct {
	m map[string]int
}

// Add a key value to the count container
func (c *CountResult) Add(k string, v int) {
	if c.m == nil {
		c.m = make(map[string]int)
	}
	c.m[k] = v
}

func (c *CountResult) String() string {
	var buffer bytes.Buffer
	for k := range c.m {
		buffer.WriteString(fmt.Sprintf("- %s (%d)\n", k, c.m[k]))
	}
	return buffer.String()
}

// Meta contains the attached metadata to the Bookmark entry
type Meta struct {
	StarsID        string `json:"stars.id"`
	StarsImageData string `json:"stars.imageData"`
	StarsIsSynced  string `json:"stars.isSynced"`
	StarsPageData  string `json:"stars.pageData"`
	StarsType      string `json:"stars.type"`
}

func (m *Meta) toSafe() (ms *MetaSafe) {
	ms = new(MetaSafe)
	ms.StarsID = m.StarsID
	ms.StarsImageData = m.StarsImageData
	ms.StarsIsSynced = m.StarsIsSynced
	ms.StarsPageData = m.StarsPageData
	ms.StarsType = m.StarsType
	return
}

// MetaSafe contains the attached metadata to the Bookmark entry w/o
// dots
type MetaSafe struct {
	StarsID        string `json:"stars_id"`
	StarsImageData string `json:"stars_imageData"`
	StarsIsSynced  string `json:"stars_isSynced"`
	StarsPageData  string `json:"stars_pageData"`
	StarsType      string `json:"stars_type"`
}

// Client is a connection builder
func Client() *elastic.Client {
	client, err := elastic.NewClient()
	if err != nil {
		panic("Unable to create a ES client")
	}

	info, err := client.ClusterHealth().Do()
	if err != nil {
		panic(fmt.Sprintf("Unable to connect a ES client %+v\n", info))
	}

	return client
}

// Count returns a map with the RootFolder name and the count
func (r *Root) Count() (c *CountResult) {
	c = new(CountResult)
	c.Add(r.Roots.BookmarkBar.Name, len(r.Roots.BookmarkBar.Children))
	c.Add(r.Roots.Other.Name, len(r.Roots.Other.Children))
	c.Add(r.Roots.Synced.Name, len(r.Roots.Synced.Children))

	return
}

// Delete drops the index
func Delete() {
	client := Client()

	_, err := client.DeleteIndex("elasticbook").Do()
	if err != nil {
		// Handle error
		panic(err)
	}
}

// timeParse converts a date (a string representation of the number of
// microseconds from the 1601/01/01
// https://chromium.googlesource.com/chromium/src/+/master/base/time/time_win.cc#56
//
// Quoting:
// From MSDN, FILETIME "Contains a 64-bit value representing the number of
// 100-nanosecond intervals since January 1, 1601 (UTC)."
func timeParse(microsecs string) string {
	t := time.Date(1601, time.January, 1, 0, 0, 0, 0, time.UTC)
	m, err := strconv.ParseInt(microsecs, 10, 64)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}
	var u int64 = 100000000000000
	du := time.Duration(u) * time.Microsecond
	f := float64(m)
	x := float64(u)
	n := f / x
	r := int64(n)
	remainder := math.Mod(f, x)
	iRem := int64(remainder)
	var i int64
	for i = 0; i < r; i++ {
		t = t.Add(du)
		fmt.Println(t)
	}

	t = t.Add(time.Duration(iRem) * time.Microsecond)

	// RFC1123 = "Mon, 02 Jan 2006 15:04:05 MST"
	return t.Format(time.RFC1123)
}

// Index takes a parsed structure and index all the Bookmarks entries
func Index(x *Root) {
	client := Client()

	if exists, _ := client.IndexExists(IndexName).Do(); !exists {
		indicesCreateResult, err := client.CreateIndex(IndexName).Do()
		if err != nil {
			// TODO: fix and check!
			fmt.Fprintf(os.Stderr, "%s\n", err.Error())
			os.Exit(1)
		} else {
			fmt.Fprintf(os.Stdout, "%s (%t)\n", IndexName, indicesCreateResult.Acknowledged)
		}
	}

	for i, b := range x.Roots.Synced.Children {
		fmt.Fprintf(os.Stdout, "%02d %s : %s\n", i, b.Name, b.URL)
		indexResponse, err := client.Index().
			Index(IndexName).
			Type("bookmark").
			BodyJson(b.toSafe()).
			Do()
		if err != nil {
			// TODO: Handle error
			panic(err)
		} else {
			fmt.Fprintf(os.Stdout, "%+v\n", indexResponse)
		}
	}
}

// Parse run the JSON parser
func Parse(b []byte) *Root {
	x := new(Root)
	err := json.Unmarshal(b, &x)
	if err != nil {
		panic(err.Error())
	}

	return x
}

// Search is the API for searching
func Search() {
	client := Client()

	termQuery := elastic.NewTermQuery("name", "slashdot")

	searchResult, err := client.Search().
		Index("elasticbook").
		Query(termQuery).
		Sort("date_added", true).
		From(0).Size(10).
		Pretty(true).
		Do()

	if err != nil {
		// Handle error
		panic(err)
	}

	fmt.Printf("Query took %d milliseconds\n", searchResult.TookInMillis)

	var ttyp Bookmark
	for _, item := range searchResult.Each(reflect.TypeOf(ttyp)) {
		if t, ok := item.(Bookmark); ok {
			fmt.Printf("Bookmark named %s: %s\n", t.Name, t.URL)
		}
	}
	fmt.Printf("Found a total of %d tweets\n", searchResult.TotalHits())

	if searchResult.Hits != nil {
		fmt.Printf("Found a total of %d tweets\n", searchResult.Hits.TotalHits)

		for _, hit := range searchResult.Hits.Hits {
			var t Bookmark
			err := json.Unmarshal(*hit.Source, &t)
			if err != nil {
				// Deserialization failed
			}

			fmt.Printf("Bookmark named %s: %s\n", t.Name, t.URL)
		}
	} else {
		// No hits
		fmt.Print("Found no Bookmarks\n")
	}
}
