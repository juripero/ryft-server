# Demo - new query combination feature - January 19, 2017

## Creating catalog data

For this demo the `ryft-313` box is used. There are Twitter data files
located under `/ryftone/twitter/` directory. The data files are relatively
small - about 3MB each.

To minimize search duration we combined these files into catalog `twitter.json`:


```{.sh}
cd /ryftone
for file in twitter/*/*; do
  curl -X POST --data-binary "@${file}" \
   -H "Content-Type: application/octet-stream" -u admin:admin \
   -s "http://localhost:8765/files?catalog=twitter1.json&file=${file}&local=true" | jq -c .
done
```

The catalog's data file size limit was set to 64 MB, so there are about 80 data files.


## Running search

The following query will be used to test performance:

```
(RECORD.text CONTAINS EXACT("Salty Dog",!CS))
AND (
  (RECORD.coordinates.coordinates CONTAINS NUMBER("40.68627529491984" < NUM < "40.8795822601648", ",", "."))
  AND
  (RECORD.coordinates.coordinates CONTAINS NUMBER("-74.17597513185498" < NUM < "-73.76652246814498", ",", "."))
)
```


### RECORD-based search

Running the search on small twitter data files take about 48 seconds and
there are 41 matches:

```{.sh}
$ ryftrest -q '(RECORD.text CONTAINS EXACT("Salty Dog",!CS)) AND ((RECORD.coordinates.coordinates CONTAINS NUMBER("40.68627529491984" < NUM < "40.8795822601648", ",", ".")) AND (RECORD.coordinates.coordinates CONTAINS NUMBER("-74.17597513185498" < NUM < "-73.76652246814498", ",", ".")))' -f 'twitter/*/*.json' -u admin:admin --search --format=json > /tmp/run1.txt 2>&1

$ cat /tmp/run1.txt | jq -c '.results | sort_by(._index.file, ._index.offset) | .[]._index'
{"file":"twitter/20170108/1483893901.json","offset":1376759,"length":3187,"fuzziness":-1}
{"file":"twitter/20170108/1483893901.json","offset":1386582,"length":3188,"fuzziness":-1}
{"file":"twitter/20170108/1483893901.json","offset":1389771,"length":3187,"fuzziness":-1}
...
{"file":"twitter/20170108/1483904702.json","offset":1777970,"length":3187,"fuzziness":-1}
{"file":"twitter/20170108/1483905601.json","offset":1814713,"length":3187,"fuzziness":-1}
{"file":"twitter/20170108/1483905601.json","offset":1824536,"length":3188,"fuzziness":-1}

$ cat /tmp/run1.txt | jq '.stats'
{
  "matches": 41,
  "totalBytes": 5185060530,
  "duration": 48196,
  "dataRate": 102.59895129207358,
  "fabricDuration": 7011,
  "fabricDataRate": 705.300049,
  "host": "ryftone-313"
}
```

If input data set is changed to catalog, the duration is a bit smaller -
just about 9 seconds (note the indexes are the same!):

```{.sh}
$ ryftrest -q '(RECORD.text CONTAINS EXACT("Salty Dog",!CS)) AND ((RECORD.coordinates.coordinates CONTAINS NUMBER("40.68627529491984" < NUM < "40.8795822601648", ",", ".")) AND (RECORD.coordinates.coordinates CONTAINS NUMBER("-74.17597513185498" < NUM < "-73.76652246814498", ",", ".")))' -f 'twitter1.json' -u admin:admin --search --format=json > /tmp/run2.txt 2>&1

$ cat /tmp/run2.txt | jq -c '.results | sort_by(._index.file, ._index.offset) | .[]._index'
{"file":"twitter/20170108/1483893901.json","offset":1376759,"length":3187,"fuzziness":-1}
{"file":"twitter/20170108/1483893901.json","offset":1386582,"length":3188,"fuzziness":-1}
{"file":"twitter/20170108/1483893901.json","offset":1389771,"length":3187,"fuzziness":-1}
...
{"file":"twitter/20170108/1483904702.json","offset":1777970,"length":3187,"fuzziness":-1}
{"file":"twitter/20170108/1483905601.json","offset":1814713,"length":3187,"fuzziness":-1}
{"file":"twitter/20170108/1483905601.json","offset":1824536,"length":3188,"fuzziness":-1}

$ cat /tmp/run2.txt | jq '.stats'
{
  "matches": 41,
  "totalBytes": 5185065237,
  "duration": 9102,
  "dataRate": 543.2721979145007,
  "fabricDuration": 6901,
  "fabricDataRate": 716.43927,
  "host": "ryftone-313"
}
```

### RAW_TEXT-based search

Let's try to use RAW_TEXT search first to narrow down the input data set.
Note the square brackets as an additional condition:

