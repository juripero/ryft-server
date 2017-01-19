The `IPv6` Search operation can be used to search for exact IPv6 addresses or
IPv6 addresses in a particular range in both structured and unstructured text
using the standard `"a:b:c:d:e:f:g:h"` format for IPv6 addresses.
The double colon (::) is also supported, per RFC guidelines.

IPv6 searches extend the general relational expression defined
[previously](./README.md#general-search-syntax) as follows:

```
(input_specifier relational_operator IPV6(expression[, options]))
```

Different ranges can be searched for by modifying the expression in the relational
expression above. There are two general expression types supported:

- `IP operator "ValueB"`
- `"ValueA" operator IP operator "ValueB"`

The box below contains of list of supported expressions. `ValueA` and `ValueB`
represent the IP addresses to compare the input data against.

- `IP = "ValueB"`
- `IP != "ValueB"` (Not equals operator)
- `IP >= "ValueB"`
- `IP > "ValueB"`
- `IP <= "ValueB"`
- `IP < "ValueB"`
- `"ValueA" <= IP <= "ValueB"`
- `"ValueA" < IP < "ValueB"`
- `"ValueA" < IP <= "ValueB"`
- `"ValueA" <= IP < "ValueB"`

In addition to `ryftprim` the following formats are supported:
- `IP == "ValueB"`
- `ValueA >= IP >= ValueB`
- `ValueA > IP > ValueB`
- `ValueA > IP >= ValueB`
- `ValueA >= IP > ValueB`

For example, to find all IP addresses greater than `1abc:2::8`,
use the following search query criteria:

```
(RAW_TEXT CONTAINS IPV6(IP > "1abc:2::8"))
```

To find all matching IPv6 addresses between `10::1` and `10::1:1`, inclusive,
in a record/field construct where the field tag is `ipaddr6`, use:

```
(RECORD.ipaddr6 CONTAINS IPV6("10::1" <= IP <= "10::1:1"))
```


# Options

The available comma-separated options for the `IPV6` primitive are:

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
- **`WIDTH`**
- `W`

so the following queries are the same:

```
(RAW_TEXT CONTAINS IPV6(IP > "1abc:2::8", SURROUNDING=5))
(RAW_TEXT CONTAINS IPV6(IP > "1abc:2::8", WIDTH=5))
(RAW_TEXT CONTAINS IPV6(IP > "1abc:2::8", W=5))
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
(RAW_TEXT CONTAINS IPV6(IP > "1abc:2::8", LINE=true))
(RAW_TEXT CONTAINS IPV6(IP > "1abc:2::8", L=true))
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
