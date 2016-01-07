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
	"github.com/gin-gonic/gin"
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

type RyftprimParams struct {
	Query         string
	Files         []string
	Surrounding   uint16
	Fuzziness     uint8
	Format        string
	CaseSensitive bool
	Fields        string
	Keys          []string
	Nodes         uint8
}

type RyftprimOut struct {
	Duration       uint64
	TotalBytes     uint64
	Matches        uint64
	FabricDataRate string
	DataRate       string
}

func ryftprim(p *RyftprimParams, n *names.Names) (ch chan error, headers chan map[interface{}]interface{}) {
	ch = make(chan error, 1)
	headers = make(chan map[interface{}]interface{})
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
			ch <- srverr.New(http.StatusBadRequest, aErr.Error())
			return
		}
		testArgs = append(testArgs, arg_query, query)

		log.Print(testArgs)
		command := exec.Command(cmd, testArgs...)

		output, err := command.CombinedOutput()
		// log.Printf("\r\n%s", output)
		// log.Printf("Duration %+v Length %+v", output[1], len(output))
		command.Run()

		if err != nil {
			ch <- srverr.NewWithDetails(http.StatusInternalServerError, err.Error(), string(output))
			return
		}

		m := make(map[interface{}]interface{})
		err = yaml.Unmarshal([]byte(output), m)
		// log.Printf("--- m:\n%v\n\n", m)
		//
		// log.Printf("MAPA Value: %v ", m["Duration"])
		if err != nil {
			ch <- srverr.NewWithDetails(http.StatusInternalServerError, err.Error(), string(output))
			return
		}
		headers <- m
		ch <- nil
	}()

	return
}

func setHeaders(c *gin.Context, m map[interface{}]interface{}) {
	log.Printf("--- m:\n%v\n\n", m)
	c.Header("X-Duration", fmt.Sprintf("%+v", m["Duration"]))
	c.Header("X-Total-Bytes", fmt.Sprintf("%+v", m["Total Bytes"]))
	c.Header("X-Matches", fmt.Sprintf("%+v", m["Matches"]))
	c.Header("X-Fabric-Data-Rate", fmt.Sprintf("%+v", m["Fabric Data Rate"]))
	c.Header("X-Data-Rate", fmt.Sprintf("%+v", m["Data Rate"]))
}
