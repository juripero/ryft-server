The Currency Search operation follows largely the same rules as the
[Number Search](./NUMBER.md) operation.
The parsing protocol diagram changes slightly to allow for configurable currency
identifiers, such as `"$"` for US currency. In addition, for currency, negative
amounts are parsed using either the minus sign or parenthetical `"()"` notation.

Currency Searches are a type of number searches
and extend the general relational expression defined
[previously](./README.md#general-search-syntax) as follows:

```
(input_specifier relational_operator CURRENCY(expression, options))
```

Different currency ranges can be searched for by modifying the expression provided.
There are two general expression types supported:

- `CUR operator1 "ValueA"`
- `"ValueA" operator1 CUR operator2 "ValueB"`

The box below contains a full list of the supported expressions.
`ValueA` and `ValueB` represent the currency values to compare the input data against.

- `CUR = "ValueA"`
- `CUR != "ValueA"` (Not equals operator)
- `CUR >= "ValueA"`
- `CUR > "ValueA"`
- `CUR <= "ValueA"`
- `CUR < "ValueA"`
- `"ValueA" <= CUR <= "ValueB"`
- `"ValueA" < CUR < "ValueB"`
- `"ValueA" < CUR <= "ValueB"`
- `"ValueA" <= CUR < "ValueB"`

Please note, the `ValueA` and `ValueB` can be quoted or unquoted.
In addition to `ryftprim` the following formats are supported:
- `CUR == "ValueA"`
- `ValueA >= CUR >= ValueB`
- `ValueA > CUR > ValueB`
- `ValueA > CUR >= ValueB`
- `ValueA >= CUR > ValueB`

As an example of a fully qualified currency search relational expression,
to find all values using the US currency system between but not including
`"$450"` and `"$10,100.50"` in a record/field construct where the field
tag is price, use:

```
(RECORD.price CONTAINS CURRENCY("$450" < CUR < "$10,100.50",
  SYMBOL="$", SEPARATOR=",", DECIMAL="."))
```

The results will contain all prices encountered in the input data that match
the specified range. For example, if a price `"$692.01"` appears in the input
data, then it will be reported as a match since it falls within the specified
range. But currency values like `-$123.00`, `$449.99` and `$100,000` would
not match.

The following aliases can be used to specify `CURRENCY` primitive as well:
- `CURRENCY`
- `MONEY`

so the following queries are the same:

```
(RECORD.price CONTAINS CURRENCY("$450" < CUR < "$10,100.50",
  SYMBOL="$", SEPARATOR=",", DECIMAL="."))
(RECORD.price CONTAINS MONEY("$450" < CUR < "$10,100.50",
  SYMBOL="$", SEPARATOR=",", DECIMAL="."))
```


# Options

The available comma-separated options for the `CURRENCY` primitive are:

- [SYMBOL](#symbol-option)
- [SEPARATOR](#separator-option)
- [DECIMAL](#decimal-option)
- [WIDTH](#width-option)
- [LINE](#line-option)


## `SYMBOL` option

The required `SYMBOL` is defined as the monetary symbol to be used in the
state machine which starts a currency match state machine. For example,
for standard US currency, a dollar sign would be specified as `"$"`.

The following aliases can be used to specify `SYMBOL` as well:
- `SYMBOL`
- `SYMB`
- `SYM`

so the following queries are the same:

```
(RECORD.price CONTAINS CURRENCY(450<CUR, SYMBOL="$", SEPARATOR=",", DECIMAL="."))
(RECORD.price CONTAINS CURRENCY(450<CUR, SYMB="$", SEPARATOR=",", DECIMAL="."))
(RECORD.price CONTAINS CURRENCY(450<CUR, SYM="$", SEPARATOR=",", DECIMAL="."))
```


## `SEPARATOR` option

The required `SEPARATOR` is defined as the separating character to use
in the state machine. For example, for standard US currency, a comma would
be specified. If other types of currencies are being searched, perhaps a
period would need to be specified. For example, US `$1,000.00` might be
written in certain European markets as `"$1.000,00"`.

Note that the `SEPARATOR` does not need to appear in the data stream,
but if it does appear, it will be considered part of the value being parsed.

Note that the `SEPARATOR` and `DECIMAL` must be different. If the same
character is specified for both, an error message will be generated.

The following aliases can be used to specify `SEPARATOR` as well:
- `SEPARATOR`
- `SEP`

so the following queries are the same:

```
(RECORD.id CONTAINS CURRENCY(450<CUR, SYM="$", SEPARATOR=",", DECIMAL="."))
(RECORD.id CONTAINS CURRENCY(450<CUR, SYM="$", SEP=",", DECIMAL="."))
```


## `DECIMAL` option

The required `DECIMAL` is defined as the decimal specifier to use. For example,
for standard US numbers, a period (decimal point) would be specified.
If other types of currencies are being searched, perhaps a comma would need
to be specified. For example, US `$1,000.00` might be written in certain
European markets as `"$1.000,00"`.

Note that the `DECIMAL` does not need to appear in the data stream,
but if it does appear, it will be considered part of the value being parsed.

Note that the `DECIMAL` and `SEPARATOR` must be different. If the same
character is specified for both, an error message will be generated.

The following aliases can be used to specify `DECIMAL` as well:
- `DECIMAL`
- `DEC`

so the following queries are the same:

```
(RECORD.id CONTAINS CURRENCY(450<CUR, SYM="$", SEP=",", DECIMAL="."))
(RECORD.id CONTAINS CURRENCY(450<CUR, SYM="$", SEP=",", DEC="."))
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
(RECORD CONTAINS CURRENCY(450<CUR, SYM="$", SEP=",", DEC=".", SURROUNDING=5))
(RECORD CONTAINS CURRENCY(450<CUR, SYM="$", SEP=",", DEC=".", WIDTH=5))
(RECORD CONTAINS CURRENCY(450<CUR, SYM="$", SEP=",", DEC=".", W=5))
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
(RECORD CONTAINS CURRENCY(450<CUR, SYM="$", SEP=",", DEC=".", LINE=true))
(RECORD CONTAINS CURRENCY(450<CUR, SYM="$", SEP=",", DEC=".", L=true))
```

Please see [boolean type](./README.md#booleans) to get the ways
the `LINE` option can be set.


# Compatible syntax

For the backward compatibility a few positional options are supported.
Names for the `SYMBOL`, `SEPARATOR` and `DECIMAL` can be omitted:

```
(input_specifier relational_operator CURRENT(expression, symbol, separator, decimal))
```

Although it is not recommended to use this syntax, it is valid to pass the
following queries:

```
(RECORD CONTAINS CURRENCY(450<CUR, "$", ",", "."))
(RECORD CONTAINS CURRENCY(450<CUR, "$", ",", ".", W=5))
```

For all cases the `SYMBOL="$"`, `SEPARATOR=","` and `DECIMAL="."`.
