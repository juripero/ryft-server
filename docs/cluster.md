This document contains some description of cluster mode features.

# Partitioning

There is an excellent description of partitioning in the
`cluster.md` document of `ryft-cluster` project.
Please take a look it.

# Busyness

This section contains description of a load balancing.

According to partitioning rules some cluster nodes contain the same data.
So the same search request can be served by `Node-A` or `Node-B` for example.
To keep both nodes loaded equally the special metric `busyness` was introduced.

For now `busyness` metric is just the number of active HTTP requests on the node
(futher it can be modified to show actual Ryft hardware loading, including
percentage completed). Having a new search request we can select `Node-A`
or `Node-B` based on this metric - the node with lowest metric will be used.

All nodes keep and update their metrics in the `consul`'s KV storage
under `busyness/` perfix. Once a new search request arrives, metrics for all
nodes are obtained from KV and all nodes are arranged - from the
lowest metric to the highest.

It's possible to use a `tolerance` parameter. Actual level will be calculated
as `level = metric / (tolerance + 1)`. This option is used to put
nodes with almost the same `busyness` metric at the same level. For example,
if `busyness/node-a=10` and `busyness/node-b=13`:
- having `tolerance=4` all nodes will be places at the same level `2`,
- having `tolerance=0` or `tolerance=1` all nodes will be at their own levels.

To customize this parameter the `--busyness-tolerance` command line option
can be used. By default it is zero.
