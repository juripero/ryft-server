# Demo - download file and get extended directory listing - March 13, 2017

## Download standalone file

The ryft server now support [GET /files](../rest/files.md#get-files) REST API to
download a file. The [file](../rest/files.md#get-files-file-parameter) option
should specify an existing file.

For example, if we upload the following data first:

```{.sh}
$ curl -X POST --data "hello world" -H "Content-Type: application/octet-stream" \
       -s "http://ryft-313:9876/files?file=test/foo/1.txt" | jq .
{
  "length": 11,
  "offset": 0,
  "path": "test/foo/1.txt"
}
```

Then we can get the file back with the following command:

```{.sh}
$ curl -s "http://ryft-313:9876/files?file=test/foo/1.txt"
hello world
```

The `Range` header is supported to download a part of a file:

```{.sh}
curl -H "Range: bytes=3-7" "http://ryft-313:9876/files?file=test/foo/1.txt"
lo wo
```


## Download a file from catalog

The same approach is used for catalogs. For example, let's upload a few data
parts with names `a.txt` and `b.txt`:

```{.sh}
curl --data "aaa hello " -H "Content-Type: application/octet-stream" \
     -s "http://ryft-313:9876/files?catalog=test/foo/2.txt&file=a.txt" | jq .
curl --data "bbb hello bbb" -H "Content-Type: application/octet-stream" \
     -s "http://ryft-313:9876/files?catalog=test/foo/2.txt&file=b.txt" | jq .
curl --data "world aaa" -H "Content-Type: application/octet-stream" \
     -s "http://ryft-313:9876/files?catalog=test/foo/2.txt&file=a.txt" | jq .
```

To download files back:

```{.sh}
$ curl -s "http://ryft-313:9876/files?catalog=test/foo/2.txt&file=a.txt"
aaa hello world aaa
$ curl -s "http://ryft-313:9876/files?catalog=test/foo/2.txt&file=b.txt"
bbb hello bbb
```

## Extended directory listing

The [GET /files](../rest/files.md#get-files) REST API endpoint also used to
get directory listing. The output is backward compatible with the previous
release and contains additional fields:
- `catalogs` is a list of catalogs. Note, the catalogs are also duplicated in
   `files` field.
- `details` is a detailed information grouped by cluster nodes. It contains
    file size, modification time and permissions.

For example:

```{.sh}
$ curl -s "http://ryft-313:9876/files/test/foo" | jq .
{
  "dir": "/test/foo",
  "files": [
    "1.txt",
    "2.txt"
  ],
  "catalogs": [
    "2.txt"
  ],
  "details": {
    "ryftone-313": {
      "1.txt": {
        "type": "file",
        "length": 23,
        "mtime": "2017-03-14T09:11:39-04:00",
        "perm": "-rw-rw-r--"
      },
      "2.txt": {
        "type": "catalog",
        "length": 83,
        "mtime": "2017-03-14T09:24:07-04:00",
        "perm": "-rw-r--r--"
      }
    }
  }
}
```

The same API endpoint is used to get catalog's content. Target path
should specify an existing catalog:

```{.sh}
$ curl -s "http://ryft-313:9876/files/test/foo/2.txt" | jq .
{
  "catalog": "/test/foo/2.txt",
  "files": [
    "a.txt",
    "b.txt"
  ],
  "details": {
    "ryftone-313": {
      "a.txt": {
        "type": "file",
        "length": 20,
        "parts": [
          {
            "length": 10,
            "offset": 0
          },
          {
            "length": 10,
            "offset": 10
          }
        ]
      },
      "b.txt": {
        "type": "file",
        "length": 13
      }
    }
  }
}
```

## How to show hidden files and directories

There is [special option](../rest/files.md#get-files-hidden-parameter)
to show hidden files and directories:

```{.sh}
$ curl -s "http://ryft-313:9876/files?hidden=true" | jq .
```

By default this option is disabled so no hidden files are shown.
