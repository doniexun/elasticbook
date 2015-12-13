package elasticbook

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// https://godoc.org/gopkg.in/olivere/elastic.v3
import "gopkg.in/olivere/elastic.v3"

// Meta contains the attached metadata to the Bookmark entry
type Meta struct {
	StarsID        string
	StarsImageData string
	StarsIsSynced  string
	StarsPageData  string
	StarsType      string
}

// Bookmark is a bookmark entry
type Bookmark struct {
	DateAdded              string
	OriginalID             string
	MetaInfo               Meta
	Name                   string
	SyncTransactionVersion string
	Type                   string
	URL                    string
}

// A sample entry:
// {
// 	"date_added": "13032201986000000",
// 	"id": "25782",
// 	"meta_info": {
// 		"stars.id": "ssc_267d2d8ea5010886",
// 		"stars.imageData": "aaa",
// 		"stars.isSynced": "true",
// 		"stars.pageData": "Ig5qbzNhUmUyOXVIc0JETQ==",
// 		"stars.type": "2"
// 	},
// 	"name": "SlashDot",
// 	"sync_transaction_version": "57627",
// 	"type": "url",
// 	"url": "http://slashdot.org/"
// }
var smpl = Bookmark{
	DateAdded:  "13032201986000000",
	OriginalID: "25782",
	MetaInfo: Meta{
		StarsID:        "ssc_267d2d8ea5010886",
		StarsImageData: "aaa",
		StarsIsSynced:  "true",
		StarsPageData:  "Ig5qbzNhUmUyOXVIc0JETQ==",
		StarsType:      "2",
	},
	Name: "SlashDot",
	SyncTransactionVersion: "57627",
	Type: "url",
	URL:  "http://slashdot.org/",
}

func client() *elastic.Client {
	client, err := elastic.NewClient()
	if err != nil {
		panic("Unable to create a ES client")
	}
	return client
}

// Sample is the sample
func Sample() {
	client := client()
	_, err := client.CreateIndex("elasticbook").Do()
	if err != nil {
		panic(err)
	}

	// Add a document to the index
	_, err = client.Index().
		Index("elasticbook").
		Type("bookmark").
		Id("1").
		BodyJson(smpl).
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
