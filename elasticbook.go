package elasticbook

// DONE: parse date
// DONE: add mapping to date
// DONE: add remote cluster
// DONE: add indexing
// DONE: add progress bar
// DONE: better query results (score)
// DONE: add counter in aliases/indices view
// DONE: add alias creation
// DONE: add alias deletion
// DONE: add default alias creation
// DONE: add alias switch
// DONE: add alias check for double/existing
// TODO: add query time ranged
// TODO: add fulltext search
// TODO: add query CLI
// TODO: add web interface
// TODO: refactor doubled code
// TODO: add Doctor for recovery/rollback default index
// TODO: add safety check when switching alias
// DONE: add 'by-index' in alias CLI
// DONE: add 'by-index' in defaultAlias CLI

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gosuri/uiprogress"
	"github.com/zeroed/elasticbook/utils"
)

// https://godoc.org/gopkg.in/olivere/elastic.v3
import "gopkg.in/olivere/elastic.v3"

// DefaultIndexName is the Elasticsearch index
const DefaultIndexName = "elasticbook"

// DefaultAliasName is the Elasticsearch alias used in Searches
const DefaultAliasName = "elasticbookdefault"

// TypeName is the type used
const TypeName = "bookmark"

// DefaultFields is where to look when looking for bookmarks
var DefaultFields = []string{"name", "url"}

// Root is the root of the Bookmarks tree
type Root struct {
	Checksum string `json:"checksum"`
	Version  int    `json:"version"`
	Roots    Roots  `json:"roots"`
}

// Count returns a map with the RootFolder name and the count
func (r *Root) Count() (c *CountResult) {
	c = new(CountResult)
	c.Add(r.Roots.BookmarkBar.Name, len(r.Roots.BookmarkBar.Children))
	c.Add(r.Roots.Other.Name, len(r.Roots.Other.Children))
	c.Add(r.Roots.Synced.Name, len(r.Roots.Synced.Children))

	return
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
			// elastic.SetInfoLog(log.New(os.Stdout, "[elastic] ", log.LstdFlags)),
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
		SetVerbose(false),
		SetURL(os.Getenv("BONSAIO_HOST")),
		SetElasticClient(client(true)))
}

// Alias creates an alias.
// It's enforced a constraint though: "No more than one index per alias"
// This means that, if the alias already exists, this method returns
// false.
func (c *Client) Alias(indexName string, aliasName string) (bool, error) {
	existingAliases, err := c.AliasNames()
	if err != nil {
		return false, err
	}

	if utils.ContainsString(existingAliases, aliasName) {
		return false, nil
	}

	client := c.client
	ack, err := client.Alias().Add(indexName, aliasName).Do()
	if err != nil {
		return false, err
	}
	return ack.Acknowledged, nil
}

// AliasNames returns the list of existing aliases (just the names,
// sorted).
// Due to the constraint enforced by Client#Alias this slice should not
// contains dupes (^_^)
func (c *Client) AliasNames() ([]string, error) {
	client := c.client
	info, err := client.Aliases().Index("_all").Do()
	if err != nil {
		return nil, err
	}

	ins := info.Indices
	var aliasNames []string
	for _, v := range ins {
		for _, x := range v.Aliases {
			aliasNames = append(aliasNames, x.AliasName)
		}
	}

	sort.Strings(aliasNames)
	return aliasNames, nil
}

// Aliases returns the list of existing aliases
func (c *Client) Aliases() ([]string, error) {
	client := c.client
	info, err := client.Aliases().Index("_all").Do()
	if err != nil {
		return nil, err
	}

	ins := info.Indices
	names := make([]string, len(ins))
	i := 0
	for k, v := range ins {
		var vs []string
		for _, x := range v.Aliases {
			vs = append(vs, x.AliasName)
		}

		c, err := client.Count(k).Do()
		if err != nil {
			return nil, err
		}

		kv := fmt.Sprintf("%s (%d): \t\t[%s]", k, c, strings.Join(vs, ", "))
		names[i] = kv
		i++
	}

	sort.Strings(names)
	return names, nil
}

// Default switch the default alias to the given index name (if it
// exists).
// Returns true if the switch is successful.
func (c *Client) Default(indexName string) (bool, error) {
	client := c.client

	ia, err := c.indexAliases()
	if err != nil {
		return false, err
	}

	if _, ok := ia[indexName]; !ok {
		err := fmt.Errorf("Index %s does not exists", indexName)
		return false, err
	}

	aliasService := client.Alias()
	for k, vs := range ia {
		if utils.ContainsString(vs, DefaultAliasName) {
			_, err := aliasService.Remove(k, DefaultAliasName).Do()
			if err != nil {
				return false, err
			}
		}
	}

	ack, err := aliasService.Add(indexName, DefaultAliasName).Do()
	if err != nil {
		return false, err
	}

	return ack.Acknowledged, nil
}

