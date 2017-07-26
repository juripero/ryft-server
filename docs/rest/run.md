The `GET /run` endpoint is used to run any command on the server side.

The command is run inside a Docker container. The Ryft user's home directory
is mounted to the container at `/ryftone` so it is also possible to upload any
script file first with `POST /files` and then execute it with `GET /run`.

See corresponding [demo](../demo/2017-05-25-run-command.md) for examples.


# Parameters

The list of supported query parameters are the following:

| Parameter | Type    | Description |
| --------- | ------- | ----------- |
| `image`   | string  | [The Docker image to run command in](#run-image-parameter). |
| `command` | string  | [The command to run](#run-command-parameter). |
| `arg`     | []string | [Array of the command arguments](#run-arg-parameter). |


## Run `image` parameter

The list of allowed images is configured via [server configuration file](../run.md#docker-configuration).
By default the `image=default` is used which is usually `alpine:latest` Docker image.


## Run `command` parameter

The command to run. Can be any system command like `sh` or `cat` or a custom script.
That custom script should be uploaded first to the Ryft user home directory. In
this case the command name should be started with dot `command=./test.sh`.

If the command is empty then the first `arg` is used as a command.

## Run `arg` parameter

Array of the command arguments. Can be specified many times:
`command=sh&arg=-c&arg=echo%20hello`.
