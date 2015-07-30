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



# How to do exact search?
Do request in browser:
```
http://192.168.56.103:8765/search/exact?query=( RAW_TEXT CONTAINS "night" )&files=passengers.txt&surrounding=10
```
Response:
```json
[{"file":"/ryftone/passengers.txt","offset":211,"length":25,"fuzziness":0,"data":"Ck1pY2hhZWwgS25pZ2h0LCAwOC0xNy0xOQ=="}
,{"file":"/ryftone/passengers.txt","offset":248,"length":25,"fuzziness":0,"data":"NTUtNDU0NSwiS25pZ2h0IEluZHVzdHJpZQ=="}
]
```

The ``data`` fields are encoded by base64: 
```haskell
base64decode("Ck1pY2hhZWwgS25pZ2h0LCAwOC0xNy0xOQ==") --> "\nMichael Knight, 08-17-19"
base64decode("NTUtNDU0NSwiS25pZ2h0IEluZHVzdHJpZQ==") --> "55-4545,"Knight Industrie"
```

# How to do fuzzy-hamming search?
Do request in browser:
```
http://192.168.56.101:8765/search/fuzzy-hamming?query=( RAW_TEXT CONTAINS "Johm" )&files=passengers.txt&surrounding=10&fuzziness=2
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

Links:
 * http://base64-encoding.online-domain-tools.com/ --- online base64 encoder/decoder
