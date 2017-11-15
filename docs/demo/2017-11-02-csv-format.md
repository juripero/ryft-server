# Demo - CSV format support - November 2, 2017

## Input data

The following test data file will be used for the demo:

```{.sh}
$ cat /ryftone/test.csv
111-hello-11:1,11:1,aa:aa,1.11
222-hello-22:2,22:2,bb:bb,2.22
333-hello-33:3,33:3,cc:cc,3.33
444-hello-44:4,44:4,dd:dd,4.44
555-hello-55:5,55:5,ee:ee,5.55
```

Note, we will use `RAW_TEXT` search and `LINE=true` option to emulate CSV record
search.


## Simple search

The simple UTF8 search gives the following result:

```{.sh}
$ ryftrest -q hello -f test.csv --line --format=utf8 --search | jq -c '.results[]'
{"_index":{"file":"test.csv","offset":0,"length":31, ...},"data":"111-hello-11:1,11:1,aa:aa,1.11\n"}
{"_index":{"file":"test.csv","offset":31,"length":31, ...},"data":"222-hello-22:2,22:2,bb:bb,2.22\n"}
{"_index":{"file":"test.csv","offset":62,"length":31, ...},"data":"333-hello-33:3,33:3,cc:cc,3.33\n"}
{"_index":{"file":"test.csv","offset":93,"length":31, ...},"data":"444-hello-44:4,44:4,dd:dd,4.44\n"}
{"_index":{"file":"test.csv","offset":124,"length":31, ...},"data":"555-hello-55:5,55:5,ee:ee,5.55\n"}
```

Let's interpret that data as CSV records:

```{.sh}
$ ryftrest -q hello -f test.csv --line --format=csv --search | jq -c '.results[]'
{"0":"111-hello-11:1","1":"11:1","2":"aa:aa","3":"1.11","_index":{"file":"test.csv","offset":0,"length":31, ...}}
{"0":"222-hello-22:2","1":"22:2","2":"bb:bb","3":"2.22","_index":{"file":"test.csv","offset":31,"length":31, ...}}
{"0":"333-hello-33:3","1":"33:3","2":"cc:cc","3":"3.33","_index":{"file":"test.csv","offset":62,"length":31, ...}}
{"0":"444-hello-44:4","1":"44:4","2":"dd:dd","3":"4.44","_index":{"file":"test.csv","offset":93,"length":31, ...}}
{"0":"555-hello-55:5","1":"55:5","2":"ee:ee","3":"5.55","_index":{"file":"test.csv","offset":124,"length":31, ...}}
```

Note, the each found record (actually line) is parsed as a few CSV columns.
With no columns name provided the keys are just the column indexes.
In our case: `"0"`, `"1"`, `"2"` and `"3"`.


## Column names

Column names can be provided in request's JSON body:

```{.sh}
$ ryftrest -q hello -f test.csv --line --format=csv --search --body '{"tweaks":{"format":{"columns":["a","b","c","d"]}}}' | jq -c '.results[]'
{"_index":{"file":"test.csv","offset":0,"length":31, ...},"a":"111-hello-11:1","b":"11:1","c":"aa:aa","d":"1.11"}
{"_index":{"file":"test.csv","offset":31,"length":31, ...},"a":"222-hello-22:2","b":"22:2","c":"bb:bb","d":"2.22"}
{"_index":{"file":"test.csv","offset":62,"length":31, ...},"a":"333-hello-33:3","b":"33:3","c":"cc:cc","d":"3.33"}
{"_index":{"file":"test.csv","offset":93,"length":31, ...},"a":"444-hello-44:4","b":"44:4","c":"dd:dd","d":"4.44"}
{"_index":{"file":"test.csv","offset":124,"length":31, ...},"a":"555-hello-55:5","b":"55:5","c":"ee:ee","d":"5.55"}
```

The default indexes are replaced with provided column names.


## Field separator

The `tweaks.format` section can be also used to change the CSV field separator:

```{.sh}
$ ryftrest -q hello -f test.csv --line --format=csv --search --body '{"tweaks":{"format":{"separator":":", "columns":["a","b","c","d"]}}}' | jq -c '.results[]'
{"_index":{"file":"test.csv","offset":0,"length":31, ...},"a":"111-hello-11","b":"1,11","c":"1,aa","d":"aa,1.11"}
{"_index":{"file":"test.csv","offset":31,"length":31, ...},"a":"222-hello-22","b":"2,22","c":"2,bb","d":"bb,2.22"}
{"_index":{"file":"test.csv","offset":62,"length":31, ...},"a":"333-hello-33","b":"3,33","c":"3,cc","d":"cc,3.33"}
{"_index":{"file":"test.csv","offset":93,"length":31, ...},"a":"444-hello-44","b":"4,44","c":"4,dd","d":"dd,4.44"}
{"_index":{"file":"test.csv","offset":124,"length":31, ...},"a":"555-hello-55","b":"5,55","c":"5,ee","d":"ee,5.55"}
```


