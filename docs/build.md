This document contains information about how to build `ryft-server`.

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

This command builds `ryft-server` and packs everything into Debian package which
can be found under `debian/` subdirectory.

Note, `make debian` should be run in the project's root directory. In this case `ryft-server` is rebuilt
and corresponding deb package is created. If you run `make` from `debian/` subdirectory then only
deb package is created - `ryft-server` is not rebuilt, just used from your `$GOPATH/bin`.

It's recommended to run `make update` periodically to update all 3rd-party
dependencies.

The command `docker-build` or `docker_build` allows to build `ryft-server` inside Docker container
using environment with fixed versions of side auxiliary packages. The side package parameters are listed
in `vendor/vendor.json` file. The build process inside Docker container performs by compiler with
fixed version and takes a time.

First of all, it is necessary to install Docker from official repository using
[installation notes](https://docs.docker.com/engine/installation/linux/docker-ce/ubuntu/).

Then it should be started the Debian package build inside Docker container:

```{.sh}
make docker-build
```

The results are packed to Debian package which can be found under `debian/` subdirectory.

## Version

Note the `ryft-server` version - it's automatically generated using current git commit.
Simple `make` is equivalent to `go build` execution with specified keys. It will will produce the
following output (exact numbers may differ):

```{.sh}
$ make
go build -ldflags "-X main.Version=0.6.1-139-g51fcf47 -X main.GitHash=51fcf47f0de217b0dfba4c4e2ed83ed172e123ae"

Version: 0.6.1-139-g51fcf47
GitHash: 51fcf47f0de217b0dfba4c4e2ed83ed172e123ae
```

where `main.Version` is the server's version and `main.GitHash` is corresponding git commit hash.

By default version variable is generated using [git describe](https://git-scm.com/docs/git-describe) tool.
But it's possible to override automatically generated version:

```{.sh}
$ make VERSION=1.2.3
go build -ldflags "-X main.Version=1.2.3 -X main.GitHash=51fcf47f0de217b0dfba4c4e2ed83ed172e123ae"
```

Same VERSION handling applies to debian builds regarding version number generation.

Automatic version number based on build and git:

```{.sh}
$ make debian
go install -ldflags "-X main.Version=0.7.0-9-g168b8c1 -X main.GitHash=168b8c1fceabe70333d5b855b9a27df219ebeb34"
```
Override automatic version number based on branch/release build requirements

```{.sh}
$ make debian VERSION=0.18.44
go install -ldflags "-X main.Version=0.18.44 -X main.GitHash=168b8c1fceabe70333d5b855b9a27df219ebeb34"
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


# Releasing

This section describes steps how to make a release build.

Before make a release please ensure:
- static/swagger.json is updated:
   - has appropriate `info.version`
   - most of clients API are correct (Go, Python, JavaScript)
- 3rd party dependencies are updated (if decided): `make fetch_vendor`
- all unit tests are OK: `make test` or `make test-cover`

Switch to `master` branch and merge all development code or create corresponding pull request and merge it.
On [GitHub Releases](https://github.com/getryft/ryft-server/releases) page
push the "Draft a new Release" button, select target branch as `master` and
set the next release tag. Enter short description. For "alpha" versions
check "This is a pre-prelease" checkbox.

Once release tag is created build the corresponding Debian package:

```{.sh}
git checkout master
git pull
make docker-build
```

Please check all the integration tests are OK in Jenkins:
- `ryft-integration-tests` [obsolete] `ryftx` tests
- `ryft-integration-tests-ryftprim` [obsolete] `ryftprim` tests
- `ryft-integration-tests-v2` new version of tests (both `ryftx` and `ryftprim`)
