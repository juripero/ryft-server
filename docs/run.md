This document contains information about how to run `ryft-server`.


# Running

Running server is quite simple assuming `$GOPATH/src/github.com/getryft/ryft-server` is current directory:

```{.sh}
./ryft-server
```

This command runs server with default arguments. Port `8765` will be used for listening by default.

To run another server instance on `9000` port just pass "address" argument:

```{.sh}
./ryft-server --address=0.0.0.0:9000
# or
./ryft-server -l:9000
```

So it's possible to run multiple server instances on the same machine.

Use `ryft-server --help` to get list of all supported arguments.


## Debug and local modes

To get detailed log messages debug mode can be used:

```{.sh}
./ryft-server --debug
```

In this mode `ryft-server` prints a lot of log messages.
This feature is useful for troubleshooting.

If consul service is not running it's possible to run `ryft-server` in local mode.
In this mode all requestes will be performed locally even `local=false` is set.
Just pass `--local-only` command line argument:

```{.sh}
./ryft-server --local-only
```

There is also no load balancing enabled in local mode.


## Logging

There is special parameter in configuration file `logging` and related section
`logging-options`. The `logging-options` contains a sets of logging configurations.
Actually `logging` is a key in that set.

```{.yaml}
logging: debug
logging-options:
  custom:
    core: debug
    core/catalogs: debug
    core/pending-jobs: debug
  debug:
    core: debug
    core/safe: debug
    core/catalogs: debug
    core/pending-jobs: debug
    core/busyness: debug
    search/ryftprim: debug
    search/ryfthttp: debug
    search/ryftmux: debug
    search/ryftdec: debug
  release:
    core: info
```

Changing the `logging` option it is possible to quickly change the full logging
configuration. There is special command line argument `--logging` which also
could be used to change logging configuration.

The logging configuration itself consists of logger names and corresponding
logging levels. By default all loggers have "info" level. The possible logging
level values are: "panic", "fatal", "error", "warn" or "warning", "info" and
"debug". It is very easy to create any logging configuration with fine-tunes logging levels.


## Keeping search results

`ryft-server` uses dedicated instance directory on `/ryftone` volume to keep temporary files.
The name of instance directory is generated using port number `/ryftone/.rest-$PORT/`
but can be customized via search configuration file (see below).

By default `ryft-server` removes all search results from `/ryftone/.rest-$PORT/`.
But it behaviour may be prevented with `--keep` flag:

```{.sh}
./ryft-server --keep
```

All temporary result files will be kept under server's instance directory.
This feature is useful for troubleshooting.


## Configuration file

`ryft-server` supports additional configuration file.
This YAML configuration file can be customized with `--config` command line option:

```{.sh}
./ryft-server --config=$path_to_yaml_config_file
```

Moreover, if configuration file is located at `/etc/ryft-server.conf`
it will be automatically used by init script when the service starts.
There is default configuration provided by debian package.

### Search configuration

Using configuration file it's possible to change the main search engine
and its options. The file format is the following:

```{.yaml}
search-backend: <search engine>
backend-options:
  <search engine options>
```

`search-backend` is the search engine name and can be one of the following:

- `ryftprim` uses *ryftprim* command line tool to access Ryft hardware (is used by default)
- `ryfthttp` uses another `ryft-server` instance to access Ryft hardware

`backend-options` is search engine specific options. For example `ryftprim` engine
supports the following options:

```{.yaml}
search-backend: ryftprim
backend-options:
  instance-name: .ryft/8765   # server instance name (.rest-$PORT by default)
  ryftprim-exec: ryftprim     # ryftprim tool path (/usr/bin/ryftprim by default)
  ryftone-mount: /ryftone     # ryftone volume (/ryftone by default)
                              # server instance directory will be: $ryftone-mount/$instance-name
```

More information about search engines can be found [here](./search/engine.md)

#### Query parser

A few options are used by query parser:

```{.yaml}
backend-options:
  ...
### query decomposition:
  compat-mode: false       # true - compatibility mode, false - generic
  optimizer-limit: 1000
  optimizer-do-not-combine: feds
```

`compat-mode` flag is used to switch REST server into "compatibility" mode.
Old search query syntax is used in this mode instead of "generic" syntax.
This mode is used to run REST server on old firmware without "generic" syntax.

