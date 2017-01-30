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
./ryft-server -l=:9000
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

There is also no any load balancing enabled in local mode.


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
    core/catalogs: debug
    core/pending-jobs: debug
    core/busyness: debug
    search/ryftprim: debug
    search/ryftone: debug
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
logging levels. By default all loggers have "info" level. It is very easy to
create any logging configuration with fine-tunes logging levels.


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
- `ryftone` uses *libryftone* library to access Ryft hardware
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
```

`local-only` is used to run `ryft-server` outside cluster. No consult dependency,
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

Signing algorithm can be customized via `algorithm` option. It can be `RS256` for
example.

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