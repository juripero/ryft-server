// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rol

import (
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"
)

func TestSearchingNightInPassengers(t *testing.T) {
	filename := filepath.Base(makePassengersFile())

	ds := RolDSCreate()
	if ok := ds.AddFile(filename); !ok {
		panic(filename + " can not be added")
	}

	idxFilename := indexFilename(filename)
	resFilename := resultsFilename(filename)
	resultDs := ds.SearchExact(resFilename, "(RAW_TEXT CONTAINS \"night\" )", 0 /* "\n" */, "", &idxFilename)

	if err := resultDs.HasErrorOccured(); err != nil {
		log.Printf("Srange error: %s\n", err.Error())
	}

	resultDs.Delete()
	ds.Delete()

	results, err := os.Open(filepath.Join("/ryftone", resFilename))
	if err != nil {
		panic(err)
	}
	defer results.Close()

	log.Println("SEARCH RESULTS:")
	if _, err := io.Copy(os.Stdout, results); err != nil {
		panic(err)
	}

	index, err := os.Open(filepath.Join("/ryftone", idxFilename))
	if err != nil {
		panic(err)
	}
	defer index.Close()

	log.Println()
	log.Println("INDEX RESULTS:")
	if _, err := io.Copy(os.Stdout, index); err != nil {
		panic(err)
	}
}

func resultsFilename(filename string) string {
	return filepath.Join("rol-results", filename+"-results.txt")
}

func indexFilename(filename string) string {
	return filepath.Join("rol-results", filename+"-index.txt")
}

const passengersDatatbase = `Name,DoB,Phone,Notes
Hannibal Smith, 10-01-1928,011-310-555-1212,A-team, baby, A-team!
DR. Thomas Magnum, 01-29-1945,310-555-2323,Magnum PI himself.
Steve McGarett, 12-30-1920,310-555-3434,The new Hawaii Five-O.
Michael Knight, 08-17-1952,011-310-555-4545,"Knight Industries Two Thousand. Kitt. He's the driver, sort of."
Stringfellow Hawke, 08-15-1944,310-555-5656,Fictional character who happens to be the chief test pilot during the development of Airwolf.
Sonny Crockett, 12-14-1949,310-555-6767,Mr. Miami Vice himself. Rico Tubbs was his partner.
Michelle Jones,07-12-1959,310-555-1213,Ms. Jones likes to spell her name many different ways.
Mishelle Jones,07-12-1959,310-555-1213,Ms. Jones proves that she likes to spell her first name differently.
Michele Jones,07-12-1959,310-555-1213,Ms. Jones once again shows that she doesn't have command over the spelling of her first name.
T,01-12-1989,310-555-9876,This guy goes by the name 'T'. No more. No less.
DJ,04-25-1985,310-555-3425,I wonder if DJ is just this guy's name or his profession?
`

func makePassengersFile() (file string) {
	f, err := ioutil.TempFile("/ryftone", "go-passengers-")
	if err != nil {
		panic(err)
	}

	_, err = f.WriteString(passengersDatatbase)
	if err != nil {
		panic(err)
	}

	f.Close()

	return f.Name()
}
