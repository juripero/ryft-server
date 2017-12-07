# Demo - backend tweaks: options, router - October 26, 2017 

## Ability to configure the search engine selection based on primitive execution

New section `router` added inside `backend-options.backend-tweaks` section.
It is a table of pairs:

     [search primitive]: [backend tool]

E.g.:

```{.yaml}
backend-options:
    backend-tweaks:
        router:
            default: ryftx
            feds,fhs: ryftprim
            pcre2: pcre2
```

If nothing is set for a search primitive the `default` value will be used.

Let's start server with the config defined above.

    ryft-server 2>&1 | grep -i "\[ryftprim\]: executing tool"

Now we'll execute queries with `ryftrest` tool and check corresponding server logs


    ./ryftrest -q "Geo" -f "training.txt" --count

server log:

    msg="[ryftprim]: executing tool" args="[-p g -q (RAW_TEXT CONTAINS EXACT(\"Geo\")) -f training.txt -e \\x0d\\x0a -v -l]" task=0000000000000003 tool=/usr/bin/ryftx

So, here we have `exact search` primitive that is not set in a `router` table, thus server will use `default` backend tool.

Use `feds` primitive and we expect `ryftprim` tool to be used:

    ryftrest -q '(RECORD CONTAINS FEDS("Geo", D=1, CASE=False))' -f "chicago.crimestat" --count

server log:

    msg="[ryftprim]: executing tool" args="[-p g -q (RECORD CONTAINS EDIT_DISTANCE(\"Geo\", DISTANCE=\"1\", CASE=\"false\", REDUCE=\"true\")) -f chicago.crimestat -e \\x0d\\x0a -v -l]" task=0000000000000001 tool=/usr/bin/ryftprim


## Custom backend options based on a per backend-mode, per engine and per primitive basis.

Server now supports `backend-mode` parameter in `/search` and `/count` REST API. This parameter allow user to define any number of sets of backend-options. `default` mode should be exist always and it will be a fallback if `backend-mode` is not passed with search request. User can add `high-performance` mode or use any other name.

`ryftrest` tool supports this parameter as `--backend-mode=[backend-mode]`.

New section `options` introduced.

```{.yaml}
backend-options:
    backend-tweaks:
        options:
            key: [options]
```

Key has a following structure `[backend-mode].[backend].[search primitive]`. This solution allow us to reduce nesting levels in a config and make it more evident. 

If `backend` parameter is omitted - backend will be defined by the `router` table.

If `backend-mode` parameter is omitted - backend-mode will be set to `default`. 

Examples:

Let's define config as:

```{.yaml}
backend-options:
    backend-tweaks:
        options:
            default.ryftx.es: ["-v1"]
            default.ryftx.ds: ["-v2"]
            defaut.ryftx: ["-v3"]
            default: ["-v4"]
            hp.ryftx.es: ["-v5"]
            hp.ryftx: ["-v6"]
            hp.ryftprim: ["-v7"]
        router:
            default: ryftx
            feds: ryftprim
            fhs: ryftprim
            pcre2: ryftprim
```

Family of `-v*` arguments have no sense for any backend, but need to show that we pass parameters and get expected results.

Start server:

    ryft-server 2>&1 | grep -i "\[ryftprim\]: executing tool" 

Search with `exact search` primitive. Expect `-v4` argument fed to `ryftx` backend.

    ryftrest -q '(RECORD CONTAINS EXACT("Geo"))' -f "chicago.crimestat" --count 

server log:

    msg="[ryftprim]: executing tool" args="[-p g -q (RECORD CONTAINS EXACT(\"Geo\")) -f chicago.crimestat -e \\x0d\\x0a -v -l -v4]" task=0000000000000001 tool=/usr/bin/ryftx

Search with `exact search` primitive and default `backend-mode`. Expect `-v1` argument.

    ryftrest -q '(RECORD CONTAINS EXACT("Geo"))' -f "chicago.crimestat" --backend-mode=default --count

server log:

    msg="[ryftprim]: executing tool" args="[-p g -q (RECORD CONTAINS EXACT(\"Geo\")) -f chicago.crimestat -e \\x0d\\x0a -v -l -v1]" task=0000000000000001 tool=/usr/bin/ryftx

Search with `exact search` primitive, backend mode `hp` and `backend` ryftprim. Expect `-v7` argument

    ryftrest -q '(RECORD CONTAINS EXACT("Geo"))' -f "chicago.crimestat" --backend-mode=hp --backend=ryftprim --count -vvv

server log:

    msg="[ryftprim]: executing tool" args="[-p g -q (RECORD CONTAINS EXACT(\"Geo\")) -f chicago.crimestat -e \\x0d\\x0a -v -l -v7]" task=0000000000000002 tool=/usr/bin/ryftprim


In real world these arguments would be like ["--rx-shard-size","64M","--rx-max-spawns","14"].