`optimizer-limit` is the maximum number of sub-queries that can be combined.
By defeault there is no such limit, i.e. `optimizer-limit=-1` means "combine all".
Zero value `optimizer-limit=0` means "do not combine at all".

`optimizer-do-not-combine` is coma-separated list of search modes that
should not be combined. Usually `FEDS` cannot be combined. Multiple
modes can be specified: `optimizer-do-not-combine=feds,fhs`.


#### Aggregation configuration

This section contains parameters related to aggregation processing:

```{.yaml}
backend-options:
  ...
  aggregations:
    optimized-tool: /usr/bin/ryft-server-aggs  # path to optimized tool (comment to disable)
    max-records-per-chunk: 16M       # maximum number of records per DATA chunk
    data-chunk-size: 1GB             # maximum DATA chunk size
    index-chunk-size: 1GB            # maximum INDEX chunk size
    concurrency: 8                   # number of parallel threads to calculate aggregation on
    engine: auto                     # aggregation engine, one of: auto, native, optimized
```

The `optimized-tool` option specifies the absolute path to the optimized tool
which can be used to calculate some aggregations. This tool is written in C
and uses no dynamic memory allocation (except for a few processing buffers)
to process large data files. The data processing is performed on chunk-by-chunk
basis with multiple threads.

The `max-records-per-chunk` option specifies the maximum number of records per
DATA processing chunk. The `G`, `M` and `K` suffices can be used to specify `1024`
multipliers.

The `data-chunk-size` and `index-chunk-size` options specify the size of DATA and
INDEX chunks in bytes. The `GB`, `MB` and `KB` suffices can be used to specify
`1024` multipliers.

The `concurrency` option specifies how many processing threads should be used
to calculate aggregations. Recommended value is the number of CPU.

The `engine` option specifies the default processing engine. Can be one of:
- `auto` automatically select appropriate aggregation engine
- `native` use native Go-implemented aggregation engine
- `optimized` use optimized C-based aggregation engine. Note, `optimized` engine doesn't support all the aggregation functions.

