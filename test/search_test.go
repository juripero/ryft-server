package main

import (
	"testing"

	"github.com/getryft/ryft-server/search"
)

// more debug info!
func init() {
	// ryfthttpServerUrl = "http://localhost:8765"
	// ryfthttpLogLevel = "debug"
	// ryftprimLogLevel = "debug"
	// printReceivedRecords = true
}

// check if expected number of records received
func checkRecordReceived(t *testing.T, r SearchResult, expected int) {

	// special case for error expected
	if expected < 0 {
		if len(r.Errors) == 0 {
			t.Errorf("error expected")
			return
		}

		if r.Stat != nil {
			t.Errorf("no received statistics expected")
			return
		}

		return // OK
	}

	if r.Stat == nil {
		t.Errorf("no received statistics")
		return
	}

	if len(r.Errors) > 0 {
		t.Errorf("%d errors received", len(r.Errors))
		return
	}

	if r.Stat.Matches != uint64(expected) {
		t.Errorf("unexpected %d matches (expected: %d)",
			r.Stat.Matches, expected)
		return
	}

	if len(r.Records) != expected {
		t.Errorf("unexpected %d records received (expected: %d)",
			len(r.Records), expected)
		return
	}

	return // OK
}

// ryftprim search (bad result)
func TestSearchPrim_Bad10(t *testing.T) {
	cfg := search.NewConfig("10", "/regression/passengers_not_found.txt")
	r := runSearch1(t.Logf, "TEST", newRyftPrim(t.Logf), cfg)
	checkRecordReceived(t, r, -1)
}

// ryftprim search
func TestSearchPrim_10(t *testing.T) {
	cfg := search.NewConfig("10", "/regression/passengers.txt")
	r := runSearch1(t.Logf, "TEST", newRyftPrim(t.Logf), cfg)
	checkRecordReceived(t, r, 12)
}

// ryftprim search
func TestSearchPrim_310(t *testing.T) {
	cfg := search.NewConfig("310", "/regression/passengers.txt")
	r := runSearch1(t.Logf, "TEST", newRyftPrim(t.Logf), cfg)
	checkRecordReceived(t, r, 11)
}

// ryftprim search
func TestSearchPrim_555(t *testing.T) {
	cfg := search.NewConfig("555", "/regression/passengers.txt")
	r := runSearch1(t.Logf, "TEST", newRyftPrim(t.Logf), cfg)
	checkRecordReceived(t, r, 11)
}

// ryftprim XML search
// check corresponding RDF file is loaded
func TestSearchPrim_XML_id1003(t *testing.T) {
	cfg := search.NewConfig(`(RECORD.id CONTAINS "1003")`, "/regression/*.pcrime")
	r := runSearch1(t.Logf, "TEST", newRyftPrim(t.Logf), cfg)
	checkRecordReceived(t, r, 2542)
}

// ryftprim XML search
// check corresponding RDF file is loaded
func TestSearchPrim_XML_id1003100(t *testing.T) {
	cfg := search.NewConfig(`(RECORD.id CONTAINS "1003100")`, "/regression/*.pcrime")
	r := runSearch1(t.Logf, "TEST", newRyftPrim(t.Logf), cfg)
	checkRecordReceived(t, r, 9)
}

// ryftprim XML search
// check corresponding RDF file is loaded
func TestSearchPrim_XML_descVEHICLE(t *testing.T) {
	cfg := search.NewConfig(`(RECORD.desc CONTAINS "VEHICLE")`, "/regression/*.pcrime")
	r := runSearch1(t.Logf, "TEST", newRyftPrim(t.Logf), cfg)
	checkRecordReceived(t, r, 672)
}

// ryfthttp search (bad result)
func TestSearchHttp_Bad10(t *testing.T) {
	cfg := search.NewConfig("10", "/regression/passengers_not_found.txt")
	r := runSearch1(t.Logf, "TEST", newRyftHttp(t.Logf), cfg)
	checkRecordReceived(t, r, -1)
}

// ryfthttp search
func TestSearchHttp_10(t *testing.T) {
	cfg := search.NewConfig("10", "/regression/passengers.txt")
	r := runSearch1(t.Logf, "TEST", newRyftHttp(t.Logf), cfg)
	checkRecordReceived(t, r, 12)
}

// ryfthttp search
func TestSearchHttp_310(t *testing.T) {
	cfg := search.NewConfig("310", "/regression/passengers.txt")
	r := runSearch1(t.Logf, "TEST", newRyftHttp(t.Logf), cfg)
	checkRecordReceived(t, r, 11)
}

// ryfthttp search
func TestSearchHttp_555(t *testing.T) {
	cfg := search.NewConfig("555", "/regression/passengers.txt")
	r := runSearch1(t.Logf, "TEST", newRyftHttp(t.Logf), cfg)
	checkRecordReceived(t, r, 11)
}

// ryfthttp XML search
// check corresponding RDF file is loaded
func TestSearchHttp_XML_id1003(t *testing.T) {
	cfg := search.NewConfig(`(RECORD.id CONTAINS "1003")`, "/regression/*.pcrime")
	r := runSearch1(t.Logf, "TEST", newRyftHttp(t.Logf), cfg)
	checkRecordReceived(t, r, 2542)
}

// ryfthttp XML search
// check corresponding RDF file is loaded
func TestSearchHttp_XML_id1003100(t *testing.T) {
	cfg := search.NewConfig(`(RECORD.id CONTAINS "1003100")`, "/regression/*.pcrime")
	r := runSearch1(t.Logf, "TEST", newRyftHttp(t.Logf), cfg)
	checkRecordReceived(t, r, 9)
}

