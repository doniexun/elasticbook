package elasticbook

// DONE: parse date
// DONE: add mapping to date
// TODO: add fulltext search
// TODO: add query CLI
// TODO: add progress bar

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/gosuri/uiprogress"
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

func (b *Bookmark) toIndexable() (bs *BookmarkIndexable) {
	bs = new(BookmarkIndexable)
	bs.DateAdded = timeParse(b.DateAdded)
	bs.OriginalID = b.OriginalID
	mis := b.MetaInfo.toIndexable()
	bs.MetaInfo = *mis
	bs.Name = b.Name
	bs.SyncTransactionVersion = b.SyncTransactionVersion
	bs.Type = b.Type
	bs.URL = b.URL
	return
}

// BookmarkIndexable is a bookmark entry with a sanitised MetaInfo
type BookmarkIndexable struct {
	DateAdded              time.Time     `json:"date_added"`
	OriginalID             string        `json:"id"`
	MetaInfo               MetaIndexable `json:"meta_info,omitempty"`
	Name                   string        `json:"name"`
	SyncTransactionVersion string        `json:"sync_transaction_version"`
	Type                   string        `json:"type"`
	URL                    string        `json:"url"`
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

// Total return the grand total of Bookmark entries parsed/indexed
func (c *CountResult) Total() int {
	var t int
	for k := range c.m {
		t += c.m[k]
	}
	return t
}

// Meta contains the attached metadata to the Bookmark entry
type Meta struct {
	StarsID        string `json:"stars.id"`
	StarsImageData string `json:"stars.imageData"`
	StarsIsSynced  string `json:"stars.isSynced"`
	StarsPageData  string `json:"stars.pageData"`
	StarsType      string `json:"stars.type"`
}

func (m *Meta) toIndexable() (ms *MetaIndexable) {
	ms = new(MetaIndexable)
	ms.StarsID = m.StarsID
	ms.StarsImageData = m.StarsImageData
	ms.StarsIsSynced = m.StarsIsSynced
	ms.StarsPageData = m.StarsPageData
	ms.StarsType = m.StarsType
	return
}

// MetaIndexable contains the attached metadata to the Bookmark entry w/o
// dots
type MetaIndexable struct {
	StarsID        string `json:"stars_id"`
	StarsImageData string `json:"stars_imageData"`
	StarsIsSynced  string `json:"stars_isSynced"`
	StarsPageData  string `json:"stars_pageData"`
	StarsType      string `json:"stars_type"`
}

// client is a connection builder
func client(remote bool) *elastic.Client {
	var clnt *elastic.Client
	var url string
	var err error

	if remote {
		url = os.Getenv("BONSAIO_HOST")
		clnt, err = elastic.NewClient(
			elastic.SetURL(url),
			elastic.SetScheme("https"),
			elastic.SetMaxRetries(5),
			elastic.SetSniff(false),
			elastic.SetHealthcheckInterval(10*time.Second),
			elastic.SetErrorLog(log.New(os.Stderr, "[elastic] ", log.LstdFlags)),
			elastic.SetInfoLog(log.New(os.Stdout, "[elastic] ", log.LstdFlags)),
			elastic.SetBasicAuth(
				os.Getenv("BONSAIO_KEY"), os.Getenv("BONSAIO_SECRET")))

	} else {
		url = elastic.DefaultURL
		clnt, err = elastic.NewClient(
			elastic.SetURL(url))
	}

	if err != nil {
		fmt.Fprintf(
			os.Stderr,
			"Unable to connect to ES cluster: (%s) %s ", url, err.Error())
		os.Exit(1)
	}

	return clnt
}

const (
	// DefaultRemote decides if the ES cluster is on Bonsai.io or it's local
	DefaultRemote = false

	// DefaultURL is the default ES local address
	DefaultURL = "http://127.0.0.1:9200"

	// DefaultVerbose decides if you wanna be bored by some noisy logs
	DefaultVerbose = false
)

// ClientOptionFunc is a function that configures a Client.
// It is used in NewClient.
type ClientOptionFunc func(*Client) error

// Client is the ElasticBook wrapper to an Elastic Client
// The "elastic" package is really inspiring!
type Client struct {
	client  *elastic.Client
	remote  bool
	url     string
	verbose bool
}

// NewClient Set up the default client
func NewClient(options ...ClientOptionFunc) (*Client, error) {
	c := &Client{
		// client:  client(false),
		remote:  DefaultRemote,
		url:     DefaultURL,
		verbose: DefaultVerbose,
	}
	for _, option := range options {
		if err := option(c); err != nil {
			return nil, err
		}
	}
	return c, nil
}

// SetElasticClient decide which kind of elastic Client use (local or remote)
func SetElasticClient(elasticClient *elastic.Client) ClientOptionFunc {
	return func(c *Client) error {
		if elasticClient != nil {
			c.client = elasticClient
			c.remote = true
		} else {
			c.client = client(false)
			c.remote = false
		}
		return nil
	}
}

// SetVerbose define the verbose logging
func SetVerbose(vvv bool) ClientOptionFunc {
	return func(c *Client) error {
		c.verbose = vvv
		return nil
	}
}

// SetURL define the current URL used (just for convenience)
func SetURL(u string) ClientOptionFunc {
	return func(c *Client) error {
		if u != "" {
			c.url = u
		} else {
			c.url = DefaultURL
		}
		return nil
	}
}

// ClientLocal connects to a local ES cluster
func ClientLocal() (*Client, error) {
	return NewClient(
		SetVerbose(true))
}

// ClientRemote connects to a remote ES cluster (Bonsai.io)
// Debug with this:
//   /_nodes/http?pretty=1
// https://github.com/olivere/elastic/wiki/Connection-Problems
func ClientRemote() (*Client, error) {
	return NewClient(
		SetVerbose(true),
		SetURL(os.Getenv("BONSAIO_HOST")),
		SetElasticClient(client(true)))
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
func (c *Client) Delete() {
	client := c.client

	r, err := client.DeleteIndex("elasticbook").Do()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}
	fmt.Fprintf(os.Stdout, "%+v\n", r)
}

// Health check the status of the cluster
func (c *Client) Health() *elastic.ClusterHealthResponse {
	cl := c.client
	info, err := cl.ClusterHealth().Do()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect a ES client %+v\n", err.Error())
		os.Exit(1)
	}

	return info
}

