There are a few REST API endpoints related to search:

- [/search](#search)
- [/count](#count)
- [/search/show](#show)

First one reports the data found. The seconds one reports
just search statistics, no any data.

The `search/show` endpoint is used to access already existing results.


# Search

The GET `/search` endpoint is used to search data on Ryft boxes.

Note, this endpoint is protected and user should provide valid credentials.
See [authentication](../auth.md) for more details.

There are a few [content types](#search-accept-header) that server can produce:
- `Accept: application/json` which is used by default
- `Accept: text/csv`
- `Accept: application/msgpack` which is used internally in cluster mode


## Search query parameters

The list of supported query parameters are the following (check detailed description below):

| Parameter     | Type    | Description |
| ------------- | ------- | ----------- |
| `query`       | string  | **Required**. [The search expression](#search-query-parameter). |
| `file`        | string  | **Required**. [The set of files or catalogs to search](#search-file-parameter). |
| `mode`        | string  | [The search mode](#search-mode-parameter). |
| `surrounding` | string  | [The data surrounding width](#search-surrounding-parameter). |
| `fuzziness`   | uint8   | [The fuzziness distance](#search-fuzziness-parameter). |
| `format`      | string  | [The structured search format](#search-format-parameter). |
| `cs`          | boolean | [The case sensitive flag](#search-cs-parameter). |
| `reduce`      | boolean | [The reduce flag for FEDS](#search-reduce-parameter). |
| `fields`      | string  | [The set of fields to get](#search-fields-parameter). |
| `transform`   | string  | [The post-process transformation](#search-transform-parameter). |
| `backend`     | string  | [The backend tool](#search-backend-parameter). |
| `backend-option`| string | [The backend tool options](#search-backend-option-parameter). |
| `data`        | string  | [The name of DATA file to keep](#search-data-and-index-parameters). |
| `index`       | string  | [The name of INDEX file to keep](#search-data-and-index-parameters). |
| `view`        | string  | [The name of VIEW file to keep](#search-data-and-index-parameters). |
| `delimiter`   | string  | [The delimiter is used to separate found records](#search-delimiter-parameter). |
| `lifetime`    | string  | [The output files lifetime](#search-lifetime-parameter). |
| `share-mode`  | string  | [The share mode used to access data files](#search-share-mode-parameter). |
| `nodes`       | int     | [The number of processing nodes](#search-nodes-parameter). |
| `local`       | boolean | [The local/cluster search flag](#search-local-parameter). |
| `stats`       | boolean | [The statistics flag](#search-stats-parameter). |
| `performance` | boolean | [Flag to report performance metrics](#search-performance-parameter). |
| `limit`       | int     | [Limit the total number of records reported](#search-limit-parameter). |
| `stream`      | boolean | **Internal** [The stream output format flag](#search-stream-parameters). |

### Search `query` parameter

The first required parameter is the search expression `query`.
It contains one or more subqueries connected using logical operators.

To execute text search for "The Batman" use the following search expression:

```
query=(RAW_TEXT CONTAINS "The Batman")
```

To execute structured search apply another keyword:

```
query=(RECORD.AlterEgo CONTAINS "The Batman")
```

Depending on [search mode](#search-mode-parameter) exact search query format may differ.
Check corresponding Ryft Open API or [short reference](../search/README.md)
for more details on search expressions.

#### Simple/Plain Queries

`ryft-server` supports simple plain queries - without any keywords.
The `query=Batman` will be automatically converted to `query=(RAW_TEXT CONTAINS "Batman")`.
NOTE: This only works for text search; it is not appropriate for structured search.

#### Complex Queries

`ryft-server` also supports complex queries containing several search expressions of different types.
For example `(RECORD.id CONTAINS "100") AND (RAW_TEXT CONTAINS DATE(MM/DD/YYYY > 04/15/2015))`.
This complex query contains two search expression: first one uses text search and the second one uses date search.
`ryft-server` will split this expression into two separate queries:
`(RECORD.id CONTAINS "100")` and `(RAW_TEXT CONTAINS DATE(MM/DD/YYYY > 04/15/2015))`.
It then calls Ryft hardware two times: the results of the first call are used as the input for the second call.

Multiple `AND` and `OR` operators are supported by the `ryft-server` within complex search queries.
Expression tree is built and each node is passed to the Ryft hardware. Then results are properly combined.

NOTE: If search query contains two or more RECORD-based expressions (of any types text, date, time, numeric...)
that query will not be split into subqueries because the Ryft hardware supports those type of queries directly.

There is also possible to use advanced text search queries to customize some parameters within search expression.
For example: `(RAW_TEXT CONTAINS FHS("555",CS=true,DIST=1,WIDTH=2)) AND (RAW_TEXT CONTAINS FEDS("777",CS=true,DIST=1,WIDTH=4))`.
The ryft server splits this expressions into two Ryft calls:
- `(RAW_TEXT CONTAINS "555")` with `fhs` search mode, `fuzziness=1` and `surrounding=2`
- `(RAW_TEXT CONTAINS "777")` with `feds` search mode, `fuzziness=1` and `surrounding=4`

This advanced search query syntax overrides the following global parameters:
- search type: `ES`, `FHS` or `FEDS` (exact search is used if fuzziness is zero)
- case sensitivity `CS=`
- fuzziness distance `DIST=`
- surrounding width `WIDTH=`

If nothing provided the global options are used by default. Any option can be omitted:
`(RAW_TEXT CONTAINS FHS("555")) AND (RAW_TEXT CONTAINS FEDS("777",CS=false))`.
See [short reference](../search/README.md) for more details.


### Search `file` parameter

The second required parameter is the set of file to search.
At least one file should be provided.

Multiple files can be provided as:

  - a list `file=1.txt&file=2.txt`
  - a wildcard: `file=*txt`

It's possible to provide catalog name as a `file` parameter. Ryft server
automatically detects catalogs and does appropriate substitutions.
Also the `catalog=` alias is supported.

Note, for backward compatibility the `files=` parameter is also supported.


### Search `mode` parameter

`ryft-server` supports several search modes:

- `es` for exact search
- `fhs` for fuzzy hamming search
- `feds` for fuzzy edit distance search
- `ds` for date search
- `ts` for time search
- `ns` for numeric search
- `cs` for currency search
- `ipv4` for IPv4 search
- `ipv6` for IPv6 search

If no search mode is specified, exact search is used **by default** for simple queries.
It is also possible to automatically detect search modes: if search query contains `DATE`
keyword then date search will be used. It's the same when `TIME` keyword is used for time search,
and so on.

In case of complex search queries, the mode specified is used for text or structured search only.
Date, time and numeric search modes will be detected automatically by corresponding keywords.

Check corresponding Ryft Open API or [short reference](./search/README.md)
for more details on search expressions.


### Search `surrounding` parameter

The number of characters in bytes up to a maximum of `64000` before the match
and after the match that will be returned when the text search is used.
For anything other than raw text, this parameter is ignored.

The `surrounding=line` should be provided for whole line surrounding.
This option is very useful for CVS files.

`surrounding=0` is used **by default**.


### Search `fuzziness` parameter

The fuzziness distance of the search up to a maximum of `255` when using a fuzzy search function.
For fuzzy hamming search, fuzziness is measured as the maximum Hamming distance allowed
in order to declare a match. For fuzzy edit distance search, fuzziness is measured
as the number of insertions, deletions or replacements required to declare a match.

`fuzziness=0` is used **by default**.


### Search `format` parameter

The input data format for the structured search.

**By default** structured search uses `format=raw` format.
That means that found data are reported as base-64 encoded raw bytes.

There are two other options: `format=xml` and `format=json`.

If input file set contains XML data, the found records could be decoded.
Just pass `format=xml` query parameter and records will be translated
from XML to JSON.

The same is true for JSON data.

See also [fields parameter](#search-fields-parameter).

For the text search there is `format=utf8` option. It interprets raw bytes as
UTF-8 string so it's easy to take a quick look at the results:

```{.json}
{
  "data": ",310-555-3425",
  "_index": {}
}
```

instead of `format=raw` - base-64 encoded raw bytes:

```{.json}
{
  "data": "LDMxMC01NTUtMzQyNQ==",
  "_index": {}
}
```

If data are not so important the `format=null` can be used.
This format tells `ryft-server` to ignore all data and to keep indexes only.


### Search `cs` parameter

The search text case-sensitive flag.

For example, if the search is case-sensitive `cs=true`, then searching for the string "John"
will not find any occurrences of "JOHN". If the same search is done with `cs=false`, then
case is ignored entirely and all possible capitalizations of the text will be found
(e.g. "jOhn" or "JOHn").

`cs=false` is used **by default**.


### Search `reduce` parameter

The search `reduce` boolean flag is used for fuzzy edit distance searches only.
The `reduce=true` tells engine to remove duplicates.
See [reduce option](../search/EDIT_DIST.md#reduce-option) for more details.

`reduce=true` is used **by default**.


### Search `fields` parameter

The comma-separated list of fields for structured search. If omitted, all fields are used.

This parameter is used to minimize structured search output or to get just subset of fields.
For example, to get identifier and date from a `*.pcrime` file pass `format=xml&fields=ID,Date`.

The same is true for JSON data: `format=json&fields=Name,AlterEgo`.


### Search `transform` parameter

This parameter specifies a post-process transformation.
Can be one of:
- `match("<expression>")`
- `replace("<expression>", "<template>")`
- `script("<script name>")`

A few transformations can be specified with several `transform` parameters.
In this case all tranformations are combined into transformation chain.

See [more details](./README.md#post-process-transformations).


### Search `backend` parameter

On those ryft-server instances where both `ryftprim` and `ryftx` backends are
present it is possible to manually select backend tool.

If `backend=ryftprim` options is used, then `ryftprim` tool will be used.
And `ryftx` tool is used in case of `backend=ryftx`.

If `backend` is empty (by default) then the most appropriate backend
is selected automatically.


### Search `backend-option` parameter

It is possible to send multiple optional flags with search backend.
All flags specified with `backend-option` are added to the end of
command line when backend tool is executed.

For example, `ryftx` can be customized with `--rx-max-spawns` and `--rx-max-spawns` flags:

```
/search?...&backend=ryftx&backend-option=--rx-max-spawns&backend-option=14&backend-option=--rx-max-spawns&backend-option=64M

# backend will be executed as
ryftx ... --rx-max-spawns 14 --rx-max-spawns 64M
```

If `ryftx` supports the `--rx-max-spawns=` syntax:

```
/search?...&backend=ryftx&backend-option=--rx-max-spawns%3D14&backend-option=--rx-max-spawns%3D64M

# backend will be executed as
ryftx ... --rx-max-spawns=14 --rx-max-spawns=64M
```

NOTE: `backend` parameter is required in order to prevent automatic selection of search backend.


### Search `data` and `index` parameters

By default, all search results are deleted from the Ryft server once they are delivered to user.
In order to preserve results thereby allowing for the ability to subsequently
"search in the previous results", two query parameters are available: `data=` and `index=`.

Using the first parameter, `data=output.dat` creates the search results on the Ryft server under `/ryftone/output.dat`.
It is possible to use that file as an input for the subsequent search call `file=output.dat`.

NOTE: It is important to use consistent file extension for the structured search
in order to let Ryft use appropriate RDF scheme!

Using the second parameter `index=index.txt` keeps the search index file under `/ryftone/index.txt`.

NOTE: According to Ryft API documentation, an index file should always have `.txt` extension!

**WARNING:** Provided data or index files will be overriden!

In some cases the output Ryft results can be additionally indexed to the
so called VIEW file - a binary index cache. This VIEW file allows quick
search result access at random position. The `view=view.bin` paramter
will save this VIEW file into `/ryftone/view.bin` file.

All filenames can contain special `{{random}}` keyword to put piece of unique
data into filename. For example `data=my-data-{{random}}.txt`.


### Search `lifetime` parameter

By default all output files should be deleted manually. But it's possible to
specify output files lifetime with `lifetime=` parameter.

For example, if `lifetime=1h` is provided then output DATA and INDEX files
will be avialable during one hour and then will be automatically deleted by
the REST service.

The following suffixes are supported:
- `h` for hours, for example `lifetime=2h`
- `m` for minutes, for example `lifetime=20m`
- `s` for seconds, for example `lifetime=200s`


### Search `delimiter` parameter

To customize output format the `delimiter=` parameter may be used. This optional
string will be used to separate found records in the output data file.

By default there is no any delimiter. To use Windows newline
just pass url-encoded `delimiter=%0D%0A`.


### Search `share-mode` parameter

By default ryft-server protects data files from simultaneous read and write.
The `share-mode` option is used to customize sharing mode.

The following sharing modes are supported:
- `share-mode=wait-up-to-10s` or `share-mode=wait-10s`.
  If data file is busy ryft-server waits up to specified timeout.
- `share-mode=skip-busy` or `share-mode=skip`.
  If data file is busy then it is removed from input fileset.
  Note, the input fileset might be empty - ryftprim reports error in this case.
- `share-mode=force-ignore` or `share-mode=ignore`.
  Force to ignore any sharing rules. Even if file is busy try to run the search.
  Note, the result might be undefined.

By default `share-mode=` is equal to `share-mode=wait-0ms` which means
report error immediately if data file is busy.


### Search `nodes` parameter

The number of Ryft processing nodes that the algorithm should use.
A minimally configured Ryft ONE ships from the factory with one processing node,
and a maximally configured Ryft ONE ships with four processing nodes.

If parameter is omitted the maximum of available nodes will be used.


### Search `local` parameter

The local/cluster flag. By default `ryft-server` uses cluster mode `local=false`.
It means `ryft-server` asks all appropriate nodes in the cluster and then combines the results.

To execute a search on single node just pass `local=true`.


### Search `stats` parameter

The statistics is not reported **by default**.
To check total number of matches and performance number just pass `stats=true`.


### Search `performance` parameter

The performance metrics are not reported **by default**.
To get performance metrics in extra statistics just pass `performance=true`.
See [this document](../perf.md) for detailed metrics description.


### Search `limit` parameter

This parameter is used to limit the total number of records reported.
There is no any limit **by default** or when `limit=0`.


### Search `stream` parameter

`ryft-server` reports results in several formats. **By default** the simple JSON object
with "results" array and "stats" object is reported. That format is used by Web-UI:

```{.json}
{
  "results": [
    {
      "Date": "04/15/2015 11:59:00 PM",
      "ID": "10034183",
      "_index": {}
    },
    {
      "Date": "04/15/2015 11:59:00 PM",
      "ID": "10034184",
      "_index": {}
    }
  ],
  "stats": {
    "matches": 2,
    "totalBytes": 6902619,
    "duration": 415,
    "dataRate": 15.862290255994683,
    "fabricDataRate": 15.86229
  }
}
```

But this format is not efficient for cluster nodes communication. We cannot decode JSON object
until whole data is received. So we use "stream" format here - a sequence of JSON "tag-object" pairs
to be able to decode input data on the fly:

```{.json}
"rec"
{
  "Date": "04/15/2015 11:59:00 PM",
  "ID": "10034183",
  "_index": {}
}

"rec"
{
  "Date": "04/15/2015 11:59:00 PM",
  "ID": "10034184",
  "_index": {}
}

"stat"
{
  "matches": 2,
  "totalBytes": 6902619,
  "duration": 1456,
  "dataRate": 4.521188500163319,
  "fabricDataRate": 4.521189
}

"end"
```


### Search `ep` parameter

Helper "error prefix" flag for cluster mode. `ep=false` is used **by default**.

To let user know which cluster node reports error `ryft-server` adds node's hostname to each error message:

`ep=false` (used by default):

```{.json}
{
    "message": "ryftprim failed ...",
    "status": 500
}
```

`ep=true` (used in cluster mode):

```{.json}
{
    "message": "[ryftone-777]: ryftprim failed ...",
   "status": 500
}
```


## Search examples

### Not structured request example

The following request:

```
/search?query=10&files=passengers.txt&surrounding=10&stats=true&local=true
```

will produce the following output:

```{.json}
{"results":[{"_index":{"file":"/passengers.txt","offset":27,"length":22,"fuzziness":0,"host":"ryftone-777"},"data":"YWwgU21pdGgsIDEwLTAxLTE5MjgsMA=="}
,{"_index":{"file":"/passengers.txt","offset":43,"length":22,"fuzziness":0,"host":"ryftone-777"},"data":"MTkyOCwwMTEtMzEwLTU1NS0xMjEyLA=="}
,{"_index":{"file":"/passengers.txt","offset":108,"length":22,"fuzziness":0,"host":"ryftone-777"},"data":"LTI5LTE5NDUsMzEwLTU1NS0yMzIzLA=="}
,{"_index":{"file":"/passengers.txt","offset":167,"length":22,"fuzziness":0,"host":"ryftone-777"},"data":"LTMwLTE5MjAsMzEwLTU1NS0zNDM0LA=="}
,{"_index":{"file":"/passengers.txt","offset":234,"length":22,"fuzziness":0,"host":"ryftone-777"},"data":"MTk1MiwwMTEtMzEwLTU1NS00NTQ1LA=="}
,{"_index":{"file":"/passengers.txt","offset":344,"length":22,"fuzziness":0,"host":"ryftone-777"},"data":"LTE1LTE5NDQsMzEwLTU1NS01NjU2LA=="}
,{"_index":{"file":"/passengers.txt","offset":478,"length":22,"fuzziness":0,"host":"ryftone-777"},"data":"LTE0LTE5NDksMzEwLTU1NS02NzY3LA=="}
,{"_index":{"file":"/passengers.txt","offset":569,"length":22,"fuzziness":0,"host":"ryftone-777"},"data":"LTEyLTE5NTksMzEwLTU1NS0xMjEzLA=="}
,{"_index":{"file":"/passengers.txt","offset":663,"length":22,"fuzziness":0,"host":"ryftone-777"},"data":"LTEyLTE5NTksMzEwLTU1NS0xMjEzLA=="}
,{"_index":{"file":"/passengers.txt","offset":770,"length":22,"fuzziness":0,"host":"ryftone-777"},"data":"LTEyLTE5NTksMzEwLTU1NS0xMjEzLA=="}
,{"_index":{"file":"/passengers.txt","offset":890,"length":22,"fuzziness":0,"host":"ryftone-777"},"data":"LTEyLTE5ODksMzEwLTU1NS05ODc2LA=="}
,{"_index":{"file":"/passengers.txt","offset":966,"length":22,"fuzziness":0,"host":"ryftone-777"},"data":"LTI1LTE5ODUsMzEwLTU1NS0zNDI1LA=="}
],"stats":{"matches":12,"totalBytes":1046,"duration":530,"dataRate":0.0018821572357753536,"fabricDataRate":0.001882}
}
```

`data` field is base-64 encoded raw bytes of found data.


### Structured request example

The following request:

```
/search?query=(RECORD.id CONTAINS "1003100")&files=*.pcrime&format=xml&fields=ID,Date&stats=true&local=true
```

will produce the following output:

```{.json}
{"results":[{"Date":"04/13/2015 11:18:00 PM","ID":"10031002","_index":{"file":"/chicago.pcrime","offset":983054,"length":678,"fuzziness":0,"host":"ryftone-777"}}
,{"Date":"04/13/2015 10:50:00 PM","ID":"10031008","_index":{"file":"/chicago.pcrime","offset":990605,"length":685,"fuzziness":0,"host":"ryftone-777"}}
,{"Date":"04/13/2015 10:45:00 PM","ID":"10031006","_index":{"file":"/chicago.pcrime","offset":991291,"length":687,"fuzziness":0,"host":"ryftone-777"}}
,{"Date":"04/13/2015 10:10:00 PM","ID":"10031003","_index":{"file":"/chicago.pcrime","offset":1006548,"length":687,"fuzziness":0,"host":"ryftone-777"}}
,{"Date":"04/13/2015 09:45:00 PM","ID":"10031004","_index":{"file":"/chicago.pcrime","offset":1020950,"length":684,"fuzziness":0,"host":"ryftone-777"}}
,{"Date":"04/13/2015 07:30:00 PM","ID":"10031009","_index":{"file":"/chicago.pcrime","offset":1079676,"length":688,"fuzziness":0,"host":"ryftone-777"}}
,{"Date":"04/13/2015 09:52:00 AM","ID":"10031001","_index":{"file":"/chicago.pcrime","offset":1333958,"length":689,"fuzziness":0,"host":"ryftone-777"}}
,{"Date":"04/13/2015 07:23:00 AM","ID":"10031000","_index":{"file":"/chicago.pcrime","offset":1373452,"length":673,"fuzziness":0,"host":"ryftone-777"}}
,{"Date":"04/11/2015 07:00:00 PM","ID":"10031005","_index":{"file":"/chicago.pcrime","offset":2096684,"length":683,"fuzziness":0,"host":"ryftone-777"}}
],"stats":{"matches":9,"totalBytes":6902619,"duration":775,"dataRate":8.494000588693925,"fabricDataRate":8.494}
}
```

## Search `Accept` header

Search endpoint produces data encoded as:
- `json` - used by default,
- `msgpack` - used internally in cluster mode
- `csv`

The output type can be specified by `Accept` HTTP header.


### Accept: text/csv

Ryft-server supports `csv` encoding accordingly to [RFC 4180](https://tools.ietf.org/html/rfc4180)

A `csv` file contains zero or more records of one or more fields per record.
Each record is separated by the newline character.
The final record may optionally be followed by a newline character.

A `csv` record can be the following types:

1. data record. The first field is "rec", then INDEX fields and data.
   Data is reported according to selected format (utf8, json, etc).

```
rec,file,offset,length,fuzziness,host,data
```

2. statistics. The first field is "stat", then STAT fields:

```
stat,matches,totalBytes,duration,dataRate,fabricDuration,fabricDataRate,host,details,extra
```

3. error. The first field is "err", then error message:

```
err,message
```

4. End of file. The first field is "end".

```
end
```

Fields which start and stop with the quote character `"` are called quoted-fields.
The beginning and ending quote are not part of the field.

Within a quoted-field a quote character followed by a second quote character is
considered a single quote. Newlines and commas may be included in a quoted-field.


For example,

```{.sh}
$ ryftrest -q hello -f test/foo/1.txt -w=10 --format=utf8 --accept=csv
rec,test/foo/1.txt,0,15,0,ryftone-313,"hello worldhel"
rec,test/foo/1.txt,2,25,0,ryftone-313,"llo worldhello worldhell"
rec,test/foo/1.txt,13,25,0,ryftone-313,ello worldhello from curl
rec,test/foo/1.txt,28,25,0,ryftone-313," from curlhello from curl"
rec,test/foo/1.txt,43,25,0,ryftone-313," from curlhello from curl"
rec,test/foo/1.txt,58,25,0,ryftone-313," from curlhello from curl"
rec,test/foo/1.txt,73,25,0,ryftone-313," from curlhello from curl"
stat,16,233,520,0.00042731945331280044,0,0,ryftone-313,null,{}
end
```

### Accept: application/json

This is default content type. It reports records, errors and statistics in
an appropriate JSON object.


### Accept: application/msgpack

This content type is used internally for communication between nodes in cluster mode.


# Count

The GET `/count` endpoint is also used to search data on Ryft boxes.
However, it does not transfer all found data, it will just print
the number of matches and associated performance numbers.

The `/count` is equivalent to `/search?limit=0`.

Note, this endpoint is protected and user should provide valid credentials.
See [authentication](../auth.md) for more details.

## Count query parameters

The list of supported query parameters are the following:

| Parameter     | Type    | Description |
| ------------- | ------- | ----------- |
| `query`       | string  | **Required**. [The search expression](#search-query-parameter). |
| `file`        | string  | **Required**. [The set of files or catalogs to search](#search-file-parameter). |
| `mode`        | string  | [The search mode](#search-mode-parameter). |
| `surrounding` | uint16  | [The data surrounding width](#search-surrounding-parameter). |
| `fuzziness`   | uint8   | [The fuzziness distance](#search-fuzziness-parameter). |
| `cs`          | boolean | [The case sensitive flag](#search-cs-parameter). |
| `reduce`      | boolean | [The reduce flag for FEDS](#search-reduce-parameter). |
| `transform`   | string  | [The post-process transformation](#search-transform-parameter). |
| `backend`     | string  | [The backend tool](#search-backend-parameter). |
| `backend-option`| string | [The backend tool options](#search-backend-option-parameter). |
| `data`        | string  | [The name of DATA file to keep](#search-data-and-index-parameters). |
| `index`       | string  | [The name of INDEX file to keep](#search-data-and-index-parameters). |
| `view`        | string  | [The name of VIEW file to keep](#search-data-and-index-parameters). |
| `delimiter`   | string  | [The delimiter is used to separate found records](#search-delimiter-parameter). |
| `lifetime`    | string  | [The output files lifetime](#search-lifetime-parameter). |
| `share-mode`  | string  | [The share mode used to access data files](#search-share-mode-parameter). |
| `nodes`       | int     | [The number of processing nodes](#search-nodes-parameter). |
| `local`       | boolean | [The local/cluster search flag](#search-local-parameter). |
| `performance` | boolean | [Flag to report performance metrics](#search-performance-parameter). |

NOTE: The `/count` parameters are absolutely the same as `/search` parameters.
Please check corresponding `/search` related sections.


## Count example

The following request:

```
/count?query=(RECORD CONTAINS "a") OR (RECORD CONTAINS "b")&files=*.pcrime&local=true
```

will report the following output:

```{.json}
{"stats": {
  "matches": 10015,
  "totalBytes": 6902619,
  "duration": 689,
  "dataRate": 9.554209660722487,
  "fabricDataRate": 9.55421
}}
```


# Show

The GET `/search/show` endpoint is used to access already existing search results.
The search results should be prepared first with `/count` or `/search&limit=` methods.

See corresponding [demo](../demo/2017-05-18-search-show.md) page.

Note, this endpoint is protected and user should provide valid credentials.
See [authentication](../auth.md) for more details.

There are a few [content types](#search-accept-header) that server can produce:
- `Accept: application/json` which is used by default
- `Accept: text/csv`
- `Accept: application/msgpack` which is used internally in cluster mode


## Show query parameters

The list of supported query parameters are almost the same as for `/search` method:

| Parameter     | Type    | Description |
| ------------- | ------- | ----------- |
| `offset`      | int     | [The first record index](#show-offset-and-count-parameters). |
| `count`       | int     | [The total number of records to show](#show-offset-and-count-parameters). |
| `format`      | string  | [The structured search format](#search-format-parameter). |
| `fields`      | string  | [The set of fields to get](#search-fields-parameter). |
| `data`        | string  | [The name of DATA file to read](#show-data-and-index-parameters). |
| `index`       | string  | [The name of INDEX file to read](#show-data-and-index-parameters). |
| `view`        | string  | [The name of VIEW file to read](#show-data-and-index-parameters). |
| `delimiter`   | string  | [The delimiter is used to separate found records](#search-delimiter-parameter). |
| `session`     | string  | [The session token](#show-session-parameter). |
| `local`       | boolean | [The local/cluster search flag](#search-local-parameter). |
| `stream`      | boolean | **Internal** [The stream output format flag](#search-stream-parameters). |


### Show `offset` and `count` parameters

These two parameters do specify the records range to show. The `offset` specifies
index of the first record to show. The `count` specifies the total number
of records to show.

For example, if "page" contains 100 records, then:
- the first page is `offset=0&count=100`
- the second page is `offset=100&count=100`
- and so on...


### Show `data` and `index` parameters

For local mode the DATA and INDEX files can be specified directly (for cluster mode
`session` should be used instead). The `index` specifies path to the INDEX file,
it should be the same as for `/search?index=`. The `data` specifies path to the DATA
file, it should be the same as for `/search?data=`.

As an optimization the VIEW file can be specified with `view` parameter.
This file should be generated by corresponding `/search?view=`.

Note all INDEX, DATA and VIEW file should be consistent, i.e. should be releated
to the same `/search` or `/count` call as well as `delimiter`.
Otherwise output is undefined.


### Show `session` parameter

The `session` token can be used instead of `data`, `index`, `view` and `delimiter`
parameters. Session contains all information about previous `/search` or `/count`
call (session token can be extracted from `stats.extra.session` field).

Moreover, session is the only way to show results in cluster mode.
