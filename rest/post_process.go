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
//	"encoding/json"

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
	for _, fi := range files {
	    if fi.Mode().IsRegular() {
			if regex.MatchString(fi.Name()) {
				if !fi.ModTime().Before(modTime) {
					if fi.ModTime().After(modTime) {
						modTime = fi.ModTime()
						FName = dir + "/" + fi.Name()
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
		if _, ok := cfg.PostExecParams["--kml-append"]; ok {
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
			match := re.FindAllStringSubmatch(cfg.Query, -1)
			myTerms	:= ""
			storedConnector := ""
			for i, vals := range match {
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
	// read the stderr and stdout buffers
	stderrBuf, _ := ioutil.ReadAll(CmdErr)
	sError := fmt.Sprintf("%s", stderrBuf)
	log.Debugf("[POST EXEC] stderr: %s", sError)

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
			newFile := server.getPostProcessingOutputPath(cfg) + "/blgeo_out_" + cfg.JobID + ".kml"
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
func makeCsvLine(data interface{}, cfg *search.Config) (string, error) {
	
	var outStr []string

	FieldNames := getCsvHeaderNames(cfg, true)
//	log.Debugf("[CSV PARSE] record: %+v", data)
	log.Debugf("[CSV PARSE] Enter makeCsvLine: Fieldnames: %q", FieldNames)

	if FieldNames[0] != "" {
		for _, item := range FieldNames {
			Hierarchy := strings.Split(item,".")
			log.Debugf("[CSV] Find Field Name: %s, depth: %d", item, len(Hierarchy))
			not_found := true
			finalVal := ""
			V := reflect.ValueOf(data)
			i := 0
			testVal := len(Hierarchy) - 1
			for ; i < len(Hierarchy) && not_found == true;  {
				if V.Kind() == reflect.Ptr {
					V = V.Elem()
					log.Debugf("[CSV]: Following ptr: kind: %s", V.Kind()) 
				}
				switch V.Kind() {
				case reflect.Map:
					log.Debugf("[CSV] processing map")
					for _, key := range V.MapKeys() {
//						log.Debugf("[CSV]   range MapKeys: %s", key.String())
						if Hierarchy[i] == key.String() {
							if i == testVal {
								V = V.MapIndex(key).Elem()
								if V.Kind() == reflect.Ptr {
									V = V.Elem()
								}	
								if V.Kind() == reflect.Slice {
									log.Debugf("[CSV] slice: %#v, len: %d, type: %T", V, V.Len(), V)
									tVal := V.Interface()
//									log.Debugf("[CSV] interface: %#v, type: %T", tVal, tVal)
									finalVal = fmt.Sprintf("%v", tVal)
									
								} else {
									finalVal = fmt.Sprintf("%s", V.String())
								}
								log.Debugf("[CSV] Match: %s %s", item, finalVal)
								not_found = false
							} else {
								log.Debugf("[CSV] map matched intermediate(%d): %s - %+v", i, key.String(), V.MapIndex(key))
								V = V.MapIndex(key).Elem()
//								log.Debugf("[CSV]    New Value: %+v", V)
								log.Debugf("[CSV]    map type: %s, kind: %s", V.Type(), V.Kind())
								i++
							}
							break
						}
					}	
					break
				default:
					log.Debugf("[CSV] Need to add type: %s", V.Kind().String())
					finalVal = ""
					not_found = false
					break
				} // end switch
				if not_found == false {
					if strings.Index(finalVal, ",") > -1 {
						finalVal = strconv.Quote(finalVal)
					}
					outStr = append(outStr, finalVal)
				}	
			}	// end for searching for one field
		}  // end for list of fields to return	
	} else {
		// No request for particular fields so give all
		//    (Note: Only returns flat JSON/XML)
		v := reflect.ValueOf(data)
		if v.Kind() == reflect.Ptr {
			v = reflect.Indirect(v)
			log.Debugf("[CSV]: Following ptr(2): kind: %s", v.Kind()) 
		}
		for _, key := range v.MapKeys() {
			f := fmt.Sprintf("%s", v.MapIndex(key))
			if strings.Index(f, ",") > -1 {
				f = strconv.Quote(f)
			}
			outStr = append(outStr, f)
			log.Debugf("[CSV PARSE] Match(1): %s '%s'", key.String(), f)
		}	
	}
	
	s := strings.Join(outStr, ",")
	return s, nil
}

// Function build directory path for post processing output files
func (server *Server) getPostProcessingOutputPath(cfg *search.Config) string {
	var middlePath string

	mountPoint, _ := server.getMountPoint()

	if info, ok := server.Config.FinalProcessor[cfg.JobType]; ok {
		// This job specifies its own path
		middlePath = info.OutDirectory
	} else {
		middlePath = server.Config.FinalProcessor["defaults"].OutDirectory
	}	
	path := mountPoint + "/" + middlePath
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.MkdirAll(path, 0775)
	}
	return  path
}

// Function to extract post processing csv file column names
func getCsvHeaderNames(cfg *search.Config, getKeys bool) []string {
	Headers	:= []string{}

	// If user specified CSV fields in tweaks, use that
	if len(cfg.CsvFields) > 0 {
		v := reflect.ValueOf(cfg.CsvFields)
		for _, key := range v.MapKeys() {
			log.Debugf("[POST EXEC] key: %s, val: %s", key.String(), v.MapIndex(key))
			if getKeys == true {
				Headers = append(Headers, fmt.Sprintf("%s", key.String()))
			} else {	
				Headers = append(Headers, fmt.Sprintf("%s", v.MapIndex(key)))
			}	
		}
	} else if len(cfg.Fields) > 0 {
		// Use --fields entry
		Headers = strings.Split(cfg.Fields,",")
	} 

	return Headers
}	