Some of these options can be overriden via corresponding request's
[tweaks](./rest/aggs.md#aggregation-processing-customization).


### Server configuration

This configuration file also contains most of the command line options
that also can be customized. Note these options may be overriden by command line.
For example if your configuration file contains:

```{.yaml}
address: :8000
```

but the server starts as:

```{.sh}
ryft-server --config=/etc/ryft-server.conf --address=:9000
```

the actual option for the address will be `:9000` since it goes last.

Using `address` option its possible to customize server's listen address.
It's equivalent to the `--address` command line option.
By default "0.0.0.0:8765" is used.

There also a few more options:

```{.yaml}
local-only: false
debug-mode: false
keep-results: false
busyness-tolerance: 0
http-timeout: 1h
processing-threads: 8
settings-path: /var/ryft/server.settings
# hostname: node-1
# instance-home: /
```

`local-only` is used to run `ryft-server` outside cluster. No consul dependency,
no load balancing enabled. It's equivalent to `--local-only` command line option.

`debug-mode` is used to enable extensive logging.
It's equivalent to `--debug` command line option.

`keep-results` is used to keep intermediate INDEX and DATA files for debugging.
It's equivalent to `--keep` command line option.

`busyness-tolerance` is used in cluster mode to customize node grouping algorithm.
See [cluster document](./cluster.md#busyness) for more details.
It's equivalent to `--busyness-tolerance` command line option.

`http-timeout` is used as read request/write response timeout for HTTP/HTTPS connections.
It's `1h` (one hour) by default.

`processing-threads` is the number of parallel threads used to handle all requests.
If zero the default system value is used. This value is used internally by Go runtime.
See [GOMAXPROCS](https://golang.org/pkg/runtime/#GOMAXPROCS) for more details.

`settings-path` is used to specify local `ryft-server` storage.

`hostname` is used to customize the hostname provided by the `ryft-server`.
By default the system's hostname is used.

`instance-home` is used to specify common home directory which is prefixed to
user's home directory or is used by default if authentication is disabled.


#### TLS server configuration

It's possible to run server with HTTPS enabled. The `--tls`, `--tls-address`,
`--tls-cert`, `--tls-key` command line options can be used or just corresponding
`tls` section in configuration file:

```{.yaml}
tls:
  enabled: false
  address: 0.0.0.0:8766
  cert-file: "<certificate file name>"
  key-file: "<key file name>"
```

The HTTPS can be enabled or disabled with boolean `enabled` flag. The listen
address can be customized via `address` option. Note the port number should be
different to the normal address.

There are two files should be provided to enabled HTTPS: certificate file `cert-file`
and corresponding certificate key `key-file`.

#### Authentication server configuration

There are a few sections related to [authentication](./auth.md).

```{.yaml}
auth-type: none
```

`auth-type` is used to select authentication provider. It can be one of:
  - `none`
  - `file`
  - `ldap`

`auth-type: none` is used to disable authentication.

If authentication is enabled, i.e. `auth-type: file` or `auth-type: ldap`,
the `auth-jwt` sections should be provided:

```{.yaml}
auth-jwt:
  algorithm: HS256
  secret: "<secret key>"
  lifetime: 2h
```

Signing algorithm can be customized via `algorithm` option. It can be `RS256`
or `HS512` for example.

Secret can be simple string `secret: "my super secret key"` or file reference
`secret: "@~/.ssh/id_rsa"` to use `~/.ssh/id_rsa` file content as a secret.

JWT token lifetime can be customized via `lifetime` option. By default it's
`lifetime: 1h`.

If simple file is used as an authentication provider `auth-type: file`,
the users credentials should be provided via:

```{.yaml}
auth-file:
  users-file: /etc/ryft-users.yaml
```

The file formats are described [here](./auth.md#simple-text-file).

If LDAP is used as an authentication provider `auth-type: ldap`, the
LDAP credentials should be provided via:

```{.yaml}
auth-ldap:
  server: ldap.forumsys.com:389
  username: "read-only-admin,dc=ryft,dc=one"
  password: "<password>"
  query: "(&(cn=%s))"
  basedn: "dc=ryft,dc=one"
#  insecure-skip-tls: true
#  insecure-skip-verify: true
```

The LDAP server address can be customized via `server` option.

The read-only user which is used to send search request to the LDAP service
can be customized via `username` and `password` options.

`query` and `basedn` can be used to specify attribute name which is used to
search and base DN.

There also a few options related to security. By default `ryft-server` tries to
connect LDAP using TLS. To disable TLS just set `insecure-skip-tls: true`.
To disable certificate verification (may be useful if LDAP uses self-signed
certificate) just set `insecure-skip-verify: true`. It is not recommended to
define these `insecure-*` options in production.

See [authentication](./auth.md) document for more details.


### Consul configuration

There is Consul-related configuration section:

```{.yaml}
consul:
  address: http://127.0.0.1:8500
  data-center: dc1
```

The `consul.address` specifies the remote consul address including schema and
port number. By default it is `http://127.0.0.1:8500`.

The `consul.datacenter` specifies the consul's data center name.
By default it is `dc1`.

Note, that consul address also can be specified via environment variable
`CONSUL_HTTP_ADDR=127.0.0.1:8500`.


### Catalog configuration

Some catalog related options can be customized via the following configuration
section:

```{.yaml}
catalogs:
  max-data-file-size: 64MB       # data file size limit: KB, MB, GB, TB
  cache-drop-timeout: 10s        # internal cache lifetime
  default-data-delim: "\n\f\n"   # default data delimiter
  temp-dir: /tmp/ryft/catalogs   # for temporary files
```

It's possible to customize catalog data size limit via `max-data-file-size`
option. If there is no more space in current catalog's data file, then new one
will be started. It's possible to use various units, for example `MB` for
megabytes (1024*1024) and `GB` for gigabytes (1024*1024*1024).

There is an internal catalog cache. Each catalog entry has it's own drop timeout
or lifetime. By default it's 10 seconds but can be changed via
`cache-drop-timeout` option. There is also possible to use various units,
for example `h` for hours or `ms` for milliseconds.

Data delimiter is used to separate different small files inside a bigger data file.
If delimiter is non empty, it will be placed each time a new file part is written
to catalog. The main purpose of this delimiter is to separate RAW text to avoid
possible collisions on the file boundaries. For structured data the data delimiter
is not so important. Anyway it can be customized via `default-data-delim` option.

Sometimes catalog need to save file content into temporary file. These
temporary files are placed in `temp-dir` directory.


### Script transformation configuration

The [script transformation](./rest/README.md#script-transformation) calls
external application or script. The list of trusted applications is configured
via `post-processing-scripts` configuration section.

```{.yaml}
post-processing-scripts:
  false:
    path: [/bin/false]
  cat:
    path: [/bin/cat]
  my_test1:
    path: [/usr/bin/jq, -c, "{lat: .lat, lon: .lon}"]
```

Each item is a script name and a `path` containing full path to the
application/script and a set of additional command line options.


### Docker configuration

There is Docker-related configuration section:

```{.yaml}
docker:
  run: ["/usr/bin/docker", "run", "--rm", "--network=none", "--volume=${RYFTHOME}:/ryftone", "--workdir=/ryftone"]
  exec: ["/usr/bin/docker", "exec", "${CONTAINER}"]
  images:
    default: ["alpine:latest"]
    alpine: ["alpine:latest"]
    ubuntu: ["ubuntu:16.04"]
    python: ["python:2.7"]
```

The list of `images` is used to restrict number of Docker images allowed.
These images should be pulled from the Docker hub with `docker pull <image>` command.

The `run` command is used to run custom command in a Docker container.

The `exec` command is used to run custom command in an already running  Docker container.

There are a few "environment" variables can be used:
- `${RYFTONE}` path to `/ryftone` partition
- `${RYFTHOME}` path to Ryft user's home directory: `/ryftone/user`
- `${RYFTUSER}` name of authenticated Ryft user
- `${CONTAINER}` Docker container identifier


### Session configuration

This configuration section customizes the session related options:

```{.yaml}
sessions:
  signing-algorithm: HS256
  secret: session-secret-key
```

The JWT token `signing-algorithm` can be one of: `HS256`, `HS384`, `HS512`,
`RS256`, `RS384`, `RS512`.

Secret can be simple string `secret: "my super secret key"` or file reference
`secret: "@~/.ssh/id_rsa"` to use `~/.ssh/id_rsa` file content as a secret.


### Ryft user configuration

Some parameters of ryft-server might be customized via Ryft user configuration file.
Every user can change some part of ryft-server behaviour uploading special `YAML`
or `JSON` configuration file:

```{.sh}
curl -s "http://ryft-host:8765/files?file=.ryft-user.yaml&offset=0" \
     -H 'Content-Type: application/octet-stream' --data \
'record-queries:
  enabled: true
  skip: ["*.txt", "*.dat"]
  json: ["*.json"]
  xml: ["*.xml"]
  csv: ["*.csv"]
'
```

By default the `default-user-config` section from main configuration file used.
But if the `/ryftone/${RYFTUSER}/.ryft-user.yaml` or `/ryftone/${RYFTUSER}/.ryft-user.json`
file is present, then it will be used instead of `default-user-config` section.

Please note, if you change parameters in main configuration file and nothing is happened
then probably there is Ryft user configuration file which overrides all parameters
from `default-user-config`.

The following parameters can be customized:
- [automatic `RECORD` replacement](#record-queries-configuration)


#### Record queries configuration

The `record-queries` subsection contains all the parameters related to
automatic `RECORD` [replacement](./search/README.md#automatic-record-replacement).

```{.yaml}
record-queries:
  enabled: true
  skip: ["*.txt", "*.dat"]
  json: ["*.json"]
  xml: ["*.xml"]
  csv: ["*.csv"]
```

This feature can be disabled by `enabled: false` option. Also this feature
can be enabled for particular backend tools, for example:

```{.yaml}
record-queries:
  enabled:
  - ryftx
  - ryftpcre2

# ... or ...

record-queries:
  enabled:
    default: true
    ryftprim: false
```

The following lists of file patterns customize extension-based file type detection:
- `skip` ignore these extensions. The `RECORD` will be kept as is.
- `json` extensions for `JSON` data. The `RECORD` will be replaced with `JRECORD`.
- `xml` extensions for `XML` data. The `RECORD` will be replaced with `XRECORD`.
- `csv` extensions for `CSV` data. The `RECORD` will be replaced with `CRECORD`.

Note, the file pattern may include directory filter, the `json: ["foo/*.json"]`
matches all JSON files in `foo` directory.

There is special case for the XML data. The XML file should be in valid format,
i.e. it should contain `<?xml ... ?>` header and a root element. For example:

```{.xml}
<?xml version = "1.0" encoding = "UTF-8" ?>
<xml_root>
  <r><c01>
    ...
```

In this case the `RECORD.c01` will be automatically replaced to `XRECORD.r.c01`.


# Debian package

Having Debian package `ryft-server-$version.deb` it's possible to install it to any compatible machine:

```{.sh}
sudo dpkg -i ryft-server-$version.deb
```

This command does all the work: installs `ryft-server` and automatically starts `ryft-server-d` service.
This service can be stopped later:

```{.sh}
sudo service ryft-server-d stop
```

and started again:

```{.sh}
sudo service ryft-server-d start
```

To uninstall debian package use:

```{.sh}
sudo dpkg -r ryft-server
```


## Log file

You can find log file of the `ryft-server-d` service at `/var/log/ryft/server.log`.

To view logs in real-time:

```{.sh}
tail -f /var/log/ryft/server.log
```


## Default backend options

There described new section in the `ryft-server.conf` that allow user to tune backend options in order to achieve better performance.

Part of config:
```{.yaml}
backend-options:
  backend-tweaks:
    options:
      ...
    router:
      ...
    abs-path:
      ...
```


### `router` section

If presented options define mapping between search primitive and backend tool.

e.g.:
```{.yaml}
backend-options:
  backend-tweaks:
    router:
      default: ryftx
      es: ryftprim
      fhs,ds: ryftprim
      prce2: prce2
```

`/search` and `/count` endpoints accept `backend` parameter, but if it is not set explicetly `router` may be used for choosing backend that fits better for current search primitive. If search primitive is ommited in the `router` table value of the `default` key will be used.

### options

Options defined in `options` section should have an `array` format and will be passed to the backend tool within a search query.
Parameters set in `backend-options` have more priority than `options` though.

User can set options for a backend tool, for a search primitive and for a combination `[backend].[search primitive]`.

User can also create set of options specifying `backend-mode` parameter. In this case key of `options` table has a structure: `[backend-mode].[backend].[search primitive]`.
e.g.:
```{.yaml}
backend-options:
  backend-tweaks:
    options:
      default: []             # default mode
      default.ryftx.es: []
      default.ryftx.ds: []
      default.ryftx: []
      hp: []                  # high-performance mode
      hp.ryftx.es: []
      hp.ryftprim.es: []
      hp.ryftprim: []
```
The rule of thumb is: the more preciesly you specify options the higher priority they have.
User can set `backend` and `backend-mode` parameters in `/search` and `count` endpoints.

Search order in config defined above:
- extract `backend-mode` from the request parameters
- extract `backend` tool from the request parameters or from the `router` table
- extract `search primitive` from the search query
- create `options` key using pattern `[backend-mode].[backend].[search primitive]` and search it in `options`.
If nothing found in `options` try `[backend].[search primitive]`, then `[search primitive]`, `[backend]`, `[backend mode]`. Finally, if nothing found use value for the `default` key.
- execute `backend` tool with a `query` and found options. E.g. `/usr/bin/ryftprim [query] ... [backend-options]`


### absolute path

This section describes the usage of absolute/relative path for the backend tool.
Only `ryftprim` tool accepts path relative to `/ryftone` patrition. The
`ryftx` and `ryftpcre2` tools accept absolute path.

If no tool is specified the `default` flag is used.

```{.yaml}
backend-options:
  backend-tweaks:
    abs-path:
      default: false
      ryftprim: false
      ryftx: true
      ryftpcre2: true
```

Other formats are also supported:

```{.yaml}
backend-options:
  backend-tweaks:
    abs-path:
    - ryftx
    - ryftpcre2
```

```{.yaml}
backend-options:
  backend-tweaks:
    abs-path: [ ryftx, ryftpcre2 ]
```