// timeParse converts a date (a string representation of the number of
// microseconds from the 1601/01/01
// https://chromium.googlesource.com/chromium/src/+/master/base/time/time_win.cc#56
//
// Quoting:
// From MSDN, FILETIME "Contains a 64-bit value representing the number of
// 100-nanosecond intervals since January 1, 1601 (UTC)."
func timeParse(microsecs string) time.Time {
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
	}

	t = t.Add(time.Duration(iRem) * time.Microsecond)

	// RFC1123 = "Mon, 02 Jan 2006 15:04:05 MST"
	// t.Format(time.RFC1123)
	return t
}

// Index takes a parsed structure and index all the Bookmarks entries
func (c *Client) Index(x *Root) {
	client := c.client

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

	var wg sync.WaitGroup
	var workForce = 5
	ch := make(chan Bookmark, workForce)

	for i := 0; i < workForce; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var b Bookmark
			var more bool

			for {
				b, more = <-ch
				if !more {
					return
				}
				if c.verbose {
					// fmt.Fprintf(os.Stdout, "%02d %s : %s\n", i, b.Name, b.URL)
				}
				_, err := client.Index().
					Index(IndexName).
					Type("bookmark").
					BodyJson(b.toIndexable()).
					Do()
				if err != nil {
					// TODO: Handle error
					fmt.Fprintf(os.Stderr, "%s\n", err.Error())
					panic(err)
				}
			}
		}()
	}

	count := x.Count().Total()
	bar := uiprogress.AddBar(count)
	bar.AppendCompleted()
	bar.PrependElapsed()
	bar.PrependFunc(func(b *uiprogress.Bar) string {
		return fmt.Sprintf("Node (%d/%d)", b.Current(), count)
	})

	uiprogress.Start()
	for _, x := range x.Roots.BookmarkBar.Children {
		ch <- x
		bar.Incr()
	}
	for _, x := range x.Roots.Synced.Children {
		ch <- x
		bar.Incr()
	}
	for _, x := range x.Roots.Other.Children {
		ch <- x
		bar.Incr()
	}

	uiprogress.Stop()
	close(ch)
	wg.Wait()

	// TODO: add BookmarkBar, Synced, Other
	// x.Roots.BookmarkBar.Children
	// x.Roots.Synced.Children
	// x.Roots.Other.Children
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
func (c *Client) Search() {
	client := c.client

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

// Version check the version of the cluster
func Version(c *elastic.Client, url string) string {
	esversion, err := c.ElasticsearchVersion(url)
	if err != nil {
		// Handle error
		panic(err)
	}
	return esversion
}
