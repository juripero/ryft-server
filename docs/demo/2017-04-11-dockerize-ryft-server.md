# Demo - Run ryft-server inside docker - April 11, 2017

Docker is a container platform and we are going to use it for running several ryft-server applications of a different versions.

For orchestrating we use `docker-compose` tool.

### Components:

In our application we have just a few of them:

* ryft-server
* consul

And we want to be able to run ryft-server in a single node mode and in a cluster mode.
We use consul for a service discovery.
We can also run several clusters on one machine. 

### Example:

Run cluster with the name `cluster1`

```{.sh}
root@ryftone-313:# docker-compose -p cluster1  up -d app
Creating cluster1_app_1

```

Check if it works

```{.sh}
ryftuser@ryftone-313:# curl -X GET "http://$(sudo docker-compose -p cluster1 port app 8765)/files?dir=/secrets/" | jq .
{
  "dir": "/secrets",
  "files": [
    "b.txt"
  ],
  "details": {
    "d974c2430df3": {
      "b.txt": {
        "type": "file",
        "length": 17,
        "mtime": "2017-04-06T13:47:05Z",
        "perm": "-rw-r--r--"
      }
    }
  }
}
```

Run second cluster with the name `cluster2`

```{.sh}
root@ryftone-313:# docker-compose -p cluster2 up -d app
Creating cluster2_app_1
```

Check that they both exist

```{.sh}
root@ryftone-313:/home/ryftuser/ryft-server-vf/3_dockerize/ryft-server/docker# docker ps -a
CONTAINER ID        IMAGE                              COMMAND                  CREATED             STATUS              PORTS                     NAMES
04302f48400b        172.16.34.3:5000/ryft/app:latest   "/ryft-server --co..."   4 minutes ago       Up 5 seconds        0.0.0.0:32785->8765/tcp   cluster2_app_1
cd6ee4e471c4        172.16.34.3:5000/ryft/app:latest   "/ryft-server --co..."   4 minutes ago       Up 1 second         0.0.0.0:32786->8765/tcp   cluster1_app_1
```

And now just stop them one by one

```{.sh}
root@ryftone-313:/home/ryftuser/ryft-server-vf/3_dockerize/ryft-server/docker# docker-compose -p cluster1 stop
Stopping cluster1_app_1 ... done
root@ryftone-313:/home/ryftuser/ryft-server-vf/3_dockerize/ryft-server/docker# docker-compose -p cluster2 stop
Stopping cluster2_app_1 ... done
```