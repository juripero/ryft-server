package main

import (
	"testing"

	"github.com/getryft/ryft-server/search"
)

// more debug info!
func init() {
	// ryftprimLogLevel = "debug"
	// ryfthttpLogLevel = "debug"
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
func TestSearchBad10(t *testing.T) {
	cfg := search.NewConfig("10", "/regression/passengers_not_found.txt")
	r := runSearch1(t.Logf, "TEST", newRyftPrim(t.Logf), cfg)
	checkRecordReceived(t, r, -1)
}

// ryftprim search
func TestSearch10(t *testing.T) {
	cfg := search.NewConfig("10", "/regression/passengers.txt")
	r := runSearch1(t.Logf, "TEST", newRyftPrim(t.Logf), cfg)
	checkRecordReceived(t, r, 12)
}

// ryftprim search
func TestSearch310(t *testing.T) {
	cfg := search.NewConfig("310", "/regression/passengers.txt")
	r := runSearch1(t.Logf, "TEST", newRyftPrim(t.Logf), cfg)
	checkRecordReceived(t, r, 11)
}

// ryftprim search
func TestSearch555(t *testing.T) {
	cfg := search.NewConfig("555", "/regression/passengers.txt")
	r := runSearch1(t.Logf, "TEST", newRyftPrim(t.Logf), cfg)
	checkRecordReceived(t, r, 11)
}

// ryftprim XML search
// check corresponding RDF file os loaded
func TestSearchXML_id1003(t *testing.T) {
	cfg := search.NewConfig(`(RECORD.id CONTAINS "1003")`, "/regression/*.pcrime")
	r := runSearch1(t.Logf, "TEST", newRyftPrim(t.Logf), cfg)
	checkRecordReceived(t, r, 2542)
}

// ryftprim XML search
// check corresponding RDF file os loaded
func TestSearchXML_id1003100(t *testing.T) {
	cfg := search.NewConfig(`(RECORD.id CONTAINS "1003100")`, "/regression/*.pcrime")
	r := runSearch1(t.Logf, "TEST", newRyftPrim(t.Logf), cfg)
	checkRecordReceived(t, r, 9)
}

// ryftprim XML search
// check corresponding RDF file os loaded
func TestSearchXML_descVEHICLE(t *testing.T) {
	cfg := search.NewConfig(`(RECORD.desc CONTAINS "VEHICLE")`, "/regression/*.pcrime")
	r := runSearch1(t.Logf, "TEST", newRyftPrim(t.Logf), cfg)
	checkRecordReceived(t, r, 672)
}
