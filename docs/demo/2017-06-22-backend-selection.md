# Demo - ryftprim/ryftx backend selection - June 22, 2017

On AWS F1 instances and on Ryft One boxes there is possible to use a few
backends:
- `ryftprim` which uses FPGA acceletaration
- `ryftx` which uses native x86 CPU
- `ryftpcre2` (soon) for PCRE2 primitives


## Automatic backend selection

By default the following table is used to automatically select backend tool:

| Primitive  | ryftprim     | ryftx | ryftpcre2  |
| ---------- | ------------ | ----- | ---------- |
| EXACT      |              |   X   |            |
| FHS (D<=1) |              |   X   |            |
| FHS (D>1)  |    X         |       |            |
| FEDS       |    X         |       |            |
| NUMBER     |              |   X   |            |
| CURRENCY   |              |   X   |            |
| IPV4       |              |   X   |            |
| IPV6       |              |   X   |            |
| DATE       |              |   X   |            |
| TIME       |              |   X   |            |
| PCRE2      | X (Ryft box) |       | X (AWS/F1) |


Let's try a few examples:

```{.sh}
$ ryftrest -p=es -q "mich" -i -f passengers.txt  | jq -c '{"backend": .extra.backend}'
{"backend":"ryftx"}
```

Please note the `backend` field in statistics `extra`. According to table the
`EXACT` primitive uses `ryftx` backend tool. The same is true for `FHS` with
distance `1`:

```{.sh}
$ ryftrest -p=fhs -d=1 -q "mich" -i -f passengers.txt  | jq -c '{"backend": .extra.backend}'
{"backend":"ryftx"}
```

But if `FHS` distance is increased to 2 or `FEDS` primitive is used, then backend is
changed to `ryftprim`:

```{.sh}
$ ryftrest -p=fhs -d=2 -q "mich" -i -f passengers.txt  | jq -c '{"backend": .extra.backend}'
{"backend":"ryftprim"}
$ ryftrest -p=feds -d=1 -q "mich" -i -f passengers.txt  | jq -c '{"backend": .extra.backend}'
{"backend":"ryftprim"}
```

In case of complex queries the query decomposition is done first. Since `ryftx`
doesn't  support boolean operations all complex queries are handled by `ryftprim`
backend.

```{.sh}
$ ryftrest -p=es -w=5 -q "mich OR mic" -i -f passengers.txt  | jq -c '{"backend": .extra.backend}'
{"backend":"ryftprim"}
$ ryftrest -p=es -w=5 -q "mich AND mic" -i -f passengers.txt | jq -c '.details[] | {"backend": .extra.backend}'
{"backend":"ryftx"}
{"backend":"ryftx"}
```

Note, the first query uses `OR` operator for `RAW_TEXT`. These boolean operation
can be handled nativelly by `ryftprim` so this backend is used in order to minimize
number of backend calls.

The second query uses `AND` operator. `ryftprim` doesn't support `AND` for `RAW_TEXT`
so query is decomposed into two simple calls. Both calls use `ryftx` as backend
according to table rule.


## Manual backend selection

There is dedicated `backend` query [option](../rest/search.md#search-backend-parameter)
in `/search` and `/count` REST API endpoints. Also the `ryftrest` supports `--backend`
command line argument.

All options can have the following values:
- "" (empty, by default) is used to select backend tool in automatic mode
- "ryftprim" or "prim" or "1" is used to select `ryftprim` tool.
- "ryftx" or "x" is used to select `ryftx` tool.
- "ryftpcre2" or "pcre2" is used to select `ryftpcre2` tool (not implemented yet).

So, even if table says to use `ryftx` backend tool we can force `ryftprim` selection:

```{.sh}
$ ryftrest -p=es -w=5 -q "mich AND mic" -i -f passengers.txt --backend=1 | jq -c '.details[] | {"backend": .extra.backend}'
{"backend":"ryftprim"}
{"backend":"ryftprim"}
```


## Configuration

The path to each backend is customized via server's [configuration file](../run.md#search-configuration).
There are the following options in `backend-options` section related:
- `ryftprim-exec` path to `ryftprim` executable, by default it is "/usr/bin/ryftprim"
- `ryftx-exec` path to `ryftx` executable, by default it is empty
- `ryftpcre2-exec` path to `ryftcpre2` executable (not implemented yet)

The configuration should contains at least one valid `ryftprim` or `ryftx` executable path.
If both are provided then backend selection rules are applied. Otherwise there is
no alternative except to use available backend tool.

On those instances there is no FPGA available (AWS i3 for example) the following
configuration should be used:
```{.yaml}
backend-options:
  ryftprim-exec:  ""                 # no FPGA
  ryftx-exec:     /usr/bin/ryftx
  ryftpcre2-exec: /usr/bin/ryftpcre2 # will be /usr/bin/ryftx soon
```

If FPGA is available (AWS F1 for example), i.e. both backends can be used:
```{.yaml}
backend-options:
  ryftprim-exec:  /usr/bin/ryftprim
  ryftx-exec:     /usr/bin/ryftx
  ryftpcre2-exec: /usr/bin/ryftpcre2 # will be /usr/bin/ryftx
```

On Ryft boxes the `PCRE2` is nativelly supported by `ryftprim`:
```{.yaml}
backend-options:
  ryftprim-exec:  /usr/bin/ryftprim
  ryftx-exec:     /usr/bin/ryftx
  ryftpcre2-exec: /usr/bin/ryftprim
```
