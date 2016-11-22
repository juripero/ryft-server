package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// test options equal
func TestOptionsEqual(t *testing.T) {
	assert.True(t, Options{}.EqualsTo(Options{}))
	assert.False(t, Options{}.EqualsTo(Options{DecimalPoint: "."}))
	assert.False(t, Options{}.EqualsTo(Options{DigitSeparator: ","}))
	assert.False(t, Options{}.EqualsTo(Options{CurrencySymbol: "$"}))
	assert.False(t, Options{}.EqualsTo(Options{Octal: true}))
	assert.False(t, Options{}.EqualsTo(Options{Reduce: true}))
	assert.False(t, Options{}.EqualsTo(Options{Case: true}))
	assert.False(t, Options{}.EqualsTo(Options{Line: true}))
	assert.False(t, Options{}.EqualsTo(Options{Width: 1}))
	assert.False(t, Options{}.EqualsTo(Options{Dist: 1}))
	assert.False(t, Options{}.EqualsTo(Options{Mode: "es"}))

	assert.Equal(t, "", Options{Case: true}.String())
	assert.Equal(t, `[fake,d=1,w=1,!cs,reduce,octal,sym="$",sep=",",dot="."]`,
		Options{
			DecimalPoint:   ".",
			DigitSeparator: ",",
			CurrencySymbol: "$",
			Octal:          true,
			Reduce:         true,
			Case:           false,
			Width:          1,
			Dist:           1,
			Mode:           "fake",
		}.String())
}

// test options set mode
func TestOptionsSetMode(t *testing.T) {
	assert.Panics(t, func() { new(Options).SetMode("bad") })

	// make fake option
	fake := func(d uint) *Options {
		return &Options{
			DecimalPoint:   ".",
			DigitSeparator: ",",
			CurrencySymbol: "$",
			Octal:          true,
			Reduce:         true,
			Case:           false,
			Width:          1,
			Dist:           d,
		}
	}

	assert.Equal(t, `[es,w=1,!cs]`, fake(1).SetMode("es").String())
	assert.Equal(t, `[fhs,d=1,w=1,!cs]`, fake(1).SetMode("fhs").String())
	assert.Equal(t, `[es,w=1,!cs]`, fake(0).SetMode("fhs").String())
	assert.Equal(t, `[feds,d=1,w=1,!cs,reduce]`, fake(1).SetMode("feds").String())
	assert.Equal(t, `[es,w=1,!cs]`, fake(0).SetMode("feds").String())
	assert.Equal(t, `[ds,w=1,!cs]`, fake(1).SetMode("ds").String())
	assert.Equal(t, `[ts,w=1,!cs]`, fake(1).SetMode("ts").String())
	assert.Equal(t, `[ns,w=1,!cs,sep=",",dot="."]`, fake(1).SetMode("ns").String())
	assert.Equal(t, `[cs,w=1,!cs,sym="$",sep=",",dot="."]`, fake(1).SetMode("cs").String())
	assert.Equal(t, `[ipv4,w=1,!cs,octal]`, fake(1).SetMode("ipv4").String())
	assert.Equal(t, `[ipv6,w=1,!cs]`, fake(1).SetMode("ipv6").String())
}

