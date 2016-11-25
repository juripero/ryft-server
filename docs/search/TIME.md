Similar to the [Date Search](./DATE.md), the Time Search operation can be
used to search for exact times or for times within a particular range
for both structured and unstructured data using the following time formats:

- `HH:MM:SS`
- `HH:MM:SS:ss`

Note: `'ss'` in the second format indicates hundredths of a second,
and if specified should always be two digits.

Time searches extend the general relational expression defined
[previously](./README.md#general-search-syntax) as follows:

```
(input_specifier relational_operator TIME(expression[, options]))
```

Similar to the Date Search, different time ranges can be searched for by modifying
the expression in the relational expression above. Again, there are two general
expression types supported:

- `TIME(TimeFormat operator ValueB)`
- `TIME(ValueA operator TimeFormat operator ValueB)`

The box below contains of list of these supported expressions. `TimeFormat` represents
the format of the times to search for and `ValueA` and `ValueB` represent times
to compare input data against.

- `TimeFormat = ValueB`
- `TimeFormat != ValueB` (Not equals operator)
- `TimeFormat >= ValueB`
- `TimeFormat > ValueB`
- `TimeFormat <= ValueB`
- `TimeFormat < ValueB`
- `ValueA <= TimeFormat <= ValueB`
- `ValueA < TimeFormat < ValueB`
- `ValueA < TimeFormat <= ValueB`
- `ValueA <= TimeFormat < ValueB`

For example, to find all times after `"09:15:00"`, use the following search query criteria:

```
(RAW_TEXT CONTAINS TIME(HH:MM:SS > 09:15:00))
```

To find all matching times between `"11:15:00"` and `"13:15:00"` but not including
those times in a record/field construct where the field tag is time, use:

```
(RECORD.time CONTAINS TIME(11:15:00 < HH:MM:SS < 13:15:00))
```

Please note, the `ValueA` and `ValueB` can be quoted or unquoted.
In addition to `ryftprim` the following formats are supported:
- `TimeFormat == ValueB`
- `ValueA >= TimeFormat >= ValueB`
- `ValueA > TimeFormat > ValueB`
- `ValueA > TimeFormat >= ValueB`
- `ValueA >= TimeFormat > ValueB`


# Options

The available comma-separated options for the `TIME` primitive are:

- [WIDTH](#width-option)
- [LINE](#line-option)


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
(RECORD CONTAINS TIME(HH:MM:SS > 09:15:00, SURROUNDING=5))
(RECORD CONTAINS TIME(HH:MM:SS > 09:15:00, WIDTH=5))
(RECORD CONTAINS TIME(HH:MM:SS > 09:15:00, W=5))
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
(RECORD CONTAINS TIME(HH:MM:SS > 09:15:00, LINE=true))
(RECORD CONTAINS TIME(HH:MM:SS > 09:15:00, L=true))
```

Please see [boolean type](./README.md#booleans) to get the ways
the `LINE` option can be set.
