This document contains short reference on search expression syntax from Ryft Open API.

Ryft supports several search modes:

- `es` for [exact search](#exact-search)
- `fhs` for [fuzzy hamming search](#fuzzy-hamming-search)
- `feds` for [fuzzy edit distance search](#fuzzy-edit-distance-search)
- `ds` for [date search](#date-search)
- `ts` for [time search](#time-search)
- `ns` for [number](#number-search) or [currency search](#currency-search)
- `rs` for [regex search](#regex-search)
- `ipv4` for [IPv4 search](#ipv4-search)

# General search syntax

A match criteria `query` parameter is used to specify how the search should be performed.
Search criteria are made up of one or more relational expressions, connected
using logical operations. The Ryft Open API defines a query language grammar,
consisting of a relational expression which takes the following form:

```
(input_specifier relational_operator expression)
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
   or within the specified Hamming distance for a fuzzy search,
   with no additional leading or trailing data. Note that this
   operator has meaning only for record- and field-based searches.
   If used with a raw text operation, an error will be generated.
   When operating against raw text data, `CONTAINS` should be used.
- `NOT_EQUALS` - The input must be anything other than expression.
   Note that this operator has meaning only for record- and field-based
   searches. If used with a raw text operation, an error will be generated.
- `CONTAINS` - The input must contain expression, and may contain
   additional leading or trailing data.
- `NOT_CONTAINS` The input must not contain expression. Note that this
   operator has meaning only for record- and field-based searches.
   If used with a raw text operation, an error will be generated.

`expression` specifies the expression to be matched. The possible values are:

- Quoted string - Any valid C language string, including backslash-escaped
   characters. For example, `"match this text\n"`. This can also
   include escaped hexadecimal characters, such as `"match this text\x0A"`,
   or `"\x48\x69\x20\x54\x68\x65\x72\x65\x21\x0A\x00"`.
   If a backslash needs to be placed in the quoted string for search
   query purposes, use the double backsplash escape sequence
   `"\\"` so that it is escaped properly.
- Wildcard - A `"?"` character is used to denote that any single character will match.
   A `"?"` can be inserted at any point(s) between quoted strings.
   For example, `"match th"?"s text\n"`.
- Any combination of the above - For example, `"match\x20th"?"s text\x0A"`,
   or `"match\x20with a wildcard right here"?"and a null at the end\x00"`.

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
(RECORD.city EQUALS "Rockville") AND (RECORD.state EQUALS "MD")
```

Parentheses can also be used to control the precedence of operations. For example:

```
((RECORD.city EQUALS "Rockville") OR (RECORD.city EQUALS "Gaithersburg"))
 AND (RECORD.state EQUALS "MD")
```


# Exact search

The exact search operation will search existing_data_set using a match criteria.
The match string can be up to 32 bytes in length. Only exact matches will be returned.
When the format of the input data is specified as raw text, the `surrounding`
parameter specifies how many characters around the match will be returned as
part of the matched data results. The `surrounding` width can be useful
to assist a human analyst or a downstream machine learning tool to determine
the contextual use of a specific matched term when
searching unstructured raw text data.

See [general search syntax](#general-search-syntax).


# Fuzzy Hamming search

The fuzzy Hamming search operation works similarly to exact search except that
matches do not have to be exact. Instead, the `fuzziness` parameter allows the
specification of a "close enough" value to indicate how close the input must be to
match criteria. The match string can be up to 32 bytes in length.
A "close enough" match is specified as a Hamming distance.

The Hamming distance between two strings of equal length is the number of positions
at which the corresponding symbols are different. As provided to the fuzzy search
operation, the Hamming distance specifies the maximum number of substitutions
that are allowed in order to declare a match. In addition, similar to
exact search, the `surrounding` mechanism can aid in downstream analysis
of contextual use of the fuzzy match results against unstructured raw text data.

Fuzzy Hamming search algorithms are highly parallelizable, so they can run
extremely quickly on accelerated hardware. Therefore, if fuzzy search
algorithms can make use of fuzzy Hamming search, then they will typically
perform at very high speed.

See [general search syntax](#general-search-syntax).


# Fuzzy Edit Distance search

Fuzzy edit distance search performs a search that does not require two strings
to be of equal length to obtain a match. Instead of considering individual symbol
differences, fuzzy edit distance search counts the minimum number of insertions,
deletions and replacements required to transform one string into another.
This can make it much more powerful than fuzzy Hamming search for certain applications.

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

See [general search syntax](#general-search-syntax).


# Date search

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

Date searches extend the general relational expression defined previously as follows:

```
(input_specifier relational_operator DATE(expression))
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

# Time search

Similar to the Date Search, the Time Search operation can be used to search for
exact times or for times within a particular range for both structured and
unstructured data using the following time formats:

- `HH:MM:SS`
- `HH:MM:SS:ss`

Note: `'ss'` in the second format indicates hundredths of a second,
and if specified should always be two digits.

Time searches extend the general relational expression defined previously as follows:

```
(input_specifier relational_operator TIME(expression))
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


# Number search

The Number Search operation can be used to search for exact numbers or numbers
in a particular range for both structured and unstructured input data.

The protocol rules are:
- No line breaks anywhere.
- The maximum number of characters is 64.
- Whitespace is not permitted within the number.
- The maximum number of consecutive digits is 16 base-10 characters.
   This means that numeric accuracy is compromised on numbers outside
   the range of approximately -253 and +253.
- The exponent value ranges when scientific notation is used are -512 and +512.
- Numeric separators are not permitted within a specified exponent.
- Infinity and "not a number" (NaN) are not accepted as numbers.

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

Number Searches extend the general relational expression defined previously as follows:

```
(input_specifier relational_operator NUMBER(expression, subitizer, decimal))
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

The subitizer is defined as the separating character to use. For example,
for standard US numbers, a comma would be specified. If other types of numbers
are being searched, such as perhaps phone numbers, than a dash would be specified.

The decimal is defined as the decimal specifier to use. For example,
for standard US numbers, a period would be specified.

Note that the subitizer and decimal must be different. If the same character
is specified for both, an error message will be generated.

As an example of a fully qualified number search relational expression,
to find all matching numbers using the US number system between
but not including `"1025"` and `"1050"` in a record/field construct
where the field tag is id, use:

```
(RECORD.id CONTAINS NUMBER("1025" < NUM < "1050", ",", "."))
```

The results will contain all numbers that are encountered in the input data that
match the specified range. For example, if a scientific notation number `"1.026e3"`
appears in the input data (which expands to `"1,026"`), then it will be reported
as a match since it falls within the specified range. Similarly, the
numbers `1049` and `1,049.9` would also match. However, the number `-1,234`
would not be a match, as it does not fall within the requested range,
nor would the number `+1,050.00001`.


# Currency search

The Currency Search operation follows largely the same rules as the Number Search operation.
The parsing protocol diagram changes slightly to allow for configurable currency
identifiers, such as `"$"` for US currency. In addition, for currency, negative
amounts are parsed using either the minus sign or parenthetical `"()"` notation.

Currency Searches are a type of number searches and extend the general relational
expression defined previously as follows:

```
(input_specifier relational_operator CURRENCY(expression, currency, subitizer, decimal))
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

For currency searches, the subitizer is defined as the optional separating character that may be
encountered when parsing currency. For example, for standard US currency, a comma would be
specified. If other types of currencies are being searched, such as perhaps certain types of European
currencies that use a period as the separator, a period would be specified.

The decimal is defined as the decimal specifier to use. For example, for standard US numbers, a period
would be specified. For other currency types, an appropriate character would be specified.

Note that the subitizer and decimal specifier must be different. If the same character is specified for
both, an error message will be generated.

As an example of a fully qualified currency search relational expression, to find all values using the US
currency system between but not including `"$450"` and `"$10,100.50"` in a record/field construct where
the field tag is price, use:

```
(RECORD.price CONTAINS CURRENCY("$450" < CUR < "$10,100.50", "$", ",", "."))
```

The results will contain all prices encountered in the input data that match the specified range. For
example, if a price `"$692.01"` appears in the input data, then it will be reported as a match since it falls
within the specified range. But currency values like `-$123.00`, `$449.99` and `$100,000` would not match.

# Regex search

Regex Searches extend the general relational expression defined previously as follows:

```
(input_specifier relational_operator REGEX(expression))
```

TBD


# IPv4 search

The IPv4 Search operation can be used to search for exact IPv4 addresses or IPv4 addresses in a
particular range in both structured and unstructured text using the standard "a.b.c.d" format for IPv4
addresses.

IPv4 Searches extend the general relational expression defined previously as follows:

```
(input_specifier relational_operator IPV4(expression))
```

Different ranges can be searched for by modifying the expression in the relational expression above.
There are two general expression types supported:

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

For example, to find all IP addresses greater than `10.11.12.13`, use the following search query criteria:

```
(RAW_TEXT CONTAINS IPV4(IP > "10.11.12.13"))
```

To find all matching IPv4 addresses adhering to `10.10.0.0/16` (that is, all IP addresses
from `10.10.0.0` through `10.10.255.255` inclusive) in a record/field construct where
the field tag is ipaddr, use:

```
(RECORD.ipaddr CONTAINS IPV4("10.10.0.0" <= IP <= "10.10.255.255"))
```
