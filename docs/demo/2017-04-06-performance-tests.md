This file contains performance metrics usage instructions.


# Introduction

Various input fileset is used:
- `RC_2016-01.data` - 32GB
- `twitter`

Various output formats:
- `raw`
- `utf8`
- `json`
- `null`


The queries are executed locally on the 313 Ryft box.
So there should be no network latency penalties.

NOTE, the performance numbers might be different for each call.


# Count vs Search operations

Let's start with a simple search for "Hello" text:

`-q '(RAW_TEXT CONTAINS ES("Hello"))' -i -f RC_2016-01.data --count`


If the `--count` operation is used, which means no found data is transferred,
then the results are the following:

```{.json}
{
  "matches": 171863,
  "totalBytes": 34164795882,
  "duration": 5942,
  "fabricDuration": 1692,
  "extra": {
    "performance": {
      "rest-count": {
        "engine": "2.524738ms",
        "prepare": "490.355µs",
        "total": "5.94624649s",
        "transfer": "5.943231397s"
      },
      "ryftprim": {
        "prepare": "58.741µs",
        "tool-exec": "5.944078502s"
      }
    }
  }
}
```

The `ryftprim` tool tells us that request took 5.942 seconds. But the bash
`time` command shows that this request took 5.957 seconds. So what this time
was spent for? Let's look at the numbers.

First of all, there is a small difference between `.duration` reported
by `ryftprim` tool and actual tool execution time `.extra.performance.ryftprim.tool-exec`
measured by ryft server. The difference is about 2 ms.

The total `.extra.performance.rest-count.total` ryft-server execution time
shows 5.946 seconds, again about 2 ms more than `ryftprim` execution time.

Perhaps remaining 11 ms (`5957-5942 = 2+2+11`) are related to curl execution
time. Although there is only 358 bytes of data in the response it also takes
some time to save this data on filesystem.

Let's run this search request with `--search` option enabled.

`-q '(RAW_TEXT CONTAINS ES("Hello"))' -i -f RC_2016-01.data --search`

The transferred data size is about 20MB. And bash `time` command shows 6.800 seconds.
Reported statistics:

```{.json}
{
  "matches": 171863,
  "totalBytes": 34164795882,
  "duration": 6190,
  "fabricDuration": 1692,
  "extra": {
    "performance": {
      "rest-search": {
        "engine": "2.558472ms",
        "prepare": "526.474µs",
        "total": "6.789210154s",
        "transfer": "6.786125208s"
      },
      "ryftprim": {
        "prepare": "68.275µs",
        "read-data": "585.937396ms",
        "tool-exec": "6.192527103s"
      }
    }
  }
}
```

There is also a 2.5 ms difference between time reported by `ryftprim` and
actual tool execution time measured by ryft server.

The same 11ms (`6800-6789=11`) is related to curl execution time.

There is a new `.extra.performance.ryftprim.read-data` time spent for reading
and parsing INDEX and DATA. And this time is relatively big - 585 ms. Note,
that this time also includes transfer to client. While INDEX and DATA are read
they also are transferring to the client.

Also please note, `.extra.performance.ryftprim.read-data` value might include part
of `ryfptrim` execution time if server's '`minimize-latency` options is `true`.


# Record-based search

Let's try record-based search and change `RAW_TEXT` to `RECORD`:

`-q '(RECORD CONTAINS ES("Hello"))' -i -f RC_2016-01.data --count`

For `--count` the amout of transferred data is quite small - 363 bytes.
Operation is done in 15.291 seconds:

```{.json}
{
  "matches": 162009,
  "totalBytes": 34164795882,
  "duration": 15275,
  "fabricDuration": 10932,
  "extra": {
    "performance": {
      "rest-count": {
        "engine": "3.049417ms",
        "prepare": "541.042µs",
        "total": "15.280633092s",
        "transfer": "15.277042633s"
      },
      "ryftprim": {
        "prepare": "49.835µs",
        "tool-exec": "15.278340991s"
      }
    }
  }
}
```

The difference between `.extra.performance.ryftprim.tool-exec` tool execution
time and the overal time reported by bash `time` command is relatively small:
`15291-15278 = 13ms`.

The `--search` operation transfers much more data - about 213MB:

`-q '(RECORD CONTAINS ES("Hello"))' -i -f RC_2016-01.data --search`

```{.json}
{
  "matches": 162009,
  "totalBytes": 34164795882,
  "duration": 15671,
  "fabricDuration": 10932,
  "extra": {
    "performance": {
      "rest-search": {
        "engine": "3.222789ms",
        "prepare": "526.15µs",
        "total": "22.553660637s",
        "transfer": "22.549911698s"
      },
      "ryftprim": {
        "prepare": "62.929µs",
        "read-data": "6.695910221s",
        "tool-exec": "15.673782556s"
      }
    }
  }
}
```

