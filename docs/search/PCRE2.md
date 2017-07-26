The `PCRE2` search performs a search that adheres strictly to the PCRE2 regular
expression rules. Ryft supports the totality of the PCRE2 specification as described
[here](http://www.pcre.org/current/doc/html/pcre2syntax.html) as of June 5, 2017.

`PCRE2` is a standards-based regular expression format that is heavily used by the
search and analytics community for a variety of important search use cases,
including cyber use cases. `PCRE2` searches extend the general relational
expression defined [previously](./README.md#general-search-syntax) as follows:

```
(input_specifier relational_operator PCRE2(expression[, options]))
```

The following aliases can be used to specify `PCRE2` primitive as well:
- **`PCRE2`**
- `REGEXP`
- `REGEX`
- `RE`

so the following queries are the same:

```
(RECORD CONTAINS PCRE2("(?i)(orange|apple)"))
(RECORD CONTAINS REGEX("(?i)(orange|apple)"))
(RECORD CONTAINS REGEX("(?i)orange|apple)"))
```

# Compatible syntax

For the backward compatibility the term `PCRE2` can be omitted:

```
(input_specifier relational_operator expression)
```

Note, there is no way to specify additional options in this syntax!
The global options will be used.


# Options

The available comma-separated options for the `PCRE2` primitive are:

- [WIDTH](#width-option)
- [LINE](#line-option)
- [FILTER](./README.md#filter-option)


Note that many query-related options, such as case-insensitive options,
are not supported in the options field, but are instead supported through
standard `PCRE2` syntax internal to the expression itself.

In addition, similar to exact search, the WIDTH and LINE options can aid in
downstream analysis of contextual use of the match results against unstructured
raw text data. A fully qualified `PCRE2` clause looking for various spellings
of the often-misspelled word "misspell" in case-insensitive fashion allowing
for one or more internal `'s'` characters would be:

```
(RAW_TEXT CONTAINS PCRE2("(?i)mis+pell"))
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
- **`WIDTH`**
- `W`

so the following queries are the same:

```
(RECORD CONTAINS PCRE2("(orange|apple)", SURROUNDING=5))
(RECORD CONTAINS PCRE2("(orange|apple)", WIDTH=5))
(RECORD CONTAINS PCRE2("(orange|apple)", W=5))
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
(RECORD CONTAINS PCRE2("(orange|apple)", LINE=true))
(RECORD CONTAINS PCRE2("(orange|apple)", L=true))
```

Please see [boolean type](./README.md#booleans) to get the ways
the `LINE` option can be set.

# See Also

- [Exact search](./EXACT.md)
- [Fuzzy Hamming search](./HAMMING.md)
- [Fuzzy Edit Distance search](./EDIT_DIST.md)
- [Date search](./DATE.md)
- [Time search](./TIME.md)
- [Number search](./NUMBER.md)
- [Currency search](./CURRENCY.md)
- [IPv4 search](./IPV4.md)
- [IPv6 search](./IPV6.md)
