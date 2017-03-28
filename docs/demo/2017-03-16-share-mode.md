# Demo - simultaneous upload/search protection - March 16, 2017

The `ryftprim` tool fails if (during the search operation) a file from
input fileset is modified by another program.

There is special [share-mode](../rest/search.md#search-share-mode-parameter)
REST API option that contols sharing.

By default there no uploads are allowed if search operation is in progress.
And vice versa, no search operatins are allowed if upload file is in progress.


## Report error immediatelly.

```{.sh}
#  -- first terminal --
curl --data "hello from curl" -H "Content-Type: application/octet-stream" \
     -s "http://ryft-313:8877/files?file=test/foo/1.txt" | jq .

#  -- second terminal ---
ryftrest -q hello -f test/foo/1.txt --format=utf8 --search --address ryft-313:8877 | jq .
```

The "file is busy" error will be reported for the second call.


## Try to wait.

We can instruct ryft-server to wait a bit for a file with `share-mode=wait-20s`:

```{.sh}
#  -- first terminal --
curl --data "hello from curl" -H "Content-Type: application/octet-stream" \
     -s "http://ryft-313:8877/files?file=test/foo/1.txt" | jq .

#  -- second terminal ---
ryftrest -q hello -f test/foo/1.txt --format=utf8 --search --address ryft-313:8877 --share-mode=wait-20s | jq .
```

Note, the second call waits until file is released and then continues.


## Skip busy files.

Busy file can be skipped:

```{.sh}
#  -- first terminal --
curl --data "hello from curl" -H "Content-Type: application/octet-stream" \
     -s "http://ryft-313:8877/files?file=test/foo/1.txt" | jq .

#  -- second terminal ---
ryftrest -q hello -f test/foo/1.txt -f test/foo/11.txt --format=utf8 --search --address ryft-313:8877 --share-mode=skip | jq .
```

Note, the two files for the second call. Result doesn't contain output from busy file.


## Ignore existing locks.

The sharing rules can be ignored:

```{.sh}
#  -- first terminal --
curl --data "hello from curl" -H "Content-Type: application/octet-stream" \
     -s "http://ryft-313:8877/files?file=test/foo/1.txt" | jq .

#  -- second terminal ---
ryftrest -q hello -f test/foo/1.txt --format=utf8 --search --address ryft-313:8877 --share-mode=ignore | jq .
```

Note, the ryftprim reports "One or more input files  could not be locked or are currently open for writing" error.
