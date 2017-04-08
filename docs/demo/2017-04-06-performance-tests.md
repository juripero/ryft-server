This file contains performance metrics usage instructions.


# Introduction

Various input fileset is used:
- `RC_2016-01.data` - 32GB

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

`-q (RAW_TEXT CONTAINS ES("Hello")) -i -f RC_2016-01.data --count`


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

`-q (RAW_TEXT CONTAINS ES("Hello")) -i -f RC_2016-01.data --search`

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
and parsing INDEX and DATA. And this time is relatively big - 585 ms.

Please note, `.extra.performance.ryftprim.read-data` value might include part
of `ryfptrim` execution time if server's '`minimize-latency` options is `true`.


# Record-based search

Let's try record-based search and change `RAW_TEXT` to `RECORD`:

`-q (RECORD CONTAINS ES("Hello")) -i -f RC_2016-01.data --count`

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

`-q (RECORD CONTAINS ES("Hello")) -i -f RC_2016-01.data --search`

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
| 26.260s    | 213MB         | `-q (RECORD CONTAINS ES("Hello")) -i -f RC_2016-01.data --search --format=raw`  |
| 25.481s    | 175MB         | `-q (RECORD CONTAINS ES("Hello")) -i -f RC_2016-01.data --search --format=utf8` |
| 26.607s    | 164MB         | `-q (RECORD CONTAINS ES("Hello")) -i -f RC_2016-01.data --search --format=json` |
| 16.446s    | 16MB          | `-q (RECORD CONTAINS ES("Hello")) -i -f RC_2016-01.data --search --format=null` |

Simple rule can be noted here: more data we have to read and transfer - longer execution.

One exception is `json` format. Although amount of data is a bit less comparing to `utf8`,
the total time is a bit bigger because we need to parse each of 162009 found record
as an JSON object. Of course it takes additional resources.

The `raw` format is also takes a bit longer comparing to `utf8`
because we need to `base64`-encode DATA of every found record.

The `null` format reports only INDEX information, no any data transferred.
This format shows minimum execution time, very close to `ryftprim` execution time.
The `.extra.performance.ryftprim.read-data` is about 554 ms.
