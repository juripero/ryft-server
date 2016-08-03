This document contains detailed description of authentication methods used by the ryft server.

The following types of authentication are supported:

- [basic](https://en.wikipedia.org/wiki/Basic_access_authentication)
- [JWT](https://jwt.io/introduction/)

To verify provided user credentials the LDAP service or simple file may be used.

The following endpoints are protected:

- [/search](./restapi.md#search)
- [/count](./restapi.md#count)
- [/files](./restapi.md#files)


# Authentication

If authentication is enabled the ryft server checks for `Authorization` HTTP header.

If `Authorization` header contains `Basic` keyword the basic authentication is used.
The ryft server extracts username and password from the header and checks the user
is authorized to access requested resources.

Otherwise if `Authorization` header contains `Bearer` keyword the JWT is used.
The ryft server extracts JWT token from the header and checks it.

There are two special endpoints for JWT authentication:

- `/login` is used to get JWT token.
- `/token/refresh` is used to refresh existing token.

## JWT Login

The `/login` endpoint expects `{"username":"login", "password":"password"}` JSON
structure as an input. If credentials are valid the JWT token is provided as a result.

For example:

```{.sh}
curl -d '{"username":"admin","password":"admin"}' "localhost:8765/login"
```

return the following:

```{.sh}
{"expire":"2016-07-11T08:13:09-04:00",
"token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE0NjgyMzkxODksImlkIjoiYWRtaW4iLCJvcmlnX2lhdCI6MTQ2ODIzNTU4OX0.X_sO1pimiDQ9XGg37PzTYIB9ohu4DJM8VG9lgqd4sqg"}
```

## JWT options

To pass JWT secret to the server the configuration file
or `--jwt-secret` command line option is used:

```{.sh}
ryft-server --jwt-secret=my-secret-key
ryft-server --jwt-secret=@my-secret-file
ryft-server --jwt-secret=hex:6D792D7365637265742D6B6579
ryft-server --jwt-secret=base64:bXktc2VjcmV0LWtleQ==
```

By default `HS256` signing algorithm is used.
It may be overriden by `--jwt-alg` command line option.
For example `--jwt-alg=RS256`.

Default token lifetime is 1 hour.
To change it use `--jwt-lifetime` command line option.
Note, the overall token refresh timeout is set to 10 lifetimes!

See corresponding section in configuration file to check all available options.


## LDAP

Most of LDAP customization can be done via configuration file.
But there is also some command line options are available.
Check `ryft-server --help` output.

First of all the LDAP server address should be provided.
See `--ldap-server` command line option.

Second, we need a special read-only account to do search requests.
This account credentials can be customized via `--ldap-user` and `--ldap-pass`
command line options.

The `--ldap-query` and `--ldap-basedn` are used to finish the LDAP search request.
Query format is used to select appropriate RDN, for example `(&(cn=%s))`.

There are a few security related options (currently in configuration file only)
that could be used to disable TLS or to disable TLS certificate verification.
Please check corresponding `auth-ldap` section of the configuration file.


## Simple text file

Simple text file may be used as a list of user credentials.

YAML format:

```{.yaml}
- username: "admin"
  password: "admin"
  home: "/"
- username: "test"
  password: "test"
  home: "/test"
  cluster-tag: "test"
- username: "foo"
  password: "foo"
  home: "/foo"
  cluster-tag: "foo"
```

JSON format:

```{.json}
[
  {"username":"admin", "password":"admin", "home":"/"},
  {"username":"test", "password":"test", "home":"/test", "cluster-tag":"test"},
  {"username":"foo", "password":"foo", "home":"/foo", "cluster-tag":"foo"}
]
```

The `home` is a directory inside `/ryftone` mount point.
It is used to separate data of various users.
An user is authorized to have access to its home only.

The `cluster-tag` is used for partitioning (see corresponding `ryft-cluster/consul.md#partitioning-setup` document).
If an user has custom paritioning rules they are located under `{cluster-tag}/partitions` KV prefix.

To run server use the following command line:

```{.sh}
ryft-server --auth=file --users-file "ryft-users.yaml"
```


# Cluster Mode

All cluster nodes should have the same list of user (or the same LDAP configured)
and the same secret key. It's important to be able to pass authentication tokens
between cluster nodes. Ryft server uses `Authorization` HTTP header "as is" to
redirect search requests.

Moreover, each user should have the same home directory on each cluster node.
