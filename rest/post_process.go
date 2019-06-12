/*
 * ============= Ryft-Customized BSD License ============
 * Copyright (c) 2018, Ryft Systems, Inc.
 * Copyright (c) 2019, BlackLynx Inc.
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
	"os"
	"os/exec"
	"io"
	"io/ioutil"
	"regexp"
	"fmt"
	"strings"
	"strconv"
	"time"
	"reflect"
	"errors"

	"github.com/getryft/ryft-server/search"

)

// Function to search a directory for the latest version of a wild card string
func getLatestFile(dir string, prefix string, ext string) (string, bool) {
	var found = false
	var FName string

	FName = ""
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Infof("Unable to read %s", dir)
		return FName, found
	}
	var modTime time.Time
	s := "^" + prefix + ".*\\." + ext + "$"
	var regex = regexp.MustCompile(s)
	log.Debugf("[POST EXEC] File regex: %s", regex.String())
	for _, fi := range files {
	    if fi.Mode().IsRegular() {
			if regex.MatchString(fi.Name()) {
				if !fi.ModTime().Before(modTime) {
					if fi.ModTime().After(modTime) {
						modTime = fi.ModTime()
						FName = dir + "/" + fi.Name()
						log.Debugf("[POST EXEC] file %s is newer", fi.Name())
						found = true
					}
				}	
			}
		}
	}
	return FName, found

}

// mark output post processing files to delete later
//   Note: Each program may have different timeout
func (server *Server) cleanupPostSession(homeDir string, cfg *search.Config) {

	addTime, _ := time.ParseDuration("0s") 
	if cfg.Lifetime > 0 {
		// Specified on command line
		addTime = cfg.Lifetime
	} else {
		// Get value from server config file entry
		if info, ok := server.Config.FinalProcessor[cfg.JobType]; ok {
			if len(info.FileLifetime) > 0 {
				if newTime, err := time.ParseDuration(info.FileLifetime); err != nil {
					log.Infof("[CONFIG] Invalid time format in %s FileLifetime(%s)",
				                                cfg.JobType, info.FileLifetime)
		    	} else {
					addTime = newTime
				}	
			}
		}	
	} 

	if addTime == 0 {
		return
	}	
	now := time.Now()

	if len(cfg.KeepJobDataAs) != 0 {
		server.addJob("delete-file", cfg.KeepJobDataAs,
			now.Add(addTime))
	}
	if len(cfg.KeepJobIndexAs) != 0 {
		server.addJob("delete-file", cfg.KeepJobIndexAs,
			now.Add(addTime))
	}
	if len(cfg.KeepJobOutputAs) != 0 {
		server.addJob("delete-file", cfg.KeepJobOutputAs,
			now.Add(addTime))
	}
}

func (server *Server) runPostCommand(cfg *search.Config) ([]string, error) {
	var	resultInfo	[]string
	var myArgs 		[]string
	var Executable	string
	var	ConfigFile	string
	log.Debugf("[POST EXEC] enter - type: %s", cfg.JobType)

	// Get ryft-server.conf parameters for target post processing job
	if info, ok := server.Config.FinalProcessor[cfg.JobType]; ok {
		// get information from ryft-server.conf and command parameters
		Executable = info.Exec
		ConfigFile = info.ConfigFile
		for k, v := range cfg.PostExecParams {
			myArgs = append(myArgs, k)
			if v.(string) != "" {
				myArgs = append(myArgs, v.(string))
			}	
		}	
	} else {
		return resultInfo, errors.New("PostCommand type not found")
	}
	// Build command struct for needs of each program
	switch cfg.JobType {
	case "blgeo":
		// blegeo specific preprocessing to setup args for blgeo call
		// if append, get previous kml output for appending
		if val, ok := cfg.PostExecParams["--kml-append"]; ok {
			log.Debugf("[POST EXEC] Replacing --kml-append %s with kml file", val)
			for i, val := range myArgs {
				if val == "--kml-append" && myArgs[i+1] == "create" {
					myArgs[i+1] = "/ryftone/jobs/blgeo_out_" + cfg.JobID + ".kml"
					break
				}	
			}	
		}

		// Add --pip if not on commandline (needed for KML boundaries)
		if _, ok := cfg.PostExecParams["--pip"]; !ok {
			re := regexp.MustCompile(`(?i) (?P<inOut>contains|not_contains)\s*pip\(vertex_file=["\\]+.*?polygons\/(?P<vFile>[^"\\]+)[^\)]*\)+\s*(?P<joiner>and|or)?`)
			log.Debugf("[POST EXEC] regex: %s", re.String())
			match := re.FindAllStringSubmatch(cfg.Query, -1)
			myTerms	:= ""
			storedConnector := ""
			for i, vals := range match {
				log.Debugf("[POST EXEC] %d: %q", i, vals)
				if i < len(match) {
					s := ""
					if strings.EqualFold(vals[1], "CONTAINS") {
						s = fmt.Sprintf("(in %s)", vals[2])
					} else {
						s = fmt.Sprintf("(out %s)", vals[2])
					}	
					if len(storedConnector) > 0 {
						myTerms = myTerms + " and "
					}
					storedConnector = vals[3]
					myTerms = myTerms + s
				}
			}	
			myArgs = append(myArgs, "--pip")
			myArgs = append(myArgs, myTerms)
		} 
		
		// add default config file if not specified on command line
		if _, ok := cfg.PostExecParams["--cfg"]; !ok {
			if _, err := os.Stat(ConfigFile); err == nil {
				myArgs = append(myArgs, "--cfg")
				myArgs = append(myArgs, ConfigFile)
			}	
		}	
		// add index file for blgeo
		myArgs = append(myArgs, "--oi-for-fusion")
		myArgs = append(myArgs, cfg.KeepJobIndexAs)
		// add data file (not needed, but blgeo won't run without)
		myArgs = append(myArgs, "-f")
		myArgs = append(myArgs, cfg.KeepJobDataAs)
		break
	default:
		log.Debugf("[POST EXEC] Using default parameter settings")
		break
	}
//	for i, val := range myArgs {
//		log.Debugf("[POST EXEC] arg: %d |%s| ", i, val)
//	}	
		
	cmd := exec.Command(Executable, myArgs...)
	log.Debugf("[POST EXEC] command: %s %s", Executable, strings.Join(myArgs, " "))
	CmdErr, err := cmd.StderrPipe()
	if err != nil {
		resultInfo = append(resultInfo, "Error on command set stderr")
		return resultInfo, err
	}
	CmdOut, err := cmd.StdoutPipe()
	if err != nil {
		resultInfo = append(resultInfo, "Error on command set stdout")
		return resultInfo, err
	}
	if err := cmd.Start(); err != nil {
		resultInfo = append(resultInfo, "Error on command execution")
		return resultInfo, err
	}
	log.Debugf("[POST EXEC] started execution - pid: %d", cmd.Process.Pid)
	// read the stderr and stdout buffers
	stderrBuf, _ := ioutil.ReadAll(CmdErr)
	sError := fmt.Sprintf("%s", stderrBuf)
	log.Debugf("[POST EXEC] stderr:  %s", sError)

	stdoutBuf, _ := ioutil.ReadAll(CmdOut)
	s := fmt.Sprintf("%s", stdoutBuf)

	if err := cmd.Wait(); err != nil {
		resultInfo = append(resultInfo, sError)
		return resultInfo, err
	}

	switch cfg.JobType {
	case "blgeo":
		// For blgeo: copy result kml out of /tmp and into jobs directory
		if oldFile, ok := getLatestFile("/tmp", "blgeo_out", "kml"); ok {
			newFile := fmt.Sprintf("/ryftone/jobs/blgeo_out_%s.kml", cfg.JobID)
			tmpFile, err := os.Open(oldFile)
			if err != nil {
				s := fmt.Sprintf("Unable to open kml file: %s", err.Error())
				resultInfo = append(resultInfo, s)
				return resultInfo, err
			}
			jobFile, err := os.Create(newFile)
			if err != nil {
				s := fmt.Sprintf("Unable to open new kml file: %s", err.Error())
				resultInfo = append(resultInfo, s)
				return resultInfo, err
			}
			defer jobFile.Close()
			_, err = io.Copy(jobFile, tmpFile)
			tmpFile.Close()
			if err != nil {
				s := fmt.Sprintf("Unable to rename output file: %s", err.Error())
				resultInfo = append(resultInfo, s)
				return resultInfo, err
			}

			cfg.KeepJobOutputAs = newFile
			log.Debugf("[POST EXEC] Renamed %s to %s", oldFile, newFile)
		}	
		break
	default:
		log.Debugf("[POST EXEC] Job post processing using default operations")
		break
	}
	resultInfo = append(resultInfo, s)
	return resultInfo, nil
}

// Function to turn results into CSV line
func makeCsvLine(data interface{}, FieldNames []string, dataType string) string {
	
	log.Debugf("[POST EXEC]: makeCsvLine(interface{}, FieldNames, %s)", dataType)
	outStr := make([]string, len(FieldNames), len(FieldNames))

	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Ptr {
		v = reflect.Indirect(v)
		log.Debugf("[POST EXEC]: In Mapkeys - handle indirect", CORE) 
	}
	vType := reflect.TypeOf(v)
	log.Debugf("[%s/MNB]: In MapKeys: name: %v, kind: %s, fields: %d", CORE, vType.Name(), vType.Kind(), vType.NumField())
//	for i := 0; i < vType.NumField(); i++ {
//		log.Debugf("[%s/MNB]: In MapKeys - Field: %d, name: %s, type: %s, kind: %s", CORE, i+1, vType.Field(i).Name, vType.Field(i).Type.Name(), vType.Field(i).Type.Kind())
//	}
	switch v.Kind() {
	case reflect.Map:	//format is csv
//		log.Debugf("[%s/MNB]: In Mapkeys - Map", CORE) 
//		if vType.Field(0).Type.Kind() == reflect.Ptr {
//			v = reflect.Indirect(vType.Field(0).Type.Kind())
//		}
		for _, key := range v.MapKeys() {
			log.Debugf("[%s/MNB]: In MapKeys: %s", CORE, key.String())
			for i, item := range FieldNames {
				if item == key.String() {
					f := fmt.Sprintf("%s", v.MapIndex(key))
					if strings.Index(f, ",") > -1 {
						f = strconv.Quote(f)
					}
					outStr[i] = f
//					log.Debugf("[%s/MNB]: appended: %s(%d)", CORE, f, i)
				}
			}	
//			xStr := fmt.Sprintf("%s:%s\n", key.String(), v.MapIndex(key))
//			csvStr = csvStr + xStr
		}
	case reflect.Struct:
		log.Debugf("[%s/MNB]: In Mapkeys - Struct kind: %v", CORE, reflect.TypeOf(v)) 
		for i := 0; i < vType.NumField(); i++ {
//		    f := vType.Field(i)
//			myS := fmt.Sprintf("Mapkeys - Field: %d, name: %s, type: %s, kind: %s", i+1, vType.Field(i).Name, vType.Field(i).Type.Name(), vType.Field(i).Type.Kind())
			log.Debugf("[%s/MNB]: In MapKeys - Field: %d, name: %s, type: %s, kind: %s", CORE, i+1, vType.Field(i).Name, vType.Field(i).Type.Name(), vType.Field(i).Type.Kind())
		}
	default:
		log.Debugf("[%s/MNB]: In Mapkeys - default kind: %v", CORE, reflect.TypeOf(v)) 
	}	
	s := strings.Join(outStr, ",")
	return s
}
