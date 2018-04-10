# Demo - Optimized aggregation - March 15, 2018

This demo describes a new optimized tool used to calculate aggregations.
This tool is written in pure C with minimum number of dynamic memory allocations.
New `ryft-server` Debian package automatically installs this tool and the
REST server uses it when it is possible.


## Optimized tool

The tool is called `ryft-server-aggs`. And it can be called manually:

```{.sh}
$ ryft-server-aggs --help
-h, --help        print this short help
-V, --version     print version if available
-q, --quiet       be quiet, disable verbose mode
-v, --verbose     enable verbose mode (also -vv and -vvv)
-X<N>, --concurrency=<N> do processing in N threads (8 by default)
--native          use "native" output format

-i<path>, --index=<path> path to INDEX file
-d<path>, --data=<path>  path to DATA file
-f<path>, --field=<path> field to access JSON data

-H<N>, --header=<N> size of DATA header in bytes
-D<N>, --delim=<N>  size of DATA delimiter in bytes
-F<N>, --footer=<N> size of DATA footer in bytes

-b<N>, --index-chunk=<N> size of INDEX chunk in bytes (512MB by default)
-B<N>, --data-chunk=<N>  size of DATA chunk in bytes (512MB by default)
-R<N>, --max-records=<N> maximum number of records per DATA chunk (16M by default)

The following signals are handled:
  SIGINT, SIGTERM - stop the tool
```

At the input the DATA and INDEX file path (`--data` and `--index`) should be provided
and at least one field `--field` to calculate statistics on. Only JSON data is
supported by optimized tool.

Also it is important to provide correct delimiter length `--delim`. This information
is used to calculate appropriate record offset in DATA file. For JSON arrays the
lengths of header and footer also can be specified.

```{.sh}
$ ryft-server-aggs --data=/ryftone/test-10M.bin --index=/ryftone/test-10M.txt -D2 --field=foo -X8 -vvv

...
[{"sum2":333333283333458493440.000000, "sum":49999995000000.000000, "min":0.000000, "max":9999999.000000, "count":10000000}]
```

The data processing is done on chunk-by-chunk basis to be able to handle very large files.
Multiple threads are used to process DATA chunks. There is dedicated `--concurrency`
command line option. If `--concurrency=0` then all processing is done on the main thread.
For any other concurrencies the DATA processing is done on the dedicated processing threads
while main thread still parses indices and does prepare next DATA chunk.

The memory mapping is used to access INDEX and DATA files. So there is no additional
data copying during "read" operation.

Many fields can be specified. In this case all these fields are processed in one
iteration.


## REST server customization

There is dedicated section in configuration file to customize aggregation processing:

```{.yaml}
backend-options:
  aggregations:
    optimized-tool: /usr/bin/ryft-server-aggs  # path to optimized tool (comment to disable)
    max-records-per-chunk: 16M       # maximum number of records per DATA chunk
    data-chunk-size: 1GB             # maximum DATA chunk size
    index-chunk-size: 1GB            # maximum INDEX chunk size
    concurrency: 8                   # number of parallel threads to calculate aggregation on
    engine: auto                     # aggregation engine, one of: auto, native, optimized
```

Almost all of these options can be overriden via corresponding request's tweaks:

```{.sh}
$ curl -s "http://localhost:8765/search/aggs?data=test-10M.bin&index=test-10M.txt&delimiter=%0d%0a&format=json&performance=true" --data '{"tweaks":{"aggs":{"concurrency":2,"engine":"auto"}}, "aggs":{"1":{"stats":{"field":"foo"}},"2":{"avg":{"field":"foo1"}}}}' | jq .stats.extra
{
  "aggregations": {
    "1": {
      "avg": 4999999.5,
      "count": 10000000,
      "max": 9999999,
      "min": 0,
      "sum": 49999995000000
    },
    "2": {
      "value": 5000000
    }
  },
  ...
}
```

The engine `native` and `optimized` can be used to compare performance of
Go and C based implementations.
