# Search backend

Search engine is an abstraction on custom search backend.

Search engine is one of:
- `ryftprim` utility
- `ryftone` library
- ryft server (HTTP/msgpack)

## Import search backends

Here is a trick how to register search engines:

```
import (
	"github.com/getryft/ryft-server/search"
	_ "github.com/getryft/ryft-server/search/ryftprim"
)
```

We need to reference `ryftprim` package so it registers its factory
in global search engine factory list.
