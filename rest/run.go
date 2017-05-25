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

package rest

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

// RunParams contains all the bound parameters for the /run endpoint.
type RunParams struct {
	Image   string   `form:"image" json:"image" msgpack:"image"`
	Command string   `form:"command" json:"command" msgpack:"command"`
	Args    []string `form:"arg" json:"args,omitempty" msgpack:"args,omitempty"`

	Local bool `form:"local" json:"local,omitempty" msgpack:"local,omitempty"`
}

// Handle /run endpoint.
func (server *Server) DoRun(ctx *gin.Context) {
	// recover from panics if any
	defer RecoverFromPanic(ctx)

	// parse request parameters
	params := RunParams{
		Image: "default",
		Local: true,
	}
	if err := binding.Form.Bind(ctx.Request, &params); err != nil {
		panic(NewError(http.StatusBadRequest, err.Error()).
			WithDetails("failed to parse request parameters"))
	}

	// get search engine
	userName, _, homeDir, _ := server.parseAuthAndHome(ctx)
	mountPoint, err := server.getMountPoint(homeDir)
	if err != nil {
		panic(NewError(http.StatusInternalServerError, err.Error()).
			WithDetails("failed to get ryftone mount point"))
	}

	// build executable command
	var args []string
	if image, ok := server.Config.DockerImages[params.Image]; !ok {
		panic(NewError(http.StatusBadRequest,
			fmt.Sprintf("image %q not found", params.Image)))
	} else {
		// copy and expand arguments
		for _, arg := range image.ExecArgs {
			arg = os.Expand(arg, func(name string) string {
				switch name {
				case "USER":
					return userName
				case "HOME":
					return filepath.Join(mountPoint, homeDir)
				case "RYFTONE":
					return mountPoint
				}

				return "" // not found
			})

			args = append(args, arg)
		}
	}

	// add +x permission
	if true {
		var cmd string
		if len(params.Command) != 0 {
			cmd = params.Command
		} else if len(params.Args) != 0 {
			cmd = params.Args[0]
		}

		if len(cmd) != 0 {
			path := filepath.Join(mountPoint, homeDir, cmd)
			if info, err := os.Stat(path); err == nil {
				err := os.Chmod(path, info.Mode()|0111) // +x
				if err != nil {
					panic(NewError(http.StatusInternalServerError, err.Error()).
						WithDetails("failed to add +x permission"))
				}
				log.WithField("path", path).Debugf("[%s]: added +x permission ()", CORE)
			}
		}
	}

	if len(params.Command) != 0 {
		args = append(args, params.Command)
	}
	args = append(args, params.Args...)
	if len(args) == 0 || len(args[0]) == 0 {
		panic(NewError(http.StatusBadRequest,
			"no any command or argument provided"))
	}

	log.WithFields(map[string]interface{}{
		"args": args,
		"user": userName,
		"home": homeDir,
	}).Infof("[%s]: start GET /run", CORE)

	// run docker image
	cmd := exec.Command(args[0], args[1:]...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		panic(NewError(http.StatusInternalServerError, err.Error()).
			WithDetails("failed to execute"))
	}

	ctx.Data(http.StatusOK, "application/octet-stream", out)
}
