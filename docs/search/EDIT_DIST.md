Edit distance search performs a search that does not require two strings to be
of equal length to obtain a match. Instead of considering individual character
differences, edit distance search counts the minimum number of insertions,
deletions and replacements required to transform one string into another.
This can make it much more powerful than [Hamming search](./HAMMING.md)
for certain applications.

For example, a fuzzy Hamming search for a search term of `"Michelle"` with a distance
of `1` would match `"Mishelle"` since the position `'c'` is changed to `'s'`.
But it would not match against `"Mischelle"`, since the bolded characters shown
don't match the same positions in the `"Michelle"` search term, evaluating to a
distance of `6`, which is greater than the desired distance `1`, so no match is declared.

On the other hand, a fuzzy edit distance search specifying the same `"Michelle"`
match string and a distance of `1` would match against the string `"Mischelle"`,
since a single `'s'` can be inserted into `"Michelle"` to arrive at `"Mischelle"`,
making it distance `1`. The string `"Michele"` would also be a match
because one `'l'` was deleted to change `"Michelle"` into `"Michele"`.
`"Mishelle"` would also be a match, because of the single replacement of an `'s'`
with the `'c'`. However, since the distance specified was only
`1`, then `"Mischele"` would not be a match, because that requires two changes:
the addition of an `'s'` and the removal of an `'l'`.
If the distance had been specified as `2`, then `"Mischele"` would also have matched,
as would the other patterns mentioned, since matches are declared whenever the calculated edit
distance between two strings is less than or equal to the desired distance.

Fuzzy edit distance is an extremely powerful search tool for a variety of data sources,
including names, addresses, medical records searching, genomic and disease research data,
common misspellings, and more. Unlike fuzzy Hamming search, fuzzy edit distance
is a more natural fuzzy search paradigm for many algorithms, since it does not
require string matches to be of the same size. The tradeoff is that
fuzzy edit distance is not as amenable to full hardware parallelization,
so algorithms reliant on it typically run slower than those that implement
fuzzy Hamming search. In addition, the match string length added
to the distance value must not exceed 32 bytes.

Edit distance searches extend the general relational expression defined
[previously](./README.md#general-search-syntax) as follows:

```
(input_specifier relational_operator EDIT_DISTANCE(expression, options))
```

A fully qualified `EDIT_DISTANCE` clause looking for the term `"albatross"`
and return 20 bytes on each side of every match encountered while using
a distance of `1` would be:

```
(RAW_TEXT CONTAINS EDIT_DISTANCE("albatross", DISTANCE="1", WIDTH="20"))
```

The following aliases can be used to specify `EDIT_DISTANCE` primitive as well:
- `EDIT_DISTANCE`
- **`EDIT_DIST`**
- `EDIT`
- `FEDS`

so the following queries are the same:

```
(RAW_TEXT CONTAINS EDIT_DISTANCE("albatross", DISTANCE="1", W=20))
(RAW_TEXT CONTAINS EDIT_DIST("albatross", DISTANCE="1", W=20))
(RAW_TEXT CONTAINS EDIT("albatross", DISTANCE="1", W=20))
(RAW_TEXT CONTAINS FEDS("albatross", DISTANCE="1", W=20))
```

# Options

The available comma-separated options for the `EDIT_DISTANCE` primitive are:

- [DISTANCE](#distance-option)
- [REDUCE](#reduce-option)
- [WIDTH](#width-option)
- [LINE](#line-option)
- [CASE](#case-option)
- [FILTER](./README.md#filter-option)

Note that the `DISTANCE` option is required, as it sets the edit distance
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
- **`DISTANCE`**
- `FUZZINESS`
- `DIST`
- `D`

so the following queries are the same:

```
(RAW_TEXT CONTAINS EDIT_DISTANCE("albatross", DISTANCE=1))
(RAW_TEXT CONTAINS EDIT_DISTANCE("albatross", FUZZINESS=1))
(RAW_TEXT CONTAINS EDIT_DISTANCE("albatross", DIST=1))
(RAW_TEXT CONTAINS EDIT_DISTANCE("albatross", D=1))
```

## `REDUCE` option

`REDUCE` determines whether or not the results will be reduced, thereby
eliminating duplicate matches. Duplicate matches are common given how the edit
distance’s Levenshtein algorithm works, where it counts insertions, deletions
and replacements when calculating fuzziness for the match. For example, if you
are searching for `"giraffe"` with distance less than or equal to one, and the
input corpus is `"That giraffe is tall,"` then three results would be generated:
1. `" giraffe"`
2. `"giraffe"`
3. `"iraffe"`

The first inserts one space before, so that is a match with distance one.

The second is an obvious exact match (distance of zero, which is less than one).

The third is missing the leading `'g'`, for a distance of one.

All three are reported. With this option set to `true`,
only one match will be reported: `"giraffe"`.

Note, although `REDUCE=false` is used by default it is usually overriden
by global `ryft-server`'s `reduce=true` option!

The following aliases can be used to specify `REDUCE` as well:
- **`REDUCE`**
- `R`

so the following queries are the same:

```
(RAW_TEXT CONTAINS EDIT_DISTANCE("albatross", D=1, REDUCE=true))
(RAW_TEXT CONTAINS EDIT_DISTANCE("albatross", D=1, R=true))
```

Please see [boolean type](./README.md#booleans) to get the ways
the `REDUCE` option can be set.


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
- **`WIDTH`**
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
- **`LINE`**
- `L`

so the following queries are the same:

```
(RAW_TEXT CONTAINS EDIT_DISTANCE("albatross", D=1, LINE=true))
(RAW_TEXT CONTAINS EDIT_DISTANCE("albatross", D=1, L=true))
```

Please see [boolean type](./README.md#booleans) to get the ways
the `LINE` option can be set.


## `CASE` option

When `CASE` is `false`, the query will be run caseinsensitive.

`CASE=true` is used by default, so searches are case sensitive.

The following aliases can be used to specify `LINE` as well:
- **`CASE`**
- `CS`

so the following queries are the same:

```
(RAW_TEXT CONTAINS EDIT_DISTANCE("albatross", D=1, CASE=false))
(RAW_TEXT CONTAINS EDIT_DISTANCE("albatross", D=1, CS=false))
```

Please see [boolean type](./README.md#booleans) to get the ways
the `CASE` option can be set.


# See Also

- [Exact search](./EXACT.md)
- [Fuzzy Hamming search](./HAMMING.md)
- [Date search](./DATE.md)
- [Time search](./TIME.md)
- [Number search](./NUMBER.md)
- [Currency search](./CURRENCY.md)
- [IPv4 search](./IPV4.md)
- [IPv6 search](./IPV6.md)
- [Regexp search](./PCRE2.md)
