The PCAP primitive family is a set of search primitives specifically geared towards cyber analytics using
an arbitrary corpus of one or more input PCAP (`.pcap`) files. There are many open source and proprietary
tools on the market today which enable PCAP analysis, but many of them perform quite poorly, and
routinely run into computational and memory bottlenecks, especially as the input corpus becomes large. 
In addition, many common tools cannot operate against raw binary PCAP data and require a very time-consuming, 
up-front transformation on the input data, resulting in a bloated set of XML or JSON data representing the individual packets.

The PCAP primitive family defined in this API does not require such unnecessary transformation steps.


# Supported Layers

The PCAP family primitives support a wide variety of operations at protocol layers 2, 3 and 4, plus a
mechanism by which any Ryft primitive (such as edit distance, PCRE2, numeric, currency, etc.) can be
used against arbitrary layer 4 payloads, thereby enabling complex querying all the way up the protocol
stack, through layer 7.

For layers 2, 3 and 4, the PCAP family extend the general relational expression as
follows:

```
[(][!]layer.field[.subfield] operator value[)]
```

Parenthetical groupings are optional. The `!` is used to represent logical inversion of the expression as
needed.

Complex expressions follow the syntax show below, created using one or more logical `AND` (`&&`) or
logical `OR` (`||`) groupings, where each side of the grouping is an expression as previously defined, and
depicted here as the quantity ...:

```
... [ && | || ...]
```

The box below shows valid layer and associated field and value information along with brief
descriptions. Many of these are modeled after similar Wireshark nomenclature and semantics:

- `frame` – a meta-layer associated with the frame headers attached to each packet 
  in a PCAP file. The associated field value can be one of: 
    - `time` – the frame time in UTC, with format: 
         Mmm DD, YYYY HH:MM:SS[.s[s[s[s[s[s[s[s[s]]]]]]]]] 
         where Mmm is the three-letter capitalized month, e.g., Jan 
    - `time_delta` – the delta time from the previous packet in the input pcap file 
         associated with the packet, as a time offset in seconds with up to nine 
         positions after the decimal point 
    - `time_delta_displayed` – the delta time from the previous packet matched in the 
         query filter associated with the operation, as a time offset in seconds with up 
         to nine positions after the decimal point 
    - `time_epoch` – the frame time as the number of seconds elapsed since midnight of 
         January 1, 1970 time_invalid – 1-bit: 1 if the frame time is invalid (out of range), 
         0 if valid 
    - `time_relative` – the delta time from the first packet in the input pcap file associated 
         with the packet, as a time offset in seconds with up to nine positions after the decimal point
- `vlan` - Virtual LAN, as defined by IEEE 802.1Q, which resides at or alongside layer 2. 
  The associated field value can be one of: 
    - `etype` - VLAN Ethernet subtype (integer or hex if leading 0x used) 
    - `tci` - 16 bit value representing all tag control information (TCI) 
    - `priority | pcp` - priority code point 3-bit (0-7) field 
    - `dei | cfi` - drop eligible indicator (DEI) one bit (0-1) field 
    - `id | vid` - VLAN identifier, lower 12 bits (0-4095) of the TCI. 
         Note: if double tagging is used, both the inner and outer tags will be searched 
- `mpls` - Multi-protocol label switching, which is a hybrid layer 2/3 protocol. The 
   associated field value can be one of: 
    - `label` - 20-bit value (0-1,048575) for the MPLS label 
    - `exp | tc` - 3-bit value (0-7) for the traffic class (and QoS and ECN) 
    - `bottom | s` - 1-bit: 1 when the current label is the last in the stack 
    - `ttl - 8-bit` (0-255) time-to-live indicator
- `eth` - Ethernet layer 2 protocol. The associated field value can be one of:
    - `src` - source address (value: 6 2-hex-digit octets separated by colons)
    - `dest` - destination address
    - `addr` - either the source or the destination address
- `ip` - IPv4 layer 3 protocol. The associated field value can be one of:
    - `src` - source address (value: `[0-255].[0-255].[0-255].[0-255]`)
    - `dest` - destination address
    - `addr` - either the source or the destination address
- `ipv6` - IPv6 layer 3 protocol. The associated field value can be one of:
    - `src` - source address (value: 8 2-hex-digit octets separated by colons)
    - `dest` - destination address
    - `addr` - either the source or the destination address
- `icmp` - ICMP protocol, as it pertains to IPv4. The associated field value can be one of:
    - `type` - the ICMP type (value: `[0-255]`)
    - `code` - the ICMP subtype (value: `[0-255]`)
- `icmpv6` - ICMP protocol, as it pertains to IPv6. The associated field value can be one of:
    - `type` - the ICMP type
    - `code` - the ICMP subtype
