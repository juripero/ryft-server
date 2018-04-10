The PCAP primitive family is a set of search primitives specifically geared towards cyber analytics using
an arbitrary corpus of one or more input PCAP (`.pcap`) files. There are many open source and proprietary
tools on the market today which enable PCAP analysis, but many of them perform quite poorly, and
routinely run into computational and memory bottlenecks, especially as the input corpus becomes large.

Ryft’s PCAP primitives solve these problems using the powerful Ryft X parallelization framework library.


# Supported Layers

The PCAP family primitives support a wide variety of operations at protocol layers 2, 3 and 4, plus a
mechanism by which any Ryft primitive (such as edit distance, PCRE2, numeric, currency, etc.) can be
used against arbitrary layer 4 payloads, thereby enabling complex querying all the way up the protocol
stack, through layer 7.

For layers 2, 3 and 4, the PCAP family extend the general relational expression defined previously as
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


# IPv4 Fragmentation Support

PCAP IPv4 (ip) support includes support for native IPv4 packet fragmentation. The library intelligently
reassembles fragmented packets before performing operations, and tabulates results using the
reassembled value. That is, if a message has been fragmented into four packets and is a ‘match’, then
only one result is reported in the tabulated results.

This is a unique capability of the Ryft architecture; most commercial and open source architectures do
not allow for intelligent reassembly for deep packet analysis or results reporting while still maintaining
fragmentation ordering for further downstream cyber analysis.

Fragmented packets are problematic, though, in terms of how they should appear in the thinned,
filtered output results, especially when considering overall performance. Currently, when an output data
file request is present, to maintain the integrity of the communication and to ensure the highest
possible performance, after internal reassembly and analysis of fragmented packets, the set of
fragmented packets precisely as they appeared in the input corpus are appended at the very end of the
rest of the filtered output. The order of the fragmented packets themselves is intact with respect to
other fragmented packets for the same communication, but is out of natural order with respect to other
packets. This decision was made to ensure that packets that are not fragmented (which is the vast
majority of packets) can be processed in parallel - as opposed to stalling - while fragmentation packets
and streams are being tabulated.

Note that downstream tools (such as Wireshark or tshark) can be used to reorder and/or sort the
thinned results based on packet time, for example, which would effectively restore natural order of all
results packets, if needed.


# TCP and UDP Payloads, with Full Layer 7 Primitive Support

Arbitrary Ryft primitives can be used against the layer 4 TCP or UDP payload, which by definition
includes data up through layer 7, using the following slight variant of the standard search primitive form:

```
(RECORD.payload relational_operator primitive(expression[, options]))
```

The `RECORD.payload` portion shown above in bold is literal. The remainder of the expression is
identical to generic search expression. This ensures that the power of any Ryft primitive can be
brought to bear on arbitrary payload data of each and every individual packet. This includes even
complex primitives such as Ryft’s edit distance primitive which tools like Wireshark do not natively
support.

However, with the exception of IPv4 fragmentation as previously described, it is important to
understand that Ryft’s PCAP implementation currently does not concatenate cross-packet payloads
together across a given stream conversation. The RECORD.payload operation works on an individual
packet by individual packet basis. If it is desired to search a payload stream in its entirety (such as, for
example, a lengthy browser session spanning many packets and many minutes), the following process is
recommended:
1. Thin the original PCAP using appropriate Ryft primitive constructs, such as by using source/destination IP address, port, etc.
2. Run the thinned results through Wireshark or tshark using their “Follow TCP Stream” feature to extract and concatenate all payloads of interest as part of the stream conversation.
3. Feed the resulting raw depacketized output through a standard Ryft search primitive.


# Results: PCAP Family

One way to make meaningful use of the Ryft PCAP family of primitives is within the confines of a larger
cyber analytics ecosystem, to quickly filter very large input data corpus that contemporary tools often
struggle to process.

Results files generated by the Ryft library for a PCAP primitive operation include the collection of source
packets that match the various complex criteria of the operation requested. The typical result file is
therefore a greatly thinned version of a much larger set of PCAP input. Since results files are themselves
valid PCAP files, they can be natively used by downstream tools such as Wireshark and tshark to perform
further analysis.
