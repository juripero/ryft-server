The Ryft REST server runs as a daemon and provides access to the Ryft hardware.
It is written in [Go](https://golang.org/) using [Gin](https://github.com/gin-gonic/gin) HTTP framework.

# Build and Run

To build `ryft-server` just use the following commands:

```{.sh}
go get -d -v github.com/getryft/ryft-server
cd $GOPATH/src/github.com/getryft/ryft-server
make
```

Running server is even simplier:

```{.sh}
./ryft-server
```

Sometimes it's useful to run multiple instances on different ports:

```{.sh}
./ryft-server 0.0.0.0:9000 --debug
```

This command runs another server instance on port `9000` in debug mode.
Debug mode is used for testing to get detailed server's log messages.

It's also possible to create Debian package:

```{.sh}
make debian
```

See [build and run](./docs/buildandrun.md) document for more details.


# API endpoints

See [here](./docs/restapi.md)


# Search mode and query decomposition

Ryft supports several search modes:

- `es` for exact search
- `fhs` for fuzzy hamming search
- `feds` for fuzzy edit distance search
- `ds` for date search
- `ts` for time search
- `ns` for numeric search

Is no any search mode provided fuzzy hamming search is used by default for simple queries.
It is also possible to automatically detect search modes: if search query contains `DATE`
keyword then date search will be used, `TIME` keyword is used for time search,
and `NUMERIC` for numeric search.

Ryft server also supports complex queries containing several search expressions of different types.
For example `(RECORD.id CONTAINS "100") AND (RECORD.date CONTAINS DATE(MM/DD/YYYY > 04/15/2015))`.
This complex query contains two search expression: first one uses text search and the second one uses date search.
Ryft server will split this expression into two separate queries:
`(RECORD.id CONTAINS "100")` and `(RECORD.date CONTAINS DATE(MM/DD/YYYY > 04/15/2015))`. It then calls
Ryft hardware two times: the results of the first call are used as the input for the second call.

Multiple `AND` and `OR` operators are supported by the ryft server within complex search queries.
Expression tree is built and each node is passed to the Ryft hardware. Then results are properly combined.

Note, if search query contains two or more expressions of the same type (text, date, time, numeric) that query
will not be splitted into subqueries because the Ryft hardware supports those type of queries directly.


# Structured search formats

By default structured search uses `raw` format. That means that found data is returned as base-64 encoded raw bytes.

There are two other options: `xml` or `json`.

If input file set contains XML data, the found records could be decoded. Just pass `format=xml` query parameter
and records will be translated from XML to JSON. Moreover, to minimize output or to get just subset of fields
the `fields=` query parameter could be used. For example to get identifier and date from a `*.pcrime` file
pass `format=xml&fields=ID,Date`.

The same is true for JSON data. Example: `format=json&fields=Name,AlterEgo`.


# Preserve search results

By default all search results are deleted from the Ryft server once they are delivered to user.
But to have "search in the previous results" feature there are two query parameters: `data=` and `index=`.

First `data=output.dat` parameter keeps the search results on the Ryft server under `/ryftone/output.dat`.
It is possible to use that file as an input for the subsequent search call `files=output.dat`.

Note, it is important to use consistent file extension for the structured search
in order to let Ryft use appropriate RDF scheme!

For now there is no way to delete such intermediate result file.
At least until `DELETE /files` API endpoint will be implemented.

Second `index=index.txt` parameter keeps the search index under `/ryftone/index.txt`.

Note, according to Ryft API documentation index file should always have `.txt` extension!
