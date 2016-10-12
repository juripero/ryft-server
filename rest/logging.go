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
	"strings"

	"github.com/getryft/ryft-server/search/ryftdec"
	"github.com/getryft/ryft-server/search/ryfthttp"
	"github.com/getryft/ryft-server/search/ryftmux"
	"github.com/getryft/ryft-server/search/ryftone"
	"github.com/getryft/ryft-server/search/ryftprim"
	"github.com/getryft/ryft-server/search/utils/catalog"

	"github.com/Sirupsen/logrus"
)

var (
	// logger instances
	log     = logrus.New()
	pjobLog = logrus.New() // pending jobs
	busyLog = logrus.New() // cluster business
)

// set logging level
func setLoggingLevel(logger string, level string) error {
	ll, err := logrus.ParseLevel(level)
	if err != nil {
		return fmt.Errorf("failed to parse level: %s", err)
	}

	switch strings.ToLower(logger) {
	case "core":
		log.Level = ll
	case "core/catalogs":
		catalog.SetLogLevel(ll)
	case "core/pending-jobs":
		pjobLog.Level = ll
	case "core/busyness":
		busyLog.Level = ll
		// TODO: more core loggers
	case "search/ryftprim":
		ryftprim.SetLogLevel(ll)
	case "search/ryftone":
		ryftone.SetLogLevel(ll)
	case "search/ryfthttp":
		ryfthttp.SetLogLevel(ll)
	case "search/ryftmux":
		ryftmux.SetLogLevel(ll)
	case "search/ryftdec":
		ryftdec.SetLogLevel(ll)
	default:
		return fmt.Errorf("'%s' is unknown logger name", logger)
	}

	return nil // OK
}

func makeDefaultLoggingOptions(level string) map[string]string {
	return map[string]string{
		"core":              level,
		"core/catalogs":     level,
		"core/pending-jobs": level,
		"core/busyness":     level,
		"search/ryftprim":   level,
		"search/ryftone":    level,
		"search/ryfthttp":   level,
		"search/ryftmux":    level,
		"search/ryftdec":    level,
	}
}
