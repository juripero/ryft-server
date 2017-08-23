# Demo - aggregation functions - August 17, 2017

This demo shows the new feature of ryft-server: aggregation functions.

Note, this is some kind of simplicity because the same post-processing can be done
via already existing `/run` endpoint (to run custom script or application).


## POST /search and /count

Now the main `/search` and `/count` endpoints support custom JSON object
as a request's body. So the following requests are the same:

```{.sh}
$ curl -s "http://localhost:8765/search?query=hello&file=1.txt" | jq -c .results[]

$ curl -s "http://localhost:8765/search" --data '{"query":"hello", "files":["1.txt"]}' | jq -c .results[]
```

The second case is preferred for long requests. Moreover, it's possible to
save request to a file and then use it:

```{.sh}
$ cat req1.json
{
	"query": "hello",
	"files": ["1.txt"]
}

$ curl -s "http://localhost:8765/search" --data @req1.json | jq -c .results[]
```

One more interesting feature here is parameter overriding. Any parameter from JSON
body can be overriden with corresponding URL query parameter. If we want use the
same `req1.json` file but change the requested query to "world" we can run:

```{.sh}
$ curl -s "http://localhost:8765/search?query=world" --data @req1.json | jq -c .results[]
```


## Aggregations parameters

Because aggregation specification is quite long there is no corresponding
URL query parameter. So aggregations can be requested only with JSON body.

The request format is very familiar. For example:

```{.sh}
ryftrest -q '(RECORD CONTAINS "hello")' -f test.json --search --format=json \
  --body '{"aggs":{"my_test":{"avg":{"field":"foo.bar"}}}}' | jq .stats
```

The requested "average" will be reported with `"my_test"` name in extra
statistics:

```{.json}
{
  "matches": 7,
  ...
  "extra": {
    "aggregations": {
      "my_test": {
        "value": 300
      }
    }
}
```


## Data formats

It is important to specify data format to perform aggregations: "XML", "JSON" or "UTF8".

For `XML` or `JSON` it is possible to access any nested field with dot-separated
notation: `{ "avg": {"field": "foo.bar.rate" }}`.

The "UTF8" format can be used for `NUMERIC` search for example. Each found record in this
case should be valid float-point number:

```{.sh}
$ ryftrest -q '(RAW_TEXT CONTAINS NUMBER(110 < NUM < 410))' -f test.json --search --format=utf8 --body '{"aggs":{"my_test":{"stats":{"field":"."}}}}' | jq .stats

{
  "matches": 3,
  ...
  "extra": {
    "aggregations": {
      "my_test": {
        "avg": 300,
        "count": 3,
        "max": 400,
        "min": 200,
        "sum": 900
      }
    }
}
```


## Aggregation functions and engines (internal)

There is an `aggregation engine` abstraction. It is used to do actual data
processing. Several aggregation functions can be referred to the same
aggregation engine.

For example, if we need just `min` and `max`:

```{.sh}
$ cat req2.json
{
	"aggs": {
		"my_min": { "min": {"field": "foo.bar" }},
		"my_max": { "max": {"field": "foo.bar" }}
	}
}

$ ryftrest -q '(RECORD CONTAINS "hello")' -f test.json --search --format=json --body @req2.json | jq .stats
{
  "matches": 7,
  ...
  "extra": {
    "aggregations": {
      "my_max": {
        "value": 500
      },
      "my_min": {
        "value": 100
      }
    }
  }
}
```

then just one aggregation engine is used. This engine calculates both minimum and maximum.
There is no need to parse, extract field for each found record twice.


## Cluster mode

All implemented aggregations work seamless in cluster mode. Each processing node
reports aggregation in some "internal" format (actually just aggregation engines
are reported) and then these intermediate aggregations are combined into final
object which is reported to user.


## Supported aggregations

The following list of aggregations are supported by ryft-server:

- [Min](../rest/aggs.md#min-aggregation)
- [Max](../rest/aggs.md#max-aggregation)
- [Sum](../rest/aggs.md#sum-aggregation)
- [Value Count](../rest/aggs.md#value-count-aggregation)
- [Avg](../rest/aggs.md#avg-aggregation)
- [Stats](../rest/aggs.md#stats-aggregation)
- [Extended Stats](../rest/aggs.md#extended-stats-aggregation)
- [Geo Bounds](../rest/aggs.md#geo-bounds-aggregation)
- [Geo Centroid](../rest/aggs.md#geo-centroid-aggregation)
