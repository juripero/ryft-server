package main

import (
	"log"
	"net/http"
	"net/url"
	"os/exec"

	"github.com/getryft/ryft-server/names"
	"github.com/getryft/ryft-server/srverr"
)

var (
	cmd                  = "ryftprim"
	ps                   = "-p"
	fuzzy_hamming_search = "fhs"
	file                 = "-f"
	query                = "-q"
	result_file          = "-od"
	index_file           = "-oi"
	surrounding          = "-w"
	nodes                = "-n"
	case_sensetivity     = "-i"
	fuzziness            = "-d"
)

func ryftprim(s *SearchParams, n names.Names) (ch chan error) {
	ch = make(chan error, 1)
	go func() {
		testArgs := make([]string, 0)

		idxFile := names.PathInRyftoneForResultDir(n.IdxFile)
		resultFile := names.PathInRyftoneForResultDir(n.ResultFile)
		// cs := s.CaseSensitive
		// nc := s.Nodes
		// sr := s.Surrounding
		// files := s.Files
		// fz := s.Fuzziness

		query, aErr := url.QueryUnescape(s.Query)
		log.Print(s)
		if aErr != nil {
			ch <- srverr.New(http.StatusBadRequest, aErr.Error())
			return
		}

		// testArgs = append(testArgs, ps)
		// testArgs = append(testArgs, fuzzy_hamming_search)
		// testArgs = append(testArgs, file)
		// testArgs = append(testArgs, files[0])
		// testArgs = append(testArgs, result_file)
		// testArgs = append(testArgs, resultFile)
		// testArgs = append(testArgs, index_file)
		// testArgs = append(testArgs, idxFile)
		//
		//
		// if s.Nodes == 0 {
		// } else {
		// 	testArgs = append(testArgs, nodes)
		// 	testArgs = append(testArgs, string(nc))
		// }

		log.Print(testArgs)
		command := exec.Command(cmd, "-p", "fhs",
			"-q",
			query,
			"-f",
			s.Files[0],
			"-od",
			resultFile,
			"-oi",
			idxFile,
			"-w",
			string(s.Surrounding),
		)
		// output, err := command.CombinedOutput()
		output, err := command.CombinedOutput()
		log.Print(output)
		command.Run()

		if err != nil {
			ch <- srverr.New(http.StatusInternalServerError, err.Error())
			return
		}
		ch <- nil
	}()

	return
}
