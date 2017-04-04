# Demo - REST API change filename endpoint - April 6, 2017

New `PUT /rename` protected endpoint provides an ability to change name of
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

### Rename in `local-only` mode

Run server with this command
```{.sh}
./ryft-server --config=/etc/ryft-server.conf --local-only
```

#### Change file name:

```{.sh}
curl -X PUT "http://localhost:8675/rename?file=/secrets/a.txt&new=/secrets2/a2.txt"
{
  "/secrets/a.txt": "OK"
}
```

#### Change directory name:

```{.sh}
curl -X PUT  "http://localhost:8675/rename?dir=/foo&new=/foo2"
{
  "/foo": "OK"
}
```

#### Change catalog name:
```{.sh}
curl -X PUT "http://localhost:8675/rename?catalog=/foo/secrets.txt&new=/foo/secrets2.txt"
{
  "/foo/secrets.txt": "OK"
}
```

#### Change file name inside a catalog:
```{.sh}
curl -X PUT "http://localhost:8675/rename?catalog=/foo/secrets.txt&file=c.txt&new=c_2.txt"
{
  "c.txt": "OK"
}
```

### Rename in a `cluster` mode

#### Change file name:

```{.sh}
curl -X PUT "http://localhost:8675/rename?file=/secrets/a.txt&new=/secrets2/a2.txt"
{
  "details": {
    "ryftone-310": {
      "length": 17,
      "path": "/secrets/b.txt"
    },
    "ryftone-313": {
      "length": 17,
      "offset": 0,
      "path": "/secrets/b.txt"
    }
  },
  "length": 17,
  "path": "/secrets/b.txt"
}
```

#### Change directory name:

```{.sh}
curl -X PUT  "http://localhost:8675/rename?dir=/foo&new=/foo2&local=false"
{
  "ryftone-310": {
    "/foo": "OK"
  },
  "ryftone-313": {
    "/foo": "OK"
  }
}
```

#### Change catalog name:
```{.sh}
curl -X PUT "http://localhost:8675/rename?catalog=/foo/secrets.txt&new=/foo/secrets2.txt"
{
  "ryftone-310": {
    "/foo/secrets.txt": "OK"
  },
  "ryftone-313": {
    "/foo/secrets.txt": "OK"
  }
}
```

#### Change file name inside a catalog:
```{.sh}
curl -X PUT "http://localhost:8675/rename?catalog=/foo/secrets.txt&file=c.txt&new=c_2.txt"
{
  "ryftone-310": {
    "c.txt": "OK"
  },
  "ryftone-313": {
    "c.txt": "OK"
  }
}
```

#### Change file name in `cluster` mode with `local=true` parameter:
```{.sh}
 curl -X PUT "http://localhost:8675/rename?file=/secrets/a.txt&new=/secrets2/a2.txt&local=true"
{
  "/secrets/a.txt": "OK"
}
```

#### Change file name using URL-path
```{.sh}
curl -X PUT "http://localhost:8675/rename/secrets3?file=c.txt&new=c2.txt"
{
  "ryftone-310": {
    "/secrets3/c.txt": "OK"
  },
  "ryftone-313": {
    "/secrets3/c.txt": "OK"
  }
}
```