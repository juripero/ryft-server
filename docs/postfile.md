Posting files

This document describes the `POST /files` API endpoint.
The main purpose of this method is to upload data to the Ryft box.


The `POST /files` API has two forms:

- [upload standalone files](#upload-standalone-files)
- [append catalog files](#append-catalog-files)

Both forms require authorization. TODO: user roles?


# Upload standalone files

The following query parameters can be used:

- `file=<path>` the file path where the data will be saved. Required.
- `offset=0` The offset inside target file where to write data. Optional.

To upload a file the following command can be used:

```{.sh}
curl -X POST --data "<file content here>" -H 'Content-Type: application/octet-stream' -s "http://localhost:8765/files?file=/test/file\{\{random\}\}.txt"
```

The file content will be saved under randomly generated name,
for example `/test/file147086a5143b3b26.txt`. Any occurence of `{{random}}`
will be replaced with hexadecimal random string.

All sub-directories will be created automatically. In the example above
the `/test/` directory will be created if it doesn't exist.

If everything is OK the result contains the following JSON object:

```{.json}
{
  "path": "/test/file147086a5143b3b26.txt",
  "length": 1046
}
```

The result contains the total number of bytes written (`length` field).
If something went wrong it's possible to upload remaining part of file
using `offset` query parameter. To continue file uploading just pass
`offset` equal to the `length` reported on previous step.

It's also possible to use multipart form data to upload file:

```{.sh}
curl -X POST -F content=@/path/to/file.txt -s "http://localhost:8765/files?file=/test/file\{\{random\}\}.txt" | jq .
```

TODO: use of gzip compressed streams and chunked transfer encoding.

If `offset` is not provided and file is already exists the error will be reported.
Use `DELETE /files` to remove unused files.


# Append catalog files

In some cases there are a bunch of files each of quite small size.
In order to make Ryft call faster these files may be combined into
the bigger so called catalog files.

The following query parameters can be used:

- `file=<path>` the file path (inside catalog) where the data will be saved. Required.
- `catalog=<path>` the catalog file path. Required.
- `offset=0` The offset inside target file where to write data. Optional.

To append a file to the catalog the following command can be used:

```{.sh}
curl -X POST --data "<file content here>" -H 'Content-Type: application/octet-stream' -s "http://localhost:8765/files?file=my.txt&catalog=/test/my-catalog"
```

If everything is OK the result contains the following JSON object:

```{.json}
{
  "data": ["/test/my-catalog.0", "/test/my-catalog.1"] // ???
  "path": "my.txt",
  "length": 1046
}
```

- `length` contains the number of bytes written.
- `data` contains a list of catalog data files. The subsequent searches should use these files as input.
- `path` contains file name relative to catalog.

Filename should be unique.

As for standalone file uploading the `offset` query parameter can be used to upload
just a part of file.

To help `ryft-server` the file length can be provided (via query parameter?).
In this case `ryft-server` can reserve required space in catalog file.
Otherwise it saves file content to temporary file and once it's uploaded
moves the whole content to catalog file.

The `GET /files?catalog=<path>` can be used to get extended catalog information.

## Implementation details

There are a few data files and one meta-data describing catalog structure.
The catalog meta-data file is proposed to be a SQLite database.

The catalog meta-data should contain the following information:

- list of items:
  - file name
  - length
  - data file path
  - offset
- list of data files:
  - data file path
  - total length
  - number of items
- other information (blacklists?)

This meta information is required to implement re-transmission feature.

In case of errors the holes in data file are possible.
Garbage collection - to remove broken files from catalog.

Search and upload simultaneous access - searching while uploading.
Lock upload till search is done.

# Error handling

In case of an error the number of written data is reported.
This value can be used to continue file uploading.


# Cluster support

Replication and sharding.

Errors from all nodes should be collected.

The minimum length is reported.

# DELETE /files

to delete files/directories and catalogs
