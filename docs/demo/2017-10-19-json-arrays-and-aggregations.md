# Demo - JSON arrays, aggregation - October 19, 2017

This demo covers a few new features:
- new `/count` output format
- support for JSON arrays
- new `/search/aggs` REST API endpoint


## New `/count` output format

Since the `0.14.0` version the output format of `/count` method has been changed
to be `/search` compatible (to be able to report errors in cluster mode).
All client's should be updated to use `.stats` instead of `.` to access statistics!

The `/count` is equivalent to `/search?limit=0`:

```{.sh}
$ ryftrest -q hello -f test1/1.txt --local --search --limit=0 -vvv
{
  "results": [],
  "stats": {
    "matches": 3,
    "totalBytes": 54,
     ...
    "extra": {
      "backend": "ryftprim",
       ...
    }
  }
}

$ ryftrest -q hello -f test1/1.txt --local --count -vvv
{
  "results": [],
  "stats": {
    "matches": 3,
    "totalBytes": 54,
     ...
    "extra": {
      "backend": "ryftprim",
       ...
    }
  }
}
```

If error occurred in cluster mode the whole request is not failed.
All successful results are still reported in statistics.
All errors are reported in `errors` field.


## Support for JSON arrays

In case the input JSON file contains root array the ryft-server is able to skip
additional `[`, `,` and `]`. JSON arrays are checked only for RECORD-based search
requests.

The following file is used as an input:

```{.sh}
$ cat /ryftone/test-a.json
[
{
  "text": "hello1",
  "foo":{"bar":100},
  "pos":"(10.1,10.1)"
},
{
  "text": "hello2",
  "foo":{"bar":200.0},
  "pos":"(20.2,20.2)"
},
{
  "text": "hello3",
  "foo":{"bar":"300"},
  "pos":{"lat":30.3, "lon":"30.3"}
},
{
  "text": "hello4",
  "foo":{"bar":"4e2"},
  "pos":[40.4,40.4]
},
{
  "text": "hello5",
  "foo":{"bar":0.5e3},
  "pos":"50.5,50.5"
},
{
  "text": "hello6",
  "no-foo":{"bar":600}
},
{
  "text": "hello7",
  "foo":{"no-bar":700}
},
{
  "text": "hello8",
  "foo":{"bar":[700]},
  "no-pos":[1, 2, 3]
}
]
```

And the following simple query works as expected:

```{.sh}
$ryftrest -q '(RECORD.text CONTAINS "hello")' -f test-a.json --local --search --format=json --data=test-data.json -vvv
{
  "results": [
    {
      ...
      "foo": {
        "bar": 100
      },
      "pos": "(10.1,10.1)",
      "text": "hello1"
    },
    {
      ...
      "foo": {
        "bar": 200
      },
      "pos": "(20.2,20.2)",
      "text": "hello2"
    },
    {
      ...
      "foo": {
        "bar": "300"
      },
      "pos": {
        "lat": 30.3,
        "lon": "30.3"
      },
      "text": "hello3"
    },
    {
      ...
      "foo": {
        "bar": "4e2"
      },
      "pos": [
        40.4,
        40.4
      ],
      "text": "hello4"
    },
    {
      ...
      "foo": {
        "bar": 500
      },
      "pos": "50.5,50.5",
      "text": "hello5"
    },
    {
      ...
      "no-foo": {
        "bar": 600
      },
      "text": "hello6"
    },
    {
      ...
      "foo": {
        "no-bar": 700
      },
      "text": "hello7"
    },
    {
      ...
      "foo": {
        "bar": [
          700
        ]
      },
      "no-pos": [
        1,
        2,
        3
      ],
      "text": "hello8"
    }
  ],
  "stats": {
    "matches": 8,
    "totalBytes": 524,
     ...
    "extra": {
      "backend": "ryftprim",
       ...
    }
  }
}
```

The results are OK, there is no "unexpected delimiter found" error message anymore.
The output DATA file still have root JSON array:

