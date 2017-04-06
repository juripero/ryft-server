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


### Rename in `local-only` mode

We start ryft-server in local-only mode. It equals to execution queries with `local=true` URL parameter.

```{.sh}
./ryft-server --config=/etc/ryft-server.conf --local-only
```

#### Response format
Response usually looks like 

Status code: `200`
```{.json}
{
    "/origin/filename": "OK"
}
```
Status code: `40x`
```{.json}
{
    "/origin/filename": "bad request error message"
}
```

If something goes wrong with the server response will be 

Status code: `50x`
```{.json}
"server error message"
```


### Examples:

#### Change file name:

Here we change filename from `a.txt` to `a2.txt` and move it from `secrets` directory into `secrets2`.
`secrets2` directory doesn't exist and it will be created during request.

```{.sh}
curl -s -X PUT "http://localhost:8675/rename?file=/secrets/a.txt&new=/secrets2/a2.txt" | jq .
{
  "/secrets/a.txt": "OK"
}
```

We can also set path to the directory that contains the file directly in the URL path string. 
```{.sh}
curl -s -X PUT "http://localhost:8675/rename/secrets?file=a.txt&new=a2.txt" | jq .
{
  "/secrets/a.txt": "OK"
}
```

#### Change directory name:
Then we rename directory `foo` into `foo2`. 

```{.sh}
curl -s -X PUT  "http://localhost:8675/rename?dir=/foo&new=/foo2" | jq .
{
  "/foo": "OK"
}
```

#### Change catalog name:
Rename catalog from `/foo/secrets.txt` to `/foo/secrets2.txt`. 

```{.sh}
curl -s -X PUT "http://localhost:8675/rename?catalog=/foo/secrets.txt&new=/foo/secrets2.txt" | jq .
{
  "/foo/secrets.txt": "OK"
}
```

#### Change file name inside a catalog:
Now we rename file `c.txt` to `c2.txt` that lays inside `/foo/secrets.txt` catalog.
Here we update records in SQL database and catalog-directory. If SQL query fails transaction will not be commited.
If something happens with the filesystem and we can't rename directory we don't try to rollback this operation somehow. 

Show files inside catalog
```
curl -X GET "http://localhost:8675/files?catalog=/foo/secrets2.txt" | jq .
{
  "catalog": "/foo/secrets2.txt",
  "files": [
    "c.txt",
    "d.txt"
  ],
  "details": {
    "ryftone-310": {
      "c.txt": {
        "type": "file",
        "length": 18
      },
      "d.txt": {
        "type": "file",
        "length": 18
      }
    },
    "ryftone-313": {
      "c.txt": {
        "type": "file",
        "length": 18
      },
      "d.txt": {
        "type": "file",
        "length": 18
      }
    }
  }
}
```

Then rename file
```{.sh}
curl -s -X PUT "http://localhost:8675/rename?catalog=/foo/secrets2.txt&file=c.txt&new=c2.txt" | jq .
{
  "c.txt": "OK"
}
```

Check files
```
curl -X GET "http://localhost:8675/files?catalog=/foo/secrets2.txt" | jq .
{
  "catalog": "/foo/secrets2.txt",
  "files": [
    "c2.txt",
    "d.txt"
  ],
  "details": {
    "ryftone-310": {
      "c2.txt": {
        "type": "file",
        "length": 18
      },
      "d.txt": {
        "type": "file",
        "length": 18
      }
    },
    "ryftone-313": {
      "c2.txt": {
        "type": "file",
        "length": 18
      },
      "d.txt": {
        "type": "file",
        "length": 18
      }
    }
  }
}
```

### Rename in a `cluster` mode
#### Response format
Response usually looks like 
Status code: `200`
```{.json}
{
    "server": {
        "/origin/filename": "OK"
    }
}
```
Status code: `40x`
```{.json}
{
    "server": {
        "/origin/filename": "bad request error message"
    }
}
```

If something goes wrong with the server response will be 

Status code: `50x`
```{.json}
"server error message"
```


It is possible to get response like. It means one node can be corrupted, but we still have `200 OK` status.
```{.json}
{
  "ryftone-310": {
    "/secrets/a.txt": "OK"
  },
  "ryftone-313": {
    "/secrets/a.txt": "no such file or directory"
  }
}

```

### Examples:
#### Change file name:

Change name from `a.txt` to `a2.txt` and move file from `secrets` to `secrets2`

```{.sh}
curl -s -X PUT "http://localhost:8675/rename?file=/secrets/a.txt&new=/secrets2/a2.txt" | jq .
{
  "ryftone-310": {
    "/secrets/a.txt": "OK"
  },
  "ryftone-313": {
    "/secrets/a.txt": "OK"
  }
}
```

#### Change directory name:

Rename directory from `foo` to `foo2`

```{.sh}
curl -s -X PUT  "http://localhost:8675/rename?dir=/foo&new=/foo2" | jq .
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

Rename catalog from `/foo/secrets.txt` to `/foo/secrets2.txt`

```{.sh}
curl -s -X PUT "http://localhost:8675/rename?catalog=/foo/secrets.txt&new=/foo/secrets2.txt" | jq .
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

Rename file `c.txt` to `c2.txt` inside catalog `/foo/secrets.txt`
```{.sh}
curl -s -X PUT "http://localhost:8675/rename?catalog=/foo/secrets.txt&file=c.txt&new=c2.txt" | jq .
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

Here we execute rename method just for one node and response will be the same as when we run server in `local-only` mode

```{.sh}
 curl -s -X PUT "http://localhost:8675/rename?file=/secrets/a.txt&new=/secrets2/a2.txt&local=true" | jq .
{
  "/secrets/a.txt": "OK"
}
```