The `EXACT` search operation will search the input data corpus
for exact matches against a string of up to 32 bytes in length.

Exact searches extend the general relational expression defined
[previously](./README.md#general-search-syntax) as follows:

```
(input_specifier relational_operator EXACT(expression[, options]))
```

A fully qualified `EXACT` clause looking for the term `"orange"` would be:

```
(RECORD CONTAINS EXACT("orange"))
```

The following aliases can be used to specify `EXACT` primitive as well:
- `EXACT`
- `ES`

so the following queries are the same:

```
(RECORD CONTAINS EXACT("orange"))
(RECORD CONTAINS ES("orange"))
```

# Compatible syntax

For the backward compatibility the term `EXACT` can be omitted:

```
(input_specifier relational_operator expression)
```

Note, there is no way to specify additional options in this syntax!
The global options will be used.

Actually the syntax can be even simplier: standalone `expression` is
converted to `(RAW_TEXT CONTAINS expression)`. For example the
`"Apple" OR "Orange"` will be converted to

```
(RAW_TEXT CONTAINS EXACT("Apple"))
OR
(RAW_TEXT CONTAINS EXACT("Orange"))
```


# Options

The available comma-separated options for the `EXACT` primitive are:

- [WIDTH](#width-option)
- [LINE](#line-option)
- [CASE](#case-option)
- [FILTER](./README.md#filter-option)


## `WIDTH` option

`WIDTH` specifies a surrounding width as an [integer value](./README.md#integers).

Surrounding width means that results will be returned which will contain
the specified number of characters (value) to the left and right of the match.

Surrounding width has no meaning for record-based searches and will be ignored.

Note that `WIDTH` and `LINE` are mutually exclusive. Option that goes last
has an effect. This behaviour differs from `ryftprim` which results an error
if both are specified in the same query.

`WIDTH=0` is used by default.

The following aliases can be used to specify `WIDTH` as well:
- `SURROUNDING`
- `WIDTH`
- `W`

so the following queries are the same:

```
(RECORD CONTAINS EXACT("orange", SURROUNDING=5))
(RECORD CONTAINS EXACT("orange", WIDTH=5))
(RECORD CONTAINS EXACT("orange", W=5))
```


## `LINE` option

When `LINE` is `true`, the query will return the (line-feed delimited) line
on which the match occurs. This can be very useful when processing line-feed
delimited comma separated value (CSV) files in order to return the entire
CSV record for which a match was found.

Should a match appear multiple times on a single line, only one match line
will be generated. If a nearby line feed cannot be found within a distance
of one chunk of either side of the chunk containing the match, then the data
results are undefined, and become implementation-specific.

The system defaults to use `WIDTH` mode, and `LINE` mode may be enabled
on a per-query basis.

Note that `WIDTH` and `LINE` are mutually exclusive. Option that goes last
has an effect. This behaviour differs from `ryftprim` which results an error
if both are specified in the same query.

`LINE=false` is used by default.

The following aliases can be used to specify `LINE` as well:
- `LINE`
- `L`

so the following queries are the same:

```
(RECORD CONTAINS EXACT("orange", LINE=true))
(RECORD CONTAINS EXACT("orange", L=true))
```

Please see [boolean type](./README.md#booleans) to get the ways
the `LINE` option can be set.


## `CASE` option

When `CASE` is `false`, the query will be run caseinsensitive.

`CASE=true` is used by default, so searches are case sensitive.

The following aliases can be used to specify `LINE` as well:
- `CASE`
- `CS`

so the following queries are the same:

```
(RECORD CONTAINS EXACT("orange", CASE=false))
(RECORD CONTAINS EXACT("orange", CS=false))
```

Please see [boolean type](./README.md#booleans) to get the ways
the `CASE` option can be set.


# See Also

- [Fuzzy Hamming search](./HAMMING.md)
- [Fuzzy Edit Distance search](./EDIT_DIST.md)
- [Date search](./DATE.md)
- [Time search](./TIME.md)
- [Number search](./NUMBER.md)
- [Currency search](./CURRENCY.md)
- [IPv4 search](./IPV4.md)
- [IPv6 search](./IPV6.md)
