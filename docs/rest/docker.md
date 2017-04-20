## Running ryft-server inside Docker

### Components:

#### Docker server

Docker is the powerful container management platform. We are going to use Docker-CE 17

#### Docker Registry

Docker Registry is a separate service that keeps containers and its meta-data. We use V2.

#### Docker Compose

Docker compose tool is used for container orchestration. It has to support corresponding docker protocol. We use version 1.11.2.
This tool allow us to run several clusters of N `ryft-server` and one `consul` node that works as a discovery service.
More info is [here](https://docs.docker.com/compose/)


### Starting from scratch

clone `https://github.com/getryft` or deliver this repository into the ryftone machine in any other way.
`docker` directory contains `docker-compose.yml` and `.env` and docker-compose use them all the time. 
`docker/Makefile` provide commands for building `ryft-server` images.

### How it works?

First of all we can not run `ryft-server` on the local machine because we need the `ryftprim` backend.
We need at least one application container `ryft-server`. `consul` will be downloaded automatically from the docker-hub.
Each `ryft-server` container has mounted `/ryftone` and `~/.ssh` directories from the host. 
`ryft-server` establishes `ssh` connection to the host machine and runs search query using `ryftprim` tool. Using `ssh` we reduce performance on transferring results, so it is not the best solution in case of performance tests.

![schema](static/docker_ryft.png)


### Manage cluster with docker-compose

you should be logged as `root` to manage `docker` service
```
sudo su
```

run application
```
docker-compose up
```

run application as daemon.
```
docker-compose up -d
```

you can see logs 

```
docker-compose logs -f
```

you can also run cluster under the special name, in this case you can start many clusters
```
docker-compose -p cluster1 up -d
```

stopping cluster
```
docker-compose -p cluster1 stop
```

delete all stopped containers
```
docker rm $(docker ps -qa)
```

docker-compose allocates ports for our `ryft-sever` dinamically, so we can find them using `docker ps` command, but it's easier to use docker-compose. `8765` is the default port for our application which it exposes. `app` is an alias for `ryft-server` in terms of docker-compose

```
docker-compose -p cluster1 port app 8765
0.0.0.0:32809
```


### Make commands

Builder container has everything to build go applications. Build it and push into the docker registry

```
make builder
```

Build go binary. You can use it in a Jenkins build job.

```
make build
```

Build container that contains go binary and push it into the docker registry 

```
make app
```

Finally we can remove all garbage created during builds 

```
make clear
```