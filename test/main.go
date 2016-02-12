package main

import (
	"bytes"
	"encoding/json"
	stdlog "log"
	"os"
	"runtime/pprof"

	// import for side effects
	"github.com/getryft/ryft-server/search"
	_ "github.com/getryft/ryft-server/search/ryfthttp"
	_ "github.com/getryft/ryft-server/search/ryftmux"
	_ "github.com/getryft/ryft-server/search/ryftone"
	_ "github.com/getryft/ryft-server/search/ryftprim"

	msgpack_codec "github.com/getryft/ryft-server/codec/msgpack.v2"
	raw_format "github.com/getryft/ryft-server/format/raw"
	xml_format "github.com/getryft/ryft-server/format/xml"
)

var (
	log Logger = stdlog.Printf
)

// logger function is used as an abstraction to unify
// access to `log.Printf` and `t.Logf` from unit tests
type Logger func(format string, args ...interface{})

// application entry point
func main() {
	defer func() {
		if e := recover(); e != nil {
			stdlog.Fatalf("FAILED")
		}
	}()

	// enable profiling
	if false {
		f, err := os.Create("search.prof")
		if err != nil {
			log("failed to create profile file: %s", err)
			panic(err)
		}

		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	// printReceivedRecords = true
	// ryftprimLogLevel = "debug"
	// ryfthttpLogLevel = "debug"

	//printSearchEngines(log)

	//search1(false) // ryftprim
	//search2(false) // HTTP
	//search3(false) // MUX
	//search4(false) // ryftone

	//count1(false) // ryftprim
	//count2(false) // HTTP
	//count3(false) // MUX
	//count4(true) // ryftone

	//files1(false) // ryftprim
	//files2(false) // HTTP
	//files3(false) // MUX

	testMsgpackFormat()
	//files3(false) // MUX
	//files4(false) // ryftone

	// formatXml()

	//testEncoder()
}

// test msgpack codec and raw format
func testMsgpackFormat() {
	idx := search.Index{}
	idx.File = "test.file.txt"
	idx.Offset = 12345
	idx.Length = 123
	idx.Fuzziness = 100
	idx.Host = "localhost"
	rec := new(search.Record)
	rec.Index = idx
	rec.Data = []byte("test data")

	b := &bytes.Buffer{}
	enc, _ := msgpack_codec.NewStreamEncoder(b)
	enc.EncodeRecord(raw_format.FromRecord(rec))
	enc.Close()

	stdlog.Printf("%s", b.String())
}

// test xml formatter
func formatXml() {
	rec := new(search.Record)
	rec.Index.File = "test.txt"
	rec.Index.Offset = 100
	//rec.Data = []byte(`<rec><ID>10034183</ID><CaseNumber>HY223673</CaseNumber><Date>04/15/2015 11:59:00 PM</Date><Block>062XX S ST LAWRENCE AVE</Block><IUCR>0486</IUCR><PrimaryType>BATTERY</PrimaryType><Description>DOMESTIC BATTERY SIMPLE</Description><LocationDescription>STREET</LocationDescription><Arrest>false</Arrest><Domestic>true</Domestic><Beat>0313</Beat><District>003</District><Ward>20</Ward><CommunityArea>42</CommunityArea><FBICode>08B</FBICode><XCoordinate>1181263</XCoordinate><YCoordinate>1863965</YCoordinate><Year>2015</Year><UpdatedOn>04/22/2015 12:47:10 PM</UpdatedOn><Latitude>41.781961688</Latitude><Longitude>-87.610984705</Longitude><Location>\"(41.781961688, -87.610984705)\"</Location></recx>`)
	rec.Data = []byte(`<rec><ID>10034183</ID><CaseNumber>HY223673</CaseNumber><Date>04/15/2015 11:59:00 PM</Date><Block>062XX S ST LAWRENCE AVE</Block><IUCR>0486</IUCR><PrimaryType>BATTERY</PrimaryType><Description>DOMESTIC BATTERY SIMPLE</Description><LocationDescription>STREET</LocationDescription><Arrest>false</Arrest><Domestic>true</Domestic><Beat>0313</Beat><District>003</District><Ward>20</Ward><CommunityArea>42</CommunityArea><FBICode>08B</FBICode><XCoordinate>1181263</XCoordinate><YCoordinate>1863965</YCoordinate><Year>2015</Year><UpdatedOn>04/22/2015 12:47:10 PM</UpdatedOn><Latitude>41.781961688</Latitude><Longitude>-87.610984705</Longitude><Location>\"(41.781961688, -87.610984705)\"</Location></rec>`)

	b, _ := json.MarshalIndent(xml_format.FromRecord(rec, []string{}), "", " ")
	log("%s", string(b))
}

// abstract seach
func search0(concurent bool, engine search.Engine) {
	// plain texts
	A := search.NewConfig(`10`, "/regression/passengers.txt")
	B := search.NewConfig(`20`, "/regression/passengers.txt")
	C := search.NewConfig(`555`, "/regression/passengers.txt")
	C.Fuzziness = 1

	// XML records
	D := search.NewConfig(`(RECORD.id CONTAINS "1003")`, "/regression/*.pcrime")
	E := search.NewConfig(`(RECORD.id CONTAINS "1003100")`, "/regression/*.pcrime")
	F := search.NewConfig(`(RECORD.desc CONTAINS "VEHICLE")`, "/regression/*.pcrime")

	_, _, _, _, _, _ = A, B, C, D, E, F

	cfgs := []search.Config{}
	cfgs = append(cfgs, *A, *B, *C)
	cfgs = append(cfgs, *D, *E, *F)
	//cfgs = append(cfgs, *A, *B, *C)
	//cfgs = append(cfgs, *D, *E, *F)

	if !concurent {
		runSearchOneByOne(log, engine, cfgs...)
	} else {
		runSearchConcurent(log, engine, cfgs...)
	}
}

// ryftprim seach
func search1(concurent bool) {
	search0(concurent, newRyftPrim(log))
}

// ryfthttp search
func search2(concurent bool) {
	search0(concurent, newRyftHttp(log))
}

// ryftmux search
func search3(concurent bool) {
	// tripple seach
	engine := newRyftMux(log,
		newRyftPrim(log),
		newRyftPrim(log),
		newRyftPrim(log),
	)

	search0(concurent, engine)
}

// ryftone search
func search4(concurent bool) {
	search0(concurent, newRyftOne(log))
}

// abstract count
func count0(concurent bool, engine search.Engine) {
	// plain texts
	A := search.NewConfig(`10`, "/regression/passengers.txt")
	B := search.NewConfig(`20`, "/regression/passengers.txt")
	C := search.NewConfig(`555`, "/regression/passengers.txt")
	C.Fuzziness = 1

	// XML records
	D := search.NewConfig(`(RECORD.id CONTAINS "1003")`, "/regression/*.pcrime")
	E := search.NewConfig(`(RECORD.id CONTAINS "1003100")`, "/regression/*.pcrime")
	F := search.NewConfig(`(RECORD.desc CONTAINS "VEHICLE")`, "/regression/*.pcrime")

	_, _, _, _, _, _ = A, B, C, D, E, F

	cfgs := []search.Config{}
	cfgs = append(cfgs, *A, *B, *C)
	cfgs = append(cfgs, *D, *E, *F)
	//cfgs = append(cfgs, *A, *B, *C)
	//cfgs = append(cfgs, *D, *E, *F)

	if !concurent {
		runCountOneByOne(log, engine, cfgs...)
	} else {
		runCountConcurent(log, engine, cfgs...)
	}
}

// ryftprim count
func count1(concurent bool) {
	count0(concurent, newRyftPrim(log))
}

// ryfthttp count
func count2(concurent bool) {
	count0(concurent, newRyftHttp(log))
}

// ryftmux count
func count3(concurent bool) {
	// tripple seach
	engine := newRyftMux(log,
		newRyftPrim(log),
		newRyftPrim(log),
		newRyftPrim(log),
	)

	count0(concurent, engine)
}

// ryftone count
func count4(concurent bool) {
	count0(concurent, newRyftOne(log))
}

// ryftprim files
func files0(concurent bool, engine search.Engine) {
	paths := []string{}
	paths = append(paths, "/")
	paths = append(paths, "/regression")
	//paths = append(paths, "/not-found")

	if !concurent {
		runFilesStepByStep(log, engine, paths...) // one-by-one
	} else {
		runFilesConcurent(log, engine, paths...) // concurent
	}
}

// ryftprim files
func files1(concurent bool) {
	files0(concurent, newRyftPrim(log))
}

// ryfthttp files
func files2(concurent bool) {
	files0(concurent, newRyftHttp(log))
}

// ryftmux files
func files3(concurent bool) {
	// tripple seach
	engine := newRyftMux(log,
		newRyftPrim(log),
		newRyftPrim(log),
		newRyftPrim(log),
	)

	files0(concurent, engine)
}

// ryftone files
func files4(concurent bool) {
	files0(concurent, newRyftOne(log))
}
