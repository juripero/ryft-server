This directory contains a few docker files.

They must be run from the root source directory!

```{.sh}
cd ${SOURCES}
docker build -t ryft.build -f docker/Dockerfile.build .
```

[Build test](./Dockerfile.build) to test building from sources.
It copies source from current directory and build them on clean Ubuntu machine.