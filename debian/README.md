## How to make ``.deb`` package

Build ``ryft-server``:

```
cd ryft-server
go build
```

Build ``.deb`` file:

```
cd ryft-server
make debian
```

or with version:

```
cd ryft-server
make debian VERSION=1.2.3
```

Note, `make debian` should be run in the project's root directory. In this case ryft-server is rebuilt
and corresponding deb package is created. If you run `make` from *debian* subdirectory then only
deb package is created - ryft-server is not rebuilt, just used from your `$GOPATH/bin`.


## How to install ``.deb`` package

```
sudo dpkg -i ryft-server-0.1_x86_64.deb
```

It will install and start ``ryft-server-d`` service.


## How to uninstall ``.deb`` package

```
sudo dpkg -r ryft-server
```

## Start & stop ``ryft-server-d`` service

```
sudo service ryft-server-d start
sudo service ryft-server-d stop
```

## Setting arguments

You can pass arguments to ryft-server daemon by creating `/etc/ryft-rest.conf` file and restarting ryft-server service.
Full list of arguments is listed by running `ryft-server --help`. Write arguments each on a separate line. For example:

```
0.0.0.0:9000
--keep
--debug
```

Daemon will start as follows: `ryft-server @/etc/ryft-rest.conf`

## Log file

You can find log file of the service that called `ryft-server-d-start.log` inside home directory.

To view logs in real-time enter following command

`tail -f $HOME/ryft-server-d-start.log`



TODO: server configuration file description! search engine and options
