# Demo - busyness metric - July 28, 2016

To show a demo we need a cluster running with at least two nodes.
In our example we assume the following nodes are avaialble: `ryft-310` and `ryft-313`.

Authentication should be enabled. `test` user should be used.
`/ryftone/test/` directory should contains a few text files (say ``passengers.txt`) on both nodes.

We also need to setup partitioning so the search requests go to both nodes.
Add `test/partitions/*.txt=ryft` to the KV storage (Consul UI can be used).

Current busyness metric can be get from KV under `busyness/` prefix.


## Busyness metric

For now the busyness metric is calculated just as the number of active search requests on the node.
It doesn't take into account actual Ryft hardware loading.


## Make some load on first nodes

There is special query option for test purposes `&fake=1m` which tell the ryft
server to send fake random data back during 1 minute.

```{.sh}
curl -s -u test:test "http://ryft-310:8765/search?query=555&files=*.txt&local=true&fake=1m" &
```

Now we can check the actual busyness metric in KV storage.


## Load balancing in action

The `ryft-310` should have bigger metric. So if we send real search query to this
node the request should be redirected to `ryft-313` node.

```{.sh}
curl -s -u test:test "http://ryft-310:8765/search?query=555&files=*.txt" | jq .
```

All indexes should contains `host=ryft-313`!


## The same metric and `tolerance` parameter

If both nodes have the same metric the actual node is selected randomly.

There is a special tolerance option. With is used to rearrange nodes based on their metrics.
Actually nodes are groupped based on metric level which is `metric / (tolerance+1)`.
Nodes within the same group are randomly shuffled.

For example, if we have the following nodes:
- `node-a` with metric `1`
- `node-b` with metric `2`
- `node-c` with metric `3`
- `node-d` with metric `4`

If `tolerance=0` all nodes will be arranged to [`node-a`, `node-b`, `node-c`, `node-d`].

If `tolerance=1` all nodes will be arranged to [`node-a`, perm(`node-b`, `node-c`), `node-d`].
Because levels are: `1/2=0`, `2/2=1`, `3/2=1`, `4/2=2`.

If `tolerance>=4` all nodes are in the same group [perm(`node-a`, `node-b`, `node-c`, `node-d`)].
