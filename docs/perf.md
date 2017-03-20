This document contains information about various performance metrics.

There are the following performance metrics:
- REST API core metrics (for `/search` and `/count` endpoints)
- search engine specific (ryftprim, ryftdec, etc)


# REST API core metrics

The `/search` and `/count` REST API endpoints consist of the following steps:
- Parse request parameters and prepare search operation
- call search engine to do search operation
- wait until all results are transferred
- send HTTP response to client

So the performance metrics `rest-search` or `rest-count` contain:
- `prepare` time between HTTP request is arrived and the search engine is called.
- `engine` time between the search engine is started and begin of transfer.
- `transfer` time between transfer begin and transfer end.
- `total` total request processing time.

```
HTTP request  -->
                    ] "prepare" search operation                      \
                    ] send asynchronous request to search "engine"     ] "total"
                    ] "transfer" of found records                     /
HTTP response <--
```

For example:

```{.json}
"performance": {
      "ryftone-313": {
        "rest-count": {
          "prepare": "1.042636ms",
          "engine": "56.932978ms",
          "transfer": "351.640492ms",
          "total": "409.616106ms"
        }
      }
    }
```


# Search engine metrics

There are many metrics for each search engine.

## `ryftprim` search engine metrics

The [ryftprim](./search/engine.md#ryftprim-search-engine) search engine
does the following steps:
- Parse request parameters and prepare command line arguments
- Do call `ryftprim` tool
- Read and transfer found data

So the performance metrics `ryftprim` contain:
- `prepare` time to check input fileset and prepare tool's command line.
- `tool-exec` the `ryftprim` tool execution time.
- `read-data` INDEX and DATA read time.

```
request  -->
             ] "prepare" search operation
             \
              |  call ryftprim tool        \
             /                              | read INDEX and DATA
                                           /
response <--
```

Note, in some cases the read operation might be started before `ryftprim` tool
is finished. So `tool-exec` and `read-data` are done in parallel.