```{.sh}
$ ryftrest -q '[RAW_TEXT CONTAINS EXACT("Salty Dog",!CS)] AND (RECORD.text CONTAINS EXACT("Salty Dog",!CS)) AND ((RECORD.coordinates.coordinates CONTAINS NUMBER("40.68627529491984" < NUM < "40.8795822601648", ",", ".")) AND (RECORD.coordinates.coordinates CONTAINS NUMBER("-74.17597513185498" < NUM < "-73.76652246814498", ",", ".")))' -f 'twitter1.json' -u admin:admin --search --format=json > /tmp/run3.txt 2>&1

$ cat /tmp/run3.txt | jq -c '.results | sort_by(._index.file, ._index.offset) | .[]._index'
{"file":"twitter/20170108/1483893901.json","offset":1376759,"length":3187,"fuzziness":-1}
{"file":"twitter/20170108/1483893901.json","offset":1386582,"length":3188,"fuzziness":-1}
{"file":"twitter/20170108/1483893901.json","offset":1389771,"length":3187,"fuzziness":-1}
...
{"file":"twitter/20170108/1483904702.json","offset":1777970,"length":3187,"fuzziness":-1}
{"file":"twitter/20170108/1483905601.json","offset":1814713,"length":3187,"fuzziness":-1}
{"file":"twitter/20170108/1483905601.json","offset":1824536,"length":3188,"fuzziness":-1}
```

$ cat /tmp/run3.txt | jq '.stats'
{
  "matches": 41,
  "totalBytes": 5233081755,
  "duration": 2485,
  "dataRate": 2008.3121389930159,
  "fabricDuration": 325,
  "fabricDataRate": 15355.863585838904,
  "details": [
    {
      "matches": 42,
      "totalBytes": 5185065237,
      "duration": 1768,
      "dataRate": 2796.868521163906,
      "fabricDuration": 259,
      "fabricDataRate": 19018.707031,
      "host": "ryftone-313"
    },
    {
      "matches": 41,
      "totalBytes": 48016518,
      "duration": 717,
      "dataRate": 63.86627612253612,
      "fabricDuration": 66,
      "fabricDataRate": 683.464478,
      "host": "ryftone-313"
    }
  ],
  "host": "ryftone-313"
}
```

We can see two Ryft calls here. The output of the first Ryft call is processed,
unique file names are extracted and the second Ryft call uses this files as
input data set. The total duration is about 2.5 seconds.

We still can try to improve this time by applying additional file filter.
Looking at the results we can notice the data is the same `20170108`.
Let's try to use this information:

```{.sh}
$ ryftrest -q '[RAW_TEXT CONTAINS EXACT("Salty Dog",!CS,FILTER="twitter/20170108/")] AND (RECORD.text CONTAINS EXACT("Salty Dog",!CS)) AND ((RECORD.coordinates.coordinates CONTAINS NUMBER("40.68627529491984" < NUM < "40.8795822601648", ",", ".")) AND (RECORD.coordinates.coordinates CONTAINS NUMBER("-74.17597513185498" < NUM < "-73.76652246814498", ",", ".")))' -f 'twitter1.json' -u admin:admin --search --format=json > /tmp/run4.txt 2>&1

$ cat /tmp/run4.txt | jq -c '.results | sort_by(._index.file, ._index.offset) | .[]._index'
{"file":"twitter/20170108/1483893901.json","offset":1376759,"length":3187,"fuzziness":-1}
{"file":"twitter/20170108/1483893901.json","offset":1386582,"length":3188,"fuzziness":-1}
{"file":"twitter/20170108/1483893901.json","offset":1389771,"length":3187,"fuzziness":-1}
...
{"file":"twitter/20170108/1483904702.json","offset":1777970,"length":3187,"fuzziness":-1}
{"file":"twitter/20170108/1483905601.json","offset":1814713,"length":3187,"fuzziness":-1}
{"file":"twitter/20170108/1483905601.json","offset":1824536,"length":3188,"fuzziness":-1}

$ cat /tmp/run4.txt | jq '.stats'
{
  "matches": 41,
  "totalBytes": 439330704,
  "duration": 1074,
  "dataRate": 390.1102502918776,
  "fabricDuration": 93,
  "fabricDataRate": 4505.144180790071,
  "details": [
    {
      "matches": 41,
      "totalBytes": 394479076,
      "duration": 463,
      "dataRate": 812.5368534360024,
      "fabricDuration": 26,
      "fabricDataRate": 14469.40625,
      "host": "ryftone-313"
    },
    {
      "matches": 41,
      "totalBytes": 44851628,
      "duration": 611,
      "dataRate": 70.00629406318727,
      "fabricDuration": 67,
      "fabricDataRate": 638.415588,
      "host": "ryftone-313"
    }
  ],
  "host": "ryftone-313"
}
```

Bingo! The total duration is about 1 second!


## Conclusion

Using square bracket feature the total performance can be impoved dramatically.
We started from 48 seconds and improved the results to just a second.

Note, that improvement is only possible on certain search queries. The first
conditional expression should significantly narrow down the input data set.
Otherwise, if input data set remains almost the same,
the performance can be even worse. For example, `[RAW_TEXT CONTAINS EXACT("New York",!CS)]`
has negative effect, since it reports all JSON files available, and Ryft
does the full search on small Twitter files.
