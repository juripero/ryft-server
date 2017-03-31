To change name of any file, directory, catalog or file inside catalog the PUT `/rename` endpoint is used.

### PUT `rename` parameters


### PUT `file` parameter

The filename to change

For standalone file it is the full file path. 
File extension could be changed.
If file moves outside from the current directory corresponding directory will be created.

For catalog it is the filename within catalog specified


### PUT `dir` parameter

The directory name to change.


### PUT `catalog` parameter

The catalog name to change

If `file` parameters is set it means file name should be changed inside this `catalog`


### PUT `new` parameter

This parameter is for new name of file, directory or catalog. It is required.


### PUT `local` parameter

This parameter allows massive renaming throught the all cluster machines. Default is `false`.