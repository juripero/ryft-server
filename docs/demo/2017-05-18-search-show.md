# Demo - New `/search/show` REST API method - May 18, 2017

This document demonstrates new feature to show already existing search results.

The regular `/count` or `/search?limit=` methods can be used to prepare search
results. These search results can be accessed with new `/search/show` REST API
method without any Ryft hardware calls.

## Input dataset

There are three nodes: `/ryftone/test1`, `/ryftone/test2` and `/ryftone/test3`.

```{.sh}
$ cat /ryftone/test1/1.txt
11111-A-hello-A-11111
22222-A-hello-A-22222
33333-A-hello-A-33333

$ cat /ryftone/test2/1.txt
11111-B-hello-B-11111
22222-B-hello-B-22222
33333-B-hello-B-33333
44444-B-hello-B-44444

$ cat /ryftone/test3/1.txt
11111-C-hello-C-11111
22222-C-hello-C-22222
33333-C-hello-C-33333
44444-C-hello-C-44444
55555-C-hello-C-55555
```

## Prepare search results

First of all we should prepare search results to access them later:

```{.sh}
$ ryftrest -q "hello" -i -f "1.txt" --address localhost:5001 --local \
  -u test:test -od 'data.txt' -oi 'index.txt' -w=5 | jq '{matches,host}'
```
```{.json}
{
  "matches": 3,
  "host": "node-1"
}
```

We have to specify INDEX and DATA files to access them later.


## Show search results (local mode)

Once search results are saved we can get any records passing corresponding
`offset` and `count` to the `/search/show` REST API method.

```{.sh}
$ curl -u test:test -s  "http://localhost:5001/search/show?data=data.txt&index=index.txt&delimiter=%0d%0a&offset=1&count=1&format=utf8&local=true" | jq .
```
```{.json}
{
  "results": [
    {
      "_index": {
        "file": "1.txt",
        "offset": 25,
        "length": 15,
        "fuzziness": 0,
        "host": "node-1"
      },
      "data": "22-A-hello-A-22"
    }
  ]
}
```

We can use various output formats, for example `RAW`:

```{.sh}
$ curl -u test:test -s  "http://localhost:5001/search/show?data=data.txt&index=index.txt&delimiter=%0d%0a&offset=0&count=2&format=raw&local=true" | jq .
```
```{.json}
{
  "results": [
    {
      "_index": {
        "file": "1.txt",
        "offset": 3,
        "length": 15,
        "fuzziness": 0,
        "host": "node-1"
      },
      "data": "MTEtQS1oZWxsby1BLTEx"
    },
    {
      "_index": {
        "file": "1.txt",
        "offset": 25,
        "length": 15,
        "fuzziness": 0,
        "host": "node-1"
      },
      "data": "MjItQS1oZWxsby1BLTIy"
    }
  ]
}
```

It's very important to provide valid parameters for the `/search/show` method.
The INDEX file and DATA file should be coherent (from the same Ryft call).
The `delimiter` parameter is also important. If delimiter does not equal to the
value specified for initial `/count` then the output data will be wrong!


## Cleanup search results

When search results are no longer required it is possible to delete them manually:

```{.sh}
$ curl -X DELETE -u test:test -s "http://localhost:5001/files?file=data.txt&file=index.txt&local=true" | jq .
```
```{.json}
{
  "data.txt": "OK",
  "index.txt": "OK"
}
```

But there is special `lifetime` parameter which allows to cleanup all output
result files automatically. So we can prepare search results with:

```{.sh}
ryftrest -q "hello" -i -f "1.txt" --address localhost:5001 --local --lifetime=1h \
  -u test:test -od 'data.txt' -oi 'index.txt' -w=5 | jq '{matches,host}'
```

All DATA and INDEX files will be available during one hour and then will be
removed by REST service itself.


## Sessions

It might be tedious to keep all these DATA and INDEX file names between
`/count` and `/search/show` calls. There is a new feature introduced called
*session*. A session contains all information about Ryft call:

- DATA file name
- INDEX file name
- delimiter
- etc...

