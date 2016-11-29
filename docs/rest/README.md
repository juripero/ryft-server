The `ryft-server` supports the following REST API endpoints:

- [/version](#version)
- [/search](./search.md#search)
- [/count](./search.md#count)
- [/files](./files.md)

The main API endpoints are [/search](./search.md#search)
and [/count](./search.md#count).

If authentication is enabled, there are also a few endpoints related to
[JWT](./auth.md).

Some endpoints are enabled in debug mode.


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

