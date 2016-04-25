package ryftdec

import (
	"testing"

	"github.com/getryft/ryft-server/search/ryftdec"
)

func TestInvalidQuery1(t *testing.T) {
	query := `(((RECORD.id CONTAINS TIME("1003")) AND (RECORD.id CONTAINS DATE("100301"))) AND (RECORD.id CONTAINS TIME("200")) AND (RECORD.id CONTAINS DATE("300")) AND (RECORD.id CONTAINS DATE("400"))`
	_, err := ryftdec.Decompose(query)
	if err == nil {
		t.Error("Expected invalid query error, got valid result")
	}
}

func TestInvalidQuery2(t *testing.T) {
	query := `((RECORD.id CONTAINS TIME("1003")) AND (RECORD.id CONTAINS DATE("100301")))) AND (RECORD.id CONTAINS TIME("200")) AND (RECORD.id CONTAINS DATE("300")) AND (RECORD.id CONTAINS DATE("400"))`
	_, err := ryftdec.Decompose(query)
	if err == nil {
		t.Error("Expected invalid query error, got valid result")
	}
}

func TestValidQuery1(t *testing.T) {
	query := `(RAW_TEXT CONTAINS "Some text0")`
	_, err := ryftdec.Decompose(query)
	if err != nil {
		t.Error("Expected valid query, got invalid")
	}
}

func TestValidQuery2(t *testing.T) {
	query := `((RAW_TEXT CONTAINS "Some text0") OR (RAW_TEXT CONTAINS "Some text1") OR (RAW_TEXT CONTAINS "Some text2"))`
	_, err := ryftdec.Decompose(query)
	if err != nil {
		t.Error("Expected valid query, got invalid")
	}
}

func TestValidQuery3(t *testing.T) {
	query := `( record.city EQUALS "Rockville" ) AND ( record.state EQUALS "MD" )`
	_, err := ryftdec.Decompose(query)
	if err != nil {
		t.Error("Expected valid query, got invalid")
	}
}

func TestValidQuery4(t *testing.T) {
	query := `( ( record.city EQUALS "Rockville" ) OR ( record.city EQUALS "Gaithersburg" ) ) AND ( record.state EQUALS "MD" )`
	_, err := ryftdec.Decompose(query)
	if err != nil {
		t.Error("Expected valid query, got invalid")
	}
}

func TestValidQuery5(t *testing.T) {
	query := `(RAW_TEXT CONTAINS DATE(MM/DD/YY > 02/28/12))`
	_, err := ryftdec.Decompose(query)
	if err != nil {
		t.Error("Expected valid query, got invalid")
	}
}

func TestValidQuery6(t *testing.T) {
	query := `(RAW_TEXT CONTAINS DATE(02/28/12 < MM/DD/YY < 01/19/15))`
	_, err := ryftdec.Decompose(query)
	if err != nil {
		t.Error("Expected valid query, got invalid")
	}
}

func TestValidQuery7(t *testing.T) {
	query := `RAW_TEXT CONTAINS TIME(HH:MM:SS > 09:15:00))`
	_, err := ryftdec.Decompose(query)
	if err != nil {
		t.Error("Expected valid query, got invalid")
	}
}

func TestValidQuery8(t *testing.T) {
	query := `(RAW_TEXT CONTAINS TIME(11:15:00 < HH:MM:SS < 13:15:00))`
	_, err := ryftdec.Decompose(query)
	if err != nil {
		t.Error("Expected valid query, got invalid")
	}
}

func TestValidQuery9(t *testing.T) {
	query := `((RAW_TEXT CONTAINS DATE(02/28/12 < MM/DD/YY < 01/19/15))  AND (RAW_TEXT CONTAINS TIME(11:15:00 < HH:MM:SS < 13:15:00)))`
	_, err := ryftdec.Decompose(query)
	if err != nil {
		t.Error("Expected valid query, got invalid")
	}
}
