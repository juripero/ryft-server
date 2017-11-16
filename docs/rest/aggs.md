The ryft server supports the following set of aggregation functions:
- [Min](#min-aggregation)
- [Max](#max-aggregation)
- [Sum](#sum-aggregation)
- [Value Count](#value-count-aggregation)
- [Avg](#avg-aggregation)
- [Stats](#stats-aggregation)
- [Extended Stats](#extended-stats-aggregation)
- [Geo Bounds](#geo-bounds-aggregation)
- [Geo Centroid](#geo-centroid-aggregation)
- [Date histogram](#date-histogram-aggregation)

The aggregations can be requested via corresponding `POST /search` or `POST /count`
methods. There should be POST body JSON object containing all required information.

Note, that it's possible to run custom post-processing script with `GET /run` method.
So if there is no required aggregation it's relatively easy to implement it manually.


# Min aggregation

This aggregation calculates minimum value over a set of found records.
The values can be extracted from specific numeric field in the record, for
example "price" or "foo.bar".

```{.json}
{"aggs" : {"min_bar" : {"min" : {"field":"foo.bar"}} }}
```

Response:

```{.json}
{
  ...
  "aggregations": {
    "min_bar": {
      "value": 100
    }
  }
}
```

The aggregation name `min_bar` also serves as the key in the JSON response.

## Missing value

If requested field is missing then the whole record is ignored. But the default
value for missing fields can be specified with `missing` option.

```{.json}
{"aggs" : {"min_bar" : {"min" : {"field":"foo.bar", "missing":-1}} }}
```


# Max aggregation

This aggregation calculates maximum value over a set of found records.
The values can be extracted from specific numeric field in the record, for
example "price" or "foo.bar".

```{.json}
{"aggs" : {"max_bar" : {"max" : {"field":"foo.bar"}} }}
```

Response:

```{.json}
{
  ...
  "aggregations": {
    "max_bar": {
      "value": 500
    }
  }
}
```

The aggregation name `max_bar` also serves as the key in the JSON response.

## Missing value

If requested field is missing then the whole record is ignored. But the default
value for missing fields can be specified with `missing` option.

```{.json}
{"aggs" : {"max_bar" : {"max" : {"field":"foo.bar", "missing":0}} }}
```


# Sum aggregation

This aggregation calculates sum of values over a set of found records.
The values can be extracted from specific numeric field in the record, for
example "price" or "foo.bar".

```{.json}
{"aggs" : {"sum_bar" : {"sum" : {"field":"foo.bar"}} }}
```

Response:

```{.json}
{
  ...
  "aggregations": {
    "sum_bar": {
      "value": 1500
    }
  }
}
```

The aggregation name `sum_bar` also serves as the key in the JSON response.

## Missing value

If requested field is missing then the whole record is ignored. But the default
value for missing fields can be specified with `missing` option.

```{.json}
{"aggs" : {"sum_bar" : {"sum" : {"field":"foo.bar", "missing":0}} }}
```


# Value Count aggregation

This aggregation just counts the number of values over a set of found records.
The values can be extracted from specific numeric field in the record, for
example "price" or "foo.bar".

```{.json}
{"aggs" : {"count_bar" : {"count" : {"field":"foo.bar"}} }}
```

Response:

```{.json}
{
  ...
  "aggregations": {
    "count_bar": {
      "value": 5
    }
  }
}
```

The aggregation name `count_bar` also serves as the key in the JSON response.


# Avg aggregation

This aggregation calculates the average value over a set of found records.
The values can be extracted from specific numeric field in the record, for
example "price" or "foo.bar".

```{.json}
{"aggs" : {"avg_bar" : {"avg" : {"field":"foo.bar"}} }}
```

Response:

```{.json}
{
  ...
  "aggregations": {
    "avg_bar": {
      "value": 300
    }
  }
}
```

The aggregation name `avg_bar` also serves as the key in the JSON response.

## Missing value

If requested field is missing then the whole record is ignored. But the default
value for missing fields can be specified with `missing` option.

```{.json}
{"aggs" : {"avg_bar" : {"avg" : {"field":"foo.bar", "missing":0}} }}
```


# Stats aggregation

This aggregation calculates the main statistics over a set of found records.
The values can be extracted from specific numeric field in the record, for
example "price" or "foo.bar".

```{.json}
{"aggs" : {"stats_bar" : {"stats" : {"field":"foo.bar"}} }}
```

Response (combination of `min`, `max`, `sum`, `avg` and `count` aggregations):

```{.json}
{
  ...
  "aggregations": {
    "stats_bar": {
        "avg": 300,
        "count": 5,
        "max": 500,
        "min": 100,
        "sum": 1500
    }
  }
}
```

The aggregation name `stats_bar` also serves as the key in the JSON response.

## Missing value

If requested field is missing then the whole record is ignored. But the default
value for missing fields can be specified with `missing` option.

```{.json}
{"aggs" : {"stats_bar" : {"stats" : {"field":"foo.bar", "missing":0}} }}
```


# Extended Stats aggregation

This aggregation calculates the extended statistics over a set of found records.
The values can be extracted from specific numeric field in the record, for
example "price" or "foo.bar".

```{.json}
{"aggs" : {"stats_bar" : {"extended_stats" : {"field":"foo.bar"}} }}
```

Response (combination of `min`, `max`, `sum`, `avg`, `count` and
`variance` aggregations):

```{.json}
{
  ...
  "aggregations": {
    "stats_bar": {
        "avg": 300,
        "count": 5,
        "max": 500,
        "min": 100,
        "std_deviation": 141.4213562373095,
        "std_deviation_bounds": {
          "lower": 17.15728752538098,
          "upper": 582.842712474619
        },
        "sum": 1500,
        "sum_of_squares": 550000,
        "variance": 20000
      }
  }
}
```

The aggregation name `stats_bar` also serves as the key in the JSON response.

## Missing value

If requested field is missing then the whole record is ignored. But the default
value for missing fields can be specified with `missing` option.

```{.json}
{"aggs" : {"stats_bar" : {"extended_stats" : {"field":"foo.bar", "missing":0}} }}
```

## Sigma value

The `std_deviation_bounds` is calculated as [`avg-sigma*std_deviation`, `avg+sigma*std_deviation`].
By default there is `sigma=2` but this value can be specified in the request JSON:

```{.json}
{"aggs" : {"stats_bar" : {"extended_stats" : {"field":"foo.bar", "sigma":3}} }}
```


# Geo Bounds aggregation

This aggregation calculates the bounding rectangle over a set of found records.
The positions can be extracted from specific field in the record, for
example "location" or "foo.bar".

```{.json}
{"aggs" : {"viewport" : {"geo_bounds" : {"field":"pos"}} }}
```

Response:

```{.json}
{
  ...
  "aggregations": {
    "viewport": {
        "bounds": {
          "bottom_right": {
            "lat": 10.1,
            "lon": 50.5
          },
          "top_left": {
            "lat": 50.5,
            "lon": 10.1
          }
        }
  }
}
```

The aggregation name `viewport` also serves as the key in the JSON response.

Instead of one `field` in "lat,lon" format it is possible to specify separate
fields for latitude and longitude:

```{.json}
{"aggs" : {"viewport" : {"geo_bounds" : {"lat":"latitude", "lon":"longitude"}} }}
```

## Geo Data format

The Geo position can be in the following formats:

- string containing `"<lat>,<lon>"`
- array of two values `[<lon>, <lat>]` (Note the order!)
- object `{"lat": <lat>, "lon": <lon>}`

Where `<lat>` and `<lon>` are valid floating point numbers.


# Geo Centroid aggregation

This aggregation calculates the simple or weighted centeroid over a set of found records.
The coordinates can be extracted from specific field in the record, for
example "location" or "foo.bar".

```{.json}
{"aggs" : {"center" : {"geo_centroid" : {"field":"pos", "weighted": true}} }}
```

Response:

```{.json}
{
  ...
  "aggregations": {
    "center": {
        "centroid": {
          "count": 5,
          "location": {
            "lat": 31.05136881163401,
            "lon": 28.16414539388629
          }
        }
  }
}
```

The aggregation name `center` also serves as the key in the JSON response.

Instead of one `field` in "lat,lon" format it is possible to specify separate
fields for latitude and longitude:

```{.json}
{"aggs" : {"center" : {"geo_centroid" : {"lat":"latitude", "lon":"longitude"}} }}
```

See [Geo Data format](#geo-data-format) for the list of supported coordinates formats.

The additional `weighted` option specifies centroid calculation algorithm.
If `weighted` is `false` (by default) then the simple average is used as centroid point.

If `weighted` is `true` then weighted average is used instead: all points are
converted to 3D space and then averaged. Averaged point is converted back
to latitude/longitude to get centroid point.
This algorithm consumes CPU resources since we need to
calculate a lot of `sin/cos` values, but the centroid point is more precisely.

# Date histogram aggregation

Date histogram splits search results on buckets and counts number of rows inside. Key of a bucket is interval and its values are rows, each contains datetime field from the interval. It is also possible to apply sub-aggregaton to each bucket.

`field` (required) - name of a field that contains date.

`interval` (required) - time interval. Possible values: `year`, `quarter`, `month`, `week`, `day`, `hour`, `minute`, `second`.

    It may also be specified with time-units: `d`, `h`, `m`, `s`, `ms`, `micros`, `nanos`.
    Fractional time values are not supported.

`offset` (optional) - the offset parameter is used to change the start value of each bucket by the specified positive (+) or negative offset (-) duration.

`time_zone` (optional) - date-times are stored in `UTC`. By default, all bucketing and rounding is also done in `UTC`. The `time_zone` parameter can be used to indicate that bucketing should use a different time zone. It may be set as UTC offset (e.g. +01:00, -08:00) or as time zone id in TZ database.

`format` (optional) - the response has `key_as_string` field, that represents time interval formatted with the `format` value. It supports jodaDate notation. E.g. "yyyy-MM-dd".

`keyed` (optional) - show buckets as a list or as a map.

`min_doc_count` (optional) - show aggregation results only for buckets that have number of documents not less then value of this parameter.

`_aggs` (optional) - sub-aggregations. You can use any single-bucket aggregaton described above.