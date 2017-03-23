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

## Short description of the CSV format

Each row in a response represents one record from the search results.

Each row in CSV format starts with the type of a record.

Symbol of comma (',') is the separator.

Each record ends with the newline character ('\n').

Each record may contains quotes and newline characters.


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
