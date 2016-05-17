This document contains information about REST API.

# Search

endpoint /search parameters:

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

## Not structured request example

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


## Structured request example

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

# Count

/count endpoint parameters:

| Method | Input type | Uri | Description |
| --- | --- | --- | --- |
| *query* | string | GET /count?query={QUERY} | String that specifying the search criteria. Required file parameter |
| *files* | string | GET /count?query={QUERY}&files={FILE} | Input data set to be searched. Comma separated list of files or directories. |
| *fuzziness* | uint8 | GET /count?query={QUERY}&files={FILE}&fuzziness={VALUE} | Specify the fuzzy search distance `[0..255]` . |
| *cs* | boolean | GET /count?query={QUERY}&files={FILE}&cs=true | Case sensitive flag. Default `false`. |
| *nodes* | int | GET /count?query={QUERY}&files={FILE}&nodes={VALUE} | Parameter that specifies nodes count `[0..4]`. Default `4`, if nodes=0 system will use default value. |
| *local* | boolean | GET /search?query={QUERY}&files={FILE}&local={VALUE} | Parameter that specifies search mode, set `true` to enable local search, set `false` for cluster mode search. Default `false`. |

## Count request example

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


# Version

endpoint parameters:

Endpoint that allows to check the current build version

## Version request example

[/version](/version)

```
{
  "git-hash": "35c358378f7c214069333004d01841f9066b8f15",
  "version": "0.5.9-76-g35c3583"
}
```


# Files

endpoint parameters:

| Method | Input type | Uri | Description |
| --- | --- | --- | --- |
| *local* | boolean | GET /files?local={VALUE} | Parameter that specifies search mode, set `true` to enable local search, set `false` for cluster mode search. Default `false` |
| *dir* | string | GET /files?&dir={VALUE} | Parameter that specifies files directory. Default `/ryftone` |


## Files request example

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
