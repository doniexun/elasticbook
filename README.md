# ElasticBook

Manage your Chrome bookmarks with Elasticsearch.

## CLI Options

### `help`

```
$ go run cmd/elasticbook/main.go -h

NAME:
   ElasticBook - Elasticsearch for your bookmarks

USAGE:
   main [global options] command [command options] [arguments...]

VERSION:
   0.0.1

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --command, -c  -c [alias|aliases|unalias|default|indices|index|count|health|parse|delete|web|persist]
   --search, -s   -s [term]
   --verbose, -V  I wanna read useless stuff
   --help, -h   show help
   --version, -v  print the version

```

### `count`

```
$ go run cmd/elasticbook/main.go -c count
- Mobile Bookmarks (33)
- Bookmarks Bar (9)
- Other Bookmarks (10669)
```

### Delete index

Sample usage (from `go run` code):

```
$ go run cmd/elasticbook/main.go -c delete
00] - elasticbook-20151226224925
01] - elasticbook-20151227201242
02] - foobar
[0-02]:  2
Want to delete the foobar index? [y/N]: y
&{Acknowledged:true}
```

### Create index

Sample usage (from `go run` code):

```
$ go run cmd/elasticbook/main.go -c index
Node (10960/10960) 10m3s [====================================================================] 100%
Index created
- Mobile Bookmarks (33)
- Bookmarks Bar (9)
- Other Bookmarks (10918)
```

### List indices

Sample usage (from `go run` code):

```
$ go run cmd/elasticbook/main.go -c indices
00] - elasticbook-20151227224924 (10956)
01] - elasticbook-20151228093443 (10960)
02] - elasticbook-20160103180734 (10979)
```

### List aliases

Sample usage (from `go run` code):

```
$ go run cmd/elasticbook/main.go -c aliases
00] - elasticbook-20151227224924 (10956)    [old]
01] - elasticbook-20151228093443 (10960)    []
02] - elasticbook-20160103180734 (10979)    [elasticbookdefault]
```

### Create alias

Sample usage (from `go run` code):

```
$ go run cmd/elasticbook/main.go -c alias
Index name: elasticbook-20151228073443
Alias name: default

00] - elasticbook-20151227224924:     []
01] - elasticbook-20151228073443:     [default]
```

## The mapping

Here the mapping used. There is a "name_suggest" for the completion.

```
{
  "elasticbook-20160111213240" : {
    "mappings" : {
      "bookmark" : {
        "properties" : {
          "date_added" : {
            "type" : "date",
            "format" : "dateOptionalTime"
          },
          "id" : {
            "type" : "string"
          },
          "meta_info" : {
            "properties" : {
              "stars_id" : {
                "type" : "string"
              },
              "stars_imageData" : {
                "type" : "string"
              },
              "stars_isSynced" : {
                "type" : "string"
              },
              "stars_pageData" : {
                "type" : "string"
              },
              "stars_type" : {
                "type" : "string"
              }
            }
          },
          "name" : {
            "type" : "string"
          },
          "name_suggest" : {
            "type" : "completion",
            "analyzer" : "simple",
            "payloads" : false,
            "preserve_separators" : true,
            "preserve_position_increments" : true,
            "max_input_length" : 50
          },
          "sync_transaction_version" : {
            "type" : "string"
          },
          "type" : {
            "type" : "string"
          },
          "url" : {
            "type" : "string"
          }
        }
      }
    }
  }
}
```

### Suggestion?

Yep, it's easy.

For instance: look for "ela" in this way:

```
POST /elasticbook-20160111213240/_suggest?pretty -d '{
    "yup" : {
        "text" : "ela",
        "completion" : {
            "field" : "name_suggest"
        }
    }
} '
```

And you'll receive this helpful suggestions:

```
{
  "_shards" : {
    "total" : 1,
    "successful" : 1,
    "failed" : 0
  },
  "yup" : [ {
    "text" : "ela",
    "offset" : 0,
    "length" : 3,
    "options" : [ {
      "text" : "ElasticSearch profiles index - Wiki",
      "score" : 14.0
    }, {
      "text" : "Blog - Bonsai - Hosted Elasticsearch",
      "score" : 14.0
    }, {
      "text" : "Elasticsearch - Installing Plugins",
      "score" : 14.0
    }, {
      "text" : "Bonsai - Hosted Elasticsearch",
      "score" : 14.0
    }, {
      "text" : "Discuss Elasticsearch, Logstash and Kibana | Elastic",
      "score" : 14.0
    } ]
  } ]
}
```

## Web interface

```
$ go run cmd/cli/main.go --web
[martini] listening on :3000 (development)
[martini] Started POST /elasticbook/search for [::1]:51415
[martini] Found a total of 20 bookmarks
[martini] Completed 200 OK in 461.663807ms
```

Base search:

![base search](https://cloud.githubusercontent.com/assets/456318/12184665/efb071a6-b596-11e5-92a9-c49dc19748e5.png)

Look for _elastic_ and found these:

![results](https://cloud.githubusercontent.com/assets/456318/12184666/efd03fcc-b596-11e5-9ec5-ced6d369ade3.png)


## Elasticsearch

- https://www.elastic.co/guide/en/elasticsearch/guide

### A POQ (plain old query)

```
$ go run cmd/elasticbook/main.go -c index

Node (9001/10870) 8m32s [=======================================================>------------]  83%
```

Once you indexed all your data, try this:

```
POST elasticbook/bookmark/_search

{
  "sort": [
    "_score",
    {
      "date_added": {
        "order": "asc"
      }
    }
  ],
  "query": {
    "filtered": {
      "query": {
        "bool": {
          "must": [
            {
              "match": {
                "name": "elastic"
              }
            }
          ]
        }
      }
    }
  }
}
```

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



