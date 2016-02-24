# Search backend

Search engine is an abstraction on custom search backend.

Search engine is one of:
- [ryftprim](./ryftprim/README.md) uses `ryftprim` utility
- [ryftone](./ryftone/README.md) uses `libryftone` library
  (to disable this search engine pass "noryftone" as go build tags)
- [ryfthttp](./ryfthttp/README.md) uses `ryft-server` (HTTP/msgpack or HTTP/json)
- [ryftmux](./ryftmux/README.md) multiplexes results from several search engines

## Import search backends

Here is a trick how to register search engines,
just use unused import for side effects:

```{.go}
import (
	"github.com/getryft/ryft-server/search"
	_ "github.com/getryft/ryft-server/search/ryftprim"
	_ "github.com/getryft/ryft-server/search/ryftone"
	_ "github.com/getryft/ryft-server/search/ryfthttp"
	_ "github.com/getryft/ryft-server/search/ryftmux"
)
```

We need to reference `ryftprim`, `ryfthttp` packages so each package
registers its factory in global search engine factory list.
