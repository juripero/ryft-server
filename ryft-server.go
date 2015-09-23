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
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/getryft/ryft-server/middleware/gzip"
	"github.com/getryft/ryft-server/names"

	"github.com/gin-gonic/gin"
)

var (
	KeepResults = false
	AuthVar     = authNone
)

const (
	authNone        = "none"
	authBasicSystem = "basic-system"
	authBasicFile   = "basic-file"
)

func readParameters() {
	portPtr := flag.Int("port", 8765, "The port of the REST-server")
	keepResultsPtr := flag.Bool("keep-results", false, "Keep results or delete after response")
	authVar := flag.String("auth", "none", "Endable or Disable BasicAuth (can be \"none\" \"base-system\" \"base-file\")")
	flag.Parse()

	names.Port = *portPtr
	KeepResults = *keepResultsPtr
	AuthVar = *authVar
}

func main() {
	log.SetFlags(log.Lmicroseconds)
	readParameters()

	r := gin.Default()

	//User credentials examples

	r.Use(gzip.Gzip(gzip.DefaultCompression))

	indexTemplate := template.Must(template.New("index").Parse(IndexHTML))
	r.SetHTMLTemplate(indexTemplate)

	switch AuthVar {
	case authNone:
		r.GET("/", func(c *gin.Context) {
			c.HTML(http.StatusOK, "index", nil)
		})

		r.GET("/search", search)
		break
	case authBasicFile:
		break
	case authBasicSystem:
		authorized := r.Group("/", gin.BasicAuth(gin.Accounts{
			"eugene": "123",
			"admin":  "admin",
		}))
		// /login endpoint
		// hit "localhost:PORT/login
		authorized.GET("/", func(c *gin.Context) {
			// get user, it was setted by the BasicAuth middleware
			_ = c.MustGet(gin.AuthUserKey).(string)
			c.HTML(http.StatusOK, "index", nil)

		})
		//
		authorized.GET("/search", func(c *gin.Context) {
			_ = c.MustGet(gin.AuthUserKey).(string)
		}, search)
		break

	}

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

// https://golang.org/src/net/http/status.go -- statuses
