package elasticbook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
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

func client() *elastic.Client {
	client, err := elastic.NewClient()
	if err != nil {
		panic("Unable to create a ES client")
	}
	return client
}

// Parse run the JSON parser
func Parse(b []byte) *Root {
	x := new(Root)
	err := json.Unmarshal(b, &x)
	if err != nil {
		panic(err.Error())
	} else {
		fmt.Fprintf(os.Stdout, "It Works!\n")
		fmt.Fprintf(os.Stdout, "\t- %s\n", x.Roots.BookmarkBar.String())
		fmt.Fprintf(os.Stdout, "\t- %s\n", x.Roots.Other.String())
		fmt.Fprintf(os.Stdout, "\t- %s\n", x.Roots.Synced.String())
	}
	return x
}

// Index takes a parsed structure and index all the Bookmarks entries
func Index(x *Root) {
	client := client()

	if exists, _ := client.IndexExists(IndexName).Do(); !exists {
		indicesCreateResult, err := client.CreateIndex(IndexName).Do()
		if err != nil {
			// TODO: fix and check!
			fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		} else {
			fmt.Fprintf(os.Stdout, "%s (%t)\n", IndexName, indicesCreateResult.Acknowledged)
		}
	}

	for i, b := range x.Roots.Synced.Children {
		fmt.Fprintf(os.Stdout, "%02d %s : %s\n", i, b.Name, b.URL)

		// body, err := json.Marshal(b)
		// fmt.Fprintf(os.Stdout, "%02d %s\n", i, body)

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

// Elastic is the sample
func Elastic() {
	client := client()
	_, err := client.CreateIndex(IndexName).Do()
	if err != nil {
		// TODO: fix and check!
		// panic(err)
	}

	// Add a document to the index
	_, err = client.Index().
		Index("elasticbook").
		Type("bookmark").
		Id("1").
		BodyJson(new(interface{})).
		// BodyJson(smpl).
		Do()
	if err != nil {
		// Handle error
		panic(err)
	}

	// Search with a term query
	termQuery := elastic.NewTermQuery("name", "slashdot")
	searchResult, err := client.Search().
		Index("elasticbook").
		Query(termQuery).
		Sort("user", true).
		From(0).Size(10).
		Pretty(true).
		Do()
	if err != nil {
		// Handle error
		panic(err)
	}

	// searchResult is of type SearchResult and returns hits, suggestions,
	// and all kinds of other information from Elasticsearch.
	fmt.Printf("Query took %d milliseconds\n", searchResult.TookInMillis)

	// Each is a convenience function that iterates over hits in a search result.
	// It makes sure you don't need to check for nil values in the response.
	// However, it ignores errors in serialization. If you want full control
	// over iterating the hits, see below.
	var ttyp Bookmark
	for _, item := range searchResult.Each(reflect.TypeOf(ttyp)) {
		if t, ok := item.(Bookmark); ok {
			fmt.Printf("Bookmark named %s: %s\n", t.Name, t.URL)
		}
	}
	// TotalHits is another convenience function that works even when something goes wrong.
	fmt.Printf("Found a total of %d tweets\n", searchResult.TotalHits())

	// Here's how you iterate through results with full control over each step.
	if searchResult.Hits != nil {
		fmt.Printf("Found a total of %d tweets\n", searchResult.Hits.TotalHits)

		// Iterate through results
		for _, hit := range searchResult.Hits.Hits {
			// hit.Index contains the name of the index

			// Deserialize hit.Source into a Tweet (could also be just a map[string]interface{}).
			var t Bookmark
			err := json.Unmarshal(*hit.Source, &t)
			if err != nil {
				// Deserialization failed
			}

			// Work with tweet
			fmt.Printf("Bookmark named %s: %s\n", t.Name, t.URL)
		}
	} else {
		// No hits
		fmt.Print("Found no tweets\n")
	}

	// Delete the index again
	_, err = client.DeleteIndex("twitter").Do()
	if err != nil {
		// Handle error
		panic(err)
	}
}
