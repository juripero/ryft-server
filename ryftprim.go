package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os/exec"

	"github.com/getryft/ryft-server/names"
	"github.com/getryft/ryft-server/srverr"
)

const (
	cmd                  = "ryftprim"
	arg_type             = "-p"
	fuzzy_hamming_search = "fhs"
	arg_separator        = "-e"
	no_separator         = ""
	arg_files            = "-f"
	arg_result_file      = "-od"
	arg_index_file       = "-oi"
	arg_surrounding      = "-w"
	arg_nodes            = "-n"
	arg_case_insensetive = "-i"
	arg_fuzziness        = "-d"
	arg_query            = "-q"
	arg_verbose          = "-v"
)

func ryftprim(s *SearchParams, n *names.Names) (ch chan error) {
	ch = make(chan error, 1)
	go func() {
		testArgs := []string{
			arg_type, fuzzy_hamming_search,
			arg_separator, no_separator,
			arg_verbose,
		}

		if !s.CaseSensitive {
			testArgs = append(testArgs, arg_case_insensetive)
		}

		if n != nil {
			idxFile := names.PathInRyftoneForResultDir(n.IdxFile)
			resultFile := names.PathInRyftoneForResultDir(n.ResultFile)

			testArgs = append(testArgs,
				arg_index_file, idxFile,
				arg_result_file, resultFile)
		}

		for _, file := range s.Files {
			testArgs = append(testArgs, arg_files, file)
		}

		if s.Nodes > 0 {
			testArgs = append(testArgs, arg_nodes, fmt.Sprintf("%d", s.Nodes))
		}

		if s.Surrounding > 0 {
			testArgs = append(testArgs, arg_surrounding, fmt.Sprintf("%d", s.Surrounding))
		}

		if s.Fuzziness > 0 {
			testArgs = append(testArgs, arg_fuzziness, fmt.Sprintf("%d", s.Fuzziness))
		}

		query, aErr := url.QueryUnescape(s.Query)

		if aErr != nil {
			ch <- srverr.New(http.StatusBadRequest, aErr.Error())
			return
		} else {
			testArgs = append(testArgs, arg_query, query)
		}

		log.Print(testArgs)
		command := exec.Command(cmd, testArgs...)

		output, err := command.CombinedOutput()
		log.Printf("\r\n%s", output)
		command.Run()

		if err != nil {
			ch <- srverr.NewWithDetails(http.StatusInternalServerError, err.Error(), string(output))
			return
		}
		ch <- nil
	}()

	return
}
