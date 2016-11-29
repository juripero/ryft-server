# Setup

This demo uses tree different ryft-servers run on the ports: `5001`, `5002` and `5003`.
Each instance uses dedicated users file - the main aim is the different home
directories for the same `test` user:

|    port | home             |
|---------|------------------|
| `:5001` | `/ryftone/test1` |
| `:5002` | `/ryftone/test2` |
| `:5003` | `/ryftone/test3` |

the get content of all home directories use `sudo ls -al /ryftone/test?` shell command.

## Consul

Run the consul agent on local node: `consul agent -dev`

Then adds the services to the consul:
```{.sh}
curl -X PUT --data '{"ID":"ryft-1","Name":"ryft-rest-api","Tags":["A","B"],"Address":"127.0.0.1","Port":5001}' "http://localhost:8500/v1/agent/service/register"
curl -X PUT --data '{"ID":"ryft-2","Name":"ryft-rest-api","Tags":["B","C"],"Address":"127.0.0.1","Port":5002}' "http://localhost:8500/v1/agent/service/register"
curl -X PUT --data '{"ID":"ryft-3","Name":"ryft-rest-api","Tags":["C","D"],"Address":"127.0.0.1","Port":5003}' "http://localhost:8500/v1/agent/service/register"
```

Add test partitions to the consul's KV storage:
```{.sh}
curl -X PUT --data 'A' "http://localhost:8500/v1/kv/test/partitions/*-a.txt"
curl -X PUT --data 'B' "http://localhost:8500/v1/kv/test/partitions/*-b.txt"
curl -X PUT --data 'C' "http://localhost:8500/v1/kv/test/partitions/*-c.txt"
curl -X PUT --data 'D' "http://localhost:8500/v1/kv/test/partitions/*-d.txt"
```

To check consul state the Web UI can be used: `http://localhost:8500/ui/#`


# Demo

This demo shows how to create and delete files locally and in cluster mode.


## Post data to each instance locally

These commands will send data to each ryft-server instance and then shows
the content of home directories:

```{.sh}
curl --data "hello1" -H "Content-Type: application/octet-stream" -s "http://localhost:5001/files?file=test.txt&local=true" -u test:test | jq -c .
curl --data "hello2" -H "Content-Type: application/octet-stream" -s "http://localhost:5002/files?file=test.txt&local=true" -u test:test | jq -c .
curl --data "hello3" -H "Content-Type: application/octet-stream" -s "http://localhost:5003/files?file=test.txt&local=true" -u test:test | jq -c .
sudo ls -al /ryftone/test?
```

Content of each file should be different.

## Delete file on each instance locally

All created files can be deleted:
```{.sh}
curl -X DELETE -s "http://localhost:5001/files?file=test.txt&local=true" -u test:test | jq -c .
curl -X DELETE -s "http://localhost:5002/files?file=test.txt&local=true" -u test:test | jq -c .
curl -X DELETE -s "http://localhost:5003/files?file=test.txt&local=true" -u test:test | jq -c .
sudo ls -al /ryftone/test?
```

## Post file to all instances in cluster mode

The following command can be used to send the same data to each ryft-server instance:
```{.sh}
curl --data "hello" -H "Content-Type: application/octet-stream" -s "http://localhost:5002/files?file=test.txt" -u test:test | jq .
sudo ls -al /ryftone/test?
```

Actually we can send this command to any instance.
Note the absense of `local=true` query parameter.

Content of each file should be the same.

## Append file content

By default (if no offset provided) data are appended:
```{.sh}
curl --data " world" -H "Content-Type: application/octet-stream" -s "http://localhost:5002/files?file=test.txt" -u test:test | jq .
sudo cat /ryftone/test1/test.txt
sudo cat /ryftone/test2/test.txt
```

## Update file content

It is possible to replace just a part of a file:
```{.sh}
curl --data "Ryft!" -H "Content-Type: application/octet-stream" -s "http://localhost:5002/files?file=test.txt&offset=6" -u test:test | jq .
sudo cat /ryftone/test1/test.txt
sudo cat /ryftone/test2/test.txt
```

## Delete file on all instances in cluster mode

The following command will delete the file on all instances:
```{.sh}
curl -X DELETE -s "http://localhost:5002/files?file=test.txt" -u test:test | jq .
sudo ls -al /ryftone/test?
```

Note the absense of `local=true` query parameter.

## Use paritioning rules

According to our partitioning rules the file will be distributed:
```{.sh}
curl --data "helloA" -H "Content-Type: application/octet-stream" -s "http://localhost:5002/files?file=test-a.txt" -u test:test | jq .
curl --data "helloB" -H "Content-Type: application/octet-stream" -s "http://localhost:5002/files?file=test-b.txt" -u test:test | jq .
curl --data "helloC" -H "Content-Type: application/octet-stream" -s "http://localhost:5002/files?file=test-c.txt" -u test:test | jq .
curl --data "helloD" -H "Content-Type: application/octet-stream" -s "http://localhost:5002/files?file=test-d.txt" -u test:test | jq .
sudo ls -al /ryftone/test?
```

Partitioning rules are applied to the delete command:
```{.sh}
curl -X DELETE -s "http://localhost:5002/files?file=test-b.txt&file=test-c.txt" -u test:test | jq .
sudo ls -al /ryftone/test?
```

## Post catalog data

The following command can be used to send the same data to each ryft-server instance:
```{.sh}
curl --data "catalog data" -H "Content-Type: application/octet-stream" -s "http://localhost:5002/files?file=test.txt&catalog=test.catalog" -u test:test | jq .
sudo ls -al /ryftone/test?
```

## Delete catalog data

```{.sh}
curl -X DELETE -s "http://localhost:5002/files?catalog=test.catalog" -u test:test | jq .
sudo ls -al /ryftone/test?
```
