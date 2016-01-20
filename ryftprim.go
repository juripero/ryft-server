package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os/exec"

	"gopkg.in/yaml.v2"

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

const (
	ryftprimKey    = "ryftprim"
	duration       = "Duration"
	totalBytes     = "Total Bytes"
	matches        = "Matches"
	fabricDataRate = "Fabric Data Rate"
	dataRate       = "Data Rate"
)

type RyftprimParams struct {
	Query         string
	Files         []string
	Surrounding   uint16
	Fuzziness     uint8
	Format        string
	CaseSensitive bool
	Fields        string
	Nodes         uint8
}

//type RyftprimStats map[string]interface{}

func ryftprim(p *RyftprimParams, n *names.Names) (ch chan error, statistic chan map[string]interface{}) {
	ch = make(chan error, 1)
	statistic = make(chan map[string]interface{})
	go func() {
		testArgs := []string{
			arg_type, fuzzy_hamming_search,
			arg_separator, no_separator,
			arg_verbose,
		}

		if !p.CaseSensitive {
			testArgs = append(testArgs, arg_case_insensetive)
		}

		if n != nil {
			idxFile := names.PathInRyftoneForResultDir(n.IdxFile)
			resultFile := names.PathInRyftoneForResultDir(n.ResultFile)

			testArgs = append(testArgs,
				arg_index_file, idxFile,
				arg_result_file, resultFile)
		}

		for _, file := range p.Files {
			testArgs = append(testArgs, arg_files, file)
		}

		if p.Nodes > 0 {
			testArgs = append(testArgs, arg_nodes, fmt.Sprintf("%d", p.Nodes))
		}

		if p.Surrounding > 0 {
			testArgs = append(testArgs, arg_surrounding, fmt.Sprintf("%d", p.Surrounding))
		}

		if p.Fuzziness > 0 {
			testArgs = append(testArgs, arg_fuzziness, fmt.Sprintf("%d", p.Fuzziness))
		}

		query, aErr := url.QueryUnescape(p.Query)

		if aErr != nil {
			// statistic <- nil
			close(statistic)
			ch <- srverr.New(http.StatusBadRequest, aErr.Error())
			return
		}
		testArgs = append(testArgs, arg_query, query)

		log.Print(testArgs)
		command := exec.Command(cmd, testArgs...)

		output, err := command.CombinedOutput()
		// log.Printf("Duration %+v Length %+v", output[1], len(output))
		command.Run()
		log.Printf("\r\n%s", output)

		if err != nil {
			// statistic <- nil
			close(statistic)
			ch <- srverr.NewWithDetails(http.StatusInternalServerError, err.Error(), string(output))
			return
		}

		stats := make(map[string]interface{})
		err = yaml.Unmarshal([]byte(output), stats)

		if err != nil {
			statistic <- nil
			ch <- srverr.NewWithDetails(http.StatusInternalServerError, "RYFTPRIM "+err.Error(), string(output))
			return
		}

		ch <- nil
		statistic <- createRyftprimStatistic(stats)
		close(statistic)
		close(ch)
	}()
	return
}

func createRyftprimStatistic(m map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"duration":       m[duration],
		"totalBytes":     m[totalBytes],
		"matches":        m[matches],
		"fabricDataRate": m[fabricDataRate],
		"dataRate":       m[dataRate],
	}
}
