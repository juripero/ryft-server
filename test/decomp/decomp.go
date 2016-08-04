package main

import (
	"bytes"
	"fmt"
)

// simple query (relational expression)
type SimpleQuery struct {
	Input      string // RAW_TEXT or RECORD
	Operator   string // CONTAINS, EQUALS, ...
	Expression string // search expression
}

// simple query as string
func (s SimpleQuery) String() string {
	return fmt.Sprintf("(%s %s %s)", s.Input, s.Operator, s.Expression)
}

// complex query
type Query struct {
	Operator  string
	Simple    *SimpleQuery
	Arguments []Query
}

func (q Query) String() string {
	var buf bytes.Buffer
	if q.Operator != "" {
		buf.WriteString(q.Operator)
		buf.WriteString(" ")
	}
	if q.Simple != nil {
		buf.WriteString(q.Simple.String())
	}

	if len(q.Arguments) > 0 {
		buf.WriteString("{")
		for i, n := range q.Arguments {
			if i != 0 {
				buf.WriteString(", ")
			}
			buf.WriteString(n.String())
		}
		buf.WriteString("}")
	}

	return buf.String()
}

func main() {
	queries := []string{
		//		`RAW_TEXT CONTAINS ?`,
		//		`RECORD EQUALS "no"`,
		//		`RECORD.id NOT_EQUALS "to"`,
		//		`RECORD.[].id NOT_EQUALS "to"`,
		//		`RAW_TEXT CONTAINS FHS("f")`,
		//		`RAW_TEXT CONTAINS FHS("f",CS = true)`,
		//		`RAW_TEXT CONTAINS FEDS( "f" , CS = true, DIST= 5, 	WIDTH =    100.50 )`,

		//		`RAW_TEXT CONTAINS DATE(MM/DD/YY > 02/28/12)`,
		//		`RECORD.date CONTAINS DATE(02/28/12 < MM/DD/YY < 01/19/15)`,
		//		`RAW_TEXT CONTAINS TIME(HH:MM:SS > 09:15:00)`,
		//		`RECORD.time CONTAINS TIME(11:15:00 < HH:MM:SS < 13:15:00)`,
		//		`RECORD.id CONTAINS NUMBER("1025" < NUM < "1050", ",", ".")`,
		//		`RECORD.price CONTAINS CURRENCY("$450" < CUR < "$10,100.50", "$", ",", ".")`,

		`( record.city EQUALS "Rockville" ) AND ( record.state EQUALS "MD" ) OR (record.xxx CONTAINS "hello" ? "world")`,
		`( record.city EQUALS "Rockville" ) AND (( record.state EQUALS "MD" ) OR (record.xxx CONTAINS "hello" ? "world"))`,

		//		`ROW_TEXT CONTAINS ?`,
		//		`RECORD EQUALZ "no"`,
		//		`RECORD. NOT_EQUALS "to"`,
		//		`RAW_TEXT CONTAINS (`,
		//		`RAW_TEXT CONTAINS FHS(`,
		//		`RAW_TEXT CONTAINS FHS(()`,
	}

	for _, q := range queries {
		p := NewParser(bytes.NewBufferString(q))
		expr, err := p.ParseQuery()
		if err != nil {
			fmt.Printf("%s: FAILED with %s\n", q, err)
		} else {
			fmt.Printf("%s => %s\n", q, expr)
		}
	}
}
