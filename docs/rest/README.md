The `ryft-server` supports the following REST API endpoints:

- [/version](#version)
- [/search](./search.md#search)
- [/count](./search.md#count)
- [/files](./files.md)
- [/rename](./rename.md)
- [/run](./run.md)

The main API endpoints are [/search](./search.md#search)
and [/count](./search.md#count).

If authentication is enabled, there are also a few endpoints related to
[JWT](./auth.md).

Some endpoints are enabled in `debug` mode only,
like [/logging/level](#logging-level)
or [/search/dry-run](./search.md#search).


# Compatibility

Since `0.10.0` version the [/search](./search.md#search) endpoint uses
`cs=true` by default. That means if no `cs` query parameter is provided
the case sensitive search will be used by default.

Since `0.10.0` the `reduce=true` is used by default (FEDS only) by `ryftrest` and `ryft-server`.
There is special REST API [option](./search.md#search-reduce-parameter) for this.
And `--reduce` and `--no-reduce` flags for the `ryftrest`.

Since `0.10.0` the [NUMBER](../search/NUMBER.md) and [CURRENCY](../search/CURRENCY.md)
search uses the following options by default:
- "$" for the `SYMBOL` option
- "," for the `SEPARATOR` option
- "." for the `DECIMAL` option

Since `0.13.0` the [limit] parameter has default value `-1` which means "no limit".
The `limit=0` means do not report any found records.


# Aggregation functions

The limited set of [aggregation functions](./aggs.md) are supported:
- [Min](./aggs.md#min-aggregation)
- [Max](./aggs.md#max-aggregation)
- [Sum](./aggs.md#sum-aggregation)
- [Value Count](./aggs.md#value-count-aggregation)
- [Avg](./aggs.md#avg-aggregation)
- [Stats](./aggs.md#stats-aggregation)
- [Extended Stats](./aggs.md#extended-stats-aggregation)
- [Geo Bounds](./aggs.md#geo-bounds-aggregation)
- [Geo Centroid](./aggs.md#geo-centroid-aggregation)


# Post-process transformations

The output data can be transformed on the server side just before
it is reported to the client. A regular expression or custom
application/script can be used to perform transformations.

The [transform](./search.md#search-transform-parameter) query option defines
a custom transformation. There may be a few transformations joined to the
transformation chain - where the output of the first transformation goes
to the input of the second transformation.

A single transformation can be one of:
- [match](#match-transformation) for regular expression match
- [replace](#replace-transformation) for regular expression replace
- [script](#script-transformation) for custom application/script

Note, the output statistic contains initial number of matches.
So we can check the number of dropped records as difference between
`Matches` and the actual number of records received.

The same is true for indexes. Indexes contain initial data position and length
without any transformations reflected.

This [page](https://regex-golang.appspot.com/assets/html/index.html)
can be used to design and check regular expressions.


## Match transformation

The regular expression `match` transformation is used as a filter.
If a record does not match the regular expression then the record is just dropped.

This transformation is defined as `transform=match("expression")` where `expression`
is a valid regular expression applied to the found record. For example, the
following transformation will report only records containing `markX` where `X`
is a digit: `transform=match("^.*mark[0-9].*$")`.


## Replace transformation

The regular expression `replace` transformation is used as a simple "match and replace".
If a record does match the regular expression then it is replaced with a template.
If a record does not match then it is leaved "as is".

This transformation is defines as `transform=replace("expression", "template")`
where `expression` is a valid regular expression and the `template` is a
replacement text. The template can use special variables `$1`, `$2`, etc. to
refer to the matched text.

For example, the following transformation will replace all `apple`s with `orange`s:
`transform=replace("^(.*)apple(.*)$", "${1}orange${2}")`.


## Script transformation

The `script` transformation uses external application or script to transform
a record. Note, this transformation is much slower comparing to regexp above.
A found record is written to the `STDIN` of the `script` and transformed record
is read from the `STDOUT`. If exit status of the `script` is non zero then
the record is dropped.

This transformation is defined as `transform=script("name")` where `name` is
a predefined script name. Valid script names should be configured via
[server's configuration file](../run.md#script-transformation-configuration).

For example, the following transformation will use `cat` to print input:
`transform=script("cat")`.

More useful scripts can be developed using `jq` or similar utilities.


# Version

The GET `/version` endpoint is used to check current `ryft-server` version.

It has no parameters, output looks like:

```{.json}
{
  "git-hash": "35c358378f7c214069333004d01841f9066b8f15",
  "version": "1.2.3"
}
```


# Logging level

The GET `/logging/level` endpoint is used to get current logging levels.
Using the POST `/logging/level` endpoint it is possible to change any
logging level on the running server.

These endpoints are available only in `debug` mode.

For example, to print current logging levels use:

```{.sh}
curl -s "http://localhost:8765/logging/level" | jq .
```

To change `core` and `search/ryftprim` logging levels:

```{.sh}
curl -X POST -s "http://localhost:8765/logging/level?core=info&search/ryftprim=error" | jq .
```
