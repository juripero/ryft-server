package ryftdec

import (
	"testing"

	"github.com/getryft/ryft-server/search/ryftdec"
)

func TestValidQuery(t *testing.T) {
	query := `((RECORD.id CONTAINS TIME("1003"))AND(RECORD.id CONTAINS DATE("100301"))) AND (RECORD.id CONTAINS TIME("200")) AND (RECORD.id CONTAINS DATE("300")) AND (RECORD.id CONTAINS DATE("400"))`
	if err := ryftdec.Validate(query); err != nil {
		t.Error("Expected valid query, got", err)
	}
}

func TestQueryWithRedundantBracket(t *testing.T) {
	query := `(((RECORD.id CONTAINS TIME("1003"))AND(RECORD.id CONTAINS DATE("100301"))) AND (RECORD.id CONTAINS TIME("200")) AND (RECORD.id CONTAINS DATE("300")) AND (RECORD.id CONTAINS DATE("400"))`
	if err := ryftdec.Validate(query); err == nil {
		t.Error("Expected invalid query, got", err)
	}
}