// Delete drops the index
func (c *Client) Delete(indexName string) {
	client := c.client

	r, err := client.DeleteIndex(indexName).Do()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}
	fmt.Fprintf(os.Stdout, "%+v\n", r)
}

// Doctor adds the Default alias to an index
func (c *Client) Doctor() {
}

// Health check the status of the cluster
func (c *Client) Health() (*elastic.ClusterHealthResponse, error) {
	cl := c.client
	return cl.ClusterHealth().Do()
}

// Indices returns the list of existing indices
func (c *Client) Indices() ([]string, error) {
	client := c.client
	ins, err := client.IndexNames()
	if err != nil {
		return nil, err
	}

	names := make([]string, len(ins))
	for i, n := range ins {
		c, err := client.Count(n).Do()
		if err != nil {
			return names, err
		}
		names[i] = fmt.Sprintf("%s (%d)", n, c)
	}

	sort.Strings(names)
	return names, nil
}

// IndexNames returns the list of existing indices (just the names)
func (c *Client) IndexNames() ([]string, error) {
	client := c.client
	ins, err := client.IndexNames()
	if err != nil {
		return nil, err
	}

	sort.Strings(ins)
	return ins, nil
}

// Index takes a parsed structure and index all the Bookmarks entries
func (c *Client) Index(x *Root) (bool, error) {
	client := c.client

	indexName := c.newIndexName()
	if exists, _ := client.IndexExists(indexName).Do(); !exists {
		_, err := client.CreateIndex(indexName).Do()
		if err != nil {
			return false, err
		}
	}

	ins, err := client.IndexNames()
	if err != nil {
		return false, err
	}

	if len(ins) == 1 {
		_, err := client.Alias().Add(indexName, DefaultIndexName).Do()
		if err != nil {
			return false, err
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
				_, err := client.Index().
					Index(indexName).
					Type(TypeName).
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

	return true, nil
}

// Parse run the JSON parser
func (c *Client) Parse(b []byte) (*Root, error) {
	x := new(Root)
	err := json.Unmarshal(b, &x)
	return x, err
}

// Search is the API for searching
func (c *Client) Search(term string) (*elastic.SearchResult, error) {
	client := c.client

	q := elastic.NewMultiMatchQuery(term, DefaultFields...).
		ZeroTermsQuery("none").
		QueryName("elasticbookSearch").
		PrefixLength(2).
		Fuzziness("AUTO").
		Type("most_fields").
		FieldWithBoost("name", float64(2))

	sr, err := client.Search().
		Index(DefaultAliasName).
		Type(TypeName).
		Query(q).
		Explain(true).
		// Sort("date_added", true). // No _score if this is enabled
		From(0).
		Size(100).
		Pretty(true).
		Do()

	return sr, err
}

// Unalias deletes an alias
func (c *Client) Unalias(aliasName string) (bool, error) {
	client := c.client
	info, err := client.Aliases().Index("_all").Do()
	if err != nil {
		return false, err
	}

	indexAliases := make(map[string][]string)

	for k, v := range info.Indices {
		var vs []string
		for _, x := range v.Aliases {
			vs = append(vs, x.AliasName)
		}
		indexAliases[k] = vs
	}

	aliasService := client.Alias()
	for k, vs := range indexAliases {
		if utils.ContainsString(vs, aliasName) {
			_, err := aliasService.Remove(k, aliasName).Do()
			if err != nil {
				return false, err
			}
		}
	}

	return true, nil
}

// URL returns the current ES client/cluster URL
func (c *Client) URL() string {
	return c.url
}

// Version check the version of the cluster
func (c *Client) Version() string {
	client := c.client
	esversion, err := client.ElasticsearchVersion(c.url)
	if err != nil {
		fmt.Fprintf(
			os.Stderr,
			"Unable to detect ES cluster version: (%s) %s ", c.url, err.Error())
		os.Exit(1)
	}

	return esversion
}

func (c *Client) indexAliases() (map[string][]string, error) {
	client := c.client
	info, err := client.Aliases().Index("_all").Do()
	if err != nil {
		return nil, err
	}

	ia := make(map[string][]string)

	for k, v := range info.Indices {
		var vs []string
		for _, x := range v.Aliases {
			vs = append(vs, x.AliasName)
		}
		ia[k] = vs
	}

	return ia, nil
}

func (c *Client) newIndexName() string {
	t := time.Now().UTC()
	s := fmt.Sprintf("%d%02d%02d%02d%02d%02d",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second())
	return fmt.Sprintf("%s-%s", DefaultIndexName, s)
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
