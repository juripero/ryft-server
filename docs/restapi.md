This document contains information about REST API.

`ryft-server` supports the following API endpoits:

- [/version](#version)
- [/search](#search)
- [/count](#count)
- [/files](#files)

The main API endpoints are `/search` and `/count`.


# Search

The GET `/search` endpoint is used to search data on Ryft boxes.

## Search query parameters

The list of supported query parameters are the following (check detailed description below):

| Parameter     | Type    | Description |
| ------------- | ------- | ----------- |
| `query`       | string  | **Required**. [The search expression](#search-query-parameter). |
| `files`       | string  | **Required**. [The set of files to search](#search-files-parameter). |
| `mode`        | string  | [The search mode](#search-mode-parameter). |
| `surrounding` | uint16  | [The data surrounding width](#search-surrounding-parameter). |
| `fuzziness`   | uint8   | [The fuzziness distance](#search-fuzziness-parameter). |
| `format`      | string  | [The structured search format](#search-format-parameter). |
| `cs`          | boolean | [The case sensitive flag](#search-cs-parameter). |
| `fields`      | string  | [The set of fields to get](#search-fields-parameter). |
| `data`        | string  | [The name of data file to keep](#search-data-and-index-parameters). |
| `index`       | string  | [The name of index file to keep](#search-data-and-index-parameters). |
| `nodes`       | int     | [The number of processing nodes](#search-nodes-parameter). |
| `local`       | boolean | [The local/cluster search flag](#search-local-parameter). |
| `stats`       | boolean | [The statistics flag](#search-stats-parameter). |
| `stream`      | boolean | **Internal** [The stream output format flag](#search-stream-and-spark-parameters). |
| `spark`       | boolean | **Internal** [The spark output format flag](#search-stream-and-spark--parameters). |
| `ep`          | boolean | **Internal** [The error prefix flag](#search-ep-parameter). |


### Search `query` parameter

The first required parameter is the search expression `query`.
It contains one or more subqueries connected using logical operators.

To do text search for the "The Batman" the following search expression is used:

```
query=(RAW_TEXT CONTAINS "The Batman")
```

For structured search another keyword should be applied:

```
query=(RECORD.AlterEgo CONTAINS "The Batman")
```

Depending on [search mode](#search-mode-parameter) exact search query format may differ.
Check corresponding Ryft Open API for more details on search expressions.

`ryft-server` supports simple plain queries - without any keywords.
The `query=Batman` will be automatically converted to `query=(RAW_TEXT CONTAINS "Batman")`.
Note, this works for text search only and not appropriate for structured search.
(Actually, the query will be `query=(RAW_TEXT CONTAINS 4261746d616e)`,
`ryft-server` uses hex encoding to avoid any possible escaping problems).

`ryft-server` also supports complex queries containing several search expressions of different types.
For example `(RECORD.id CONTAINS "100") AND (RECORD.date CONTAINS DATE(MM/DD/YYYY > 04/15/2015))`.
This complex query contains two search expression: first one uses text search and the second one uses date search.
`ryft-server` will split this expression into two separate queries:
`(RECORD.id CONTAINS "100")` and `(RECORD.date CONTAINS DATE(MM/DD/YYYY > 04/15/2015))`.
It then calls Ryft hardware two times: the results of the first call are used as the input for the second call.

Multiple `AND` and `OR` operators are supported by the `ryft-server` within complex search queries.
Expression tree is built and each node is passed to the Ryft hardware. Then results are properly combined.

Note, if search query contains two or more expressions of the same type (text, date, time, numeric) that query
will not be splitted into subqueries because the Ryft hardware supports those type of queries directly.


### Search `files` parameter

The second required parameter is the set of file to search.
At least one file should be provided.

Multiple files can be provided as:

  - a list `files=1.txt&files=2.txt`
  - a wildcard: `files=*txt`


### Search `mode` parameter

`ryft-server` supports several search modes:

- `es` for exact search
- `fhs` for fuzzy hamming search
- `feds` for fuzzy edit distance search
- `ds` for date search
- `ts` for time search
- `ns` for numeric search

If no any search mode provided fuzzy hamming search is used **by default** for simple queries.
It is also possible to automatically detect search modes: if search query contains `DATE`
keyword then date search will be used, `TIME` keyword is used for time search,
and `NUMERIC` for numeric search.

In case of complex search queries provided mode is used for text or structured search only.
Date, time and numeric search modes will be detected automatically by corresponding keywords.

Note, the fuzzy edit distance search mode removes duplicates by default (`-r` option of ryftprim).


### Search `surrounding` parameter

The number of characters in bytes up to a maximum of `262144` before the match
and after the match that will be returned when the text search is used.
For anything other than raw text, this parameter is ignored.

`surrounding=0` is used **by default**.


### Search `fuzziness` parameter

The fuzziness of the search up to a maximum of `255` when using a fuzzy search function.
For fuzzy hamming search, fuzziness is measured as the maximum Hamming distance allowed
in order to declare a match. For fuzzy edit distance search, fuzziness is measured
as the number of insertions, deletions or replacements required to declare a match.

`fuzziness=0` is used **by default**.


### Search `format` parameter

The input data format for the structured search.

**By default** structured search uses `format=raw` format.
That means that found data are reported as base-64 encoded raw bytes.

There are two other options: `format=xml` and `format=json`.

If input file set contains XML data, the found records could be decoded. Just pass `format=xml` query parameter
and records will be translated from XML to JSON.

The same is true for JSON data. 

See also [fields parameter](#search-fields-parameter).


### Search `cs` parameter

The search text case-sensitive flag.

For example, if the search is case-sensitive `cs=true`, then searching for the string "John"
will not find any occurrences of "JOHN". If the same search is done with `cs=false`, then
case is ignored entirely and all possible capitalizations of the text will be found
(e.g. "jOhn" or "JOHn").

`cs=false` is used **by default**.


### Search `fields` parameter

The coma-separated list of fields for structured search. If omitted all fields are used.

This parameter is used to minimize structured search output or to get just subset of fields.
For example to get identifier and date from a `*.pcrime` file pass `format=xml&fields=ID,Date`.

The same is true for JSON data: `format=json&fields=Name,AlterEgo`.


### Search `data` and `index` parameters

By default all search results are deleted from the Ryft server once they are delivered to user.
In order to preserve results thereby allowing for the ability to subsequently
"search in the previous results", there are two query parameters available: `data=` and `index=`.

Using the first parameter, `data=output.dat` creates the search results on the Ryft server under `/ryftone/output.dat`.
It is possible to use that file as an input for the subsequent search call `files=output.dat`.

Note, it is important to use consistent file extension for the structured search
in order to let Ryft use appropriate RDF scheme!

Using the second parameter `index=index.txt` keeps the search index file under `/ryftone/index.txt`.

Note, according to Ryft API documentation index file should always have `.txt` extension!

**WARNING:** Provided data or index files will be overriden!


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


### Search `stream` and `spark` parameters

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
until whole data will be received. So we use "stream" format here - a sequence of JSON "tag-object" pairs
to be able to decode input data of the fly:

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

"spark" format is even shorter than "stream". It reports sequence of records (one by one):

```{.json}
{
  "Date": "04/15/2015 11:59:00 PM",
  "ID": "10034183",
  "_index": {}
}
{
  "Date": "04/15/2015 11:59:00 PM",
  "ID": "10034184",
  "_index": {}
}
```

The "spark" format is obsolete for now and will be removed soon.


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
/search?query=10&files=passengers.txt&surrounding=10&fuzziness=0&stats=true&local=true
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
/search?query=(RECORD.id CONTAINS "1003100")&files=/*.pcrime&fuzziness=0&format=xml&fields=ID,Date&stats=true&local=true
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


# Count

The GET `/count` endpoint is also used to search data on Ryft boxes.
However, it does not transfer all found data, it will just print
the number of matches and associated performance numbers.

## Count query parameters

The list of supported query parameters are the following:

| Parameter     | Type    | Description |
| ------------- | ------- | ----------- |
| `query`       | string  | **Required**. [The search expression](#search-query-parameter). |
| `files`       | string  | **Required**. [The set of files to search](#search-files-parameter). |
| `mode`        | string  | [The search mode](#search-mode-parameter). |
| `surrounding` | uint16  | [The data surrounding width](#search-surrounding-parameter). |
| `fuzziness`   | uint8   | [The fuzziness distance](#search-fuzziness-parameter). |
| `cs`          | boolean | [The case sensitive flag](#search-cs-parameter). |
| `data`        | string  | [The name of data file to keep](#search-data-and-index-parameters). |
| `index`       | string  | [The name of index file to keep](#search-data-and-index-parameters). |
| `nodes`       | int     | [The number of processing nodes](#search-nodes-parameter). |
| `local`       | boolean | [The local/cluster search flag](#search-local-parameter). |

Note, most of the `/count` parameters are absolutely the same as `/search` parameters.
Please check corresponding `/search` related sections.


## Count example

The following request:

```
/count?query=(RECORD CONTAINS "a") OR (RECORD CONTAINS "b")&files=*.pcrime&local=true
```

will report the following output:

```{.json}
{
  "matches": 10015,
  "totalBytes": 6902619,
  "duration": 689,
  "dataRate": 9.554209660722487,
  "fabricDataRate": 9.55421
}
```


# Files

The GET `/files` endpoint is used to get Ryft box directory content.
The name of all subdirectories and files are reported.


## Files query parameters

The list of supported query parameters are the following:

| Parameter | Type    | Description |
| --------- | ------- | ----------- |
| `dir`     | string  | [The directory to get content of](#files-dir-parameter). |
| `local`   | boolean | [The local/cluster flag](#search-local-parameter). |


### Files `dir` parameter

The directory to get content of. Root directory `dir=/` is used **by default**.

The directory name should be relative to the Ryft volume.
The `dir=/test` request will report content of `/ryftone/test` directory on the Ryft box.


## Files example

The following request:

```
/files?dir=/&local=true
```

will print the root `/ryftone` content:

```{.json}
{
  "dir": "/",
  "files": [
    "chicago.pcrime",
    "passengers.txt"
  ],
  "folders": [
    "demo",
    "regression",
    "test"
  ]
}
```


# Version

The GET `/version` endpoint is used to check current `ryft-server` version.

It has no parameters, output looks like:

```{.json}
{
  "git-hash": "35c358378f7c214069333004d01841f9066b8f15",
  "version": "1.2.3"
}
```
