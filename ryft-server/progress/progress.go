package progress

import (
	"net/http"

	"github.com/DataArt/ryft-rest-api/rol"
	"github.com/DataArt/ryft-rest-api/ryft-server/binding"
	"github.com/DataArt/ryft-rest-api/ryft-server/names"
	"github.com/DataArt/ryft-rest-api/ryft-server/srverr"
)

func Progress(s *binding.Search, n names.Names) (ch chan error) {
	ch = make(chan error, 1)
	go func() {
		ds := rol.RolDSCreate()
		defer ds.Delete()

		for _, f := range s.Files {
			ok := ds.AddFile(f)
			if !ok {
				ch <- srverr.New(http.StatusNotFound, "Could not add file "+f)
				return
			}
		}

		idxFile := names.PathInRyftoneForResultDir(n.IdxFile)
		resultsDs := func() *rol.RolDS {
			if s.Fuzziness == 0 {
				return ds.SearchExact(names.PathInRyftoneForResultDir(n.ResultFile), s.Query, s.Surrounding, "", &idxFile)
			} else {
				return ds.SearchFuzzyHamming(names.PathInRyftoneForResultDir(n.ResultFile), s.Query, s.Surrounding, s.Fuzziness, "", &idxFile)
			}
		}()
		defer resultsDs.Delete()

		if err := resultsDs.HasErrorOccured(); err != nil {
			if !err.IsStrangeError() {
				ch <- srverr.New(http.StatusInternalServerError, err.Error())
				return
			}
		}

		ch <- nil

	}()
	return
}
