package main

// func StreamJson(resultsFile, idxFile *os.File, w io.Writer, completion chan error, sleepiness time.Duration) {
// wEncoder := json.NewEncoder(w)
// idxRecords := make(chan IdxRecord, 64)
// dropConnection := make(chan struct{}, 1)
// 	go func() {
// 		for {
// 			select {
// 			case <-completion:
// 				// log.Println("** streaming completion")
// 				recordsScan(idxFile, idxRecords, sleepiness)
// 				close(idxRecords)
// 				return

// 			case <-dropConnection:
// 				close(idxRecords)
// 				return

// 			default:
// 				// log.Println("** streaming continue")
// 				recordsScan(idxFile, idxRecords, sleepiness)
// 			}
// 		}
// 	}()

// 	if _, err := w.Write([]byte("[")); err != nil {
// 		return
// 	}
// 	defer func() {
// 		if _, err := w.Write([]byte("]")); err != nil {
// 			return
// 		}
// 	}()

// 	var err error
// 	firstIteration := true
// 	for r := range idxRecords {
// 		if !firstIteration {
// 			w.Write([]byte(","))
// 		}

// 		r.Data = readDataBlock(resultsFile, r.Length, sleepiness)

// 		if err = wEncoder.Encode(r); err != nil {
// 			log.Printf("Encoding error: %s", err.Error())
// 			dropConnection <- struct{}{}
// 			return
// 		}

// 		firstIteration = false
// 	}
// }

// func recordsScanner(idx *os.File, watcher *fsnotify.Watcher, ch chan error, records chan IdxRecord) {

// 	for {
// 		var line string
// 		n, _ := fmt.Fscanln(idx, &line)

// 		if n == 0 {
// 			break // waiting for write event
// 		}

// 		r, _ := NewIdxRecord(line)
// 		records <- r
// 	}

// 	select {
// 	case e := <-watcher.Events:
// 		if e.Op&fsnotify.Write == fsnotify.Write && e.Name == ResultsDirPath(n.IdxFile) {
// 			continue
// 		}
// 	case err := <-watcher.Errors:
// 		log.Printf("records watching error: %s", err)

// 	}
// }

// func streamJson(idx, res *os.File, w io.Writer, watcher *fsnotify.Watcher, ch chan error) {
// 	wEncoder := json.NewEncoder(w)
// 	idxRecords := make(chan IdxRecord, 64)
// 	dropConnection := make(chan struct{}, 1)

// 	//go recordsScanner(idx, watcher, ch, idxRecords)

// }

// func recordsScan(r io.Reader, recordsChan chan IdxRecord, sleepiness time.Duration) {
// 	for {
// 		var line string
// 		n, _ := fmt.Fscanln(r, &line)
// 		if n == 0 {
// 			//log.Printf("** number of lines = 0, with error: %s", e.Error())
// 			time.Sleep(sleepiness)
// 			break
// 		}
// 		// else {
// 		// 	log.Printf("** Scanned line %s", line)
// 		// }

// 		r, err := NewIdxRecord(line)
// 		if err != nil {
// 			break
// 		}

// 		recordsChan <- r
// 	}
// }

// func readDataBlock(r io.Reader, length uint16, sleepiness time.Duration) (result []byte) {
// 	var total uint16 = 0
// 	for total < length {
// 		data := make([]byte, length-total)
// 		n, _ := r.Read(data)
// 		if n != 0 {
// 			result = append(result, data...)
// 			total = total + uint16(n)
// 		} else {
// 			time.Sleep(sleepiness)
// 		}
// 	}
// 	return
// }
