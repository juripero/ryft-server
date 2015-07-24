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



# How to do search?
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

Links:
 * http://base64-encoding.online-domain-tools.com/ --- online base64 encoder/decoder
