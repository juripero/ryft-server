The Date Search operation allows for exact date searches and for searching
for dates within a range for both structured and unstructured data using
the following date formats:

- `YYYY/MM/DD`
- `YY/MM/DD`
- `DD/MM/YYYY`
- `DD/MM/YY`
- `MM/DD/YYYY`
- `MM/DD/YY`

Note: The `"/"` character in the above list of formats can be replaced by any
other single character delimiter. For example, `YYYY-MM-DD` and `MM_DD_YYYY`
are both acceptable date formats.

Date searches extend the general relational expression defined
[previously](./README.md#general-search-syntax) as follows:

```
(input_specifier relational_operator DATE(expression[, options]))
```

Different date ranges can be searched for by modifying the expression provided.
There are two general expression types supported:

- `DATE(DateFormat operator ValueB)`
- `DATE(ValueA operator DateFormat operator ValueB)`

The box below contains a full list of the supported expressions.
`DateFormat` represents the format of the dates to search for
and `ValueA` and `ValueB` represent dates to compare input data against.

- `DateFormat = ValueB`
- `DateFormat != ValueB` (Not equals operator)
- `DateFormat >= ValueB`
- `DateFormat > ValueB`
- `DateFormat <= ValueB`
- `DateFormat < ValueB`
- `ValueA <= DateFormat <= ValueB`
- `ValueA < DateFormat < ValueB`
- `ValueA < DateFormat <= ValueB`
- `ValueA <= DateFormat < ValueB`

For example, to find all dates after `"02/28/12"` in unstructured raw text,
use the following search query criteria:

```
(RAW_TEXT CONTAINS DATE(MM/DD/YY > 02/28/12))
```

To find all matching dates between `"02/28/12"` and `"01/19/15"` but not including
those two dates in a record/field construct where the field tag is date, use:

```
(RECORD.date CONTAINS DATE(02/28/12 < MM/DD/YY < 01/19/15))
```

Please note, the `ValueA` and `ValueB` can be quoted or unquoted.
In addition to `ryftprim` the following formats are supported:
- `DateFormat == ValueB`
- `ValueA >= DateFormat >= ValueB`
- `ValueA > DateFormat > ValueB`
- `ValueA > DateFormat >= ValueB`
- `ValueA >= DateFormat > ValueB`


# Options

The available comma-separated options for the `DATE` primitive are:

- [WIDTH](#width-option)
- [LINE](#line-option)
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
(RECORD CONTAINS DATE(MM/DD/YY > 02/28/12, SURROUNDING=5))
(RECORD CONTAINS DATE(MM/DD/YY > 02/28/12, WIDTH=5))
(RECORD CONTAINS DATE(MM/DD/YY > 02/28/12, W=5))
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
(RECORD CONTAINS DATE(MM/DD/YY > 02/28/12, LINE=true))
(RECORD CONTAINS DATE(MM/DD/YY > 02/28/12, L=true))
```

Please see [boolean type](./README.md#booleans) to get the ways
the `LINE` option can be set.


# See Also

- [Exact search](./EXACT.md)
- [Fuzzy Hamming search](./HAMMING.md)
- [Fuzzy Edit Distance search](./EDIT_DIST.md)
- [Time search](./TIME.md)
- [Number search](./NUMBER.md)
- [Currency search](./CURRENCY.md)
- [IPv4 search](./IPV4.md)
- [IPv6 search](./IPV6.md)
