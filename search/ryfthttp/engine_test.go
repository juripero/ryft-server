package ryfthttp

import (
	"testing"

	"github.com/getryft/ryft-server/search"
	"github.com/stretchr/testify/assert"
)

// test prepare search url
func TestEnginePrepareSearchUrl(t *testing.T) {
	check := func(cfg *search.Config, url string, local bool, expected string) {
		engine, err := NewEngine(map[string]interface{}{
			"server-url": url,
			"local-only": local,
		})
		if assert.NoError(t, err) {
			url := engine.prepareSearchUrl(cfg)
			assert.EqualValues(t, expected, url.String())
		}
	}

	cfg := search.NewConfig("hello", "1.txt", "2.txt")
	check(cfg, "http://localhost:12345", false,
		"http://localhost:12345/count?cs=false&ep=true&file=1.txt&file=2.txt&format=null&local=false&query=hello&stats=true&stream=true")
	check(cfg, "http://localhost:12345", true,
		"http://localhost:12345/count?cs=false&ep=true&file=1.txt&file=2.txt&format=null&local=true&query=hello&stats=true&stream=true")
	cfg.Files = nil

	cfg.Query = "hel lo"
	check(cfg, "http://localhost:12345", false,
		"http://localhost:12345/count?cs=false&ep=true&format=null&local=false&query=hel+lo&stats=true&stream=true")
	cfg.Query = "hel\nlo"
	check(cfg, "http://localhost:12345", false,
		"http://localhost:12345/count?cs=false&ep=true&format=null&local=false&query=hel%0Alo&stats=true&stream=true")
	cfg.Query = "hello"

	cfg.Mode = "fhs"
	check(cfg, "http://localhost:12345", false,
		"http://localhost:12345/count?cs=false&ep=true&format=null&local=false&mode=fhs&query=hello&stats=true&stream=true")
	check(cfg, "http://localhost:12345", true,
		"http://localhost:12345/count?cs=false&ep=true&format=null&local=true&mode=fhs&query=hello&stats=true&stream=true")
	cfg.Mode = ""

	cfg.CaseSensitive = true
	check(cfg, "http://localhost:12345", false,
		"http://localhost:12345/count?cs=true&ep=true&format=null&local=false&query=hello&stats=true&stream=true")
	check(cfg, "http://localhost:12345", true,
		"http://localhost:12345/count?cs=true&ep=true&format=null&local=true&query=hello&stats=true&stream=true")
	cfg.CaseSensitive = false

	cfg.Surrounding = 2
	check(cfg, "http://localhost:12345", false,
		"http://localhost:12345/count?cs=false&ep=true&format=null&local=false&query=hello&stats=true&stream=true&surrounding=2")
	check(cfg, "http://localhost:12345", true,
		"http://localhost:12345/count?cs=false&ep=true&format=null&local=true&query=hello&stats=true&stream=true&surrounding=2")
	cfg.Surrounding = 0

	cfg.Fuzziness = 1
	check(cfg, "http://localhost:12345", false,
		"http://localhost:12345/count?cs=false&ep=true&format=null&fuzziness=1&local=false&query=hello&stats=true&stream=true")
	check(cfg, "http://localhost:12345", true,
		"http://localhost:12345/count?cs=false&ep=true&format=null&fuzziness=1&local=true&query=hello&stats=true&stream=true")
	cfg.Fuzziness = 0

	cfg.Nodes = 3
	check(cfg, "http://localhost:12345", false,
		"http://localhost:12345/count?cs=false&ep=true&format=null&local=false&nodes=3&query=hello&stats=true&stream=true")
	check(cfg, "http://localhost:12345", true,
		"http://localhost:12345/count?cs=false&ep=true&format=null&local=true&nodes=3&query=hello&stats=true&stream=true")
	cfg.Nodes = 0

	cfg.KeepDataAs = "data.bin"
	check(cfg, "http://localhost:12345", false,
		"http://localhost:12345/count?cs=false&data=data.bin&ep=true&format=null&local=false&query=hello&stats=true&stream=true")
	check(cfg, "http://localhost:12345", true,
		"http://localhost:12345/count?cs=false&data=data.bin&ep=true&format=null&local=true&query=hello&stats=true&stream=true")
	cfg.KeepDataAs = ""

	cfg.KeepIndexAs = "index.txt"
	check(cfg, "http://localhost:12345", false,
		"http://localhost:12345/count?cs=false&ep=true&format=null&index=index.txt&local=false&query=hello&stats=true&stream=true")
	check(cfg, "http://localhost:12345", true,
		"http://localhost:12345/count?cs=false&ep=true&format=null&index=index.txt&local=true&query=hello&stats=true&stream=true")
	cfg.KeepIndexAs = ""

	cfg.Limit = 100
	check(cfg, "http://localhost:12345", false,
		"http://localhost:12345/count?cs=false&ep=true&format=null&limit=100&local=false&query=hello&stats=true&stream=true")
	check(cfg, "http://localhost:12345", true,
		"http://localhost:12345/count?cs=false&ep=true&format=null&limit=100&local=true&query=hello&stats=true&stream=true")
	cfg.Limit = 0

	cfg.ReportIndex = true
	cfg.ReportData = true
	check(cfg, "http://localhost:12345", false,
		"http://localhost:12345/search?cs=false&ep=true&format=raw&local=false&query=hello&stats=true&stream=true")
	check(cfg, "http://localhost:12345", true,
		"http://localhost:12345/search?cs=false&ep=true&format=raw&local=true&query=hello&stats=true&stream=true")

	cfg.ReportIndex = true
	cfg.ReportData = false
	check(cfg, "http://localhost:12345", false,
		"http://localhost:12345/search?cs=false&ep=true&format=null&local=false&query=hello&stats=true&stream=true")
	check(cfg, "http://localhost:12345", true,
		"http://localhost:12345/search?cs=false&ep=true&format=null&local=true&query=hello&stats=true&stream=true")
}

// test prepare files url
func TestEnginePrepareFilesUrl(t *testing.T) {
	check := func(dir string, url string, local bool, expected string) {
		engine, err := NewEngine(map[string]interface{}{
			"server-url": url,
			"local-only": local,
		})
		if assert.NoError(t, err) {
			url := engine.prepareFilesUrl(dir)
			assert.EqualValues(t, expected, url.String())
		}
	}

	check("foo", "http://localhost:12345", false,
		"http://localhost:12345/files?dir=foo&local=false")
	check("foo", "http://localhost:12345", true,
		"http://localhost:12345/files?dir=foo&local=true")
}
