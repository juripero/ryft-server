
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
http://192.168.56.101:8765/search/fuzzy-hamming?query=( RAW_TEXT CONTAINS "Johm" )&files=passengers.txt&surrounding=10&fuzziness=2
```
Response:

```json
[{"file":"/ryftone/passengers.txt","offset":430,"length":24,"fuzziness":2,"data":"ZyB0aGUgZGV2ZWxvcG1lbnQgb2YgQWly"}
,{"file":"/ryftone/passengers.txt","offset":551,"length":24,"fuzziness":2,"data":"Ck1pY2hlbGxlIEpvbmVzLDA3LTEyLTE5"}
,{"file":"/ryftone/passengers.txt","offset":585,"length":24,"fuzziness":2,"data":"LTEyMTMsTXMuIEpvbmVzIGxpa2VzIHRv"}
,{"file":"/ryftone/passengers.txt","offset":645,"length":24,"fuzziness":2,"data":"Ck1pc2hlbGxlIEpvbmVzLDA3LTEyLTE5"}
,{"file":"/ryftone/passengers.txt","offset":679,"length":24,"fuzziness":2,"data":"LTEyMTMsTXMuIEpvbmVzIHByb3ZlcyB0"}
,{"file":"/ryftone/passengers.txt","offset":752,"length":24,"fuzziness":2,"data":"LgpNaWNoZWxlIEpvbmVzLDA3LTEyLTE5"}
,{"file":"/ryftone/passengers.txt","offset":786,"length":24,"fuzziness":2,"data":"LTEyMTMsTXMuIEpvbmVzIG9uY2UgYWdh"}
,{"file":"/ryftone/passengers.txt","offset":831,"length":24,"fuzziness":2,"data":"c24ndCBoYXZlIGNvbW1hbmQgb3ZlciB0"}
,{"file":"/ryftone/passengers.txt","offset":933,"length":24,"fuzziness":2,"data":"bmFtZSAnVCcuIE5vIG1vcmUuIE5vIGxl"}
]
```

The ``data`` fields are encoded by base64: 
```
ZyB0aGUgZGV2ZWxvcG1lbnQgb2YgQWly --> "g the development of Air"
Ck1pY2hlbGxlIEpvbmVzLDA3LTEyLTE5 --> "\nMichelle Jones,07-12-19"
LTEyMTMsTXMuIEpvbmVzIGxpa2VzIHRv --> "-1213,Ms. Jones likes to"
LTEyMTMsTXMuIEpvbmVzIHByb3ZlcyB0 --> "-1213,Ms. Jones proves t"
...

```



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



Links:
 * http://base64-encoding.online-domain-tools.com/ --- online base64 encoder/decoder
