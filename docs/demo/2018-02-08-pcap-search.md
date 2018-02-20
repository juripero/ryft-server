# Demo - PCAP search - February 8, 2018

## New REST API endpoints

There are a few new REST API endpoints to support PCAP search:
- [GET `/pcap/search`](../rest/search.md#pcap-search)
- [GET `/pcap/count`](../rest/search.md#pcap-search)

Note, since there is no INDEX file provided by the `ryftx` the `/pcap/search`
endpoint works only with the `/pcap/search?limit=0` option.

There is no complex queries supported for the PCAP search so the input search
query is forwarded to the `ryftx` "as is".


## `ryftrest` tool

The `ryftrest` tool also supports the PCAP search. Just set the search mode to
`-p pcap`:

```{.sh}
ryftrest -vvv -p pcap -f maccdc2012_00000.pcap -q "ip.src == 192.168.229.254 and ip.dest == 192.168.202.79"
```

As usual the output DATA file can be saved to disk with `-od` option:

```{.sh}
ryftrest -vvv -p pcap -f maccdc2012_00000.pcap -q "ip.src == 192.168.229.254 and ip.dest == 192.168.202.79" -od test/out.pcap
```


## `tshark` post-processing

There is dedicated Docker image `ryft/alpine-tshark` containing `tshark` tool.
This tool can be run on the server side using `/run` REST API endpoint.

Please, ensure the server configuration contains:

```{.yaml}
docker:
  ...
  images:
    ...
    tshark: ["ryft/alpine-tshark"]
```

So having any PCAP file it's possible to run `tshark` tool:

```{.sh}
curl -s 'http://localhost:8765/run?image=tshark&arg=-r&arg=test/out.pcap&arg=-qnz&arg=io,phs'

===================================================================
Protocol Hierarchy Statistics
Filter:

eth                                      frames:20642 bytes:3120824
  vlan                                   frames:20642 bytes:3120824
    ip                                   frames:20642 bytes:3120824
      tcp                                frames:20637 bytes:3120036
        ssl                              frames:10923 bytes:2426356
          tcp.segments                   frames:5 bytes:6139
            ssl                          frames:5 bytes:6139
        ssh                              frames:12 bytes:2408
      icmp                               frames:5 bytes:788
===================================================================
```

or

```{.sh}
$ curl -s 'http://localhost:8765/run?image=tshark&arg=-r&arg=test/out.pcap&arg=-qnz&arg=conv,ip'
================================================================================
IPv4 Conversations
Filter:<No Filter>
                                               |       <-      | |       ->      | |     Total     |    Relative    |   Duration   |
                                               | Frames  Bytes | | Frames  Bytes | | Frames  Bytes |      Start     |              |
192.168.202.79       <-> 192.168.229.254        20642   3120824       0         0   20642   3120824  2186583226.001226902   -2186587322.0000
================================================================================
```

Note, the `ryft-alpine-tshark` Docker image has been built during ryft-server Debian package
installation in post-install script.
