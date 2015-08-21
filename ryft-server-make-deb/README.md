## Status of the package maker

Under active development.

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

## How to install ``.deb`` package

```
sudo dpkg -i ryft-server-0.1_x86_64.deb

```


## How to uninstall ``.deb`` package

```
sudo dpkg -r ryft-server
```