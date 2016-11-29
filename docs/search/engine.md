This document contains detailed search engine description.

# Abstract search engine

The Ryft hardware can be accessed via `libryftone` library or `ryftprim` command line tool.
To unify access to Ryft hardware a search engine abstraction was designed. Moreover,
this conception was naturally fit in cluster mode.

Search engine interface looks like:

```{.go}
type Engine interface {
	Search(cfg *Config) (*Result, error)
	Count(cfg *Config) (*Result, error)
	Files(dir string) (*DirInfo, error)

	Options() map[string]interface{}
}
```

Most of the methods are related to the corresponding REST API endpoints:
[/search](../rest/search.md#search), [/count](../rest/search.md#count) and [/files](../rest/files.md).
The `Options()` method is used to get search engine's internal options which initially could be
customized via [search configuration file](../buildandrun.md#search-configuration).

The `ryft-server` uses this abstract search engine in its main code
so we can easily change actual search engine implementation.
For example, for test purposes it's quite easy to create fake search engine
which will use simple `grep` tool on local filesystem.

There is the list of search engine implementations:

- [ryftprim](#ryftprim-search-engine) uses `ryftprim` command line tool
- [ryftone](#ryftone-search-engine) uses `libryftone` library from Ryft Open API
- [ryfthttp](#ryfthttp-search-engine) uses another `ryft-server` instance
- [ryftmux](#ryftmux-search-engine) multiplexes results from several search engines
- [ryftdec](#ryftdec-search-engine) decomposes complex search queries

The `ryftprim` and `ryftone` search engines are used for local search.
The `ryftdec` stays ahead and translates complex search queries into several calls to backend:

```
                               ------------
                         --->  | ryftprim |  --->  /usr/bin/ryftprim
                        /      ------------
      -----------      /
--->  | ryftdec |  ----           - OR -
      -----------      \
                        \      ------------
                         --->  | ryftone  |  --->  /usr/lib/libryftone.so
                               ------------
```

In cluster mode we use a set of `ryfthttp` search engines to access remote Ryft boxes
and `ryftmux` to multiplex all the results received:

```
                                ------------
                          --->  | *local*  |  --->  local search engine
                         /      ------------
                        /
      -----------      /        ------------
--->  | ryftmux |  ---------->  | ryfthttp |  --->  cluster node #1
      -----------      \        ------------
                        \           ...
                         \      ------------
                          --->  | ryfthttp |  --->  cluster node #N
                                ------------
```


## Import search backends

Search engine implementation should register its factory in global list
to let user code create this search engine by name. Usually factory registration
is placed into `init()` function. Here is a trick how to register search engines,
just use unused import for side effects:

```{.go}
import (
	"github.com/getryft/ryft-server/search"
	_ "github.com/getryft/ryft-server/search/ryftprim"
	_ "github.com/getryft/ryft-server/search/ryftone"
	_ "github.com/getryft/ryft-server/search/ryfthttp"
)
```


# `ryftprim` search engine

The `ryftprim` search engine uses `ryftprim` command line tool to access Ryft hardware.
Actually `ryftprim` tool uses `libryftone` library internally, but due to thread-safe
limitations of `libryftone` we have to spawn separate process per each search call.

Implementation sends corresponding search command via the `ryftprim` tool and then
parses generated index and data files.

## `ryftprim` options

The `ryftprim` search engine supports the following options (they can be customized
via [search configuration file](../buildandrun.md#search-configuration)):

- `instance-name` - the search engine instance name. This name is used to distinguish different instances.
- `ryftprim-exec` - path to the `ryftprim` tool. By default it is `/usr/bin/ryftprim`.
- `ryftprim-legacy` - use `ryftprim` legacy mode to get machine-readable statistics. By default it is `true`.
- `ryftone-mount` - Ryft main volume. By default it is `/ryftone`.
- `open-poll` - open file poll timeout. By default it is "50ms".
  The engine will try to open index or data file many times using this timeout.
- `read-poll` - read file poll timeout. By default it is "50ms".
  The engine will try to read index or data file many times using this timeout.
- `read-limit` - the limit of read attempts. By default it is "100".
  After 100 read fails the engine will stop reading and return an error.
- `keep-files` - keep intermediate data and index files in order to implement server's
  [--keep option](../buildandrun.md#keeping-search-results).
- `index-host` - cluster's node name. Engine marks all found record indexes with
  cluster node's name. This feature let user distinguish where record come from.

Note, the `ryftprim` search engine's working directory is `$ryftone-mount/$instance-name`.


# `ryftone` search engine

The `ryftone` search engine uses `libryftone` library to access Ryft hardware. See
[Ryft Open API](http://info.ryft.com/acton/attachment/17117/f-0002/1/-/-/-/-/Ryft-Open-API-Library-User-Guide.pdf).

Implementation is very similar to `ryftprim` search engine. It also sends corresponding
search command to `libryftone` library and then parses generated index and data files.

For now, due to thread-safe issue it is not recommended to use `ryftone` search engine.
To ignore `libryftone` linking just pass "noryftone" tag to the `go build`:

```{.sh}
go build -tags "noryftone"
```

This is done by default by the `Makefile`.

## `ryftone` options

The `ryftone` search engine supports the same options as [ryftprim](#ryftprim-options) above
except for `ryftprim-exec`.


# `ryfthttp` search engine

The `ryfthttp` search engine uses another `ryft-server` instance to access Ryft hardware.
This search engine is used in cluster mode to forward search queries to remote Ryft boxes.

## `ryfthttp` options

The `ryfthttp` search engine supports the following options:

- `server-url` - remote `ryft-server` address including host name and port.
  By default it is "http://localhost:8765".
- `local-only` - flag to use local search on remote Ryft box. Related to `local=` server's query parameter.
- `skip-stat` - flag to skip statistics. Related to `stats=` server's query parameter.
- `index-host` - cluster's node name. Engine marks all found record indexes with
  cluster node's name. This feature let user distinguish where record come from.
  Usually this option is not applied because all found indexes should be already marked
  by the local search engine (`ryftprim` or `ryftone`).

Usually these options are filled by the `ryft-server` automatically when cluster configuration is built.
But it's also possible to customize them via [search configuration file](../buildandrun.md#search-configuration).


# `ryftmux` search engine

The `ryftmux` search engine does multiplexing results from several search engines.
For example, one `ryftprim` and a few `ryfthttp` from cluster's nodes.

This search engine is configured and used by the `ryft-server` internally.


# `ryftdec` search engine

The `ryftdec` search engine is a kind of filter. It uses backend for its work - another search engine instance.
The main purpose of `ryftdec` is to decompose complex search query into several simple sub-queries.
These simple sub-queries are forwarded to backend and the results are properly combined.

For example let's process this complex [search query](../rest/search.md#search-query-parameter):

```
(RECORD.id CONTAINS "10030") AND (RECORD.date CONTAINS DATE(MM/DD/YYYY > 04/13/2015)) AND (RECORD.date CONTAINS TIME(HH:MM:SS > 11:20:00))
```

This query contains three sub-queries of different types. The `ryftdec` decomposes
complex expression into the following tree:

- `AND` operator
  - A: `(RECORD.id CONTAINS "10030")`
  - `AND` operator
    - B: `(RECORD.date CONTAINS DATE(MM/DD/YYYY > 04/13/2015))`
    - C: `(RECORD.date CONTAINS TIME(HH:MM:SS > 11:20:00))`

The expression `A` will be called first as a normal search. The result of `A`
will be used as input for the `B` sub-query (date search). And result of `B`
will be used as input for the `C` sub-query (time search).

Currently `AND` and `OR` operators are supported.

For structured search it's important to keep temporary file extension coherent
to the input! For example, if input contains `*.pcrime` mask the temporary file should also
have `.pcrime` extension. Otherwise Ryft hardware won't use corresponding RDF scheme.
