
# Installing Dependencies & Building

1. See installation instructions for golang environment — https://golang.org/doc/install
2. ``go get github.com/getryft/ryft-server``
3. ``go install``
4. Get ``ryft-server`` binary from ``$GOPATH/bin``

# Running & Command Line Parameters

```
ryft-server -port=8765
```
Default value ``port`` is ``8765``
# Packaging into deb file

https://github.com/getryft/ryft-server/blob/master/ryft-server-make-deb/README.md

# Keeping search results

By default REST-server removes search results from ``/ryftone/RyftServer-PORT/``. But it behaviour may be prevented:

```
ryft-server -keep-results
```
Please pay attention that REST-server removes ``/ryftone/RyftServer-PORT`` when it starts.

# Index

```
http://52.4.187.202:8765
```

# How to do a search?
Do request in browser:

```
http://192.168.56.101:8765/search?query=( RAW_TEXT CONTAINS "Johm" )&files=passengers.txt&surrounding=10&fuzziness=2

```

# How to do search by field's value?

```
http://52.20.99.136:8765/search?query=(RECORD.id%20EQUALS%20%2210034183%22)&files=*.pcrime&surrounding=10&fuzziness=0&format=xml

```

# Parameters (TODO)
* ``query`` is the string specifying the search criteria.
* ``files``  is the input data set to be searched
* ``fuzziness`` Specify the fuzzy search distance [0..255]
* ``cs``
* ``format`` is the parameter for the structed search. Specify the search format.
* ``surrounding`` width when generating results. For example, a value of 2 means that 2 + * characters before and after a search match will be included with data result


# Good requests for tests

```
curl "http://ryft-emulator:8765/search?query=(RAW_TEXT%20CONTAINS%20%22bin%22)&files=jdk-8u45-linux-x64.tar.gz&fuzziness=2&surrounding=10"
```

# Links
 * http://base64-encoding.online-domain-tools.com/ --- online base64 encoder/decoder
 * http://msgpack.org/ --- link to msgpack official
 * https://github.com/ugorji/go --- link to msgpack library
