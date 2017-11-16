# Demo - Date histogram aggregation - November 16, 2017

Histogram is a multi-bucket aggregation. It represents frequency distribution of a set of numeric data.
Date histrogram aggregation is similar to the usual histohram aggregation, but it can be applied only on date fields.


# Date histohram in action

## Request fields

`field` (required) - name of a field that contains date.

`interval` (required) - time interval. Possible values: `year`, `quarter`, `month`, `week`, `day`, `hour`, `minute`, `second`.
    It may also be specified with time-units: `d`, `h`, `m`, `s`, `ms`, `micros`, `nanos`.
    Fractional time values are not supported.

`offset` (optional) - the offset parameter is used to change the start value of each bucket by the specified positive (+) or negative offset (-) duration.

`time_zone` (optional) - date-times are stored in `UTC`. By default, all bucketing and rounding is also done in `UTC`. The `time_zone` parameter can be used to indicate that bucketing should use a different time zone. It may be set as UTC offset (e.g. +01:00, -08:00) or as time zone id in TZ database.

`format` (optional) - the response has `key_as_string` field, that represents time interval formatted with the `format` value. It supports jodaDate notation. E.g. "yyyy-MM-dd".

`keyed` (optional) - show buckets as a list or as a map.

`min_doc_count` (optional) - show aggregation results only for buckets that have number of documents not less then value of this parameter.

`_aggs` (optional) - it is possible to run sub-aggregations on elements of each bucket. All possible aggregations are described in [documentation](../rest/aggs.md) and shown in [this demo](./2017-08-17-aggregations.md).

## Examples
File we in search
```{.sh}
cat datehist.json
{"Num": 1,"Created": "04/15/2015 10:00:00 PM", "Updated":"04/15/2015 10:00:00 PM", "Descr": "111"}
{"Num": 1,"Created": "04/15/2015 10:01:00 PM", "Updated":"01/15/2015 10:00:00 PM", "Descr": "111"}
{"Num": 1,"Created": "04/15/2015 10:01:00 PM", "Updated":"02/15/2015 10:00:00 PM", "Descr": "111"}
{"Num": 1,"Created": "04/15/2015 10:01:00 PM", "Updated":"03/15/2015 10:00:00 PM", "Descr": "112"}
{"Num": 1,"Created": "04/15/2015 10:02:00 PM", "Updated":"07/15/2015 10:00:00 PM", "Descr": "112"}
{"Num": 1,"Created": "04/15/2015 10:02:00 PM", "Updated":"10/15/2015 10:00:00 PM", "Descr": "112"}
{"Num": 1,"Created": "04/15/2016 10:02:00 PM", "Updated":"04/15/2015 10:00:00 PM", "Descr": "112"}
{"Num": 1,"Created": "04/15/2016 10:00:00 PM", "Updated":"04/15/2015 10:00:00 PM", "Descr": "112"}
{"Num": 1,"Created": "04/15/2016 10:00:00 PM", "Updated":"04/15/2015 10:00:00 PM", "Descr": "112"}
{"Num": 1,"Created": "04/15/2016 10:00:00 PM", "Updated":"04/15/2016 10:00:00 PM", "Descr": "112"}
{"Num": 1,"Created": "04/15/2017 10:00:00 PM", "Updated":"04/15/2017 10:00:00 PM", "Descr": "113"}
{"Num": 1,"Created": "08/15/2017 10:00:00 PM", "Updated":"04/15/2017 10:00:00 PM", "Descr": "113"}
{"Num": 1,"Created": "08/15/2017 10:00:00 PM", "Updated":"04/15/2017 10:00:00 PM", "Descr": "113"}
{"Num": 1,"Created": "08/15/2017 10:00:00 PM", "Updated":"04/15/2017 10:00:00 PM", "Descr": "113"}
{"Num": 1,"Created": "09/15/2017 09:00:00 PM", "Updated":"04/15/2017 10:00:00 PM", "Descr": "113"}
{"Num": 1,"Created": "08/15/2017 08:22:00 PM", "Updated":"04/15/2017 10:00:11 PM", "Descr": "113"}
{"Num": 1,"Created": "08/15/2017 10:23:00 PM", "Updated":"04/15/2017 10:00:06 PM", "Descr": "113"}
{"Num": 1,"Created": "09/15/2017 10:25:00 PM", "Updated":"04/15/2017 10:00:01 PM", "Descr": "113"}
```