// ryfthttp XML search
// check corresponding RDF file is loaded
func TestSearchHttp_XML_descVEHICLE(t *testing.T) {
	cfg := search.NewConfig(`(RECORD.desc CONTAINS "VEHICLE")`, "/regression/*.pcrime")
	r := runSearch1(t.Logf, "TEST", newRyftHttp(t.Logf), cfg)
	checkRecordReceived(t, r, 672)
}

// ryftmux search (bad result)
func TestSearchMux_Bad10(t *testing.T) {
	cfg := search.NewConfig("10", "/regression/passengers_not_found.txt")
	s := newRyftMux(t.Logf, newRyftPrim(t.Logf), newRyftHttp(t.Logf))
	r := runSearch1(t.Logf, "TEST", s, cfg)
	checkRecordReceived(t, r, 2*-1)
}

// ryftmux search
func TestSearchMux_10(t *testing.T) {
	cfg := search.NewConfig("10", "/regression/passengers.txt")
	s := newRyftMux(t.Logf, newRyftPrim(t.Logf), newRyftHttp(t.Logf))
	r := runSearch1(t.Logf, "TEST", s, cfg)
	checkRecordReceived(t, r, 2*12)
}

// ryftmux search
func TestSearchMux_310(t *testing.T) {
	cfg := search.NewConfig("310", "/regression/passengers.txt")
	s := newRyftMux(t.Logf, newRyftPrim(t.Logf), newRyftHttp(t.Logf))
	r := runSearch1(t.Logf, "TEST", s, cfg)
	checkRecordReceived(t, r, 2*11)
}

// ryftmux search
func TestSearchMux_555(t *testing.T) {
	cfg := search.NewConfig("555", "/regression/passengers.txt")
	s := newRyftMux(t.Logf, newRyftPrim(t.Logf), newRyftHttp(t.Logf))
	r := runSearch1(t.Logf, "TEST", s, cfg)
	checkRecordReceived(t, r, 2*11)
}

// ryftmux XML search
// check corresponding RDF file is loaded
func TestSearchMux_XML_id1003(t *testing.T) {
	cfg := search.NewConfig(`(RECORD.id CONTAINS "1003")`, "/regression/*.pcrime")
	s := newRyftMux(t.Logf, newRyftPrim(t.Logf), newRyftHttp(t.Logf))
	r := runSearch1(t.Logf, "TEST", s, cfg)
	checkRecordReceived(t, r, 2*2542)
}

// ryftmux XML search
// check corresponding RDF file is loaded
func TestSearchMux_XML_id1003100(t *testing.T) {
	cfg := search.NewConfig(`(RECORD.id CONTAINS "1003100")`, "/regression/*.pcrime")
	s := newRyftMux(t.Logf, newRyftPrim(t.Logf), newRyftHttp(t.Logf))
	r := runSearch1(t.Logf, "TEST", s, cfg)
	checkRecordReceived(t, r, 2*9)
}

// ryftmux XML search
// check corresponding RDF file is loaded
func TestSearchMux_XML_descVEHICLE(t *testing.T) {
	cfg := search.NewConfig(`(RECORD.desc CONTAINS "VEHICLE")`, "/regression/*.pcrime")
	s := newRyftMux(t.Logf, newRyftPrim(t.Logf), newRyftHttp(t.Logf))
	r := runSearch1(t.Logf, "TEST", s, cfg)
	checkRecordReceived(t, r, 2*672)
}

// ryftone search (bad result)
func TestSearchOne_Bad10(t *testing.T) {
	cfg := search.NewConfig("10", "/regression/passengers_not_found.txt")
	r := runSearch1(t.Logf, "TEST", newRyftOne(t.Logf), cfg)
	checkRecordReceived(t, r, -1)
}

// ryftone search
func TestSearchOne_10(t *testing.T) {
	cfg := search.NewConfig("10", "/regression/passengers.txt")
	r := runSearch1(t.Logf, "TEST", newRyftOne(t.Logf), cfg)
	checkRecordReceived(t, r, 12)
}

// ryftone search
func TestSearchOne_310(t *testing.T) {
	cfg := search.NewConfig("310", "/regression/passengers.txt")
	r := runSearch1(t.Logf, "TEST", newRyftOne(t.Logf), cfg)
	checkRecordReceived(t, r, 11)
}

// ryftone search
func TestSearchOne_555(t *testing.T) {
	cfg := search.NewConfig("555", "/regression/passengers.txt")
	r := runSearch1(t.Logf, "TEST", newRyftOne(t.Logf), cfg)
	checkRecordReceived(t, r, 11)
}

// ryftone XML search
// check corresponding RDF file is loaded
func TestSearchOne_XML_id1003(t *testing.T) {
	cfg := search.NewConfig(`(RECORD.id CONTAINS "1003")`, "/regression/*.pcrime")
	r := runSearch1(t.Logf, "TEST", newRyftOne(t.Logf), cfg)
	checkRecordReceived(t, r, 2542)
}

// ryftone XML search
// check corresponding RDF file is loaded
func TestSearchOne_XML_id1003100(t *testing.T) {
	cfg := search.NewConfig(`(RECORD.id CONTAINS "1003100")`, "/regression/*.pcrime")
	r := runSearch1(t.Logf, "TEST", newRyftOne(t.Logf), cfg)
	checkRecordReceived(t, r, 9)
}

// ryftone XML search
// check corresponding RDF file is loaded
func TestSearchOne_XML_descVEHICLE(t *testing.T) {
	cfg := search.NewConfig(`(RECORD.desc CONTAINS "VEHICLE")`, "/regression/*.pcrime")
	r := runSearch1(t.Logf, "TEST", newRyftOne(t.Logf), cfg)
	checkRecordReceived(t, r, 672)
}