For now each `/count` or `/search` call reports corresponding session token
in its `stat.extra.session` field. Session token is `base64` encoded data
(actually it is [JWT](https://jwt.io/) token).


```{.sh}
ryftrest -q "hello" -i -f "1.txt" --address localhost:5001 --local --lifetime=1h \
  -u test:test -od 'data.txt' -oi 'index.txt' -w=5 | jq .
```
```{.json}
{
  "matches": 3,
  "totalBytes": 67,
  "duration": 18,
  "dataRate": 0.003549787733289931,
  "fabricDuration": 17,
  "fabricDataRate": 0.00355,
  "host": "node-1",
  "extra": {
    "session": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpbmZvIjpbeyJkYXRhIjoiZGF0YS50eHQiLCJkZWxpbSI6IlxyXG4iLCJpbmRleCI6ImluZGV4LnR4dCIsIm1hdGNoZXMiOjMsInZpZXciOiIiLCJ3aWR0aCI6NX1dfQ.k07U6zhyj2CP4eiphGdFCpJvrmtPDkkDoc3aI4y9j-Q"
  }
}
```

So now, instead of passing all the file names we can pass one session token:

```{.sh}
$ export SESSION=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpbmZvIjpbeyJkYXRhIjoiZGF0YS50eHQiLCJkZWxpbSI6IlxyXG4iLCJpbmRleCI6ImluZGV4LnR4dCIsIm1hdGNoZXMiOjMsInZpZXciOiIiLCJ3aWR0aCI6NX1dfQ.k07U6zhyj2CP4eiphGdFCpJvrmtPDkkDoc3aI4y9j-Q
$ curl -u test:test -s  "http://localhost:5001/search/show?&offset=1&count=1&format=utf8&local=true&session=${SESSION}" | jq .
```
```{.json}
{
  "results": [
    {
      "_index": {
        "file": "1.txt",
        "offset": 25,
        "length": 15,
        "fuzziness": 0,
        "host": "node-1"
      },
      "data": "22-A-hello-A-22"
    }
  ]
}
```

Having session token it's possible to use random file names. Just pass
`{{random}}` keyword in file template:

```{.sh}
ryftrest -q "hello" -i -f "1.txt" --address localhost:5001 --local --lifetime=1h \
  -u test:test -od 'data-{{random}}.txt' -oi 'index-{{random}}.txt' -w=5 | jq .
```
```{.json}
{
  "matches": 3,
  "totalBytes": 67,
  "duration": 18,
  "dataRate": 0.003549787733289931,
  "fabricDuration": 17,
  "fabricDataRate": 0.00355,
  "host": "node-1",
  "extra": {
    "session": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpbmZvIjpbeyJkYXRhIjoiZGF0YS0xNGJmYWRkMjNmNjAzNWE0LnR4dCIsImRlbGltIjoiXHJcbiIsImluZGV4IjoiaW5kZXgtMTRiZmFkZDIzZjYwNzMyOS50eHQiLCJtYXRjaGVzIjozLCJ2aWV3IjoiIiwid2lkdGgiOjV9XX0.Sjdq1rh9qTE-EBTbfh8-OP9P2v64Yiv4MDSBZbJDMmg"
  }
}
```
Random file names can be very helpful especially with non-zero `lifetime`.

One important thing about session: it's the only way to get search results
in cluster mode. Session token keeps all information about each node these
search results belongs to.


## Show search results (cluster mode)

There are a three ryft server instances running on different ports: `5001`, `5002`, `5003`.
These instances uses different configuration. It's important to specify different
home directory for the `test` user per each instance.
See [POST files](./2016-09-29-post-files.md) demo for more details.

To prepare cluster search results just replace `--local` with `--cluster`:

```{.sh}
ryftrest -q "hello" -i -f "1.txt" --address localhost:5001 --cluster --lifetime=1h \
  -u test:test -od 'data-{{random}}.txt' -oi 'index-{{random}}.txt' -w=5 | jq .
```
```{.json}
{
  "matches": 12,
  "totalBytes": 267,
  "duration": 18,
  "dataRate": 0.015435187645207822,
  "fabricDuration": 17,
  "fabricDataRate": 0.015435,
  "details": [
    {
      "matches": 3,
      "totalBytes": 67,
      "duration": 18,
      "dataRate": 0.003549787733289931,
      "fabricDuration": 17,
      "fabricDataRate": 0.00355,
      "host": "node-1"
    },
    {
      "matches": 5,
      "totalBytes": 111,
      "duration": 17,
      "dataRate": 0.0062269323012408085,
      "fabricDuration": 16,
      "fabricDataRate": 0.006227,
      "host": "node-3"
    },
    {
      "matches": 4,
      "totalBytes": 89,
      "duration": 15,
      "dataRate": 0.005658467610677084,
      "fabricDuration": 15,
      "fabricDataRate": 0.005658,
      "host": "node-2"
    }
  ],
  "extra": {
    "session": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpbmZvIjpbeyJkYXRhIjoiZGF0YS0xNGJmYWU4NDA1ZmRkNjUyLnR4dCIsImRlbGltIjoiXHJcbiIsImluZGV4IjoiaW5kZXgtMTRiZmFlODQwNWZlMWYxZC50eHQiLCJsb2NhdGlvbiI6Imh0dHA6Ly8xMjcuMC4wLjE6NTAwMSIsIm1hdGNoZXMiOjMsIm5vZGUiOiJ1YnVudHUtdm0iLCJ2aWV3IjoiIiwid2lkdGgiOjV9LHsiZGF0YSI6ImRhdGEtMTRiZmFlODQwNWZkZDY1Mi50eHQiLCJkZWxpbSI6IlxyXG4iLCJpbmRleCI6ImluZGV4LTE0YmZhZTg0MDVmZTFmMWQudHh0IiwibG9jYXRpb24iOiJodHRwOi8vMTI3LjAuMC4xOjUwMDMiLCJtYXRjaGVzIjo1LCJub2RlIjoidWJ1bnR1LXZtIiwidmlldyI6IiIsIndpZHRoIjo1fSx7ImRhdGEiOiJkYXRhLTE0YmZhZTg0MDVmZGQ2NTIudHh0IiwiZGVsaW0iOiJcclxuIiwiaW5kZXgiOiJpbmRleC0xNGJmYWU4NDA1ZmUxZjFkLnR4dCIsImxvY2F0aW9uIjoiaHR0cDovLzEyNy4wLjAuMTo1MDAyIiwibWF0Y2hlcyI6NCwibm9kZSI6InVidW50dS12bSIsInZpZXciOiIiLCJ3aWR0aCI6NX1dfQ.LOX3Bx42AwQ2WUJ6qSjkCOk-s2gRGGqrORoZyUiQlsI"
  }
}
```

As we can see the session token is much bigger because now it contains all
the information for all cluster nodes.

Now let's talk a bit about indexing. We have various number of matches per
each node. `[3, 5, 4]` in our case. To get search results the flat indexing
is used:
- we can get `12=3+5+4` records total
- records `[0..2]` come from the first cluster node
- records `[3..7]` come from the second cluster node
- records `[8..11]` come from the third cluster node

The following query reports records from the first cluster node:

```{.sh}
$ curl -u test:test -s  "http://localhost:5001/search/show?&offset=1&count=1&format=utf8&session=${SESSION}" | jq .
```
```{.json}
{
  "results": [
    {
      "_index": {
        "file": "1.txt",
        "offset": 25,
        "length": 15,
        "fuzziness": 0,
        "host": "node-1"
      },
      "data": "22-A-hello-A-22"
    }
  ]
}
```

The following query reports records from the second and the third cluster nodes:

```{.sh}
$ curl -u test:test -s  "http://localhost:5001/search/show?&offset=7&count=2&format=utf8&session=${SESSION}" | jq .
```
```{.json}
{
  "results": [
    {
      "_index": {
        "file": "1.txt",
        "offset": 91,
        "length": 15,
        "fuzziness": 0,
        "host": "node-3"
      },
      "data": "55-C-hello-C-55"
    },
    {
      "_index": {
        "file": "1.txt",
        "offset": 3,
        "length": 15,
        "fuzziness": 0,
        "host": "node-2"
      },
      "data": "11-B-hello-B-11"
    }
  ]
}
```

Note, having these cluster session token we can send `/search/show` request
to any node and will have the same result (if records order is not taken into account).


## TODO

There is no VIEW file optimization yet. A `view` file is binary index
on Ryft's INDEX and DATA file to speedup random access.

The `ryftrest` tool also can be upgraded to support `--show` mode.