## format

`format` field accept jodaTime notation, otherwise "yyyy-MM-ddTHH:mm:ss.SSSZZ" format pattern will be used.
Unformatted results:
```{.sh}
ryftrest -q '(RECORD CONTAINS "11")' -f datehist.json --count --format=json --body '{"aggs":{"my_test":{"date_histogram":{"field":"Created","interval":"year"}}}}' --address=0.0.0.0:8778 | jq .stats.extra.aggregations
{
  "my_test": {
    "buckets": [
      {
        "key_as_string": "2015-01-01T00:00:00.000+00:00",
        "key": 1420070400000,
        "doc_count": 6
      },
      {
        "key_as_string": "2016-01-01T00:00:00.000+00:00",
        "key": 1451606400000,
        "doc_count": 4
      },
      {
        "key_as_string": "2017-01-01T00:00:00.000+00:00",
        "key": 1483228800000,
        "doc_count": 8
      }
    ]
  }
}
```

Results formatted with `yyyy` pattern
```{.sh}
ryftrest -q '(RECORD CONTAINS "11")' -f datehist.json --count --format=json --body '{"aggs":{"my_test":{"date_histogram":{"field":"Created","interval":"year", "format":"yyyy"}}}}' --address=0.0.0.0:8778 | jq .stats.extra.aggregations
{
  "my_test": {
    "buckets": [
      {
        "key_as_string": "2015",
        "key": 1420070400000,
        "doc_count": 6
      },
      {
        "key_as_string": "2016",
        "key": 1451606400000,
        "doc_count": 4
      },
      {
        "key_as_string": "2017",
        "key": 1483228800000,
        "doc_count": 8
      }
    ]
  }
}
```

## interval

`interval` is a required field for this aggregation. Devide results with `week` interval.
```{.sh}
ryftrest -q '(RECORD CONTAINS "11")' -f datehist.json --count --format=json --body '{"aggs":{"my_test":{"date_histogram":{"field":"Created","interval":"week", "format": "yyyy-MM-dd"}}}}' --address=0.0.0.0:8778 | jq .stats.extra.aggregations
{
  "my_test": {
    "buckets": [
      {
        "key_as_string": "2015-04-12",
        "key": 1428796800000,
        "doc_count": 6
      },
      {
        "key_as_string": "2016-04-10",
        "key": 1460246400000,
        "doc_count": 4
      },
      {
        "key_as_string": "2017-04-09",
        "key": 1.491696e+12,
        "doc_count": 1
      },
      {
        "key_as_string": "2017-08-13",
        "key": 1502582400000,
        "doc_count": 5
      },
      {
        "key_as_string": "2017-09-10",
        "key": 1505001600000,
        "doc_count": 2
      }
    ]
  }
}
```

Same idea for a 100 days interval.
```{.sh}
ryftrest -q '(RECORD CONTAINS "11")' -f datehist.json --count --format=json --body '{"aggs":{"my_test":{"date_histogram":{"field":"Updated","interval":"100d", "format":"yyyy-MM-dd"}}}}' --address=0.0.0.0:8778 | jq .stats.extra.aggregations
{
  "my_test": {
    "buckets": [
      {
        "key_as_string": "2015-01-03",
        "key": 1420243200000,
        "doc_count": 3
      },
      {
        "key_as_string": "2015-04-13",
        "key": 1428883200000,
        "doc_count": 5
      },
      {
        "key_as_string": "2015-07-22",
        "key": 1437523200000,
        "doc_count": 1
      },
      {
        "key_as_string": "2016-02-07",
        "key": 1454803200000,
        "doc_count": 1
      },
      {
        "key_as_string": "2017-03-13",
        "key": 1489363200000,
        "doc_count": 8
      }
    ]
  }
}
```

