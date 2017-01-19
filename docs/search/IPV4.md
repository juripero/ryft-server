The IPv4 Search operation can be used to search for exact IPv4 addresses or
IPv4 addresses in a particular range in both structured and unstructured text
using the standard "a.b.c.d" format for IPv4 addresses.

IPv4 Searches extend the general relational expression defined
[previously](./README.md#general-search-syntax) as follows:

```
(input_specifier relational_operator IPV4(expression[, options]))
```

Different ranges can be searched for by modifying the expression in the
relational expression above. There are two general expression types supported:

- `IP operator "ValueB"`
- `"ValueA" operator IP operator "ValueB"`

The box below contains of list of supported expressions.
`ValueA` and `ValueB` represent the IP addresses to compare the input data against.

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

For example, to find all IP addresses greater than `10.11.12.13`,
use the following search query criteria:

```
(RAW_TEXT CONTAINS IPV4(IP > "10.11.12.13"))
```

To find all matching IPv4 addresses adhering to `10.10.0.0/16` (that is, all
IP addresses from `10.10.0.0` through `10.10.255.255` inclusive) in a
record/field construct where the field tag is `ipaddr`, use:

```
(RECORD.ipaddr CONTAINS IPV4("10.10.0.0" <= IP <= "10.10.255.255"))
```


# Options

The available comma-separated options for the `IPV4` primitive are:

- [OCTAL](#octal-option)
- [WIDTH](#width-option)
- [LINE](#line-option)
- [FILTER](./README.md#filter-option)


## `OCTAL` option

The default behavior of the `IPv4` search primitive is to parse all octets
encountered as decimal. When option `OCTAL=true` is specified, any octets
encountered with leading `0`'s will be parsed as octal quantities instead
of decimal quantities for comparison purposes.

This can be important in certain cyber analysis scenarios, where obfuscation
attempts using octal notation can sometimes fool contemporary systems into
decoding incorrect IP addresses if they do not properly parse octal quantities.
This is also useful when it is known that certain input datasets may not be
strictly adhering to a certain type of leading zero parsing rule.

`OCTAL=0` is used by default.

The following aliases can be used to specify `OCTAL` as well:
- `USE_OCTAL`
- **`OCTAL`**
- `OCT`

so the following queries are the same:

```
(RAW_TEXT CONTAINS IPV4(IP > "10.11.12.13", USE_OCTAL))
(RAW_TEXT CONTAINS IPV4(IP > "10.11.12.13", OCTAL=true))
(RAW_TEXT CONTAINS IPV4(IP > "10.11.12.13", OCT=true))
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
(RAW_TEXT CONTAINS IPV4(IP > "10.11.12.13", SURROUNDING=5))
(RAW_TEXT CONTAINS IPV4(IP > "10.11.12.13", WIDTH=5))
(RAW_TEXT CONTAINS IPV4(IP > "10.11.12.13", W=5))
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
(RAW_TEXT CONTAINS IPV4(IP > "10.11.12.13", LINE=true))
(RAW_TEXT CONTAINS IPV4(IP > "10.11.12.13", L=true))
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
- [IPv6 search](./IPV6.md)
