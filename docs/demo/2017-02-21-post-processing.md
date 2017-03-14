# Demo - new post-process transformation feature - February 21, 2017

There are a few post-process transformations supported: `match`, `replace` and
a `script` call.

We will use the special test JSON file to show the feature.
If no transformations are applied the `RAW_TEXT` output look like:

```{.sh}
$ ryftrest -q "hello" -f test.json -w=5 --format=utf8 --search -u admin:admin | jq -c .results[].data
"1111-hello-1111"
"2222-hello-2222"
"3333-hello-3333"
"4444-hello-4444"
"5555-hello-5555"
```


## Regexp match

The first kind of post-process transformation is a regexp match filter.
If found data is not matched it is dropped. The regexp is used as a filter
expression. Please note the new `--transform 'match("<expression>")'` option.
This option prints odd numbers only:

```{.sh}
$ ryftrest -q "hello" -f test.json -w=5 --format=utf8 --search -u admin:admin \
  --transform 'match("^[13579].*$")' \
  | jq -c .results[].data
"1111-hello-1111"
"3333-hello-3333"
"5555-hello-5555"
```


## Regexp replace

The second kind of post-process transformation is regexp replace.
We can do "match and replace" with `--transform 'replace("<expression>", "<template>")'` option.
Let's replace all `hello`s with `bye`s.

```{.sh}
$ ryftrest -q 'RAW_TEXT CONTAINS "hello"' -f test.json -w=5 --format=utf8 --search -u admin:admin \
  --transform 'replace("^(.*)hello(.*)$", "${1}bye${2}")' \
  | jq -c .results[].data
"1111-bye-1111"
"2222-bye-2222"
"3333-bye-3333"
"4444-bye-4444"
"5555-bye-5555"
```


## Call a script

A set of predefined scripts can be used to do a transformation.
The scripts are configured in server's [configuration file](../run.md#script-transformation-configuration).

For example, we can use RECORD-based search and the `jq` tool:

```{.sh}
ryftrest -q 'RECORD.text CONTAINS "hello"' -f test.json --format=utf8 --search -u admin:admin \
  --transform 'script("jq_ab")' \
  | jq -r -c .results[].data | jq -c .
{"a+b":11,"a":10,"b":1}
{"a+b":22,"a":20,"b":2}
{"a+b":33,"a":30,"b":3}
{"a+b":44,"a":40,"b":4}
{"a+b":55,"a":50,"b":5}
```

the `jq_ab` script adds two fields `a+b` and is configured as:

```{.yaml}
jq_ab:
  path: [/usr/bin/jq, -c, "{\"a+b\": (.a + .b), \"a\": .a, \"b\": .b}"]
```

In general, any script can be used. It takes input via STDIN and prints
output to STDOUT. If exit code is not zero the record is dropped.


## Transformation chain

A few transformations can be combined into the transformation chain:
we can use `match` and `replace` for the same data:

```{.sh}
$ ryftrest -q 'RAW_TEXT CONTAINS "hello"' -f test.json -w=5 --format=utf8 --search -u admin:admin \
  --transform 'match("^[13579].*$")' \
  --transform 'replace("^(.*)hello(.*)$", "${1}bye${2}")' \
  | jq -c .results[].data
"1111-bye-1111"
"3333-bye-3333"
"5555-bye-5555"
```

In this case each found record is processed by two transformations:
If the first transformation matches, then the second transformation is applied.
If the first transformation does not match, then the record is dropped.


## See Also

The `ryftrest` utility supports `--transform` command line option.

The `/search` and `/count` REST API supports `transform` [query parameter](../rest/search.md#search-transform-parameter).

A set of predefined scripts are defined in [configuration file](../run.md#script-transformation-configuration).