Although `ryftprim` execution time is almost the same - 15.673 seconds.
Total time reported by bash's `time` is much bigger - 25.801 seconds.

There are about 10 seconds extra. And this time consists of the following parts:
- `.extra.performance.ryftprim.read-data` takes 6.695 seconds
- curl transfer and output save takes `25.801-22.553 = 3.248` seconds

So we should pay additional attention to the read and transfer procedures.
Current values look too expensive.


# Let's check how the output format impacts performance.

All previous requests use default `format=raw`. Let's try other formats
and summarize the results in the table:

| Total time | Transfer size | Search request |
| ---------- | ------------- | -------------- |
| 26.260s    | 213MB         | `-q '(RECORD CONTAINS ES("Hello"))' -i -f RC_2016-01.data --search --format=raw`  |
| 25.481s    | 175MB         | `-q '(RECORD CONTAINS ES("Hello"))' -i -f RC_2016-01.data --search --format=utf8` |
| 26.607s    | 164MB         | `-q '(RECORD CONTAINS ES("Hello"))' -i -f RC_2016-01.data --search --format=json` |
| 16.446s    | 16MB          | `-q '(RECORD CONTAINS ES("Hello"))' -i -f RC_2016-01.data --search --format=null` |

Simple rule can be noted here: more data we have to read and transfer - longer execution.

One exception is `json` format. Although amount of data is a bit less comparing to `utf8`,
the total time is a bit bigger because we need to parse each of 162009 found record
as an JSON object. Of course it takes additional resources.

The `raw` format is also takes a bit longer comparing to `utf8`
because we need to `base64`-encode DATA of every found record.

The `null` format reports only INDEX information, no data transferred.
This format shows minimum execution time, very close to `ryftprim` execution time.
The `.extra.performance.ryftprim.read-data` is about 554 ms.


# Another option to check is Content-Type

We can specify JSON or CSV Content-Type with `--accept` option:

| Total time | Transfer size | Search request |
| ---------- | ------------- | -------------- |
| 24.199s    | 213MB         | `-q '(RECORD CONTAINS ES("Hello"))' -i -f RC_2016-01.data --search --accept json` |
| 26.765s    | 203MB         | `-q '(RECORD CONTAINS ES("Hello"))' -i -f RC_2016-01.data --search --accept csv` |

As we can see even if the transfer size is a bit less the CSV output takes a bit more time.


# Query combination

Additional effort is spent by ryft server for query combination: index unwinding
duplicate removal etc. Let's try `AND` oeprator first:

`-q '{RECORD CONTAINS ES("Hello")} AND {RECORD CONTAINS ES("Hello")}' -i -f RC_2016-01.data --search`

The total time is 24.450 seconds.
There are two `ryftptim` calls. For the second call we search the same keyword,
but the input set size is much less. Key points are:
- 15.943s - first `ryftprim` call
- 0.928s - second `ryftprim` call
- 6.855 - read and transfer

Detailed performance metrics look as:

```{.json}
{
  "rest-search": {
    "engine": "1.665316ms",
    "prepare": "577.487µs",
    "total": "24.265254396s",
    "transfer": "24.263011593s"
  },
  "ryftdec": {
    "final-post-proc": {
      "build-items": "93.911238ms",
      "sort-items": "37.369532ms",
      "total": "6.85503049s",
      "transform": "0s",
      "write-data": "0s",
      "write-index": "0s"
    },
    "intermediate-steps": [
      {
        "post-proc": "160.483155ms",
        "ryftprim": {
          "prepare": "70.93µs",
          "tool-exec": "15.943403028s"
        },
        "total": "15.943723118s"
      },
      {
        "post-proc": "173.482287ms",
        "ryftprim": {
          "prepare": "64.376µs",
          "tool-exec": "928.707587ms"
        },
        "total": "929.00126ms"
      }
    ],
    "prepare": "1.667609ms"
  }
}
```

There is `.ryftdec.final-post-proc` performance metrics section containing
information related to query decomposition. As we can see the data read and
transfer time is still remarkable big.

For the `OR` operator

`-q '{RECORD CONTAINS ES("Hello")} OR {RECORD CONTAINS ES("Hello")}' -i -f RC_2016-01.data --search`

the input file set for the both `ryftprim` calls are the same and total processing
time is increased up to 40.198 seconds:
- 16.467s - first `ryftprim` call
- 16.002s - second `ryftprim` call
- 7.143s - read and transfer

Detailed performance metrics:

```{.json}
{
  "rest-search": {
    "engine": "1.546922ms",
    "prepare": "553.074µs",
    "total": "40.153056708s",
    "transfer": "40.150956712s"
  },
  "ryftdec": {
    "final-post-proc": {
      "build-items": "49.974347ms",
      "sort-items": "106.634527ms",
      "total": "7.143789449s",
      "transform": "0s",
      "write-data": "0s",
      "write-index": "0s"
    },
    "intermediate-steps": [
      {
        "ryftprim": {
          "prepare": "96.276µs",
          "tool-exec": "16.467981271s"
        },
        "total": "16.468327481s"
      },
      {
        "post-proc": "315.352256ms",
        "ryftprim": {
          "prepare": "61.336µs",
          "tool-exec": "16.002242787s"
        },
        "total": "16.002528272s"
      }
    ],
    "prepare": "1.554462ms"
  }
}
```