- `udp - UDP layer 4 protocol. The associated field value can be one of:
    - `srcport` - the source port (value: `[0-65535]`)
    - `dstport` - the destination port
    - `port` - either the source or the destination port
- `tcp` - TCP layer 4 protocol. The associated field value can be one of:
    - `srcport` - the source port (value: `[0-65535]`)
    - `dstport` - the destination port
    - `port` - either the source or the destination port
    - `flags` - the TCP flags construct, with supported subfields:
        - `fin` - the FIN bit
        - `syn` - the SYN bit
        - `reset` - the RST bit

The supported operators vary depending on the field and value, but in general adhere to:
- Equality: `==`, `eq`
- Inequality: `!=`, `ne`
- Greater than: `>`, `gt`
- Greater than or equal to: `>=`, `ge`
- Less than: `<`, `lt`
- Less than or equal to: `<=`, `le`

- Not:  `!, not`
Note that the “Not” operations are currently supported for use directly alongside boolean types only, 
such as tcp.flags.[fin|syn|reset], and do not propagate through complex expressions.

# IPv4 Fragmentation Support

PCAP IPv4 (ip) support includes support for native IPv4 packet fragmentation. The library intelligently
The library intelligently handles IP fragmentation packets as native, individualized packets, or as 
combined, de-fragmented payload data, depending on the types of operations requested in the associated 
query clause. For example, if a message has been fragmented into four packets and results in a payload 
match, then only one result is reported in the tabulated results, yet all packets that made up the 
reassembled quantity remain available for detailed analysis.

The library ensures that output PCAP data files maintain the precise order and timing of the fragments 
with respect to other packets in the source corpus. This means that downstream tools (such as Wireshark 
or tshark) can continue to operate on the output PCAP data natively if needed.

# TCP and UDP Payloads, with Full Layer 7 Primitive Support

Arbitrary Ryft primitives can be used against the layer 4 TCP or UDP payload, which by definition
includes data up through layer 7, using the following slight variant of the standard search primitive form:

```
(RECORD.payload relational_operator primitive(expression[, options]))
```

The `RECORD.payload` portion shown above in bold is literal. The remainder of the expression is
identical to generic search expression. This ensures that the power of any API primitive can be
brought to bear on arbitrary payload data of each and every individual packet. This includes even
complex primitives such as the edit distance primitive which tools like Wireshark do not natively
support.

However, with the exception of IPv4 fragmentation as previously described, it is important to
understand that the library's PCAP implementation currently does not concatenate cross-packet payloads
together across a given stream conversation. The RECORD.payload operation works on an individual
packet by individual packet basis. If it is desired to search a payload stream in its entirety (such as, for
example, a lengthy browser session spanning many packets and many minutes), the following process is
recommended:
1. Thin the original PCAP using appropriate API primitive constructs, such as by using source/destination IP address, port, etc.
2. Run the thinned results through Wireshark or tshark using their “Follow TCP Stream” feature to extract and concatenate all payloads of interest as part of the stream conversation.
3. Feed the resulting raw depacketized output through a second set of API payload search primitive.

The workflow noted above can be quite time consuming as tools like Wireshark and tshark are not full 
parallel architectures. Furthermore, they often require significant CPU and RAM resources, and can therefore 
be prohibitive to use on very large input PCAP corpus data.

Thankfully, the BlackLynx APIs provide a mechanism by which the TCP streams, UDP streams, and payloads can be 
generated in parallel alongside any PCAP operation. This is achieved by using an appropriately crafted global 
option string entry.

The global option string entry to use is l4_stream_dump, which takes sub-parameters that specify the direction 
(transmit as tx, receive as rx, or both), whether (on) or not (off) to output payload in addition to the 
isolated packet stream, and finally an output filename prefix for the file(s) that will be generated. The precise 
format is:

```
l4_stream_dump="tx|rx|both:on|off:prefix"
```

The output filenames used will start with the prefix specified, which can include full or relative path detail 
as desired. For example, if the files should be dumped to /tmp and start with foo, then the prefix used might be 
specified as /tmp/foo. The full output filename is defined as:

```
prefix_packet_nnn_tx|rx|both.pcap|payload
```

where:
  - prefix is the prefix value from the associated l4_stream_dump global option string
  - _packet_ is literal and always appears in the filename
  - nnn is the integer packet number from the associated input corpus file of the first match that resulted in the stream being followed
  - tx|rx|both is the tx|rx|both selection from the associated l4_stream_dump global option string
  - . is the literal extension separator and always appears in the filename
  - pcap|payload is the extension. One output file with extension pcap is always generated and represents the extracted stream. 
    A separate file with extension payload is generated as the concatenated L4 payloads, but only if the value on was 
    specified as previously defined in the second sub-parameter of the associated l4_stream_dump global option string
    
Note that use of the l4_stream_dump global option string does not preclude the normal generation of output data and 
output index files. All such features can be used on combination, during the same operation, as desired.

# Results: PCAP Family
One way to make meaningful use of the PCAP family of primitives is within the confines of a larger cyber analytics 
ecosystem, to quickly filter very large input data corpus that contemporary tools often struggle to process.

The data results file generated by the PCAP primitive operation include the collection of source packets that match 
the various complex criteria of the operation requested. The typical result file is therefore a greatly thinned version 
of a much larger set of PCAP input. Since results files are themselves valid PCAP files, they can be natively used by 
downstream tools such as Wireshark and tshark to perform further analysis.

In addition to output data file generation, the PCAP primitive also supports output index file generation. The index 
file output is a set of comma-separated value lines, one line per match. The columns are, left to right:
- filename where the match was found,
- byte offset (where offset 0 is the first byte of the file) of the start of the packet where the match was found,
- the length of the packet where the match was found,
- the distance, which is usually reported as 0, except for Hamming and Edit Distance payload invocations,
- the layer 4 payload match qualifier(s) as MQ='"data1"[ "data2"][ ...]' where dataN is the relevant portion(s) 
  of the query string that applied, if match qualifiers are selected in the global options, and
- the associated packet number as PN='xx' where xx is the packet number, if match qualifiers are selected in the global options
