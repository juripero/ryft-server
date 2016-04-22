//Tests for follwing tree
//Expression: 'AND'
//  Expression: 'AND'
//    Expression: '(RECORD.id CONTAINS TIME("1003"))'
//    Expression: '(RECORD.id CONTAINS DATE("100301"))'
//  Expression: 'AND'
//    Expression: '(RECORD.id CONTAINS TIME("200"))'
//    Expression: 'AND'
//      Expression: '(RECORD.id CONTAINS DATE("300"))'
//      Expression: '(RECORD.id CONTAINS DATE("400"))'

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
	result, _ := ryftdec.Decompose(query)

	RootNodeChildren(t, result)
	FirstLevelNodeChildren(t, result)
	FirstLevelNodeExpression(t, result)
	SecondLevelNodeExpression(t, result)
	ThirdLevelNodeExpression(t, result)
}

func RootNodeChildren(t *testing.T, result *ryftdec.Node) {
	if len(result.SubNodes) != 2 {
		t.Error("Expected 2 subnodes for root node, got", len(result.SubNodes))
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

func SecondLevelNodeExpression(t *testing.T, result *ryftdec.Node) {
	node := result.SubNodes[0].SubNodes[0]

	if node.Expression != `(RECORD.id CONTAINS TIME("1003"))` {
		t.Error(`Expected (RECORD.id CONTAINS TIME("1003")), got`, result.Expression)
	}
}

func ThirdLevelNodeExpression(t *testing.T, result *ryftdec.Node) {
	node := result.SubNodes[1].SubNodes[1].SubNodes[0]

	if node.Expression != `(RECORD.id CONTAINS DATE("300"))` {
		t.Error(`Expected (RECORD.id CONTAINS DATE("300")), got`, node.Expression)
	}
}
