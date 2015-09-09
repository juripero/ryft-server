
# How to build the REST-server?

1. See installation instructions for golang environment — https://golang.org/doc/install
2. ``mkdir -pv $GOPATH/src/github.com/DataArt/``
3. ``cd $GOPATH/src/github.com/DataArt/``
4. ``git clone https://github.com/DataArt/ryft-rest-api.git``
5. ``cd ryft-server-api``
7. ``go get``
8. ``go install``
9. Get ``ryft-server`` binary from ``$GOPATH/bin``

# How to run the REST-server?

```
ryft-server -port=8765
```
Default value ``port`` is ``8765``

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

[{"_index":{"file":"/ryftone/passengers.txt","fuzziness":2,"length":24,"offset":430},"base64":"ZyB0aGUgZGV2ZWxvcG1lbnQgb2YgQWly"}
,{"_index":{"file":"/ryftone/passengers.txt","fuzziness":2,"length":24,"offset":551},"base64":"Ck1pY2hlbGxlIEpvbmVzLDA3LTEyLTE5"}
,{"_index":{"file":"/ryftone/passengers.txt","fuzziness":2,"length":24,"offset":585},"base64":"LTEyMTMsTXMuIEpvbmVzIGxpa2VzIHRv"}
,{"_index":{"file":"/ryftone/passengers.txt","fuzziness":2,"length":24,"offset":645},"base64":"Ck1pc2hlbGxlIEpvbmVzLDA3LTEyLTE5"}
,{"_index":{"file":"/ryftone/passengers.txt","fuzziness":2,"length":24,"offset":679},"base64":"LTEyMTMsTXMuIEpvbmVzIHByb3ZlcyB0"}
,{"_index":{"file":"/ryftone/passengers.txt","fuzziness":2,"length":24,"offset":752},"base64":"LgpNaWNoZWxlIEpvbmVzLDA3LTEyLTE5"}
,{"_index":{"file":"/ryftone/passengers.txt","fuzziness":2,"length":24,"offset":786},"base64":"LTEyMTMsTXMuIEpvbmVzIG9uY2UgYWdh"}
,{"_index":{"file":"/ryftone/passengers.txt","fuzziness":2,"length":24,"offset":831},"base64":"c24ndCBoYXZlIGNvbW1hbmQgb3ZlciB0"}
,{"_index":{"file":"/ryftone/passengers.txt","fuzziness":2,"length":24,"offset":933},"base64":"bmFtZSAnVCcuIE5vIG1vcmUuIE5vIGxl"}
]
```

# How to do search by field's value?

```
http://52.20.99.136:8765/search?query=(RECORD.id%20EQUALS%20%2210034183%22)&files=*.pcrime&surrounding=10&fuzziness=0&format=xml

