# Demo - backend options - August 17, 2017

This demo shows how backend options may be passed into search backend with search request.

## /search and /count endpoints

Now both `/search` and `/count` endpoints can accept `backend-option` parameter.
`backend-option` keeps a list of backend options. You should remember that as ryft-server can automatically switch backend you need to set `backend` parameter manually.

```{.sh}
./ryftrest -a app1:8765 -q "mish" -f "*" --backend "ryftx" --backend-option "--rx-shard-size" --backend-option "64M"
```

Ryft-Server log:

```
time="2017-08-23T14:25:48Z" level=info msg="[ryftprim]: executing tool" args="[-p g -q (RAW_TEXT CONTAINS EXACT(\"mish\")) -f passengers.txt -f ryftrest -e \\x0d\\x0a -v -l --rx-shard-size 64M]" task=00000000000001a9 tool=/usr/bin/ryftx
```

You can add any number of backend options consistently.
