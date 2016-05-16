# RyftOne search engine

## WARNING libryftone library is not thread-safe yet!

### Multiple requests to libryftone from one process at the same time DO NOT WORK!


RyftOne search engine is based on `libryftone` library.
See [manual](http://info.ryft.com/acton/attachment/17117/f-0002/1/-/-/-/-/Ryft-Open-API-Library-User-Guide.pdf).

Note: implementation is very similar to RyftPrim search engine!
Need to combine both to avoid code duplication!

To ignore libryftone linking just pass "noryftone" tag to the go build:

```
go build -tags "noryftone"
```

This might be helpful to build `ryft-server` without `libryftone` installed.
