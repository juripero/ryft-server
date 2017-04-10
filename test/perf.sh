#!/bin/bash

# print usage info
usage() {
	cat <<EOF
runs various performance tests to get metrics...

Usage: $0 [options]

--test|--run <cmd> Pass the <cmd> to the ryftrest and measure its performance.

--out-dir <dir>    Output directory. Current directory is used by default.
-a|--address=<addr> Specifies the ryft-server address.
                      "http://localhost:8765" by default.
-u|--user=<cred>   Use user credentials, "username:password".
  |--auth
--accept=<fmt>     Accept format can be "json" or "csv".

-h|--help          Prints this short help message.
EOF
}

# print error message $1 and exit
fail() {
	echo "ERROR: $1"
	exit 1
}

# default values
ADDRESS=
AUTH_USER=
ACCEPT=
OUTDIR=.
TESTS=()

# parse options
while [[ $# > 0 ]]; do
	case "$1" in
	--test=*|--run=*)
		TESTS=("${TESTS[@]}" "${1#*=}")
		shift
		;;
	--test|--run)
		TESTS=("${TESTS[@]}" "$2")
		shift 2
		;;
	--out-dir=*)
		OUTDIR="${1#*=}"
		shift
		;;
	--out-dir)
		OUTDIR="$2"
		shift 2
		;;
	-a=*|--address=*)
		ADDRESS="${1#*=}"
		shift
		;;
	-a|--address)
		ADDRESS="$2"
		shift 2
		;;
	-u=*|--user=*|--auth=*)
		AUTH_USER="${1#*=}"
		shift
		;;
	-u|--user|--auth)
		AUTH_USER="$2"
		shift 2
		;;
	--accept=*)
		ACCEPT="${1#*=}"
		shift
		;;
	--accept)
		ACCEPT="$2"
		shift 2
		;;
	-h|--help)
		usage
		exit 0
		;;
	*) # unknown option
		fail "'$1' is unknown option, run '$0 --help' for help"
		;;
	esac
done

DEFAULT_OPTS=()
[[ ! -z "$ADDRESS" ]] && DEFAULT_OPTS=("${DEFAULT_OPTS[@]}" --address "$ADDRESS")
[[ ! -z "$AUTH_USER" ]] && DEFAULT_OPTS=("${DEFAULT_OPTS[@]}" --user "$AUTH_USER")
[[ ! -z "$ACCEPT" ]] && DEFAULT_OPTS=("${DEFAULT_OPTS[@]}" --accept "$ACCEPT")

mkdir -p "$OUTDIR" || fail "failed to create output directory"

TEST_ID=0

# do a test
do_test() {
	TEST_ID=$((TEST_ID+1))
	OUT="${OUTDIR}/result-${TEST_ID}.txt"
	ARGS=("${DEFAULT_OPTS[@]}")
	JQ_STATS=.
	for arg in "$@"; do
		ARGS=("${ARGS[@]}" "$arg")
		if [ "$arg" = "--search" ]; then
			# for --search adjust STATS location
			JQ_STATS=.stats
		fi
	done
	echo "==================================================================="
	echo "testing: $@"
	TIMEFORMAT='  done in %3lR'
	time { ryftrest "${ARGS[@]}" --performance > "$OUT" 2>&1; }
	echo "  output file: $OUT (`stat --printf=%s $OUT` bytes)"
	cat "$OUT" | jq -r "${JQ_STATS} | \"  matches:\(.matches), bytes:\(.totalBytes), duration:\(.duration), fabric:\(.fabricDuration)\""
	cat "$OUT" | jq "${JQ_STATS} | .extra.performance"
	echo " "
	echo " "
}

# run custom tests (DOES NOT WORK yet)
for test_cmd in "${TESTS[@]}"; do
	do_test $test_cmd
done

# Reddit data + RAW_TEXT search
false && {
FILES="-f RC_2016-01.data"
do_test -q '(RAW_TEXT CONTAINS ES("Hello"))' -i $FILES --count
#do_test -q '(RAW_TEXT CONTAINS ES("Hello"))' -i $FILES --search
do_test -q '(RAW_TEXT CONTAINS ES("Bill"))'  -i $FILES --count
do_test -q '(RAW_TEXT CONTAINS ES("Trump"))' -i $FILES --count
}

# Reddit data + RECORD search
false && {
FILES="-f RC_2016-01.data"
#do_test -q '(RECORD CONTAINS ES("Hello"))' -i $FILES --count
do_test -q '(RECORD CONTAINS ES("Hello"))' -i $FILES --search --format=raw
do_test -q '(RECORD CONTAINS ES("Hello"))' -i $FILES --search --format=utf8
do_test -q '(RECORD CONTAINS ES("Hello"))' -i $FILES --search --format=json
do_test -q '(RECORD CONTAINS ES("Hello"))' -i $FILES --search --format=null
#do_test -q '(RECORD CONTAINS ES("Bill"))'  -i $FILES --format=utf8 --search
#do_test -q '(RECORD CONTAINS ES("Trump"))' -i $FILES --format=utf8 --search
}

false && {
FILES="-f RC_2016-01.data"
#do_test -q '(RECORD CONTAINS ES("Hello"))' -i $FILES --search --accept=json
#do_test -q '(RECORD CONTAINS ES("Hello"))' -i $FILES --search --accept=csv
#do_test -q '{RECORD CONTAINS ES("Hello")} AND {RECORD CONTAINS ES("Hello")}' -i $FILES --search
#do_test -q '{RECORD CONTAINS ES("Hello")} OR {RECORD CONTAINS ES("Hello")}' -i $FILES --search
do_test -q '{RECORD CONTAINS ES("Hello")} OR {RECORD CONTAINS ES("Hello")}' -i $FILES --search -oi p-test.txt
do_test -q '{RECORD CONTAINS ES("Hello")} OR {RECORD CONTAINS ES("Hello")}' -i $FILES --search -od p-test.data
do_test -q '{RECORD CONTAINS ES("Hello")} OR {RECORD CONTAINS ES("Hello")}' -i $FILES --search -oi p-test.txt -od p-test.data
}

false && {
FILES="-f RC_2016-01.data"
do_test -q '(RECORD CONTAINS ES("Hello"))' -i $FILES --search --transform 'match("^.*$")'
do_test -q '(RECORD CONTAINS ES("Hello"))' -i $FILES --search --transform 'replace("^(.*)$", "$1")'
do_test -q '(RECORD CONTAINS ES("Hello"))' -i $FILES --search --transform 'script("cat",-)'
}

false && {
FILES="-c twitter1.json"
do_test -q '(RAW_TEXT CONTAINS ES("Trump"))' -i $FILES --search
do_test -q '(RECORD CONTAINS ES("Trump"))' -i $FILES --search
}
