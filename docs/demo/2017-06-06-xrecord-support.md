# Demo - support for XRECORD and CRECORD - June 6, 2017

On AWS F1 instances the `ryftprim` supports new search input keywords:
`XRECORD` and `CRECORD` for XML and CSV data respectivelly.

New `ryft-server` query parser now supports these keywords for record search operations.
`XRECORD` and `CRECORD` can be used in the same way as `RECORD` does.

`(XRECORD.bla.bla.bla CONTAINS "hello")`


## `RECORD` to `XRECORD`/`CRECORD` automatic adjustment

For the AWS F1 instances it is possible to replace `RECORD` to `XRECORD`
or `CRECORD` keyword based on the content of input file. So all the
clients can still use `RECORD` as usual and `ryft-server` will
automatically adjust it.

If query parser gets `RECORD` keyword as an input in search query it detects
type of requested file in the following steps:
- the extension of file is checked first
- the content of file is checked then

It is recommended to specify adjustment rules based on extension because
this method is much faster.


### XML file detection

The file extension is checked first. If it is related to the list of known XML
extensions the file will be used as XML.

If no extension is matched then the content of file is checked.
The XML file should start with `<?xml` pattern.

### CSV file detection

The same for CSV: file extension is checked first then the content of file.
The CSV file should contain at least two lines (each at least two rows) of valid CSV data.


### Adjustment configuration

The list of known XML and CSV file extensions can be customized via configuration
file. There is a dedicated section in `ryft-server.conf` [configuration file](../run.md#record-queries-configuration)
but this configuration might be [overriden](#ryft-user-configuration).

```{.yaml}
record-queries:
  enabled: true                    # true - replace RECORD with XRECORD or CRECORD based on input data
  skip: ["*.txt"]                  # file patterns to skip
  json: ["*.json"]                 # file patterns for JSON data
  xml: ["*.xml"]                   # file patterns for XML data
  csv: ["*.csv"]                   # file patterns for CSV data
```

Of course we can disable this `RECORD` adjustment feature by single `record-queries.enabled` flag.
If this falg is `false` (`false` is used by default) the query will be used "as is".
If flag is `true` then adjustment operation is in action.

The query will be used "as is" also when the file extension is matched to one of
`record-queries.skip` or `record-queries.json` pattern (there is no dedicated
`JRECORD` keyword for JSON data).

Note, it is also important to provide valid set of `record-queries.skip` or `record-queries.json` patterns.
If file extension is not mached, each input file will be opened a few times to check file content.
And this usually will impact overall performance in not a good way.


## Ryft user configuration

There are a few ways to customize `record-queries` configuration section.
First there is `default-user-config` section in ryft server [configuration file](../run.md#ryft-user-configuration).
This configuration is used if no custom Ryft user configuration is provided.

```{.yaml}
default-user-config:
  record-queries:
    enabled: true                    # true - replace RECORD with XRECORD or CRECORD based on input data
    skip: ["*.txt"]                  # file patterns to skip
    json: ["*.json"]                 # file patterns for JSON data
    xml: ["*.xml"]                   # file patterns for XML data
    csv: ["*.csv"]                   # file patterns for CSV data
```

If there is `${RYFTHOME}/.ryft-user.yaml` YAML file located in Ryft user's home
directory then user configuration will be loaded from this file. Using this
file user is able to customize the list extensions related to XML or CSV data
by himself without modifying server's configuration file. Usual POST /files
can be used.

```{.yaml}
record-queries:
  enabled: false                   # true - replace RECORD with XRECORD or CRECORD based on input data
  skip: ["*.txt"]                  # file patterns to skip
  json: ["*.json"]                 # file patterns for JSON data
  xml: ["*.xml"]                   # file patterns for XML data
  csv: ["*.csv"]                   # file patterns for CSV data
```

Another option is JSON file `${RYFTHOME}/.ryft-user.json`. This file is used
in second (if no YAML file exists).

```{.json}
{
  "record-queries": {
    "enabled": true,
    "skip": ["*.txt"],
    "json": ["*.json"],
    "xml": ["*.xml"],
    "csv": ["*.csv"]
  }
}
```
