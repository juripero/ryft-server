The Number Search operation can be used to search for exact numbers or numbers
in a particular range for both structured and unstructured input data.

The protocol rules are:
- No line breaks anywhere.
- The maximum number of characters is 64.
- Whitespace is not permitted within the number.
- The maximum number of consecutive digits is 16 base-10 characters.
   This means that numeric accuracy is compromised on numbers outside
   the range of approximately `-2^53` and `+2^53`.
- The exponent value ranges when scientific notation is used are `-512` and `+512`.
- Numeric separators are not permitted within a specified exponent.
- Infinity and "not a number" (`NaN`) are not accepted as numbers.
- Leading zeros are automatically handled as valid as part of the digit portion of the number.

When a series of characters violates these parsing rules, the result
will be the last valid sequence within the series, which allows for
partial matches, and can be an excellent tool for analyzing
potentially dirty data.

This general solution allows for numbers to be represented in arbitrary forms, including:

- Configurable separators, such as perhaps specifying a comma representing
   a thousands separator for the US number system (e.g., `"7,000"`),
   or specifying a dash separating fields in a phone number or a social security
   number (e.g., `"1-800-555-1212"` or `"123-45-6789"`).
- Configurable decimals, such as perhaps specifying a period to represent
   the decimal for the US number system (e.g., `"7.2"`).
- A minus sign to specify a negative value, such as `"-7.2"`.
- Scientific notation, such as `"-2e3"`, `"-2e-3"`, `"-2.2E+2"`, etc.
- Data results returned for a particular number matching specified criteria will
   be truncated after a total of 64 characters.

