package ryftdec

import (
	"testing"

	"github.com/getryft/ryft-server/search/ryftdec"
)

func TestMisformattedQuery(t *testing.T) {
	query := `((RECORD.id CONTAINS TIME("1003"))AND(RECORD.id CONTAINS DATE("100301"))) AND (RECORD.id CONTAINS TIME("200")) AND (RECORD.id CONTAINS DATE("300")) AND (RECORD.id CONTAINS DATE("400"))`
	DecompositionTests(t, query)
}

func TestRegularQuery(t *testing.T) {
	query := `((RECORD.id CONTAINS TIME("1003")) AND (RECORD.id CONTAINS DATE("100301"))) AND (RECORD.id CONTAINS TIME("200")) AND (RECORD.id CONTAINS DATE("300")) AND (RECORD.id CONTAINS DATE("400"))`
	DecompositionTests(t, query)
}

func DecompositionTests(t *testing.T, query string) {
	result := ryftdec.Decompose(query)

	RootNodeChildren(t, result)
	FirstLevelNodeChildren(t, result)
	FirstLevelNodeExpression(t, result)
	SecondLevelNodeChildren(t, result)
	ThirdLevelNodeExpression(t, result)
}

func RootNodeChildren(t *testing.T, result *ryftdec.Node) {
	if len(result.SubNodes) != 1 {
		t.Error("Expected 1 subnodes for root node, got", len(result.SubNodes))
	}
}

func FirstLevelNodeExpression(t *testing.T, result *ryftdec.Node) {
	node := result.SubNodes[0]

	if node.Expression != "AND" {
		t.Error("Expected AND, got", node.Expression)
	}
}

func FirstLevelNodeChildren(t *testing.T, result *ryftdec.Node) {
	node := result.SubNodes[0]

	if len(node.SubNodes) != 2 {
		t.Error("Expected 2 subnodes for first level node, got", len(result.SubNodes))
	}
}

func SecondLevelNodeChildren(t *testing.T, result *ryftdec.Node) {
	node := result.SubNodes[0].SubNodes[0]

	if len(node.SubNodes) != 2 {
		t.Error("Expected 2 subnodes for second level node, got", len(result.SubNodes))
	}
}

func ThirdLevelNodeExpression(t *testing.T, result *ryftdec.Node) {
	node := result.SubNodes[0].SubNodes[0].SubNodes[0]

	if node.Expression != `(RECORD.id CONTAINS TIME("1003"))` {
		t.Error(`Expected RECORD.id CONTAINS TIME("1003"), got`, node.Expression)
	}
}