// test for set options
func TestOptionsSetOpt(t *testing.T) {
	// positive case
	check := func(option string, expected string) {
		opt := new(Options)
		opt.Mode = "test"
		opt.Case = true
		_, err := opt.Set(option, "")
		if assert.NoError(t, err) {
			assert.Equal(t, expected, opt.String())
		}
	}

	// negative case
	bad := func(option string, expectedError string) {
		opt := new(Options)
		_, err := opt.Set(option, "")
		if assert.Error(t, err) {
			assert.Contains(t, err.Error(), expectedError)
		}
	}

	// DISTANCE
	check(` FUZZINESS = 2 `, `[test,d=2]`)
	check(`DISTANCE=2`, `[test,d=2]`)
	check(`DIST = 2`, `[test,d=2]`)
	check(`d=2`, `[test,d=2]`)
	check(`D = " 2 "`, `[test,d=2]`) // as string
	check(`D = "2"`, `[test,d=2]`)   // as string
	check(`!d`, `[test]`)            // not
	bad(`D=tru`, "found instead of integer value")
	bad(`D=1.23`, "found instead of integer value")
	bad(`D=,`, "found instead of integer value")
	bad(`D=100000`, "is out of range") // [0..64K)
	bad(`D=65536`, "is out of range")  // [0..64K)
	bad(`D=-1`, "is out of range")     // [0..64K)
	bad(`!D=`, "extra data at the end")
	bad(`D=""`, "failed to parse integer")
	bad(`D no`, "found instead of =")

	// WIDTH
	check(` SURROUNDING = 2 `, `[test,w=2]`)
	check(`WIDTH=2`, `[test,w=2]`)
	check(`w = 2`, `[test,w=2]`)
	check(`W = " 3 "`, `[test,w=3]`) // as string
	check(`W = "3"`, `[test,w=3]`)   // as string
	check(`!W`, `[test]`)            // not
	bad(`W=tru`, "found instead of integer value")
	bad(`W=1.23`, "found instead of integer value")
	bad(`W=,`, "found instead of integer value")
	bad(`W=100000`, "is out of range") // [0..64K)
	bad(`W=65536`, "is out of range")  // [0..64K)
	bad(`W=-1`, "is out of range")     // [0..64K)
	bad(`!W=`, "extra data at the end")
	bad(`W=""`, "failed to parse integer")
	bad(`W no`, "found instead of =")

	// LINE
	check(` LINE = true `, `[test,line]`)
	check(`L=true`, `[test,line]`)
	check(`L = "true"`, `[test,line]`)
	check(`L = " T "`, `[test,line]`)
	check(`L = t`, `[test,line]`)
	check(`L = 1`, `[test,line]`)
	check(`L = " 1 "`, `[test,line]`)
	check(`L = TRUE`, `[test,line]`)
	check(`L = True`, `[test,line]`)
	check(`L`, `[test,line]`) // shortcut
	check(`!L`, `[test]`)     // not
	bad(`L=tru`, "failed to parse boolean")
	bad(`L=1.23`, "found instead of boolean value")
	bad(`L=,`, "found instead of boolean value")
	bad(`!,`, "no valid option name found")
	bad(`!L=,`, "extra data at the end")
	bad(`L no`, "extra data at the end")

	// CASE
	check(` CASE = true `, `[test]`)
	check(`CS=false`, `[test,!cs]`)
	check(`CS = "false"`, `[test,!cs]`)
	check(`CS = " F "`, `[test,!cs]`)
	check(`CS = f`, `[test,!cs]`)
	check(`CS = 0`, `[test,!cs]`)
	check(`CS = " 0 "`, `[test,!cs]`)
	check(`CS = FALSE`, `[test,!cs]`)
	check(`CS = False`, `[test,!cs]`)
	check(`CS`, `[test]`)      // shortcut
	check(`!CS`, `[test,!cs]`) // not
	bad(`CS=tru`, "failed to parse boolean")
	bad(`CS=1.23`, "found instead of boolean value")
	bad(`CS=,`, "found instead of boolean value")
	bad(`!,`, "no valid option name found")
	bad(`!CS=,`, "extra data at the end")
	bad(`CS no`, "extra data at the end")

	// REDUCE
	check(` REDUCE = true `, `[test,reduce]`)
	check(`R=true`, `[test,reduce]`)
	check(`R = "true"`, `[test,reduce]`)
	check(`R = " T "`, `[test,reduce]`)
	check(`R = t`, `[test,reduce]`)
	check(`R = 1`, `[test,reduce]`)
	check(`R = " 1 "`, `[test,reduce]`)
	check(`R = TRUE`, `[test,reduce]`)
	check(`R = True`, `[test,reduce]`)
	check(`R`, `[test,reduce]`) // shortcut
	check(`!R`, `[test]`)       // not
	bad(`R=tru`, "failed to parse boolean")
	bad(`R=1.23`, "found instead of boolean value")
	bad(`R=,`, "found instead of boolean value")
	bad(`!,`, "no valid option name found")
	bad(`!R=,`, "extra data at the end")
	bad(`R no`, "extra data at the end")

	// test for OCTAL options parsing (generic queries)
	check(` USE_OCTAL = true `, `[test,octal]`)
	check(` OCTAL = true `, `[test,octal]`)
	check(`OCT=true`, `[test,octal]`)
	check(`OCT = "true"`, `[test,octal]`)
	check(`OCT = " T "`, `[test,octal]`)
	check(`OCT = t`, `[test,octal]`)
	check(`OCT = 1`, `[test,octal]`)
	check(`OCT = " 1 "`, `[test,octal]`)
	check(`OCT = TRUE`, `[test,octal]`)
	check(`OCT = True`, `[test,octal]`)
	check(`OCT`, `[test,octal]`) // shortcut
	check(`!OCT`, `[test]`)      // not
	bad(`OCT=tru`, "failed to parse boolean")
	bad(`OCT=1.23`, "found instead of boolean value")
	bad(`OCT=,`, "found instead of boolean value")
	bad(`!,`, "no valid option name found")
	bad(`!OCT=,`, "extra data at the end")
	bad(`OCT no`, "extra data at the end")

	// SYMBOL
	check(` SYMBOL = "$" `, `[test,sym="$"]`)
	check(` SYMB = 1 `, `[test,sym="1"]`)
	check(` SYM = 1.23 `, `[test,sym="1.23"]`)
	bad(`SYM=,`, "found instead of string value")
	bad(`!SYM`, "is not supported for string option")
	bad(`SYM no`, "found instead of =")

	// SEPARATOR
	check(` SEPARATOR = "," `, `[test,sep=","]`)
	check(` SEP = 1 `, `[test,sep="1"]`)
	check(` SEP = 1.23 `, `[test,sep="1.23"]`)
	bad(`SEP=,`, "found instead of string value")
	bad(`!SEP`, "is not supported for string option")
	bad(`SEP no`, "found instead of =")

	// DECIMAL
	check(` DECIMAL = "." `, `[test,dot="."]`)
	check(` DEC = 1 `, `[test,dot="1"]`)
	check(` DEC = 1.23 `, `[test,dot="1.23"]`)
	bad(`DEC=,`, "found instead of string value")
	bad(`!DEC`, "is not supported for string option")
	bad(`DEC no`, "found instead of =")

	bad(`BAD no`, "unknown option")
}
