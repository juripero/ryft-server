# backend options tweaks
d
There described new section in the `ryft-server.conf` that allow user to tune backend options in order to achieve better performance.

Part of config:
```{.yaml}
backend-options:
    backend-tweaks:
        options:
            ...
        router:
            ...
```

## `router` section

If presented options define mapping between search primitive and backend tool.

e.g.:
```{.yaml}
router:
    default: ryftx
    es: ryftprim
    fhs,ds: ryftprim
    prce2: prce2
```

`/search` and `/count` endpoints accept `backend` parameter, but if it is not set explicetly `router` may be used for choosing backend that fits better for current search primitive. If search primitive is ommited in the `router` table value of the `default` key will be used. 

## options

Options defined in `options` section should have an `array` format and will be passed to the backend tool within a search query. 
Parameters set in `backend-options` have more priority than `options` though.

User can set options for a backend tool, for a search primitive and for a combination `[backend].[search primitive]`. 

User can also create set of options specifying `backend-mode` parameter. In this case key of `options` table has a structure: `[backend-mode].[backend].[search primitive]`. 
e.g.:
```{.yaml}
options:
    default: []             # default mode
    default.ryftx.es: []
    default.ryftx.ds: []
    default.ryftx: []
    hp: []                  # high-performance mode
    hp.ryftx.es: []  
    hp.ryftprim.es: []
    hp.ryftprim: []
```
The rule of thumb is: the more preciesly you specify options the higher priority they have.
User can set `backend` and `backend-mode` parameters in `/search` and `count` endpoints.

Search order in config defined above:
- extract `backend-mode` from the request parameters
- extract `backend` tool from the request parameters or from the `router` table
- extract `search primitive` from the search query
- create `options` key using pattern `[backend-mode].[backend].[search primitive]` and search it in `options`. 
If nothing found in `options` try `[backend].[search primitive]`, then `[search primitive]`, `[backend]`, `[backend mode]`. Finally, if nothing found use value for the `default` key.
- execute `backend` tool with a `query` and found options. E.g. `/usr/bin/ryftprim [query] ... [backend-options]`

## debug

We added `debug-internals` flag of boolean type on the first level of `ryft-server.conf`. 
If presented and set to `true` response from `/search` and `/count` endpoints has `debug` section with used backend tool and its arguments.