## time_zone

`time_zone` in  ISO 8601 UTC offset
```{.sh}
ryftrest -q '(RECORD CONTAINS "11")' -f datehist.json --count --format=json --body '{"aggs":{"my_test":{"date_histogram":{"field":"Created","interval":"year", "time_zone": "+01:00", "format":"yyyy-MM-dd ZZ"}}}}' --address=0.0.0.0:8778 | jq .stats.extra.aggregations
{
  "my_test": {
    "buckets": [
      {
        "key_as_string": "2015-01-01 +01:00",
        "key": 1420066800000,
        "doc_count": 6
      },
      {
        "key_as_string": "2016-01-01 +01:00",
        "key": 1451602800000,
        "doc_count": 4
      },
      {
        "key_as_string": "2017-01-01 +01:00",
        "key": 1483225200000,
        "doc_count": 8
      }
    ]
  }
}
```

`time_zone` a time zone id, an identifier used in the TZ database
```{.sh}
ryftrest -q '(RECORD CONTAINS "11")' -f datehist.json --count --format=json --body '{"aggs":{"my_test":{"date_histogram":{"field":"Created","interval":"year", "time_zone": "Asia/Pontianak", "format":"yyyy-MM-dd ZZ"}}}}' --address=0.0.0.0:8778 | jq .stats.extra.aggregations
{
  "my_test": {
    "buckets": [
      {
        "key_as_string": "2015-01-01 +07:00",
        "key": 1420045200000,
        "doc_count": 6
      },
      {
        "key_as_string": "2016-01-01 +07:00",
        "key": 1451581200000,
        "doc_count": 4
      },
      {
        "key_as_string": "2017-01-01 +07:00",
        "key": 1483203600000,
        "doc_count": 8
      }
    ]
  }
}
```

## offset

Lets get initial request without the "offset" field

```{.sh}
ryftrest -q '(RECORD CONTAINS "11")' -f datehist.json --count --format=json --body '{"aggs":{"my_test":{"date_histogram":{"field":"Created","interval":"month"}}}}' --address=0.0.0.0:8778 | jq .stats.extra.aggregations
{
  "my_test": {
    "buckets": [
      {
        "key_as_string": "2015-04-01T00:00:00.000+00:00",
        "key": 1427846400000,
        "doc_count": 6
      },
      {
        "key_as_string": "2016-04-01T00:00:00.000+00:00",
        "key": 1459468800000,
        "doc_count": 4
      },
      {
        "key_as_string": "2017-04-01T00:00:00.000+00:00",
        "key": 1491004800000,
        "doc_count": 1
      },
      {
        "key_as_string": "2017-08-01T00:00:00.000+00:00",
        "key": 1501545600000,
        "doc_count": 5
      },
      {
        "key_as_string": "2017-09-01T00:00:00.000+00:00",
        "key": 1.504224e+12,
        "doc_count": 2
      }
    ]
  }
}
```

Set `offset` using date math like `offset=+20d+1h+2m+3s+4ms+5micros+6nanos`

```{.sh}
ryftrest -q '(RECORD CONTAINS "11")' -f datehist.json --count --format=json --body '{"aggs":{"my_test":{"date_histogram":{"field":"Created","interval":"month", "offset": "+20d+1h+2m+3s+4ms+5micros+6nanos"}}}}' --address=0.0.0.0:8778 |  jq .stats.extra.aggregations
{
  "my_test": {
    "buckets": [
      {
        "key_as_string": "2015-04-01T01:02:03.004+00:00",
        "key": 1427850123004,
        "doc_count": 6
      },
      {
        "key_as_string": "2016-04-01T01:02:03.004+00:00",
        "key": 1459472523004,
        "doc_count": 4
      },
      {
        "key_as_string": "2017-04-01T01:02:03.004+00:00",
        "key": 1491008523004,
        "doc_count": 1
      },
      {
        "key_as_string": "2017-08-01T01:02:03.004+00:00",
        "key": 1501549323004,
        "doc_count": 5
      },
      {
        "key_as_string": "2017-09-01T01:02:03.004+00:00",
        "key": 1504227723004,
        "doc_count": 2
      }
    ]
  }
}
```

