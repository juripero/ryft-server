/*
 * ============= Ryft-Customized BSD License ============
 * Copyright (c) 2015, Ryft Systems, Inc.
 * All rights reserved.
 * Redistribution and use in source and binary forms, with or without modification,
 * are permitted provided that the following conditions are met:
 *
 * 1. Redistributions of source code must retain the above copyright notice,
 *   this list of conditions and the following disclaimer.
 * 2. Redistributions in binary form must reproduce the above copyright notice,
 *   this list of conditions and the following disclaimer in the documentation and/or
 *   other materials provided with the distribution.
 * 3. All advertising materials mentioning features or use of this software must display the following acknowledgement:
 *   This product includes software developed by Ryft Systems, Inc.
 * 4. Neither the name of Ryft Systems, Inc. nor the names of its contributors may be used
 *   to endorse or promote products derived from this software without specific prior written permission.
 *
 * THIS SOFTWARE IS PROVIDED BY RYFT SYSTEMS, INC. ''AS IS'' AND ANY
 * EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
 * WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
 * DISCLAIMED. IN NO EVENT SHALL RYFT SYSTEMS, INC. BE LIABLE FOR ANY
 * DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
 * (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
 * LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND
 * ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
 * (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
 * SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
 * ============
 */

package main

import (
	"errors"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/getryft/ryft-server/middleware/auth"
	"github.com/getryft/ryft-server/middleware/gzip"
	"github.com/getryft/ryft-server/names"

	"github.com/gin-gonic/gin"
)

var (
	//KeepResults - Flag that enables/disables log saving
	KeepResults                   bool
	authType, fileName, groupName string
	portPtr                       int
	flagArgs                      []string
)

const (
	none            = "none"
	authBasicSystem = "basic-system"
	authBasicFile   = "basic-file"
)

func readParameters() {
	//Port number
	flag.IntVar(&portPtr, "port", 8765, "The port of the REST-server")
	flag.IntVar(&portPtr, "p", 8765, "The port of the REST-server (shorthand)")
	//keep-results
	flag.BoolVar(&KeepResults, "keep-results", false, "Keep results or delete after response")
	flag.BoolVar(&KeepResults, "k", false, "Keep results or delete after response (shorthand)")
	//Auth type
	flag.StringVar(&authType, "auth", none, "Endable or Disable BasicAuth (can be \"none\" \"basic-system -g *usergroup*\" \"basic-file -f *filename*\")")
	flag.StringVar(&authType, "a", none, "Endable or Disable BasicAuth (can be \"none\" \"basic-system -g *usergroup*\" \"basic-file -f *filename*\") (shorthand)")
	//Users group
	flag.StringVar(&groupName, "users-group", "", "Add user group for the \"basic-system\" ")
	flag.StringVar(&groupName, "g", "", "Add user group for the \"basic-system\" (shorthand)")
	//Users file
	flag.StringVar(&fileName, "users-file", "", "Add user file for the \"basic-file\")")
	flag.StringVar(&fileName, "f", "", "Add user file for the \"basic-file\") (shorthand)")

	flag.Parse()
	flagArgs = flag.Args()
}

func main() {
	log.SetFlags(log.Lmicroseconds)

	readParameters()

	r := gin.Default()

	//User credentials examples

	indexTemplate := template.Must(template.New("index").Parse(IndexHTML))
	r.SetHTMLTemplate(indexTemplate)

	switch authType {
	case authBasicFile:
		if fileName == "" {
			log.Printf("File parameter users-file is required when using basic-file authentication.")
			os.Exit(1)
		}
		auth, err := auth.AuthBasicFile(fileName)
		if err != nil {
			log.Printf("Error reading users-file %v", err)
			os.Exit(1)
		}
		r.Use(auth)
		break
	case authBasicSystem:
		settings, err := parseParams(flagArgs)
		if err != nil {
			log.Printf(err.Error())
			os.Exit(1)
		} else {

			r.Use(auth.BasicAuthLDAP(settings))
		}

		break

	}

	r.Use(gzip.Gzip(gzip.DefaultCompression))
	//Setting routes
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index", nil)
	})
	r.GET("/search", search)

	// Clean previously created folder
	if err := os.RemoveAll(names.ResultsDirPath()); err != nil {
		log.Printf("Could not delete %s with error %s", names.ResultsDirPath(), err.Error())
		os.Exit(1)
	}

	// Create folder for results cache
	if err := os.MkdirAll(names.ResultsDirPath(), 0777); err != nil {
		log.Printf("Could not create directory %s with error %s", names.ResultsDirPath(), err.Error())
		os.Exit(1)
	}

	// Name Generator will produce unique file names for each new results files
	names.StartNamesGenerator()
	r.Run(fmt.Sprintf(":%d", names.Port))

}
func parseParams(flagArgs []string) (auth.Settings, error) {
	var settings auth.Settings
	var url, port, userPrefix, userPostfix string
	for _, s := range flagArgs {
		if strings.Contains(s, "url=") {
			fmt.Println(s + "\n")
			url = strings.Replace(s, "url=", "", 1)
		} else if strings.Contains(s, "usr=") {
			userPostfix = "1"
			userPrefix = "1"
			fmt.Println(s + "\n")
		} else if strings.Contains(s, "port=") {
			port = strings.Replace(s, "port=", "", 1)
			fmt.Println(s + "\n")
		}
	}
	if url != "" && port != "" && userPrefix != "" && userPostfix != "" {
		settings = auth.Settings{
			port,
			url,
			auth.UserDN{
				userPrefix,
				userPostfix,
			},
		}
		fmt.Println("noerror")
		return settings, nil
	} else {
		fmt.Println("error")
		return settings, errors.New("Invalid parameters")
	}
}

// https://golang.org/src/net/http/status.go -- statuses
