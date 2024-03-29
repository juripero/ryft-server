# Demo - authentication and expansion on search type - July 21, 2016

## Authentication and multi-tenancy support

Current ryft server version supports multi user mode. Each user is authorized
to access dedicated subfolder of `/ryftone` mount point.

Let's have a three users: `admin`, `test` and `John`.
Corresponding credentials file can looks like:

```{.yaml}
- username: "admin"
  password: "admin"
  home: "/"
- username: "test"
  password: "test"
  home: "/test"
- username: "John"
  password: "pass"
  home: "/john"
```

Here we have `admin` who can access the whole `/ryftone` and two simple users
`test` and `John` who is able to access only their home directories `/ryftone/test`
and `/ryftone/john`.

To run ryft server with authentication enabled just use the following command:

```{.sh}
ryft-server --auth=file --users-file=users.yaml
```

Now we can access [/search](../rest/search.md#search), [/count](../rest/search.md#count)
and [/files](../rest/files.md) endpoints only if we provide valid user
credentials. Otherwise `401 Unauthorized` HTTP status will be reported:

```{.sh}
ryftrest -q '(RAW_TEXT CONTAINS "555")' -f '*.txt' -v
```

To provide user credentials special `-u` or `--user` flag  is supported:

```{.sh}
ryftrest -q '(RAW_TEXT CONTAINS "555")' -f '*.txt' -u 'admin:admin' -vv
```

Now the command should print a few records found.


### Home directories

Let's create a few text files for each users:

```{.sh}
mkdir -p /ryftone/test
echo " test 555 test " > /ryftone/test/test.txt
echo " test 777 test " >> /ryftone/test/test.txt

mkdir -p /ryftone/john
echo " john 555 john " > /ryftone/john/john.txt
echo " john 777 john " >> /ryftone/john/john.txt
```

Running the same command with various credentials we can see different results.

If we run as `test` user:

```{.sh}
ryftrest -q '555' -f '*.txt' -u 'test:test' -w 5 -vv --no-stats --format=utf8
```

The output will be:

```{.json}
{
  "results": [
    {
      "data": "test 555 test",
      "_index": {
        "host": "ryftone-vm-selaptop",
        "fuzziness": 0,
        "length": 13,
        "offset": 1,
        "file": "/test.txt"
      }
    }
  ]
}
```

If we run the same command as `John` user:

```{.sh}
ryftrest -q '555' -f '*.txt' -u 'John:pass' -w 5 -vv --no-stats --format=utf8
```

The output will be different:

```{.json}
{
  "results": [
    {
      "data": "john 555 john",
      "_index": {
        "host": "ryftone-vm-selaptop",
        "fuzziness": 0,
        "length": 13,
        "offset": 1,
        "file": "/john.txt"
      }
    }
  ]
}
```

Please note the file path in the `_index`. The filename is relative to user's home
directory. `John` knows nothing about `test`, and `test` knows nothing about `John`'s files.

We can also check [/files](../rest/files.md) endpoint to get directory content for each user:

```{.sh}
curl -s -u 'test:test' "http://localhost:8765/files" | jq .
```
```{.json}
{
  "folders": [],
  "files": [
    "test.txt"
  ],
  "dir": "/"
}
```

```{.sh}
curl -s -u 'John:pass' "http://localhost:8765/files" | jq .
```
```{.json}
{
  "folders": [],
  "files": [
    "john.txt"
  ],
  "dir": "/"
}
```


### JSON Web Tokens

[JSON Web Token](https://jwt.io/) is an alternative way of authentication among
with the basic authentication described above. Ryft server supports both authentication
methods: if HTTP request contains `Authorization: Basic ...` keyword then
basic authentication is used, if  HTTP request contains `Authorization: Bearer ...`
keyword then JWT is used.

In order to support JWT ryft server exports two endpoints. First `/login` is used
to provide user credentials and get valid token back. The second `/token/refresh` is
used to refresh existing token. By default token's lifetime is one hour but this
timeout can be customized via `--jwt-lifetime` command line option.

```{.sh}
curl -s -d '{"username":"test", "password":"test"}' "http://localhost:8765/login"
```
```{.json}
{
"expire":"2016-07-19T08:55:21-04:00",
"token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE0Njg5MzI5MjEsImlkIjoidGVzdCIsIm9yaWdfaWF0IjoxNDY4OTI5MzIxfQ.4hp5JSxGEWrRKNW0SWi5O2-LTTu5hQK178D0AzUVKuI"
}
```

Having this token it's possible to access ryft server without user credentials:

```{.sh}
export TOKEN=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE0Njg5MzI5MjEsImlkIjoidGVzdCIsIm9yaWdfaWF0IjoxNDY4OTI5MzIxfQ.4hp5JSxGEWrRKNW0SWi5O2-LTTu5hQK178D0AzUVKuI
curl -H "Authorization: Bearer $TOKEN" http://localhost:8765/files
```

Signing algorithm and secret can be customized via corresponding `--jwt-alg` and
`--jwt-secret` command line options. For example, to use RSA the
following command should be used:

```{.sh}
ryft-server --jwt-alg=RS256 --jwt-secret=@/home/ryftuser/.ssh/id_rsa
```


## Expansion of search types and Regex search

Previous ryft server allowed one text search type per request - hamming or edit distance.
New feature allows to pass various text search types for each sub-expression.
For example: `(RAW_TEXT CONTAINS FHS("555",CS=true,DIST=1,WIDTH=2)) AND (RAW_TEXT CONTAINS FEDS("777",CS=true,DIST=1,WIDTH=4))`.
The ryft server splits this expressions into two Ryft calls:
- `(RAW_TEXT CONTAINS "555")` with `FHS` search type, `fuzziness=1` and `surrounding=2`
- `(RAW_TEXT CONTAINS "777")` with `FEDS` search type, `fuzziness=1` and `surrounding=4`.

New syntax overrides the following global parameters:
- search type: `FHS` or `FEDS` (exact search is used if fuzziness is zero)
- case sensitivity 
- fuzziness distance
- surrounding width

If nothing provided the global options are used by default. We can omit case sensitivity flag:
`(RAW_TEXT CONTAINS FHS("555")) AND (RAW_TEXT CONTAINS FEDS("777",CS=false))`

Here are some examples:

```{.sh}
# -d 0 and -w 10 are overriden by search expression
ryftrest -q '(RAW_TEXT CONTAINS FHS("John",DIST=2,WIDTH=3))' -f '*.txt' -d 0 -w 10 --local --format=utf8 -vv -u 'test:test'

# split into two Ryft calls: FHS and FEDS
ryftrest -q '(RAW_TEXT CONTAINS FHS("John",DIST=2,WIDTH=3)) AND (RAW_TEXT CONTAINS FEDS("sony",DIST=1,WIDTH=4))' -f '*.txt' -i -d 0 -w 10 --local --format=utf8 -vv -u 'test:test'
```

There is also new Rexeg search type support:

```{.sh}
ryftrest -q '(RAW_TEXT CONTAINS REGEX("[JS]on[nyes]",PCRE_OPTION_DEFAULT))' -f '*.txt' -u 'test:test' -vv --no-stats --local --format=utf8
```
