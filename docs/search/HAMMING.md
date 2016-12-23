The fuzzy Hamming search operation works similarly to [exact search](./EXACT.md)
except that matches do not have to be exact. Instead, the `DISTANCE` option
allows the specification of a "close enough” value to indicate how close
the input must be to the match string contained in the match criteria.
The match string can be up to 32 bytes in length.

A “close enough” match is specified as a Hamming distance.
The Hamming distance between two strings of equal length is the number
of positions at which the corresponding characters are different. As provided
to the Hamming search operation, the Hamming distance specifies the maximum
number of character substitutions that are allowed in order to declare a match.

Hamming searches extend the general relational expression defined
[previously](./README.md#general-search-syntax) as follows:

```
(input_specifier relational_operator HAMMING(expression, options))
```

A fully qualified `HAMMING` clause looking for the term `"albatross"`
and return matching lines using a distance of `1` would be:

```
(RAW_TEXT CONTAINS HAMMING("albatross", DISTANCE="1", LINE="true"))
```

The following aliases can be used to specify `HAMMING` primitive as well:
- `HAMMING`
- `FHS`

so the following queries are the same:

```
(RAW_TEXT CONTAINS HAMMING("albatross", DISTANCE="1", LINE="true"))
(RAW_TEXT CONTAINS FHS("albatross", DISTANCE="1", LINE="true"))
```

Hamming search algorithms are highly parallelizable, so they can run extremely
quickly on accelerated hardware. Therefore, if search algorithms can make use
of Hamming search, then they will typically perform at very high speed,
often at the same speed as [exact searches](./EXACT.md).


# Options

The available comma-separated options for the `HAMMING` primitive are:

- [DISTANCE](#distance-option)
- [WIDTH](#width-option)
- [LINE](#line-option)
- [CASE](#case-option)
- [FILTER](./README.md#filter-option)

Note that the `DISTANCE` option is required, as it sets the Hamming distance
that will be used. If `DISTANCE` is specified as zero the search mode
will be automatically changed to [EXACT](./EXACT.md).


## `DISTANCE` option

`DISTANCE` specifies the fuzzy search distance as an [integer value](./README.md#integers).

A match is determined if the distance between the match and the search expression
is less than or equal to this `DISTANCE` option.

For unstructured `RAW_TEXT` queries, the resulting distance value for each match
encountered is written to the output index file, if specified, which allows for
downstream analytics tools to know how close the match was to the request.

For `RECORD` based searching (where input_specifier is `RECORD` or `RECORD.<field_name>`),
the distance is not reported in the output index file, since the purpose of
`RECORD` based searching is to return the entire matching `RECORD` which can often
have more than one internal match, each of which could have a different distance.

`DISTANCE=0` is used by default.

The following aliases can be used to specify `DISTANCE` as well:
- `DISTANCE`
- `FUZZINESS`
- `DIST`
- `D`

so the following queries are the same:

```
(RAW_TEXT CONTAINS HAMMING("albatross", DISTANCE=1))
(RAW_TEXT CONTAINS HAMMING("albatross", FUZZINESS=1))
(RAW_TEXT CONTAINS HAMMING("albatross", DIST=1))
(RAW_TEXT CONTAINS HAMMING("albatross", D=1))
```


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
(RAW_TEXT CONTAINS HAMMING("albatross", D=1, SURROUNDING=5))
(RAW_TEXT CONTAINS HAMMING("albatross", D=1, WIDTH=5))
(RAW_TEXT CONTAINS HAMMING("albatross", D=1, W=5))
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
(RAW_TEXT CONTAINS HAMMING("albatross", D=1, LINE=true))
(RAW_TEXT CONTAINS HAMMING("albatross", D=1, L=true))
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
(RAW_TEXT CONTAINS HAMMING("albatross", D=1, CASE=false))
(RAW_TEXT CONTAINS HAMMING("albatross", D=1, CS=false))
```

Please see [boolean type](./README.md#booleans) to get the ways
the `CASE` option can be set.


# See Also

- [Exact search](./EXACT.md)
- [Fuzzy Edit Distance search](./EDIT_DIST.md)
- [Date search](./DATE.md)
- [Time search](./TIME.md)
- [Number search](./NUMBER.md)
- [Currency search](./CURRENCY.md)
- [IPv4 search](./IPV4.md)
- [IPv6 search](./IPV6.md)
