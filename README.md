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

There is also some information about [search engine](./docs/search.md) implementation.


# REST API

`ryft-server` supports a few REST endpoints:

  - [/version](./docs/restapi.md#version)
  - [/search](./docs/restapi.md#search)
  - [/count](./docs/restapi.md#count)
  - [/files](./docs/restapi.md#files)

All examples assume the `ryft-server` host name is `ryftone-777`.

The main API endpoints are `/search` and `/count`. Both have almost the same parameters.
However, the /count endpoint does not transfer all found data, it will just print the number of matches and associated performance numbers.
The minimum required parameters are search query and the set of files to search, so simplest request is:

```{.sh}
curl "http://ryftone-777:8765/search?query=Joe&files=*.txt"
```

Of course it's possible to customize search. The following command will capture data surrounding of 5 bytes 
and fuzzy edit distance search (fuzziness=2) instead of fuzzy hamming search which is used by default:

```{.sh}
curl "http://ryftone-777:8765/search?query=Joe&files=*.txt&mode=feds&surrounding=5&fuzziness=2"
# - or -
curl --get --data-urlencode 'query=(RAW_TEXT CONTAINS "Joe")' \
  "http://ryftone-777:8765/search?files=*.txt&mode=feds&surrounding=5&fuzziness=2"
```

By default cluster mode is used. To execute a search on a single node use `local` query parameter:

```{.sh}
curl "http://ryftone-777:8765/search?query=Joe&files=*.txt&local=true"
```

To get human readably data (instead of base-64 encoded bytes) `format=utf8` can be used.
This parameter asks `ryft-server` to interpret found bytes as UTF-8 string:

```{.sh}
curl "http://ryftone-777:8765/search?query=Joe&files=*.txt&format=utf8"
```

The `/version` endpoint is used to check server's version:

```{.sh}
curl "http://ryftone-777:8765/version"
```

This request prints current server version and corresponding git hash number.
This information is extremelly useful for bug reporting.


See [REST API](./docs/restapi.md) document for more details.


# Command line tools

## ryftrest tool

`ryftrest` is a simple bash script which syntax is very similar to native `ryftprim` tool.
But there are a few differences (try `ryftrest --help` for detailed syntax):

- `ryftrest` can send requests to remote Ryft boxes (via `--address` option)
- `ryftrest` supports complex search queries (because `ryft-server` does)
- `ryftrest` can print found data records

The last feature in conjunction with `jq` JSON command line processor may be very useful:

```{.sh}
ryftrest -q '(RECORD.id CONTAINS "100310")' -f '*.pcrime' --local --format=xml --fields=ID,Date | jq ".results[].Date"
```

This command will print extracted list of date strings.
