# Demo - Run custom command on server side - May 25, 2017

There is a new `GET /run` REST API method which runs custom command on server-side.

The Docker is used to run this command in an isolated environment. So the Ryft user
have access to its home directory only.


## Run system commands

The command is run in the Ryft users home directory. For example, the `test` user
home directory on the Ryft box is `/ryftone/home/test`. This directory is mounted
to the Docker container at `/ryftone` and this is default working directory.

So the `ls /ryftone/home/test` and `curl -u test:test -s "http://localhost:8765/run?command=ls"`
should produce the same list of files.

Note, that running `pwd` command which print current directory will show:

```{.sh}
$ curl -u test:test -s "http://localhost:8765/run?command=pwd"
/ryftone
```

It's possible to pass any number of command arguments with `arg` option:

```{.sh}
$ curl -u test:test -s "http://localhost:8765/run?command=ls&arg=-al"
...
```

It's also possible to run shell commands:

```{.sh}
$ curl -u test:test -s "http://localhost:8765/run?command=sh&arg=-c&arg=echo%20hello%20from%20shell"
hello from shell
```


## Customize docker image

By default there `alpine:latest` Docker image is used. It's very small but has
limited set of features. For example, we cannot use Bash shell:

```{.sh}
$ curl -u test:test -s "http://localhost:8765/run?command=bash&arg=-c&arg=echo%20hello%20from%20Bash"
{
    "status": 500,
    "message": "exit status 127",
    "details": "failed to execute"
}
```

But we can specify another Docker image with `image=` option to run bash commands:

```{.sh}
$ curl -u test:test -s "http://localhost:8765/run?image=ubuntu&command=bash&arg=-c&arg=echo%20hello%20from%20Bash"
hello from Bash
```

The set of allowed Docker images is specified via corresponding [configuration section](../run.md#docker-configuration):

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

There is also `docker.run` option which allows to customize Docker run command.
In our case, for example, we use `--rm` option to automatically remove container
once command is finished.

Environment variables can be used. So we mount Ryft user home directory with:
`--volume=${RYFTHOME}:/ryftone` option.

To sum up, the final Docker command consists of the following parts:

```
[docker.run] [docker.images[query.image]] query.command [query.arg]
```


## Run custom command

Inside Docker container we have access to Ryft user home directory
so we can upload any script:

```{.sh}
$ cat test.sh
#!/bin/sh

for i in 1 2 3 4 5 ; do
  echo $i
done

$ curl -s "http://localhost:8765/files?file=test.sh&offset=0&local=true" \
    -u test:test --data-binary @test.sh -H "Content-Type: application/octet-stream"
{"length":51,"offset":0,"path":"test.sh"}
```

Once the script is uploaded we can run it remotely:

```{.sh}
$ curl -u test:test -s "http://localhost:8765/run?command=sh&arg=test.sh"
1
2
3
4
5
```

We even can omit `sh` command (please note `./` at the begin of script name,
we should respect Linux rules since `/ryftone` is not in our `$PATH`):

```{.sh}
$ curl -u test:test -s "http://localhost:8765/run?command=./test.sh"
1
2
3
4
5
```


### Run python script

The python script can be run in the same way. First upload script:

```{.sh}
$ cat test.py
#!/usr/bin/python

if __name__ == "__main__":
  print "Hello from Python"

$ curl -s "http://localhost:8765/files?file=test.py&offset=0&local=true" \
    -u test:test --data-binary @test.py -H "Content-Type: application/octet-stream"
{"length":75,"offset":0,"path":"test.py"}
```

And then just run it:

```{.sh}
$ curl -u test:test -s "http://localhost:8765/run?image=python&command=python&arg=test.py"
Hello from Python
```

Note the `python` image! Neither `alpine` or `ubuntu` don't have Python installed by default.