## Field filter

As for JSON or XML data we can get only required fields and skip others:

```{.sh}
$ ryftrest -q hello -f test.csv --line --format=csv --search --body '{"tweaks":{"format":{"separator":":", "columns":["a","b","c","d"]}}}' --fields 'a,d' | jq -c '.results[]'
{"_index":{"file":"test.csv","offset":0,"length":31, ...},"a":"111-hello-11","d":"aa,1.11"}
{"_index":{"file":"test.csv","offset":31,"length":31, ...},"a":"222-hello-22","d":"bb,2.22"}
{"_index":{"file":"test.csv","offset":62,"length":31, ...},"a":"333-hello-33","d":"cc,3.33"}
{"_index":{"file":"test.csv","offset":93,"length":31, ...},"a":"444-hello-44","d":"dd,4.44"}
{"_index":{"file":"test.csv","offset":124,"length":31, ...},"a":"555-hello-55","d":"ee,5.55"}
```


## Array output

Since the CSV record is an array of fields we can get it as is.
There is special `array` flag in `tweaks.format` section:

```{.sh}
$ ryftrest -q hello -f test.csv --line --format=csv --search --body '{"tweaks":{"format":{"separator":":", "array":true}}}' | jq -c '.results[]'
{"_csv":["111-hello-11","1,11","1,aa","aa,1.11"],"_index":{"file":"test.csv","offset":0,"length":31, ...}}
{"_csv":["222-hello-22","2,22","2,bb","bb,2.22"],"_index":{"file":"test.csv","offset":31,"length":31, ...}}
{"_csv":["333-hello-33","3,33","3,cc","cc,3.33"],"_index":{"file":"test.csv","offset":62,"length":31, ...}}
{"_csv":["444-hello-44","4,44","4,dd","dd,4.44"],"_index":{"file":"test.csv","offset":93,"length":31, ...}}
{"_csv":["555-hello-55","5,55","5,ee","ee,5.55"],"_index":{"file":"test.csv","offset":124,"length":31, ...}}
```

The found record is saved as an array under `"_csv"` key.

Array output is very helpful with the `CSV` output format:

```{.sh}
$ ryftrest -q hello -f test.csv --line --format=csv --search --body '{"tweaks":{"format":{"separator":":", "array":true}}}' --accept=csv | head -n5
rec,test.csv,0,31,-1,ubuntu-vm,111-hello-11,"1,11","1,aa","aa,1.11"
rec,test.csv,31,31,-1,ubuntu-vm,222-hello-22,"2,22","2,bb","bb,2.22"
rec,test.csv,62,31,-1,ubuntu-vm,333-hello-33,"3,33","3,cc","cc,3.33"
rec,test.csv,93,31,-1,ubuntu-vm,444-hello-44,"4,44","4,dd","dd,4.44"
rec,test.csv,124,31,-1,ubuntu-vm,555-hello-55,"5,55","5,ee","ee,5.55"
```

Without `"array":true` flag the CSV output will look weird - each
record will be reported as stringified JSON object.


## Aggregations

As for JSON or XML data we can run aggregations on CSV data too:

```{.sh}
$ ryftrest -q hello -f test.csv --line --format=csv --search --body '{"aggs":{"my_stat":{"stats":{"field":"[3]"}}}}' | jq '.stats.extra.aggregations'
{
  "my_stat": {
    "avg": 3.33,
    "count": 5,
    "max": 5.55,
    "min": 1.11,
    "sum": 16.65
  }
}
```

Note, the field is expressed as `"[3]"` which actually means the column index `#3`.
If column names are provided then field can be expressed with appropriate column:

```{.sh}
$ ryftrest -q hello -f test.csv --line --format=csv --search --body '{"tweaks":{"format":{"columns":["a","b","c","d"]}}, "aggs":{"my_stat":{"stats":{"field":".d"}}}}' | jq '.stats.extra.aggregations'
{
  "my_stat": {
    "avg": 3.33,
    "count": 5,
    "max": 5.55,
    "min": 1.11,
    "sum": 16.65
  }
}
```

The `".d"` field will be internally converted to the column index `#3`.
