The GET `/files` endpoint is used to get Ryft box directory content
The name of all subdirectories and files are reported.
Also this method is used to download standalone file or catalog's part.

The POST `/files` endpoint is used to upload a file to Ryft box.
The catalog feature is supported to upload a bunch of small files.

To delete any file, directory or catalog the DELETE `/files` endpoint is used.

To change name of any file, directory, catalog or file inside catalog the PUT `/files/name` endpoint is used.

Note, these endpoints are protected and user should provide valid credentials.
See [authentication](../auth.md) for more details.


## GET Files

The same GET `/files` method is used to get directory content
or to download a file.

The list of supported query parameters to get directory content are the following:

| Parameter | Type    | Description |
| --------- | ------- | ----------- |
| `dir`     | string  | [The directory to get content of](#get-files-dir-parameter). |
| `hidden`  | boolean | [The report hidden files flag](#get-files-hidden-parameter). |
| `local`   | boolean | [The local/cluster flag](#search-local-parameter). |

Note, the `dir` should specify valid directory path.


The list of supported query parameters to get catalog's content are the following:

| Parameter | Type    | Description |
| --------- | ------- | ----------- |
| `catalog` | string  | [The catalog name](#get-files-catalog-parameter). |
| `local`   | boolean | [The local/cluster flag](#search-local-parameter). |

Note, the `file` parameter should be empty and `catalog` should specify
valid catalog path.


The list of supported query parameters to download a standalone file are the following:

| Parameter | Type    | Description |
| --------- | ------- | ----------- |
| `dir`     | string  | [The directory where the file is located](#get-files-dir-parameter). |
| `file`    | string  | [The filename to download](#get-files-file-parameter). |

The following queries are the same:

```{.sh}
curl 'http://localhost:8765/files?dir=foo&file=test.txt'
curl 'http://localhost:8765/files?file=foo/test.txt'
curl 'http://localhost:8765/files/foo?file=test.txt'
```


The list of supported query parameters to download a file from catalog are the following:

| Parameter | Type    | Description |
| --------- | ------- | ----------- |
| `dir`     | string  | [The directory where the catalog is located](#get-files-dir-parameter). |
| `catalog` | string  | [The catalog name](#get-files-catalog-parameter). |
| `file`    | string  | [The filename (inside catalog)](#get-files-file-parameter). |

The following queries are the same:

```{.sh}
curl 'http://localhost:8765/files?dir=foo&catalog=test.catalog&file=test.txt'
curl 'http://localhost:8765/files?catalog=foo/test.catalog&file=test.txt'
curl 'http://localhost:8765/files/foo?catalog=test.catalog&file=test.txt'
```


Moreover the [Content-Range](https://www.w3.org/Protocols/rfc2616/rfc2616-sec14.html)
and [If-Modified-Since](https://www.w3.org/Protocols/rfc2616/rfc2616-sec14.html)
headers are supported for downloading.
So it's possible to download only required part of a file.


## POST Files

The list of supported query parameters for the POST standalone files are the following:

| Parameter | Type    | Description |
| --------- | ------- | ----------- |
| `file`    | string  | [The filename to upload](#post-files-file-parameter). |
| `offset`  | integer | [The optional position of uploaded chunk](#post-files-offset-parameter). |
| `length`  | integer | [The optional length of uploaded chunk](#post-files-length-parameter). |
| `lifetime`| string  | [The optional lifetime of the uploaded file](#post-files-lifetime-parameter). |
| `local`   | boolean | [The optional local/cluster flag](#search-local-parameter). |

The list of supported query parameters for the POST files to catalog:

| Parameter | Type    | Description |
| --------- | ------- | ----------- |
| `catalog` | string  | [The catalog name to upload to](#post-files-catalog-parameter). |
| `delimiter`| string | [The data delimiter to use](#post-files-delimiter-parameter). |
| `file`    | string  | [The filename to upload](#post-files-file-parameter). |
| `offset`  | integer | [The position of uploaded chunk](#post-files-offset-parameter). |
| `length`  | integer | [The length of uploaded chunk](#post-files-length-parameter). |
| `lifetime`| string  | [The optional lifetime of the uploaded file](#post-files-lifetime-parameter). |
|`share-mode`| string | [The share mode used to access data files](#post-files-share-mode-parameter). |
| `local`   | boolean | [The local/cluster flag](#search-local-parameter). |


## DELETE Files

The list of supported query parameters for the DELETE endpoint are the following:

| Parameter | Type    | Description |
| --------- | ------- | ----------- |
| `dir`     | string  | [The directory to delete](#delete-files-parameters). |
| `file`    | string  | [The standalone file to delete](#delete-files-parameters). |
| `catalog` | string  | [The catalog to delete](#delete-files-parameters). |
| `local`   | boolean | [The local/cluster flag](#search-local-parameter). |


### GET files `dir` parameter

The directory to get content of. Root directory `dir=/` is used **by default**.

This parameter also can be used for downloading and it specifies the path
where requested file or catalog is located.

The directory name should be relative to the Ryft volume and user's home.
The `dir=/foo` request will report content of `/ryftone/test/foo` directory on the Ryft box.


### GET files `hidden` parameter

The flag to report hidden files. The `hidden=false` is used **by default**.
That means all the hidden files are not reported.


### GET files `file` parameter

The filename to download.

For standalone file it is the full file path relative to directory `dir` specified.

For catalog it is the filename within catalog specified.


### GET files `catalog` parameter

The catalog to download a part of.

It is the full catalog path relative to directory `dir` specified.


### POST files content

To upload a file the content should be provided.
There are two supported `Content-Type` headers:

- `application/octet-stream`
- `multipart/form-data` - actual file content should be provided via `file` key.

For example:

```{.sh}
$ curl -X POST --data "hello" -H 'Content-Type: application/octet-stream' -s "http://localhost:8765/files?file=/test/file.txt" | jq .
#   -- OR --
$ curl -X POST -F file=@/path/to/file.txt -s "http://localhost:8765/files?file=/test/file.txt&length=100" | jq .
```


### POST files `catalog` parameter

If `catalog` parameter is provided then file will be appended to that catalog
file instead of standalone file uploading. This feature is used to upload a
bunch of small files to a bigger catalog data file.

The `catalog` parameter contains full catalog's path.
All non-existsing sub-directories will be created automatically.

Special keyword `{{random}}` can be used to generate unique catalog names.
This keyword will be replaced with some unique hexadecimal string.
For example, `catalog=foo-{{random}}.catalog` will be replaced to something like
`foo-aabbccddeeff.catalog`. Anyway the actual catalog name will be reported in
the response body.

### POST files `delimiter` parameter

Data delimiter is used in catalog files as a separator between different file
parts. It is very important specially for RAW text files to use something like `delimiter=%0a`.
Otherwise unexpected text matches can be found on file part boundaries.

If no delimiter is provided the default value will be used.
The default delimiter can be customized via ryft-server's
[configuration file](../run.md#catalog-configuration).

Once provided the delimiter cannot be changed for the same catalog.


### POST files `file` parameter

To upload a file the `file` parameter should be provided.
It contains full path of the uploaded data content. For example, if `file=bar/foo.txt`
then the data will be saved under `/ryftone/test/bar/foo.txt` (assuming user's
home directory is `test`).

If `catalog` parameter is not specified the `file` parameter contains full path.
All non-existsing sub-directories will be created automatically.

Special keyword `{{random}}` can be used to generate unique filenames.
This keyword will be replaced with some unique hexadecimal string.
For example, `file=foo-{{random}}.txt` will be replaced to something like
`foo-aabbccddeeff.txt`. Anyway the actual filename will be reported in
the response body.


### POST files `offset` parameter

It's possible to upload just a part of file. If `offset` query parameter is
present then the data will be saved using this offset as write position in
destination file.

Using this parameter it's possible to continue upload of failed data.
Or just split file and upload it in chunks.


### POST files `length` parameter

This optional parameters is used to specify uploading data length in bytes.
This parameter can help ryft server to avoid extra data copy. So if it's
possible this parameter should be provided.


### POST files `lifetime` parameter

This optional parameters is used to specify lifetime of the uploaded data.
If this parameter is provided the file or catalog will be deleted after
specified amount of time. For example if `lifetime=1h` is provided the file
will be availeble during a hour and then will be automatically removed.


### POST files `share-mode` parameter

By default ryft-server protects data files from simultaneous read and write.
The `share-mode` option is used to customize sharing mode.

The following sharing modes are supported:
- `share-mode=wait-up-to-10s` or `share-mode=wait-10s`.
  If data file is busy ryft-server waits up to specified timeout.
- `share-mode=force-ignore` or `share-mode=ignore`.
  Force to ignore any sharing rules. Even if file is busy try to upload the data.
  Note, it might be dangerous and data might be corrupted.

By default `share-mode=` is equal to `share-mode=wait-0ms` which means
report error immediately if data file is busy with search.



### PUT `files/name` parameters


### PUT `file` parameter

The filename to change

For standalone file it is the full file path. 
File extension could be changed.
If file moves outside from the current directory corresponding directory will be created.

For catalog it is the filename within catalog specified


### PUT `directory` parameter

The directory name to change.


### PUT `catalog` parameter

The catalog name to change

If `file` parameters is set it means file name should be changed inside this `catalog`


### PUT `new` parameter

This parameter is for new name of file, directory or catalog.



### DELETE files parameters

It's possible to specify file, directory or catalog to delete.
Multiple parameters can be used together.

Also wildcards are supported. To delete all JSON files just pass `file=*.json`.

All the names should be relative to the Ryft volume and user's home.
The `file=bar/foo.txt` request will delete `/ryftone/test/bar/foo.txt` on the Ryft box
(assuming user's home directory is *test*).


## Files example

The following request:

```
GET /files?dir=/&local=true
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

The following request:

```
DELETE /files?dir=demo&file=*.pcrime&file=p*.txt&local=true
```

will delete specified nodes.
