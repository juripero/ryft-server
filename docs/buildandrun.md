This document contains information about how to build and run `ryft-server`.

# Cloning & Building

> The instructions below assume you have a properly configured GO dev environment with GOPATH and GOROOT env variables configured.
> If you start from scratch we recommend to use this [automated installer](https://github.com/demon-xxi/tools).

> To use `go get` command with private repositories use the following setting to force SSH protocol instead of HTTPS:
> `git config --global url."git@github.com:".insteadOf "https://github.com/"`
> Make sure you have configured [SSH token authentication](https://help.github.com/articles/generating-an-ssh-key/) for GitHub.

To clone `ryft-server` source files use the following commands (downloading may take a while because of dependencies):

```{.sh}
go get -d -v github.com/getryft/ryft-server
cd $GOPATH/src/github.com/getryft/ryft-server
```

All commands below assume you are in `$GOPATH/src/github.com/getryft/ryft-server` directory.
To build server `make` tool is used, just run:

```{.sh}
make
```

The `ryft-server` executable will be created. You can run it without any arguments.
`ryft-server` listens on `8765` port by default.

To build `ryft-server` from another (say experimental) branch use combination of the following commands:

```{.sh}
git checkout <branch-name>
go get -d -v
make
```

Also it's possible to build Debian package:

```{.sh}
make debian
```

This command builds `ryft-servers` and packs everything into Debian package which
can be found under `debian/` subdirectory.

Note, `make debian` should be run in the project's root directory. In this case `ryft-server` is rebuilt
and corresponding deb package is created. If you run `make` from `debian/` subdirectory then only
deb package is created - `ryft-server` is not rebuilt, just used from your `$GOPATH/bin`.


## Version

Note the `ryft-server` version - it's automatically generated using current git commit.
Simple `make` will produce the following output (exact numbers may differ):

```{.sh}
$ make
go build -ldflags "-X main.Version=0.6.1-139-g51fcf47 -X main.GitHash=51fcf47f0de217b0dfba4c4e2ed83ed172e123ae"
```

where `main.Version` is the server's version and `main.GitHash` is corresponding git commit hash.

By default version variable is generated using [git describe](https://git-scm.com/docs/git-describe) tool.
But it's possible to override automatically generated version:

```{.sh}
$ make VERSION=1.2.3
go build -ldflags "-X main.Version=1.2.3 -X main.GitHash=51fcf47f0de217b0dfba4c4e2ed83ed172e123ae"
```


## Build tags

On some development hosts there is no `libryftone` installed by default. In this case we can build
`ryft-server` without `ryftone` search engine support (only `ryftprim` will be available):

```{.sh}
make GO_TAGS=noryftone
```

Note, for now `GO_TAGS=noryftone` tag is defined automatically by the `Makefile`.
To disable this behaviour just pass empty tag:

```{.sh}
make GO_TAGS=
```


# Running

Running server is quite simple assuming `$GOPATH/src/github.com/getryft/ryft-server` is current directory:

```{.sh}
./ryft-server
```

This command runs server with default arguments. Port `8765` will be used for listening by default.

To run another server instance on `9000` port just pass "address" argument:

```{.sh}
./ryft-server 0.0.0.0:9000
```

So it's possible to run multiple server instances on the same machine.

Use `ryft-server --help` to get list of all supported arguments.


## Debug mode

To get detailed log messages debug mode can be used:

```{.sh}
./ryft-server --debug
```

In this mode `ryft-server` prints a lot of log messages.
This feature is useful for troubleshooting.


## Keeping search results

`ryft-server` uses dedicated instance directory on `/ryftone` volume to keep temporary files.
The name of instance directory is generated using port number `/ryftone/RyftServer-$PORT/`
but can be customized via search configuration file (see below).

By default `ryft-server` removes all search results from `/ryftone/RyftServer-$PORT/`.
But it behaviour may be prevented with `--keep` flag:

```{.sh}
./ryft-server --keep
```

All temporary result files will be kept under server's instance directory.
This feature is useful for troubleshooting.


## Search configuration

`ryft-server` supports additional configuration file related to search.
This YAML configuration file can be customized with `--config` flag:

```{.sh}
./ryft-server --config=$path_to_yaml_config_file
```

Using search configuration file it's possible to change the main search engine
and its options. The file format is the following:

```{.yaml}
searchBackend: <search engine>
backendOptions:
  <search engine options>
```

`searchBackend` is the search engine name and can be one of the following:

- `ryftprim` uses *ryftprim* command line tool to access Ryft hardware (is used by default)
- `ryftone` uses *libryftone* library to access Ryft hardware
- `ryfthttp` uses another `ryft-server` instance to access Ryft hardware

`backendOptions` is search engine specific options. For example `ryftprim` engine
supports the following options:

```{.yaml}
searchBackend: ryftprim
backendOptions:
  instance-name: .ryft/8765   # server instance name (RyftServer-$PORT by default)
  ryftprim-exec: ryftprim     # ryftprim tool path (/usr/bin/ryftprim by default)
  ryftone-mount: /ryftone     # ryftone volume (/ryftone by default)
                              # server instance directory will be: $ryftone-mount/$instance-name
```

More information about search engines can be found [here](./search.md)


## Authentication and security

**TBD** need to add security **TBD**

The following flags are supported:

```
  -a, --auth=AUTH  Authentication type: none, file, ldap.
  --users-file=USERS-FILE
                   File with user credentials. Required for --auth=file.
  --ldap-server=LDAP-SERVER
                   LDAP Server address:port. Required for --auth=ldap.
  --ldap-user=LDAP-USER
                   LDAP username for binding. Required for --auth=ldap.
  --ldap-pass=LDAP-PASS
                   LDAP password for binding. Required for --auth=ldap.
  --ldap-query="(&(uid=%s))"
                   LDAP user lookup query. Defauls is '(&(uid=%s))'. Required for --auth=ldap.
  --ldap-basedn=LDAP-BASEDN
                   LDAP BaseDN for lookups.'. Required for --auth=ldap.

  -t, --tls          
                    Enable TLS/SSL. Default 'false'.
  --tls-crt=TLS-CRT  
                    Certificate file. Required for --tls=true.
  --tls-key=TLS-KEY  
                    Key-file. Required for --tls=true.
  --tls-address=0.0.0.0:8766  
                     Address:port to listen on HTTPS. Default is 0.0.0.0:8766
```


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


## Setting arguments

You can pass arguments to `ryft-server` daemon by creating `/etc/ryft-rest.conf` file and restarting `ryft-server-d` service.
Write flags and arguments (see above) each on a separate line. For example:

```
0.0.0.0:9000
--debug
--keep
```

Daemon will start as follows:

```{.sh}
ryft-server @/etc/ryft-rest.conf
```


## Log file

You can find log file of the `ryft-server-d` service that called `ryft-server-d-start.log` inside home directory of `ryftuser`.

To view logs in real-time:

```{.sh}
tail -f ~/ryft-server-d-start.log
```
