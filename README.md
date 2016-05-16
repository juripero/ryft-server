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
Ryft server will split this expresssion into two separate queries:
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

Check the `swagger.json` for detailed information.

## Search endpoint /search parameters :

| Method | Input type | Uri | Description |
| --- | --- | --- | --- |
| *query* | string | GET /search?query={QUERY} | The search expression. Required. |
| *files* | string | GET /search?query={QUERY}&files={FILE} | Input data set to be searched. Comma separated list of files or directories. Could contain wildcards. |
| *fuzziness* | uint8 | GET /search?query={QUERY}&files={FILE}&fuzziness={VALUE} | The fuzzy search distance `[0..255]`. |
| *cs* | boolean | GET /search?query={QUERY}&files={FILE}&cs=true | Case sensitive flag. Default `false`. |
| *format* | string | GET /search?query={QUERY}&files={FILE}&apm;format={FORMAT} | Parameter for the structured search. Specify the input data format `xml`, `json` or `raw` (Default). |
| *surrounding* | uint16 | GET /search?query={QUERY}&files={FILE}&surrounding={VALUE} | Parameter that specifies the number of characters before the match and after the match that will be returned when the input specifier type is raw text |
| *fields* | string | GET /search?query={QUERY}&files={FILE}&format=xml&fields={FIELDS...} | For structured search specify the list of required fields. If omitted all fields are used. |
| *data* | string | GET /search?query={QUERY}&files={FILE}&format=xml&data={dataFile} | Name of results data file to keep. WARNING: file will be overriden! |
| *index* | string | GET /search?query={QUERY}&files={FILE}&format=xml&index={indexFile} | Name of results index file to keep. WARNING: file will be overriden! |
| *nodes* | int | GET /search?query={QUERY}&files={FILE}&nodes={VALUE} | Parameter that specifies nodes count `[0..4]`. Default `4`, if nodes=0 system will use default value. |
| *local* | boolean | GET /search?query={QUERY}&files={FILE}&local={VALUE} | Parameter that specifies cluster mode, set `true` to enable local search, set `false` for cluster mode search. Default `false`. |
| *stats* | boolean | GET /search?query={QUERY}&files={FILE}&stats={VALUE} | Parameter that enables including statistics . Default `false`. |
| *stream* | boolean | GET /search?query={QUERY}&files={FILE}&stream={VALUE} | Parameter that specifies response format. Internally used in cluster mode. Default `false`. |
| *spark* | boolean | GET /search?query={QUERY}&files={FILE}&local={VALUE} | Parameter that specifies response format. Recommended to use with Spark. Default `false`. |
| *ep* | boolean | GET /search?query={QUERY}&files={FILE}&local={VALUE} | Error Prefix. Parameter that specifies error prefix to find out from which node error comes. Recommended to use in cluster mode. Default `false`. |

### Not structured request example

[/search?query=10&files=passengers.txt&surrounding=10&fuzziness=0&local=false](/search?query=10&files=passengers.txt&surrounding=10&fuzziness=0&local=false)

```
[
  {
    "_index": {
      "file": "/ryftone/passengers.txt",
      "offset": 27,
      "length": 22,
      "fuzziness": 0
    },
    "data": "YWwgU21pdGgsIDEwLTAxLTE5MjgsMA=="
  },
  {
    "_index": {
      "file": "/ryftone/passengers.txt",
      "offset": 43,
      "length": 22,
      "fuzziness": 0
    },
    "data": "MTkyOCwwMTEtMzEwLTU1NS0xMjEyLA=="
  }
]
```

`data` is *base64* encoded bytes of search results.


### Structured request example

[/search?query=(RECORD.id EQUALS "10034183")&files=*.pcrime&surrounding=10&fuzziness=0&format=xml&local=true](/search?query=(RECORD.id%20EQUALS%20%2210034183%22)&files=*.pcrime&surrounding=10&fuzziness=0&format=xml&local=true)

```
[
  {
    "Arrest": "false",
    "Beat": "0313",
    "Block": "062XX S ST LAWRENCE AVE",
    "CaseNumber": "HY223673",
    "CommunityArea": "42",
    "Date": "04/15/2015 11:59:00 PM",
    "Description": "DOMESTIC BATTERY SIMPLE",
    "District": "003",
    "Domestic": "true",
    "FBICode": "08B",
    "ID": "10034183",
    "IUCR": "0486",
    "Latitude": "41.781961688",
    "Location": "\"(41.781961688, -87.610984705)\"",
    "LocationDescription": "STREET",
    "Longitude": "-87.610984705",
    "PrimaryType": "BATTERY",
    "UpdatedOn": "04/22/2015 12:47:10 PM",
    "Ward": "20",
    "XCoordinate": "1181263",
    "YCoordinate": "1863965",
    "Year": "2015",
    "_index": {
      "file": "/ryftone/chicago.pcrime",
      "offset": 0,
      "length": 693,
      "fuzziness": 0
    }
  }
]
```

## Count endpoint

| Method | Input type | Uri | Description |
| --- | --- | --- | --- |
| *query* | string | GET /count?query={QUERY} | String that specifying the search criteria. Required file parameter |
| *files* | string | GET /count?query={QUERY}&files={FILE} | Input data set to be searched. Comma separated list of files or directories. |
| *fuzziness* | uint8 | GET /count?query={QUERY}&files={FILE}&fuzziness={VALUE} | Specify the fuzzy search distance `[0..255]` . |
| *cs* | boolean | GET /count?query={QUERY}&files={FILE}&cs=true | Case sensitive flag. Default `false`. |
| *nodes* | int | GET /count?query={QUERY}&files={FILE}&nodes={VALUE} | Parameter that specifies nodes count `[0..4]`. Default `4`, if nodes=0 system will use default value. |
| *local* | boolean | GET /search?query={QUERY}&files={FILE}&local={VALUE} | Parameter that specifies search mode, set `true` to enable local search, set `false` for cluster mode search. Default `false`. |

### Count request example

[/count?query=(RECORD CONTAINS "a")OR(RECORD CONTAINS "b")&files=*.pcrime&local=true](/count?query=(RECORD%20CONTAINS%20%22a%22)OR(RECORD%20CONTAINS%20%22b%22)&files=*.pcrime&local=true)

```
{
	"matches": 10000,
	"totalBytes": 6892667,
	"duration": 2071,
	"dataRate": 3.174002,
	"fabricDataRate": 3.174002
}
```


## Version endpoint

Endpoint that allows to check the current build version

### Version request example

[/version](/version)

```
{
  "git-hash": "35c358378f7c214069333004d01841f9066b8f15",
  "version": "0.5.9-76-g35c3583"
}
```


## Files endpoint
| Method | Input type | Uri | Description |
| --- | --- | --- | --- |
| *local* | boolean | GET /files?local={VALUE} | Parameter that specifies search mode, set `true` to enable local search, set `false` for cluster mode search. Default `false` |
| *dir* | string | GET /files?&dir={VALUE} | Parameter that specifies files directory. Default `/ryftone` |


### Files request example

[/files](/files)

```
{
  "dir": "/",
  "files": [
    "chicago.pcrime",
    "passengers.txt"
  ],
  "folders": [
    "RyftServer-8765",
    "RyftServer-9000",
    "demo",
    "regression",
    "test",
    "tmp"
  ]
}
```

