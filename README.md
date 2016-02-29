# Clonning & Building

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

For packaging into deb file see instructions [here](./debian/README.md).

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

# API endpoints

## Search endpoint /search parameters :

| Method | Input type | Uri | Description |
| --- | --- | --- | --- |
| *query* | string | GET /search?query={QUERY} | String that specifying the search criteria. Required file parameter |
| *files* | string | GET /search?query={QUERY}&files={FILE} | Input data set to be searched. Comma separated list of files or directories. |
| *fuzziness* | uint8 | GET /search?query={QUERY}&files={FILE}&fuzziness={VALUE} | Specify the fuzzy search distance `[0..255]`. |
| *cs* | string | GET /search?query={QUERY}&files={FILE}&cs=true | Case sensitive flag. Default `false`. |
| *format* | string | GET /search?query={QUERY}&files={FILE}&apm;format={FORMAT} | Parameter for the structed search. Specify the input data format `xml` or `raw`(Default). |
| *surroinding* | uint16 | GET /search?query={QUERY}&files={FILE}&surrounding={VALUE} | Parameter that specifies the number of characters before the match and after the match that will be returned when the input specifier type is raw text |
| *fields* | string | GET /search?query={QUERY}&files={FILE}&format=xml&fields={FIELDS...} | Parametr that specifies needed keys in result. Required format=xml. |
| *nodes* | string | GET /search?query={QUERY}&files={FILE}&nodes={VALUE} | Parameter that specifies nodes count `[0..4]`. Default `4`, if nodes=0 system will use default value. |
| *local* | boolean | GET /search?query={QUERY}&files={FILE}&local={VALUE} | Parameter that specifies search mode, set `true` to enable local search, set `false` for cluster mode search. Default `false`. |
| *stats* | boolean | GET /search?query={QUERY}&files={FILE}&stats={VALUE} | Parameter that enables including statistics . Default `false`. |
| *stream* | boolean | GET /search?query={QUERY}&files={FILE}&stream={VALUE} | Parameter that specifies response format. Recomended to use with cluster mode. Default `false`. |
| *spark* | boolean | GET /search?query={QUERY}&files={FILE}&local={VALUE} | Parameter that specifies response format. Recomended to use with Spark. Default `false`. |
| *ep* | boolean | GET /search?query={QUERY}&files={FILE}&local={VALUE} | Error Prefix. Parameter that specifies error prefix to find out from which node error comes. Recomended to use with cluster mode. Default `false`. |

### Not structed request example

[/search?query=(RAW_TEXT CONTAINS "10")&files=passengers.txt&surrounding=10&fuzziness=0&local=false](/search?query=(RAW_TEXT%20CONTAINS%20%2210%22)&files=passengers.txt&surrounding=10&fuzziness=0&local=false)

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


### Structed request example

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
| *cs* | string | GET /count?query={QUERY}&files={FILE}&cs=true | Case sensitive flag. Default `false`. |
| *nodes* | string | GET /count?query={QUERY}&files={FILE}&nodes={VALUE} | Parameter that specifies nodes count `[0..4]`. Default `4`, if nodes=0 system will use default value. |
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

[/version]
(/version)

```
{
  "git-hash": "35c358378f7c214069333004d01841f9066b8f15",
  "version": "0.5.9-76-g35c3583"
}

```
