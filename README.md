The Ryft REST server runs as a daemon and provides access to the Ryft hardware.
It is written in [Go](https://golang.org/) using [Gin](https://github.com/gin-gonic/gin) HTTP framework.

# Build and Run

To build `ryft-server` just use the following commands:

```{.sh}
go get -d -v github.com/getryft/ryft-server
cd $GOPATH/src/github.com/getryft/ryft-server
make
```

Running server is even simpler:

```{.sh}
./ryft-server
```

Sometimes it's useful to run multiple instances on different ports:

```{.sh}
./ryft-server -l=:9000 --debug
```

This command runs another server instance on port `9000` in debug mode.
Debug mode is used for testing to get detailed server's log messages.

It's also possible to create a Debian package:

```{.sh}
make debian
```

Most of the options can be customized via configuration file.
See [build and run](./docs/buildandrun.md) document for more details.

There is also some information about [search engine](./docs/search/engine.md)
implementation and [authnetication](./docs/auth.md).


# REST API

`ryft-server` supports a few REST endpoints:

  - [/version](./docs/rest/README.md#version)
  - [/search](./docs/rest/search.md#search)
  - [/count](./docs/rest/search.md#count)
  - [/files](./docs/rest/files.md)

All examples assume the `ryft-server` host name is `ryftone-777`.

The main API endpoints are `/search` and `/count`. Both have almost the same parameters.
However, the `/count` endpoint does not transfer all found data.
Instead, it prints the number of matches and associated performance numbers.
The minimum required parameters are search query and the set of files to search.
Here's an example of a simple request:

```{.sh}
curl "http://ryftone-777:8765/search?query=Joe&files=*.txt"
```

Of course it's possible to customize the search. The following command captures 5 bytes of data surrounding the search value, and executtes a fuzzy edit distance search (fuzziness=2) instead of fuzzy hamming search, which is used by default:

```{.sh}
curl "http://ryftone-777:8765/search?query=Joe&files=*.txt&mode=feds&surrounding=5&fuzziness=2"
# - or -
curl --get --data-urlencode 'query=(RAW_TEXT CONTAINS "Joe")' \
  "http://ryftone-777:8765/search?files=*.txt&mode=feds&surrounding=5&fuzziness=2"
```

By default, cluster mode is used. To execute a search on a single node use `local` query parameter:

```{.sh}
curl "http://ryftone-777:8765/search?query=Joe&files=*.txt&local=true"
```

To get human readable data (instead of base-64 encoded bytes) `format=utf8` can be used.
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


See [REST API](./docs/rest/README.md) document for more details.

Some endpoints are protected. See [authentication](./docs/auth.md) document
and [corresponding demo](./docs/demo/2016-07-21-authentication.md).


# Command line tools

## ryftrest tool

`ryftrest` is a simple bash script with syntax that is very similar to the native `ryftprim` tool.
But there are a few differences (try `ryftrest --help` for detailed syntax):

- `ryftrest` can send requests to remote Ryft boxes (via `--address` option)
- `ryftrest` supports complex search queries (because `ryft-server` does)
- `ryftrest` can print found data records

The last feature in conjunction with `jq` JSON command line processor may be very useful:

```{.sh}
ryftrest -q '(RECORD.id CONTAINS "100310")' -f '*.pcrime' --local --format=xml --fields=ID,Date | jq ".results[].Date"
```

This command will print extracted list of date strings.

The `ryftrest` tool can be used to do any requests that `ryftprim` does and even more:

```{.sh}
ryftrest -q '(RAW_TEXT CONTAINS NUMBER("35" < NUM <= "50", ",", "."))' -f '*.pcrime' -vv
ryftrest -q '(RAW_TEXT CONTAINS CURRENCY("$300" < CUR <= "$500", "$", ",", "."))' -f '*.pcrime' -vv
ryftrest -q '(RAW_TEXT CONTAINS REGEX("$[3-5]00", PCRE_OPTION_DEFAULT))' -f '*.pcrime' -vv
ryftrest -q '((RECORD.Date CONTAINS TIME(11:50:00 < HH:MM:SS <= 11:55:00)) \
    AND(RECORD.ID CONTAINS FEDS("10029", CS=false, DIST=3, WIDTH=0)) \
    AND(RECORD.Description CONTAINS FHS("vehycle", CS=false, DIST=1, WIDTH=0))' -f '*.pcrime' -vv
```

For more detailed examples see:
[ryftrest sample 1](./docs/demo/2016-04-28-ryftrest.md) and [ryftrest sample 2](./docs/demo/2016-05-12-ryftrest.md)
