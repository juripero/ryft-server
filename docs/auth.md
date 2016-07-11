This document contains detailed authentication description.

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
The ryft server extracts username and password from the header and checks them.

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

## JWT secret

To pass JWT secret to the server the `--jwt-secret` command line argument is used:

```{.sh}
ryft-server --jwt-secret=my-secret-key
ryft-server --jwt-secret=@my-secret-file
ryft-server --jwt-secret=hex:6D792D7365637265742D6B6579
ryft-server --jwt-secret=base64:bXktc2VjcmV0LWtleQ==
```

## LDAP

TBD

## Simple text file

simple text file may be used as a list of user credentials.

YAML format:

```{.yaml}
- username: "admin"
  password: "admin"
  home: "/"
- username: "test"
  password: "test"
  home: "/test"
- username: "foo"
  password: "foo"
  home: "/foo"
```

JSON format:

```{.json}
[
  {"username":"admin", "password":"admin", "home":"/"},
  {"username":"test", "password":"test", "home":"/test"},
  {"username":"foo", "password":"foo", "home":"/foo"}
]
```

To run server use the following command line:

```{.sh}
ryft-server --auth=file --users-file ryft-auth.yaml
```
