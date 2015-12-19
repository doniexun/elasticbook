# ElasticBook

Manage your Chrome bookmarks with Elasticsearch.

## Elasticsearch

- https://www.elastic.co/guide/en/elasticsearch/guide

## Elastic

Elasticsearch client for Go.

- https://github.com/olivere/elastic/tree/release-branch.v3

## Martini

Classy web framework for Go.

- https://github.com/go-martini/martini

## BoltDB

An embedded key/value database for Go.

- https://github.com/boltdb/bolt

## Utils

### JQ

jq is a lightweight and flexible command-line JSON processor.

- https://stedolan.github.io/jq/
- https://stedolan.github.io/jq/manual/
- https://github.com/stedolan/jq

```
$ jq keys bookmarks_20151213.json
[
  "checksum",
  "roots",
  "version"
]
```

```
$ jq '.roots | keys' bookmarks_20151213.json
[
  "bookmark_bar",
  "other",
  "sync_transaction_version",
  "synced"
]
```

```
$ jq '.roots.bookmark_bar | keys' bookmarks_20151213.json
[
  "children",
  "date_added",
  "date_modified",
  "id",
  "name",
  "type"
]

$ jq '.roots.other | keys' bookmarks_20151213.json
[
  "children",
  "date_added",
  "date_modified",
  "id",
  "name",
  "type"
]
```

```
$ jq '.roots.other.children | length' bookmarks_20151213.json
10622
```

```
$ jq '.roots.other.children | .[42] | keys' bookmarks_20151213.json
[
  "date_added",
  "id",
  "meta_info",
  "name",
  "sync_transaction_version",
  "type",
  "url"
]
```

```
$ jq '.roots.other.children | .[] | .name ' bookmarks_20151213.json
...
...
...
```

```
$ jq '.roots.other.children | .[] | select(.url == "https://golang.org/")' bookmarks_20151215.json
{
  "date_added": "13094604045096757",
  "id": "37151",
  "name": "The Go Programming Language",
  "sync_transaction_version": "32917",
  "type": "url",
  "url": "https://golang.org/"
  }
```