Number Searches extend the general relational expression defined
[previously](./README.md#general-search-syntax) as follows:

```
(input_specifier relational_operator NUMBER(expression, options))
```

Different number ranges can be searched for by modifying the expression provided.
There are two general expression types supported:

- `NUM operator1 "ValueA"`
- `"ValueA" operator1 NUM operator2 "ValueB"`

The box below contains a full list of the supported expressions.
`ValueA` and `ValueB` represent the numbers to compare the input data against.

- `NUM = "ValueA"`
- `NUM != "ValueA"` (Not equals operator)
- `NUM >= "ValueA"`
- `NUM > "ValueA"`
- `NUM <= "ValueA"`
- `NUM < "ValueA"`
- `"ValueA" <= NUM <= "ValueB"``
- `"ValueA" < NUM < "ValueB"`
- `"ValueA" < NUM <= "ValueB"`
- `"ValueA" <= NUM < "ValueB"`

Please note, the `ValueA` and `ValueB` can be quoted or unquoted.
In addition to `ryftprim` the following formats are supported:
- `NUM == "ValueA"`
- `ValueA >= NUM >= ValueB`
- `ValueA > NUM > ValueB`
- `ValueA > NUM >= ValueB`
- `ValueA >= NUM > ValueB`

As an example of a fully qualified number search relational expression,
to find all matching numbers using the US number system between
but not including `"1025"` and `"1050"` in a record/field construct
where the field tag is id, use:

```
(RECORD.id CONTAINS NUMBER("1025" < NUM < "1050", SEPARATOR=",", DECIMAL="."))
```

The results will contain all numbers that are encountered in the input data that
match the specified range. For example, if a scientific notation number `"1.026e3"`
appears in the input data (which expands to `"1,026"`), then it will be reported
as a match since it falls within the specified range. Similarly, the
numbers `1049` and `1,049.9` would also match. However, the number `-1,234`
would not be a match, as it does not fall within the requested range,
nor would the number `+1,050.00001`.

The following aliases can be used to specify `NUMBER` primitive as well:
- **`NUMBER`**
- `NUMERIC`

so the following queries are the same:

```
(RECORD.id CONTAINS NUMBER("1025" < NUM < "1050", SEPARATOR=",", DECIMAL="."))
(RECORD.id CONTAINS NUMERIC("1025" < NUM < "1050", SEPARATOR=",", DECIMAL="."))
```


# Options

The available comma-separated options for the `NUMBER` primitive are:

- [SEPARATOR](#separator-option)
- [DECIMAL](#decimal-option)
- [WIDTH](#width-option)
- [LINE](#line-option)
- [FILTER](./README.md#filter-option)


## `SEPARATOR` option

The required `SEPARATOR` is defined as the separating character to use.
For example, for standard US numbers, a comma would be specified.
If other types of numbers are being searched, such as perhaps phone numbers,
than a dash would be specified.

Note that the `SEPARATOR` does not need to appear in the data stream,
but if it does appear, it will be considered part of the value being parsed.

Note that the `SEPARATOR` and `DECIMAL` must be different. If the same
character is specified for both, an error message will be generated.

The following aliases can be used to specify `SEPARATOR` as well:
- **`SEPARATOR`**
- `SEP`

so the following queries are the same:

```
(RECORD.id CONTAINS NUMBER(1025<NUM<1050, SEPARATOR=",", DECIMAL="."))
(RECORD.id CONTAINS NUMBER(1025<NUM<1050, SEP=",", DECIMAL="."))
```

The `SEPARATOR=","` is used by default, i.e. if nothing is provided.
This behaviour differs from `ryftprim`!


## `DECIMAL` option

The required `DECIMAL` is defined as the decimal specifier to use. For example,
for standard US numbers, a period (decimal point) would be specified.

Note that the `DECIMAL` does not need to appear in the data stream,
but if it does appear, it will be considered part of the value being parsed.

Note that the `DECIMAL` and `SEPARATOR` must be different. If the same
character is specified for both, an error message will be generated.

The following aliases can be used to specify `DECIMAL` as well:
- **`DECIMAL`**
- `DEC`

so the following queries are the same:

```
(RECORD.id CONTAINS NUMBER(1025<NUM<1050, SEP=",", DECIMAL="."))
(RECORD.id CONTAINS NUMBER(1025<NUM<1050, SEP=",", DEC="."))
```

The `DECIMAL="."` is used by default, i.e. if nothing is provided.
This behaviour differs from `ryftprim`!


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
(RECORD CONTAINS NUMBER(1025<NUM<1050, SEP=",", DEC=".", SURROUNDING=5))
(RECORD CONTAINS NUMBER(1025<NUM<1050, SEP=",", DEC=".", WIDTH=5))
(RECORD CONTAINS NUMBER(1025<NUM<1050, SEP=",", DEC=".", W=5))
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
(RECORD CONTAINS NUMBER(1025<NUM<1050, SEP=",", DEC=".", LINE=true))
(RECORD CONTAINS NUMBER(1025<NUM<1050, SEP=",", DEC=".", L=true))
```

Please see [boolean type](./README.md#booleans) to get the ways
the `LINE` option can be set.


# Compatible syntax

For the backward compatibility a few positional options are supported.
Names for the `SEPARATOR` and `DECIMAL` can be omitted:

```
(input_specifier relational_operator NUMBER(expression, separator, decimal))
```

Although it is not recommended to use this syntax, it is valid to pass the
following queries:

```
(RECORD CONTAINS NUMBER(1025<NUM<1050, ",", "."))
(RECORD CONTAINS NUMBER(1025<NUM<1050, ",", ".", W=5))
```

For all cases the `SEPARATOR=","` and `DECIMAL="."`.


# See Also

- [Exact search](./EXACT.md)
- [Fuzzy Hamming search](./HAMMING.md)
- [Fuzzy Edit Distance search](./EDIT_DIST.md)
- [Date search](./DATE.md)
- [Time search](./TIME.md)
- [Currency search](./CURRENCY.md)
- [IPv4 search](./IPV4.md)
- [IPv6 search](./IPV6.md)
- [Regexp search](./PCRE2.md)
