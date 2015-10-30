
# Installing Dependencies & Building

1. See installation instructions for golang environment — https://golang.org/doc/install
2. ``go get github.com/getryft/ryft-server``
3. ``go install``
4. Get ``ryft-server`` binary from ``$GOPATH/bin``

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
# Packaging into deb file

https://github.com/getryft/ryft-server/blob/master/ryft-server-make-deb/README.md

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
| *fuzziness* | uint8 | GET /search?query={QUERY}&files={FILE}&fuzziness={VALUE} | Specify the fuzzy search distance [0..255] . |
| *cs* | string | GET /search?query={QUERY}&files={FILE}&cs=true | Case sensitive flag. Default 'false'. |
| *format* | string | GET /search?query={QUERY}&files={FILE}&apm;format={FORMAT} | Parameter for the structed search. Specify the input data format 'xml' or 'raw'(Default). |
| *surroinding* | uint16 | GET /search?query={QUERY}&files={FILE}&surrounding={VALUE} | Parameter that specifies the number of characters before the match and after the match that will be returned when the input specifier type is raw text |
| *fields* | string | GET /search?query={QUERY}&files={FILE}&format=xml&fields={FIELDS...} | Parametr that specifies needed keys in result. Required format=xml. |
| *nodes* | string | GET /search?query={QUERY}&files={FILE}&nodes={VALUE} | Parameter that specifies nodes count [0..4]. Default 4, if nodes=0 system will use default value. |

### Not structed request example

[/search?query=(RAW_TEXT CONTAINS "10")&files=passengers.txt&surrounding=10&fuzziness=0](/search?query=(RAW_TEXT%20CONTAINS%20%2210%22)&files=passengers.txt&surrounding=10&fuzziness=0)

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

[/search?query=(RECORD.id EQUALS "10034183")&files=*.pcrime&surrounding=10&fuzziness=0&format=xml](/search?query=(RECORD.id%20EQUALS%20%2210034183%22)&files=*.pcrime&surrounding=10&fuzziness=0&format=xml)

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
| *fuzziness* | uint8 | GET /count?query={QUERY}&files={FILE}&fuzziness={VALUE} | Specify the fuzzy search distance [0..255] . |
| *cs* | string | GET /count?query={QUERY}&files={FILE}&cs=true | Case sensitive flag. Default 'false'. |
| *nodes* | string | GET /count?query={QUERY}&files={FILE}&nodes={VALUE} | Parameter that specifies nodes count [0..4]. Default 4, if nodes=0 system will use default value. |

### Count request example

[/count?query=(RECORD CONTAINS "a")OR(RECORD CONTAINS "b")&files=*.pcrime](/count?query=(RECORD%20CONTAINS%20%22a%22)OR(RECORD%20CONTAINS%20%22b%22)&files=*.pcrime)

```
"Matching: 10000"
```
