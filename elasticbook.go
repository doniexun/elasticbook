package elasticbook

import (
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

// Meta contains the attached metadata to the Bookmark entry
type Meta struct {
	StarsID        string `json:"stars.id"`
	StarsImageData string `json:"stars.imageData"`
	StarsIsSynced  string `json:"stars.isSynced"`
	StarsPageData  string `json:"stars.pageData"`
	StarsType      string `json:"stars.type"`
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

func client() *elastic.Client {
	client, err := elastic.NewClient()
	if err != nil {
		panic("Unable to create a ES client")
	}
	return client
}

// Parse run the JSON parser
func Parse(b []byte) {
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
	return
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
