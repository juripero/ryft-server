This document contains short description of search syntax used for
various search types:

- [Exact search](./EXACT.md)
- [Fuzzy Hamming search](./HAMMING.md)
- [Fuzzy Edit Distance search](./EDIT_DIST.md)
- [Date search](./DATE.md)
- [Time search](./TIME.md)
- [Number search](./NUMBER.md)
- [Currency search](./CURRENCY.md)
- [IPv4 search](./IPV4.md)
- [IPv6 search](./IPV6.md)


# General search syntax

A match criteria `query` parameter is used to specify how the search should be performed.
Search criteria are made up of one or more relational expressions, connected
using logical operations. The Ryft Open API defines a query language grammar,
consisting of a relational expression which takes the following form:

```
(input_specifier relational_operator primitive(expression[, options]))
```

`input_specifier` specifies how the input data is arranged. The possible values are:

- `RAW_TEXT` - The input is a sequence of raw bytes with no implicit formatting or grouping.
- `RECORD` - The input is a series of records. Search all records.
- `RECORD.<field_name>` - The input is a series of records.
   Search only the field called `<field_name>` in each record.
   Note: for JSON input records, multiple field names can be specified
   with `'.'` separators between them to specify a field hierarchy,
   or with `'[]'` separators to specify array hierarchy.


`relational_operator` specifies how the input relates to the expression. The possible values are:

- `EQUALS` - The input must match expression either exactly for an exact search
   or within the specified distance for a fuzzy search,
   with no additional leading or trailing data. Note that this
   operator has meaning only for record- and field-based searches.
   If used with raw text input, an error will be generated.
   When searching raw text data, `CONTAINS` should be used instead of `EQUALS`.
- `NOT_EQUALS` - The input must be anything other than expression.
   Note that this operator has meaning only for record- and field-based
   searches. If used with raw text input, an error will be generated. 
- `CONTAINS` - The input must contain expression,
   and may contain additional leading or trailing data.
- `NOT_CONTAINS` The input must not contain expression. Note that this
   operator has meaning only for record- and field-based searches.
   If used with raw text input, an error will be generated.


`primitive` specifies the search primitive associated with the clause. The possible values are:

- [EXACT](./EXACT.md) - Search for an exact match.
- [HAMMING](./HAMMING.md) - Perform a fuzzy search using the Hamming distance algorithm.
- [EDIT_DISTANCE](./EDIT_DIST.md) - Perform a fuzzy search using the edit distance (Levenshtein) algorithm.
- [DATE](./DATE.md) - Search for a date or a range of dates.
- [TIME](./TIME.md) - Search for a time or a range of times.
- [NUMBER](./NUMBER.md) - Search for a number or a range of numbers.
- [CURRENCY](./CURRENCY.md) - Search for a monetary value or a range of monetary values.
- [IPV4](./IPV4.md) - Search for an IPv4 address or a range of IPv4 addresses.
- [IPV6](./IPV6.md) - Search for an IPv6 address or a range of IPv6 addresses.


`expression` specifies the expression to be matched. The possible values are:

- Quoted string - Any valid C language string, including backslash-escaped
   characters. For example, `"match this text\n"`. This can also
   include escaped hexadecimal characters, such as `"match this text\x0A"`,
   or `"\x48\x69\x20\x54\x68\x65\x72\x65\x21\x0A\x00"`.
   If a backslash needs to be placed in the quoted string for search
   query purposes, use the double backsplash escape sequence `"\\"`.
- Wildcard - A `"?"` character is used to denote that any **single** character will match.
   A `"?"` can be inserted at any point(s) **between** quoted strings.
   For example, `"match th"?"s text\n"`.
- Any combination of the above - For example, `"match\x20th"?"s text\x0A"`,
   or `"match\x20with a wildcard right here"?"and a null at the end\x00"`.

`options` specify a comma-separated list of options that can further qualify
the request for certain primitives. The possible values are specific for
each search type.

Note that it is permissible to include valid but extraneous options,
in which case they will be ignored. For example, if a `DISTANCE` options
is specified with an `EXACT` primitive, the `DISTANCE` option will
be ignored and the search will still run. 

`logical_operator` allows for complex collections of relational expressions. The possible values are:

- `AND` The logical expression `(a AND b)` evaluates to true only if both the
   relational expression `a` evaluates to true and the relational
   expression `b` evaluates to true.
- `OR` The logical expression `(a OR b)` evaluates to true if either the
   relational expression `a` evaluates to true or the relational expression
   `b` evaluates to true.
- `XOR` The logical expression `(a XOR b)` evaluates to true if either the
   relational expression `a` evaluates to true or the relational expression
   `b` evaluates to true, but not both.

Multiple relational expressions can be combined using the logical
operators `AND`, `OR`, and `XOR`. For example:

```
(RECORD.city EQUALS EXACT("Rockville")) AND (RECORD.state EQUALS EXACT("MD"))
```

Parentheses can also be used to control the precedence of operations.
Additional whitespace is allowable, which can simplify comprehension.
For example:

```
( (RECORD.city EQUALS EXACT("Rockville")) OR (RECORD.city EQUALS EXACT("Gaithersburg")) )
 AND (RECORD.state EQUALS EXACT("MD"))
```

# Option types

Each search options is of specific type:
- [Integer](#Integers)
- [Boolean](#Booleans)
- [String](#Strings)


## Integers

Values can be quoted or unquoted.

Examples: `W=3`, `W="5"`.


## Booleans

Boolean values can be parsed from:

- "1", "t", "T", "true", "TRUE", "True" as `true`
- "0", "f", "F", "false", "FALSE", "False" as `false`

Note, the double quotes can be omitted.

Examples: `L=true`, `L=1`, `L="T"`, `L=False`.

The booleans can be also defined as just names with
optional negation: `L, !CS` means `L=true, CS=false`.


## Strings

Values should be quoted.

Examples, `SYMBOL="$"`


# Generic options

Some options are supported by each search type. For example `WIDTH` or `LINE`.
Some options are used by ryft REST server internally such as `FILTER` in case
of search in catalogs.


## `FILTER` option

`FILTER` specifies a regular expression as a [string value](./README.md#strings).
This option is used with catalog's only. It filters resulting catalog file parts
by name.

For example if catalog contains many `*.txt` text and `*.bin` binary file parts
then the results can be narrowed down by corresponding regular expression:
- `FILTER=".*\.txt"` for the text files
- `FILTER=".*\.bin"` for the binary files

Any regular expression can be used to specify complex filename filtering rules
like date ranging etc.

`FILTER=""` is used by default. Empty filter means **all** file parts.

The following aliases can be used to specify `FILTER` as well:
- `FILE_FILTER`
- `FILTER`
- `FF`

so the following queries are the same:

```
(RECORD CONTAINS EXACT("orange", FILE_FILTER=".*txt"))
(RECORD CONTAINS EXACT("orange", FILTER=".*txt"))
(RECORD CONTAINS EXACT("orange", FF=".*txt"))
```
