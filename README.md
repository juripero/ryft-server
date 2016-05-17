# Cloning & Building

> The instructions below assume you have a properly configured GO dev environment with GOPATH and GOROOT env variables configured.
> If you starty from scratch we recommend to use this [automated installer](https://github.com/demon-xxi/tools).

> To use `go get` command with private repositories use the following setting to force SSH protocol instead of HTTPS:
> `git config --global url."git@github.com:".insteadOf "https://github.com/"`
> Make sure you have configured [SSH token authentication](https://help.github.com/articles/generating-an-ssh-key/) for GitHub.

```bash
go get github.com/getryft/ryft-server
cd $GOPATH/src/github.com/getryft/ryft-server
make
```

To change git branch use combination of commands:
```bash
cd $GOPATH/src/github.com/getryft/ryft-server
git checkout <branch-name>
go get
```

For packaging into deb file just run `make debian`, see detailed instructions [here](./debian/README.md).

# Running & Command Line Parameters

```
usage: ryft-server [<flags>] [<address>]

Flags:
  --help           Show context-sensitive help (also try --help-long and --help-man).
  -k, --keep       Keep search results temporary files.
  -d, --debug      Run http server in debug mode.
  -a, --auth=AUTH  Authentication type: none, file, ldap.
  --users-file=USERS-FILE
                   File with user credentials. Required for --auth=file.
  --ldap-server=LDAP-SERVER
                   LDAP Server address:port. Required for --auth=ldap.
  --ldap-user=LDAP-USER
                   LDAP username for binding. Required for --auth=ldap.
  --ldap-pass=LDAP-PASS
                   LDAP password for binding. Required for --auth=ldap.
  --ldap-query="(&(uid=%s))"
                   LDAP user lookup query. Defauls is '(&(uid=%s))'. Required for --auth=ldap.
  --ldap-basedn=LDAP-BASEDN
                   LDAP BaseDN for lookups.'. Required for --auth=ldap.

  -t, --tls          
                    Enable TLS/SSL. Default 'false'.
  --tls-crt=TLS-CRT  
                    Certificate file. Required for --tls=true.
  --tls-key=TLS-KEY  
                    Key-file. Required for --tls=true.
  --tls-address=0.0.0.0:8766  
                     Address:port to listen on HTTPS. Default is 0.0.0.0:8766

Args:
  [<address>]  Address:port to listen on. Default is 0.0.0.0:8765.

```
Default value ``port`` is ``8765``

# Keeping search results

By default REST-server removes search results from ``/ryftone/RyftServer-PORT/``. But it behaviour may be prevented:

```
ryft-server --keep
```

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


# API endpoints

See [here](./docs/restapi.md)