[
    {
        "_index": {
            "file": "/ryftone/chicago.pcrime",
            "fuzziness": 0,
            "length": 693,
            "offset": 0
        },
        "attrs": null,
        "base64": "PHJlYz48SUQ+MTAwMzQxODM8L0lEPjxDYXNlTnVtYmVyPkhZMjIzNjczPC9DYXNlTnVtYmVyPjxEYXRlPjA0LzE1LzIwMTUgMTE6NTk6MDAgUE08L0RhdGU+PEJsb2NrPjA2MlhYIFMgU1QgTEFXUkVOQ0UgQVZFPC9CbG9jaz48SVVDUj4wNDg2PC9JVUNSPjxQcmltYXJ5VHlwZT5CQVRURVJZPC9QcmltYXJ5VHlwZT48RGVzY3JpcHRpb24+RE9NRVNUSUMgQkFUVEVSWSBTSU1QTEU8L0Rlc2NyaXB0aW9uPjxMb2NhdGlvbkRlc2NyaXB0aW9uPlNUUkVFVDwvTG9jYXRpb25EZXNjcmlwdGlvbj48QXJyZXN0PmZhbHNlPC9BcnJlc3Q+PERvbWVzdGljPnRydWU8L0RvbWVzdGljPjxCZWF0PjAzMTM8L0JlYXQ+PERpc3RyaWN0PjAwMzwvRGlzdHJpY3Q+PFdhcmQ+MjA8L1dhcmQ+PENvbW11bml0eUFyZWE+NDI8L0NvbW11bml0eUFyZWE+PEZCSUNvZGU+MDhCPC9GQklDb2RlPjxYQ29vcmRpbmF0ZT4xMTgxMjYzPC9YQ29vcmRpbmF0ZT48WUNvb3JkaW5hdGU+MTg2Mzk2NTwvWUNvb3JkaW5hdGU+PFllYXI+MjAxNTwvWWVhcj48VXBkYXRlZE9uPjA0LzIyLzIwMTUgMTI6NDc6MTAgUE08L1VwZGF0ZWRPbj48TGF0aXR1ZGU+NDEuNzgxOTYxNjg4PC9MYXRpdHVkZT48TG9uZ2l0dWRlPi04Ny42MTA5ODQ3MDU8L0xvbmdpdHVkZT48TG9jYXRpb24+Iig0MS43ODE5NjE2ODgsIC04Ny42MTA5ODQ3MDUpIjwvTG9jYXRpb24+PC9yZWM+",
        "childs": [
            {
                "attrs": null,
                "childs": [
                    {
                        "data": "10034183",
                        "type": "chardata"
                    }
                ],
                "name": "ID",
                "namespace": "",
                "type": "element"
            },
            {
                "attrs": null,
                "childs": [
                    {
                        "data": "HY223673",
                        "type": "chardata"
                    }
                ],
                "name": "CaseNumber",
                "namespace": "",
                "type": "element"
            },
            {
                "attrs": null,
                "childs": [
                    {
                        "data": "04/15/2015 11:59:00 PM",
                        "type": "chardata"
                    }
                ],
                "name": "Date",
                "namespace": "",
                "type": "element"
            },
            {
                "attrs": null,
                "childs": [
                    {
                        "data": "062XX S ST LAWRENCE AVE",
                        "type": "chardata"
                    }
                ],
                "name": "Block",
                "namespace": "",
                "type": "element"
            },
            {
                "attrs": null,
                "childs": [
                    {
                        "data": "0486",
                        "type": "chardata"
                    }
                ],
                "name": "IUCR",
                "namespace": "",
                "type": "element"
            },
            {
                "attrs": null,
                "childs": [
                    {
                        "data": "BATTERY",
                        "type": "chardata"
                    }
                ],
                "name": "PrimaryType",
                "namespace": "",
                "type": "element"
            },
            {
                "attrs": null,
                "childs": [
                    {
                        "data": "DOMESTIC BATTERY SIMPLE",
                        "type": "chardata"
                    }
                ],
                "name": "Description",
                "namespace": "",
                "type": "element"
            },
            {
                "attrs": null,
                "childs": [
                    {
                        "data": "STREET",
                        "type": "chardata"
                    }
                ],
                "name": "LocationDescription",
                "namespace": "",
                "type": "element"
            },
            {
                "attrs": null,
                "childs": [
                    {
                        "data": "false",
                        "type": "chardata"
                    }
                ],
                "name": "Arrest",
                "namespace": "",
                "type": "element"
            },
            {
                "attrs": null,
                "childs": [
                    {
                        "data": "true",
                        "type": "chardata"
                    }
                ],
                "name": "Domestic",
                "namespace": "",
                "type": "element"
            },
            {
                "attrs": null,
                "childs": [
                    {
                        "data": "0313",
                        "type": "chardata"
                    }
                ],
                "name": "Beat",
                "namespace": "",
                "type": "element"
            },
            {
                "attrs": null,
                "childs": [
                    {
                        "data": "003",
                        "type": "chardata"
                    }
                ],
                "name": "District",
                "namespace": "",
                "type": "element"
            },
            {
                "attrs": null,
                "childs": [
                    {
                        "data": "20",
                        "type": "chardata"
                    }
                ],
                "name": "Ward",
                "namespace": "",
                "type": "element"
            },
            {
                "attrs": null,
                "childs": [
                    {
                        "data": "42",
                        "type": "chardata"
                    }
                ],
                "name": "CommunityArea",
                "namespace": "",
                "type": "element"
            },
            {
                "attrs": null,
                "childs": [
                    {
                        "data": "08B",
                        "type": "chardata"
                    }
                ],
                "name": "FBICode",
                "namespace": "",
                "type": "element"
            },
            {
                "attrs": null,
                "childs": [
                    {
                        "data": "1181263",
                        "type": "chardata"
                    }
                ],
                "name": "XCoordinate",
                "namespace": "",
                "type": "element"
            },
            {
                "attrs": null,
                "childs": [
                    {
                        "data": "1863965",
                        "type": "chardata"
                    }
                ],
                "name": "YCoordinate",
                "namespace": "",
                "type": "element"
            },
            {
                "attrs": null,
                "childs": [
                    {
                        "data": "2015",
                        "type": "chardata"
                    }
                ],
                "name": "Year",
                "namespace": "",
                "type": "element"
            },
            {
                "attrs": null,
                "childs": [
                    {
                        "data": "04/22/2015 12:47:10 PM",
                        "type": "chardata"
                    }
                ],
                "name": "UpdatedOn",
                "namespace": "",
                "type": "element"
            },
            {
                "attrs": null,
                "childs": [
                    {
                        "data": "41.781961688",
                        "type": "chardata"
                    }
                ],
                "name": "Latitude",
                "namespace": "",
                "type": "element"
            },
            {
                "attrs": null,
                "childs": [
                    {
                        "data": "-87.610984705",
                        "type": "chardata"
                    }
                ],
                "name": "Longitude",
                "namespace": "",
                "type": "element"
            },
            {
                "attrs": null,
                "childs": [
                    {
                        "data": "\"(41.781961688, -87.610984705)\"",
                        "type": "chardata"
                    }
                ],
                "name": "Location",
                "namespace": "",
                "type": "element"
            }
        ],
        "name": "rec",
        "namespace": "",
        "type": "element"
    }
]

```

# Parameters

TODO


# How to check error and success results?
```
http://192.168.56.103:8765/search/test-ok 
[{"number":0},{"number":1},{"number":2},...,{"number":98},{"number":99},{"number":100}]
```

```
http://192.168.56.103:8765/search/test-fail
{
    "message": "Test error",
    "status": 500
}
```

# Good requests for tests

```
curl "http://ryft-emulator:8765/search?query=(RAW_TEXT%20CONTAINS%20%22bin%22)&files=jdk-8u45-linux-x64.tar.gzfuzziness=2&surrounding=10"
```



# Links
 * http://base64-encoding.online-domain-tools.com/ --- online base64 encoder/decoder
