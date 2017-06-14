To change name of any file, directory, catalog or file inside catalog the PUT `/rename` endpoint is used.

## PUT `rename` parameters

The list of supported query parameters are the following:

| Parameter | Type    | Description |
| --------- | ------- | ----------- |
| `file`    | string  | [The filename to change](#put-file-parameter). |
| `dir`     | string  | [The directory name to change](#put-dir-parameter). |
| `catalog` | string  | [The catalog name to change](#put-catalog-parameter). |
| `new`     | string  | [The new name](#put-new-parameter). |
| `local`   | boolean | [The local/cluster flag](#search-local-parameter). |


### PUT `file` parameter

The filename to change.

For standalone file it is the full file path.
File extension can not be changed.
If file moves outside from the current directory corresponding directory will be created.

For catalog it is the filename within catalog specified


### PUT `dir` parameter

The directory name to change.


### PUT `catalog` parameter

The catalog name to change.
Catalog extension can not be changed.

If `file` parameters is set it means file name should be changed inside this `catalog`


### PUT `new` parameter

This parameter is for new name of file, directory or catalog. It is required.


### PUT `local` parameter

This parameter allows massive renaming throught the all cluster machines. Default is `false`.



## URL-path

You can also set path to the source file, directory or catalog right in URL-path. 
```
/rename?file=/path/to/file.txt&...
```
equals to
```
/rename/path/to/?file=file.txt&...
```
