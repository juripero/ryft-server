# RyftDEC search engine

In terms of ryft-server RyftDEC is a search engine. The main purpose is to
decompose complex search query into several simple sub-queries.
RyftDEC uses another search engine as a backend and performs each sub-query
using that backend.

For example let's process this complex search query:

```
(RECORD.id CONTAINS "10030") AND (RECORD.date CONTAINS DATE(MM/DD/YYYY > 04/13/2015)) AND (RECORD.date CONTAINS TIME(HH:MM:SS > 11:20:00))
```

RyftDEC decomposes this complex expression into the following tree:

- `AND` operator
  - A: `(RECORD.id CONTAINS "10030")`
  - `AND` operator
    - B: `(RECORD.date CONTAINS DATE(MM/DD/YYYY > 04/13/2015))`
    - C: `(RECORD.date CONTAINS TIME(HH:MM:SS > 11:20:00))`

The expression `A` will be called first as a normal search. The result of `A`
will be used as input for the `B` sub-query (date search). And result of `B`
will be used as input for the `C` sub-query (time search).

Currently `AND` operator is supported. `OR` and `XOR` is not supported yet.

## Extension

For structured search it's important to keep temporary file extension coherent
to the input. If input contains `*.pcrime` mask the temporary file should also
have `.pcrime` extension. Otherwise ryft won't use correspopnding RDF scheme.
