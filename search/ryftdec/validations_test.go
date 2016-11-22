package ryftdec

/*
import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInvalidQueries(t *testing.T) {
	queries := []string{
		"", " ", "   ",
		"(", ")", "((", "))", ")(",
		"TEST", " TEST ", " TEST   FOO   TEST  ",
		"AND", " AND ", "   AND  ", ` ( "AND" ) `,
		`))OR((`,
		`() AND ()`,
		`() AND (OR)`,
		`() AND () AND ()`,
		`() AND OR" "() MOR ()`,
		`(((RECORD.id CONTAINS TIME("1003")) AND (RECORD.id CONTAINS DATE("100301"))) AND (RECORD.id CONTAINS TIME("200")) AND (RECORD.id CONTAINS DATE("300")) AND (RECORD.id CONTAINS DATE("400"))`,
		`((RECORD.id CONTAINS TIME("1003")) AND (RECORD.id CONTAINS DATE("100301")))) AND (RECORD.id CONTAINS TIME("200")) AND (RECORD.id CONTAINS DATE("300")) AND (RECORD.id CONTAINS DATE("400"))`,
		`(RAW_TEXT CONTAINS FHS("text",CS=123, DIST=100, WIDTH=2000))`,
	}

	for _, q := range queries {
		_, err := Decompose(q, decomposerOptions())
		assert.Error(t, err, "Invalid query: '%s'", q)
	}
}

func TestValidQueries(t *testing.T) {
	queries := []string{
		`(RAW_TEXT CONTAINS "Some text0")`,
		`(RAW_TEXT CONTAINS aabbccddeeff)`,
		`((RAW_TEXT CONTAINS "Some text0"))`,
		`(RAW_TEXT CONTAINS "Some text0") OR (RAW_TEXT CONTAINS "Some text1") OR (RAW_TEXT CONTAINS "Some text2")`,
		`((RAW_TEXT CONTAINS "Some text0") OR (RAW_TEXT CONTAINS "Some text1") OR (RAW_TEXT CONTAINS "Some text2"))`,
		`(( record.city EQUALS "Rockville" ) AND ( record.state EQUALS "MD" ))`,
		`(( ( record.city EQUALS "Rockville" ) OR ( record.city EQUALS "Gaithersburg" ) ) AND ( record.state EQUALS "MD" ))`,
		`(RAW_TEXT CONTAINS DATE(MM/DD/YY > 02/28/12))`,
		`(RAW_TEXT CONTAINS DATE(02/28/12 < MM/DD/YY < 01/19/15))`,
		`(RAW_TEXT CONTAINS TIME(HH:MM:SS > 09:15:00))`,
		`(RAW_TEXT CONTAINS TIME(11:15:00 < HH:MM:SS < 13:15:00))`,
		`((RAW_TEXT CONTAINS DATE(02/28/12 < MM/DD/YY < 01/19/15))  AND (RAW_TEXT CONTAINS TIME(11:15:00 < HH:MM:SS < 13:15:00)))`,
		`(RECORD.Name.Actors.[].Name CONTAINS "Christian")`,
		`((RECORD.Date CONTAINS DATE(MM/DD/YYYY = 04/15/2015))AND(RECORD.Date CONTAINS TIME(HH:MM:SS = 11:59:00)))`,
		`((RECORD.Date CONTAINS DATE(MM/DD/YYYY = 04/15/2015))AND(RECORD.Date CONTAINS TIME(HH:MM:SS = 11:59:00)))`,
		`((RECORD.Date CONTAINS DATE(MM/DD/YYYY <= 04/15/2015))AND(RECORD.Date CONTAINS TIME(HH:MM:SS <= 11:59:00)))`,
		`((RECORD.Date CONTAINS DATE(MM/DD/YYYY<=04/15/2015))AND(RECORD.Date CONTAINS TIME(HH:MM:SS<=11:59:00)))`,
		`((RECORD.Date CONTAINS DATE(MM/DD/YYYY>=04/15/2015))AND(RECORD.Date CONTAINS TIME(HH:MM:SS>=11:59:00)))`,
		`((RECORD.Date CONTAINS DATE(MM/DD/YYYY >= 04/15/2015))AND(RECORD.Date CONTAINS TIME(HH:MM:SS >= 11:59:00)))`,
		`((RECORD.Date CONTAINS DATE(MM/DD/YYYY != 04/15/2015))AND(RECORD.Date CONTAINS TIME(HH:MM:SS != 11:59:00)))`,
		`((RECORD.Date CONTAINS DATE(MM/DD/YYYY!=04/15/2015))AND(RECORD.Date CONTAINS TIME(HH:MM:SS!=11:59:00)))`,
		`(RECORD.doc.play_name NOT_CONTAINS "King Lear")`,
		`(RAW_TEXT NOT_EQUALS "?")`,
		`(RAW_TEXT CONTAINS "?")`,
		`(RAW_TEXT CONTAINS ?)`,
		`(RAW_TEXT CONTAINS "he"??"o")`,
		`(RECORD.price CONTAINS CURRENCY("$450" < CUR < "$10,100.50", "$", ",", "."))`,
		`(RECORD.body CONTAINS FHS("test", CS=true, DIST=10, WIDTH=100))`,
		`(RECORD.body CONTAINS FEDS("test", CS=false, DIST=10, WIDTH=100))`,
		`(RECORD.body CONTAINS FEDS("test",CS=false,DIST=10,WIDTH-100))`,
		`(RECORD.body CONTAINS FEDS('test',CS=false,DIST=10,WIDTH=100))`,
		`(RECORD.body CONTAINS FEDS('test',CS=false,DIST=10,WIDTH=100)) AND (RECORD.body CONTAINS FHS("test", CS=true, DIST=10, WIDTH=100))`,
		`(RECORD.body CONTAINS "FEDS")`,
		`(RECORD.body CONTAINS REGEX("\w+", CASELESS))`,
		`((RECORD.body CONTAINS "DATE()") AND (RAW_TEXT CONTAINS DATE(MM/DD/YYYY!=04/15/2015)))`,
		` (RAW_TEXT CONTAINS "Some text0")`,
	}

	for _, q := range queries {
		_, err := Decompose(q, decomposerOptions())
		assert.NoError(t, err, "Valid query: `%s`", q)
	}
}

// test extension detection
func TestDetectExtension(t *testing.T) {

	type ExtFileSet struct {
		fileNames   []string
		dataOut     string
		expectedExt string
		expectedErr bool
	}

	data := []ExtFileSet{
		{[]string{}, "out.txt", ".txt", false},
		{[]string{"a.txt"}, "", ".txt", false},
		{[]string{"a.txt", "b.txt"}, "", ".txt", false},
		{[]string{"a.dat", "b.dat"}, "", ".dat", false},
		{[]string{"a.txt", "b.dat"}, "", "", true},
		{[]string{"a.txt", "b.dat"}, "c.jpeg", "", true},
		{[]string{}, "", "", false},
		{[]string{"foo/a.txt", "my.test/b.txt"}, "", ".txt", false},
		{[]string{"foo/a.txt", "my.test/b.txt"}, "data.txt", ".txt", false},
		{[]string{"foo/*.txt", "my.test/*txt"}, "", ".txt", false},
		{[]string{"foo/*.txt", "my.test/*"}, "data.txt", ".txt", false},
		{[]string{"my.test/*"}, "data.txt", ".txt", false},
		{[]string{"nyctaxi/xml/2015/yellow/*"}, "ryftnyctest.nxml", ".nxml", false},
	}

	for _, d := range data {
		ext, err := detectExtension(d.fileNames, d.dataOut)
		if d.expectedErr {
			assert.Error(t, err, "data:%s out:%s", d.fileNames, d.dataOut)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, d.expectedExt, ext)
		}
	}

}
*/
