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

#### Change file name:

```{.sh}
```

#### Change directory name:

```{.sh}
```

#### Change catalog name:
```{.sh}
```

#### Change file name inside a catalog:
```{.sh}
```

#### Empty request
```{.sh}
```