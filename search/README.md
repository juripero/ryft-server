# Search backend

Search engine is an abstraction on custom search backend.

Search engine is one of:
- `ryftprim` uses ryftprim utility
- `ryftone` uses ryftone library (NOT IMPLEMENTED yet)
- `ryfthttp` uses ryft HTTP server (HTTP/msgpack)
- `ryftmux` multiplexes results from several engines

## Import search backends

Here is a trick how to register search engines,
just use unused import for side effects:

```
import (
	"github.com/getryft/ryft-server/search"
	_ "github.com/getryft/ryft-server/search/ryftprim"
	_ "github.com/getryft/ryft-server/search/ryfthttp"
	_ "github.com/getryft/ryft-server/search/ryftmux"
)
```

We need to reference `ryftprim`, `ryfthttp` packages so each package
registers its factory in global search engine factory list.