```{.sh}
$ head /ryftone/out.json
[
{
  "text": "hello1",
  "foo":{"bar":100},
  "pos":"(10.1,10.1)"
}
,
{
  "text": "hello2",
  "foo":{"bar":200.0},

$ tail /ryftone/out.json
  "foo":{"no-bar":700}
}
,
{
  "text": "hello8",
  "foo":{"bar":[700]},
  "no-pos":[1, 2, 3]
}

]
```

JSON arrays are properly handled within complex queries when `AND` or `OR` boolean
operations are used:

```{.sh}
$ryftrest -q '{RECORD.text CONTAINS "hello"} AND {RECORD.text CONTAINS "hello"}' -f test-a.json --local --search --format=json --data=out.json -vvv
{
  "results": [
    ...
  ],
  "stats": {
    "matches": 8,
    "totalBytes": 1072,
     ...
    "details": [
      {
        "matches": 8,
        "totalBytes": 524,
         ...
      },
      {
        "matches": 8,
        "totalBytes": 548,
         ...
      }
    ]
  }
}

$ ryftrest -q '{RECORD.text CONTAINS "hello"} OR {RECORD.text CONTAINS "hello"}' -f test-a.json --local --search --format=json --data=out.json -vvv
{
  "results": [
    ...
  ],
  "stats": {
    "matches": 8,
    "totalBytes": 1048,
     ...
    "details": [
      {
        "matches": 8,
        "totalBytes": 524,
         ...
      },
      {
        "matches": 8,
        "totalBytes": 524,
         ...
      }
    ]
  }
}
```

Note the total matches for `OR` is still 8 because duplicates are eliminated.
The output DATA file still have root array element.


Aggregations also can be be calculated on such type of DATA:

```{.sh}
$ ryftrest -q '{RECORD.text CONTAINS "hello"} OR {RECORD.text CONTAINS "hello"}' -f test-a.json --local --search --format=json --data=out.json -vvv --body '{"aggs":{"my_stat":{"stats":{"field":"foo.bar"}}}}'
{
  "results": [
    ...
  ],
  "stats": {
    "matches": 8,
    "totalBytes": 1048,
     ...
    "details": [
      {
        "matches": 8,
        "totalBytes": 524,
         ...
      },
      {
        "matches": 8,
        "totalBytes": 524,
         ...
      }
    ],
    "extra": {
      "aggregations": {
        "my_stat": {
          "avg": 300,
          "count": 5,
          "max": 700,
          "min": 100,
          "sum": 1500
        }
      }
    }
  }
}
```


## New `/search/aggs` REST API endpoint

There is a new REST API endpoint to run aggregation on already existing results.
It is very similar to `/search/show` when we can get results later.

We can save search results with `/count` and then use `/search/aggs` method
to calculate various aggregations.

```{.sh}
$ export SESSION=`ryftrest -q '{RECORD.text CONTAINS "hello"}' -f test.json --local --count --format=json --data=out.json --index=out.txt | jq -r  .stats.extra.session`
$ curl -s "http://localhost:8765/search/aggs?session=$SESSION&format=json" -d '{"aggs":{"my_stat":{"stats":{"field":"foo.bar"}}}}' | jq .{
  "results": [],
  "stats": {
     ...
    "extra": {
      "aggregations": {
        "my_stat": {
          "avg": 300,
          "count": 5,
          "max": 500,
          "min": 100,
          "sum": 1500
        }
      }
    }
  }
}

```

Note, the JSON arrays are not supported by `/search/show` and `/search/aggs` yet.
There is no `RECORD` indicator in the DATA/INDEX results to check `[` at the begin of file.


## New `ignore-missing-files` query option

The `/search` and `/count` supports new option: `ignore-missing-files=true`.
In this case the ryft server reports empty statistics instead of error message.

```{.sh}
$ curl -s "http://localhost:8765/count?local=true&query=hello&format=json&stats=true&ignore-missing-files=true" | jq .
{
  "results": [],
  "stats": {
    "matches": 0,
    "totalBytes": 0,
     ...
  }
}
```
