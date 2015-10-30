## How to make ``.deb`` package

Build ``ryft-server``:

```
cd ryft-server
go build
```

Build ``.deb`` file:

```
./ryft-server-make-deb
```
or with version
```
./ryft-server-make-deb 0.2
```


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

## Setting parameters

You can pass parameters to ryft-server daemon by creating `/etc/ryft-rest.conf` file and restarting ryft-server service. 
Full list of parameters is listed by running `ryft-server --help`. 
Daemon will start as follows: `ryft-server @/etc/ryft-rest.conf`