The following point need to be explained:

| Metric        | AND     | OR       |
| ------------- | ------- | -------- |
| `build-items` | 93.91ms | 49.97ms  |
| `sort-items`  | 37.36ms | 106.63ms |

For the `AND` operation the `build-items` time is greater because there are
index unwinding present (we need to unwind indexes for the second `ryftprim` results).
The `OR` results do not require index to be unwinded.

But the `OR` results need to be processed to remove duplicates. That is why
the `sort-items` is greater for the `OR`.


# Query combination and file save

If we want to save processed results with `-oi` and `-od` options then the
results are the following:

| Total time | Write Index | Write Data | Search request |
| ---------- | ----------- | ---------- | -------------- |
| 42.546s    | 0.222s      | 0s         | `-q '{RECORD CONTAINS ES("Hello")} OR {RECORD CONTAINS ES("Hello")}' -i -f RC_2016-01.data --search -oi p-test.txt` |
| 45.022s    | 0s          | 4.050s     | `-q '{RECORD CONTAINS ES("Hello")} OR {RECORD CONTAINS ES("Hello")}' -i -f RC_2016-01.data --search -od p-test.data` |
| 46.909s    | 0.241s      | 4.001s     | `-q '{RECORD CONTAINS ES("Hello")} OR {RECORD CONTAINS ES("Hello")}' -i -f RC_2016-01.data --search -oi p-test.txt -od p-test.data` |

We can see that writing output INDEX file is quite fast,
but writing output DATA file (213MB) is expensive.


# Transform feature

Transform feature allows us to do minimal operations on found results.
There are two transform features built-in in ryft server: `match` and `replace`.
Also it's possible to call predefined external script.

| Total time | Ryfptrim | Transform | Search request |
| ---------- | -------- | --------- | -------------- |
| 24.340s    | 15.819s  | 4.806s    | `-q '(RECORD CONTAINS ES("Hello"))' -i -f RC_2016-01.data --search --transform 'match("^.*$")'` |
| 24.436s    | 15.959s  | 4.692s    | `-q '(RECORD CONTAINS ES("Hello"))' -i -f RC_2016-01.data --search --transform 'replace("^(.*)$", "$1")'` |
| 6m0.414s   | 15.812s  | 5m41.953s | `-q '(RECORD CONTAINS ES("Hello"))' -i -f RC_2016-01.data --search --transform 'script("cat",-)'` |

As we can see transform takes siginificant amount of time to process 162009
matches (213MB in total).

Please note, that there is a RHFS compatibility issue. That's why calling an
external script is so slow now. Once the ryft server does fork to call a script,
RHFS flushes all open file descriptors and this impacts performance dramatically.


# Catalog search

Catalog search in general equals to the regular file search. Additional step
here is index unwinding. We will use twitter monthly catalog:

```{.sh}
cd /ryftone
for file in twitter/201701*/*; do
  curl -X POST --data-binary "@${file}" \
    -H "Content-Type: application/octet-stream" \
    -s "http://localhost:8765/files?catalog=twitter1.json&file=${file}&local=true" | jq -c .
done
```

Catalog size is about14 GB.

So let's run search on catalog:

`-q (RAW_TEXT CONTAINS ES("Trump")) -i -c twitter1.json --search`

The request is done in 6.442s. There are 472689 matches found and total data
transferred size is about 61MB.

The performance looks as:

```{.json}
{
  "rest-search": {
    "engine": "48.561045ms",
    "prepare": "531.732µs",
    "total": "6.332789016s",
    "transfer": "6.283696239s"
  },
  "ryftdec": {
    "final-post-proc": {
      "build-items": "281.559474ms",
      "sort-items": "9.966594ms",
      "total": "2.084752159s",
      "transform": "0s",
      "write-data": "0s",
      "write-index": "0s"
    },
    "intermediate-steps": [
      {
        "post-proc": "614.303904ms",
        "ryftprim": {
          "prepare": "2.334068ms",
          "tool-exec": "3.56702764s"
        },
        "total": "3.570958449s"
      }
    ],
    "prepare": "48.574025ms"
  }
}
```


# Conclusion

There are a few items that we shold take a look:
- `read-data` performance metric for `ryftprim` and `ryftdec` search engines
- `write-data` performance metric for `ryftdec` search engine
- transform feature and especially call an external script

These items should be optimized in futher release
or at least objectives why it's not possible should be found.

The following items are not covered in the current phase:
- performance of posting files
