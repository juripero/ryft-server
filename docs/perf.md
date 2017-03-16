This document contains information about various performance metrics.

There are the following performance metrics:
- REST API core metrics (for `/search` and `/count` endpoints)
- search engine specific


# REST API core metrics

The `/search` and `/count` REST API endpoints consist of the following steps:
- HTTP request is arrived - parse request and prepare search operation
- call search engine to do search operation
- wait until all results are transferred
- send HTTP response to client

So the performance metrics `rest-search` or `rest-count` contain:
- `prepare` time between HTTP request is arrived and the search engine is called.
- `engine` time between the search engine is started and begin of transfer.
- `transfer` time between transfer begin and transfer end.
- `total` total request processing time.

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