## keyed

```{.sh}
ryftrest -q '(RECORD CONTAINS "11")' -f datehist.json --count --format=json --body '{"aggs":{"my_test":{"date_histogram":{"field":"Created","interval":"month", "offset": "+20d+1h+2m+3s+4ms+5micros+6nanos", "keyed": true}}}}' --address=0.0.0.0:8778 |  jq .stats.extra.aggregations
{
  "my_test": {
    "buckets": {
      "2017-09-01T01:02:03.004+00:00": {
        "key_as_string": "2017-09-01T01:02:03.004+00:00",
        "key": 1504227723004,
        "doc_count": 2
      },
      "2017-08-01T01:02:03.004+00:00": {
        "key_as_string": "2017-08-01T01:02:03.004+00:00",
        "key": 1501549323004,
        "doc_count": 5
      },
      "2017-04-01T01:02:03.004+00:00": {
        "key_as_string": "2017-04-01T01:02:03.004+00:00",
        "key": 1491008523004,
        "doc_count": 1
      },
      "2016-04-01T01:02:03.004+00:00": {
        "key_as_string": "2016-04-01T01:02:03.004+00:00",
        "key": 1459472523004,
        "doc_count": 4
      },
      "2015-04-01T01:02:03.004+00:00": {
        "key_as_string": "2015-04-01T01:02:03.004+00:00",
        "key": 1427850123004,
        "doc_count": 6
      }
    }
  }
}
```

## min_doc_count
```{.sh}
ryftrest -q '(RECORD CONTAINS "11")' -f datehist.json --count --format=json --body '{"aggs":{"my_test":{"date_histogram":{"field":"Created","interval":"month", "min_doc_count": 5}}}}}' --address=0.0.0.0:8778 |  jq .stats.extra.aggregations
{
  "my_test": {
    "buckets": [
      {
        "doc_count": 6,
        "key": 1427846400000,
        "key_as_string": "2015-04-01T00:00:00.000+00:00"
      },
      {
        "doc_count": 5,
        "key": 1501545600000,
        "key_as_string": "2017-08-01T00:00:00.000+00:00"
      }
    ]
  }
}
```

# sub aggregations

```{.sh}
ryftrest -q '(RECORD CONTAINS "11")' -f datehist.json --count --format=json --body '{"aggs":{"my_test":{"date_histogram":{"field":"Created","interval":"year", "_aggs":{"bucket_aggr":{"stats":{"fi
eld": "Num"}}}}}}}' --address=0.0.0.0:8778 | jq .stats.extra.aggregations
{
  "my_test": {
    "buckets": [
      {
        "bucket_aggr": {
          "avg": 1,
          "count": 6,
          "max": 1,
          "min": 1,
          "sum": 6
        },
        "doc_count": 6,
        "key": 1420070400000,
        "key_as_string": "2015-01-01T00:00:00.000+00:00"
      },
      {
        "bucket_aggr": {
          "avg": 1,
          "count": 4,
          "max": 1,
          "min": 1,
          "sum": 4
        },
        "doc_count": 4,
        "key": 1451606400000,
        "key_as_string": "2016-01-01T00:00:00.000+00:00"
      },
      {
        "bucket_aggr": {
          "avg": 1,
          "count": 8,
          "max": 1,
          "min": 1,
          "sum": 8
        },
        "doc_count": 8,
        "key": 1483228800000,
        "key_as_string": "2017-01-01T00:00:00.000+00:00"
      }
    ]
  }
}
```