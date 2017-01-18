# Demo - new query parser features - January 12, 2017

The new lexical parser was developed instead of regexp based.

The parser contains the following features:
- [flexible syntax](#flexible-syntax)
- [new query optimization rules](#query-optimization-rules)
- [new documentation](../search/README.md)


## Helper "dry-run" REST API endpoint

There is special REST API endpoint `/search/dry-run` that can be used to
test new query parser. This endpoint accepts the same parameters as
regular `/search` but does not call ryftprim. Instead it just prints
parsed and optimized queries.

We will use this API endpoint to show some new parser features. For example,
simple request `http://localhost:8765/search/dry-run?query=hello&file=*.txt`
produces the following response:

```{.json}
{
    "engine": {
        // ... search engine properties ...
    },
    "final": {
        "new-expr": "(RAW_TEXT CONTAINS EXACT(\"hello\"))",
        "old-expr": "(RAW_TEXT CONTAINS \"hello\")",
        "options": {
            "case": true,
            "mode": "es"
        },
        "structured": false
    },
    "parsed": {
        "new-expr": "(RAW_TEXT CONTAINS EXACT(\"hello\"))",
        "old-expr": "(RAW_TEXT CONTAINS \"hello\")",
        "options": {
            "case": true,
            "mode": "es"
        },
        "structured": false
    },
    "request": {
        "query": "hello",
        "files": [
            "*.txt"
        ],
        "case": true
    }
}
```

Resulting JSON contains:
- `"request"` input request from URL query parameters
- `"parsed"` fully parsed but not optimized query
- `"final"'` final (optimized) query


## Compatibility mode

Both parsed and final queries contains two search expressions:
- "new-expr" uses new "generic" syntax
- "old-expr" uses old "compatible" syntax

By default the new "generic" syntax is used and ryftprim is called with
`-p g` argument. But it is possible to run REST server in "compatible" mode.
In this case "old-expr" will be used with corresponding search mode.
The "compatible" mode can be used with old firmware version.

There is special parameter in [configuration](../run.md#search-configuration)
to enable/disable "compatibility" mode:
```{.yaml}
backend-options:
  compat-mode: true    # true - compatibility mode, false - generic
```


## Flexible syntax

New query parser support all search modes:
- [Exact search](../search/EXACT.md)
- [Fuzzy Hamming search](../search/HAMMING.md)
- [Fuzzy Edit Distance search](../search/EDIT_DIST.md)
- [Date search](../search/DATE.md)
- [Time search](../search/TIME.md)
- [Number search](../search/NUMBER.md)
- [Currency search](../search/CURRENCY.md)
- [IPv4 search](../search/IPV4.md)
- [IPv6 search](../search/IPV6.md)

Note, the RegExp search is not supported yet since there is no specifications
available for the REGEXP syntax and its parameters.


### Old and new syntax

Both syntax are supported. The following queries are parsed to the same "final" query:

- `(RAW_TEXT CONTAINS "hello")`
- `(RAW_TEXT CONTAINS EXACT("hello"))`
- `(RAW_TEXT CONTAINS ES("hello"))`


### Plain simple query

The search query can be even shorter. **Single** word or quoted text is
automatically converted to `(RAW_TEXT CONTAINS )` query:
- `hello`
- `"hello world"`


### Inline options

It is possible to override global search options using new query syntax:
`(RAW_TEXT CONTAINS EXACT("hello", WIDTH=5, CASE=false))`
will override global `WIDTH` and `CASE`.

Each search type has its own set of supported options.

For `FHS` and `FEDS` search if distance is not provided (or provided as zero)
the search type will be automatically changed to `EXACT`. The following queries
are the same:
- `(RAW_TEXT CONTAINS FHS("hello"))`
- `(RAW_TEXT CONTAINS FEDS("hello", DIST=0))`
- `(RAW_TEXT CONTAINS EXACT("hello"))`

Each option supports a few aliases. For example distance can be specified as:
- `(RAW_TEXT CONTAINS FEDS("hello", DISTANCE=10))`
- `(RAW_TEXT CONTAINS FEDS("hello", FUZZINESS=10))`
- `(RAW_TEXT CONTAINS FEDS("hello", DIST=10))`
- `(RAW_TEXT CONTAINS FEDS("hello", D=10))`

Value can be optionally quoted:
- `(RAW_TEXT CONTAINS FEDS("hello", D="10"))`
- `(RAW_TEXT CONTAINS FEDS("hello", D=10))`

Boolean options can be specified in the following ways:
- `(RAW_TEXT CONTAINS EXACT("hello", CS="False"))`
- `(RAW_TEXT CONTAINS EXACT("hello", CS=False))`
- `(RAW_TEXT CONTAINS EXACT("hello", CS=0))`
- `(RAW_TEXT CONTAINS EXACT("hello", !CS))`


### Width and Line

`WIDTH` and `LINE` options can be provided at the same time.
The last one has an effect. (ryftprim shows error in that case).

Moreover, for consistency with the REST API it is also possible
to specify `WIDTH=line` instead of `LINE=true`.


### New operators

The DATE, TIME, IPV4 and IPV6 search support additional operators:
- `IP == "ValueB"`
- `ValueA >= IP >= ValueB`
- `ValueA > IP > ValueB`
- `ValueA > IP >= ValueB`
- `ValueA >= IP > ValueB`

These operators will be automatically converted to Ryft-supported form.
For example: `RAW_TEXT CONTAINS IPv4("1.1.1.1" >= IP > "8.8.8.8")` will be
converted to `(RAW_TEXT CONTAINS IPV4("8.8.8.8" < IP <= "1.1.1.1"))`.


## Query optimization rules

The optimization rules are the following:
- combine all RECORD-based sub-queries `(RECORD.text CONTAINS FHS("A",d=1)) AND (RECORD.id CONTAINS NUMBER(NUM > 100))`
- search type based exception list - can be configured via `optimizer-do-not-combine` option: `(RECORD.text CONTAINS FEDS("A",d=1)) AND (RECORD.id CONTAINS NUMBER(NUM > 100))`
- combine all RAW_TEXT-based sub-queries containing `OR`: `(RAW_TEXT CONTAINS "Apple") OR (RAW_TEXT CONTAINS "Orange")`
- do NOT combine RAW_TEXT-based sub-queries containing `AND`: `(RAW_TEXT CONTAINS "Apple") AND (RAW_TEXT CONTAINS "Orange")`

There are also [curly braces](../search/README.md#curly-braces) supported.
The curly braces break optimization rules.
It can be used for manual optimization.


## Line option

In order to support `LINE` options there are a several changes on each
product part:
- `ryftrest` supports `--line` switch, actually means `-w=line`.
- REST service supports `surrounding=line`, see [corresnponding documentation page](../rest/search.md#search-surrounding-parameter)
- Swagger interface updated, `surrounding` parameter is of `string` type.


Example:

```{.sh}
ryftrest -f regression/wikipe*bin -q '"babe ruth"' -u admin:admin --search --format=utf8 --line | jq .results
```
