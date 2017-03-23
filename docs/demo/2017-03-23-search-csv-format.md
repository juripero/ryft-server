# Demo - Search endpoint response in the CSV format - March 23, 2017

For this demonstration we need to upload file test/foo/1.txt.

## Encoding types

Ryft-server's `search` endpoint now supports three types of encoding:
- json
- msgpack
- csv

You can change response format sending `Accept` header with the appropriate encoding type.
For `csv` it is `text/csv`.
We added support of this header into ryftrest script with the `--accept` parameter.

Swagger web-ui also supports `accept` header selector.

## Short description of the CSV format

Each row in a response represents one record from the search results.
Thus in a response we have different types of data.

In order to differ these records each row in CSV format starts with the type of a record.

Symbol of comma (',') is the separator.

Each record ends with the newline character ('\n').

Each record may contains quotes and newline characters.

We used RFC 4180 as a base.


1. data record has type `rec`. Then fields from INDEX (filename,offset,length,fuzziness)
and data in format defined in the `format` parameter.

2. statistics has type `stat`. Then fields from STAT (matches, total_bytes, duration, data_rate, fabric_duration, fabric_data_rate, host, details, extra)

3. error has type `err`. Then string with the error message.

4. `end` denotes the end of the response body.


## examples
#### record

```{.sh}
bash ryftrest -q 'hello' -w 10 -f test/foo/1.txt --format=utf8 --search --no-stats --address localhost:9786 --accept csv
rec,test/foo/1.txt,0,15,0,ryftone-313,"hello world
hel"
rec,test/foo/1.txt,2,25,0,ryftone-313,"llo world
hello worldhell"
rec,test/foo/1.txt,13,25,0,ryftone-313,ello worldhello from curl
rec,test/foo/1.txt,28,25,0,ryftone-313," from curlhello from curl"
rec,test/foo/1.txt,43,25,0,ryftone-313," from curlhello from curl"
rec,test/foo/1.txt,58,25,0,ryftone-313," from curlhello from curl"
rec,test/foo/1.txt,73,25,0,ryftone-313," from curlhello from curl"
rec,test/foo/1.txt,88,25,0,ryftone-313," from curlhello from curl"
rec,test/foo/1.txt,103,25,0,ryftone-313," from curlhello from curl"
rec,test/foo/1.txt,118,25,0,ryftone-313," from curlhello from curl"
rec,test/foo/1.txt,133,25,0,ryftone-313," from curlhello from curl"
rec,test/foo/1.txt,148,25,0,ryftone-313," from curlhello from curl"
rec,test/foo/1.txt,163,25,0,ryftone-313," from curlhello from curl"
rec,test/foo/1.txt,178,25,0,ryftone-313," from curlhello from curl"
rec,test/foo/1.txt,193,25,0,ryftone-313," from curlhello from curl"
rec,test/foo/1.txt,208,25,0,ryftone-313," from curlhello from curl"
rec,test/foo/1.txt,230,22,0,ryftone-313,url310310 hello 310310
end
```

#### statistics
```{.sh}
bash ryftrest -q test -w 10 -f test/foo/1.txt --format=utf8 --search --address localhost:9786 --accept csv
stat,0,252,294,0.0008174351283482144,0,0,ryftone-313,null,{}
end
```

#### error
```{.sh}
bash ryftrest -q 'hkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkk' -w 10 -f test/foo/1.txt --format=utf8 --search --no-stats --address localhost:9786 --accept csv
err,"ryftprim failed with exit status 156
ERROR:  An error occurred - results are not valid!
"
err,"500 ryftprim failed with exit status 156
ERROR:  An error occurred - results are not valid!
"
end
```

#### record + statistics
```{.sh}
bash ryftrest -q hello -w 10 -f test/foo/1.txt --format=utf8 --search --address localhost:9786 --accept csv
rec,test/foo/1.txt,0,15,0,ryftone-313,"hello world
hel"
rec,test/foo/1.txt,2,25,0,ryftone-313,"llo world
hello worldhell"
rec,test/foo/1.txt,13,25,0,ryftone-313,ello worldhello from curl
rec,test/foo/1.txt,28,25,0,ryftone-313," from curlhello from curl"
rec,test/foo/1.txt,43,25,0,ryftone-313," from curlhello from curl"
rec,test/foo/1.txt,58,25,0,ryftone-313," from curlhello from curl"
rec,test/foo/1.txt,73,25,0,ryftone-313," from curlhello from curl"
rec,test/foo/1.txt,88,25,0,ryftone-313," from curlhello from curl"
rec,test/foo/1.txt,103,25,0,ryftone-313," from curlhello from curl"
rec,test/foo/1.txt,118,25,0,ryftone-313," from curlhello from curl"
rec,test/foo/1.txt,133,25,0,ryftone-313," from curlhello from curl"
rec,test/foo/1.txt,148,25,0,ryftone-313," from curlhello from curl"
rec,test/foo/1.txt,163,25,0,ryftone-313," from curlhello from curl"
rec,test/foo/1.txt,178,25,0,ryftone-313," from curlhello from curl"
rec,test/foo/1.txt,193,25,0,ryftone-313," from curlhello from curl"
rec,test/foo/1.txt,208,25,0,ryftone-313," from curlhello from curl"
rec,test/foo/1.txt,230,22,0,ryftone-313,url310310 hello 310310
stat,17,252,724,0.0003319418891358771,0,0,ryftone-313,null,{}
end
```

#### record in XML format

Response has json-encoded `data` field

```{.sh}

bash ryftrest -q '(RECORD CONTAINS "electronic")' -i -w 10 -f regression/chicago.ryftpcrime --format=xml --search --address localhost:9786 --accept=csv --fields=Arrest,Beat
rec,regression/chicago.ryftpcrime,71061,700,-1,ryftone-313,"{""Arrest"":""false"",""Beat"":""2432""}"
rec,regression/chicago.ryftpcrime,71762,705,-1,ryftone-313,"{""Arrest"":""false"",""Beat"":""1731""}"
rec,regression/chicago.ryftpcrime,91109,700,-1,ryftone-313,"{""Arrest"":""false"",""Beat"":""0531""}"
rec,regression/chicago.ryftpcrime,173904,703,-1,ryftone-313,"{""Arrest"":""false"",""Beat"":""1421""}"
rec,regression/chicago.ryftpcrime,184280,703,-1,ryftone-313,"{""Arrest"":""false"",""Beat"":""2423""}"
rec,regression/chicago.ryftpcrime,344315,703,-1,ryftone-313,"{""Arrest"":""false"",""Beat"":""1622""}"
rec,regression/chicago.ryftpcrime,353300,706,-1,ryftone-313,"{""Arrest"":""false"",""Beat"":""1821""}"
rec,regression/chicago.ryftpcrime,400385,701,-1,ryftone-313,"{""Arrest"":""false"",""Beat"":""0834""}"
...omitted...
rec,regression/chicago.ryftpcrime,425869,699,-1,ryftone-313,"{""Arrest"":""false"",""Beat"":""2411""}"
rec,regression/chicago.ryftpcrime,702628,705,-1,ryftone-313,"{""Arrest"":""false"",""Beat"":""2515""}"
stat,72,6892667,306,21.48156695895725,1,6573.359375,ryftone-313,null,{}
end
```
