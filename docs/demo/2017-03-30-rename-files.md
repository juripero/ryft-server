# Demo - REST API change filename endpoint - March 30, 2017

New `PUT files/name` protected endpoint provides an ability to change name of
- file
- directory
- catalog
- file inside catalog

### Parameters:

At least one of parameters (`file`, `directory`, `catalog`) should be presented.

#### `file`

This param should have full path to the source file. 

You can't change file extension using this endpoint.

If it is used with the `catalog` parameter file will be renamed inside catalog.


#### `dir`

This params should have full path to the source directory.


#### `catalog`

This params keeps full path to the catalog file. 

If it is used with the `file` parameter catalog will not be renamed.


#### `new`

This params represents new file name. It can have name of the file, directory or catalog.

This parameter is the only required.


### Examples:

#### Change file name:

```{.sh}
curl -X PUT "http://localhost:8675/files/name?file=/test/secrets/a.txt&new=/test/secrets/a2.txt"
{
  "/ryftone/test/secrets/a.txt": "OK"
}
```

#### Change directory name:

```{.sh}
curl -X PUT "http://localhost:8675/files/name?dir=/test/foo&new=/test/foo2"
{
  "/ryftone/test/foo": "OK"
}
```

#### Change catalog name:
```{.sh}
curl -X PUT "http://localhost:8675/files/name?catalog=test/foo/secrets.txt&new=/test/foo/secrets2.txt"
{
  "/ryftone/test/foo/secrets.txt": "OK"
}
```

#### Change file name inside a catalog:
```{.sh}
curl -X PUT "http://localhost:8675/files/name?catalog=test/foo/secrets.txt&file=a.txt&new=a2.txt"
{
  "/ryftone/test/foo/secrets.txt": "OK"
}
```

#### Empty request
```{.sh}
curl -X PUT "http://localhost:8675/files/name?new=/test/foo2"
{
  "status": 400,
  "message": "missing source filename"
}
```