The `ryft-server` supports the following REST API endpoints:

- [/version](#version)
- [/search](./search.md#search)
- [/count](./search.md#count)
- [/files](./files.md)

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
