package main

const IndexHTML = `
<!DOCTYPE html>
<html>
    <head>
      <meta charset="utf-8"/>
        <title>index</title>
      <style>
      html, button, input, select, textarea {
          font-family: Helvetica-Neue,Arial,"Nimbus Sans L",sans-serif;
      }
      body {
          color: #444444;
      }
      #wrap {
          margin: 0 auto;
          max-width: 1100px;
          min-width: 960px;
          overflow: hidden;
          padding: 0 0 0 20px;
          width: 85%;
      }
      #page-title {
          margin-top: 33px;
      }
      #content {
          background-color: #FFFFFF;
          border: 1px solid #CCCCCC;
          border-radius: 2px 2px 2px 2px;
          float: none;
          font-size: 13px;
          line-height: 18px;
          /*margin: 20px 0 0 27%;*/
          padding: 20px 3%;
          vertical-align: top;
      }
      #content > div {
          display: none;
      }
      #content img {
          background-color: #EDEDED;
          border: 1px solid #CCCCCC;
          border-radius: 2px 2px 2px 2px;
          max-width: 700px;
          padding: 20px;
      }
      #content a:link {
          color: #0072A3;
          text-decoration: none;
      }
      #content a:link:hover {
          text-decoration: underline;
      }
      #content p.description {
          background-color: #DBEAF9;
          border: 1px solid #B5CEE7;
          font-style: italic;
          line-height: 15px;
          margin: 10px 0 8px;
          padding: 7px;
      }
      #content ul {
          margin: 0 0 0 30px;
      }
      #content ul li {
          padding: 0;
          margin: 2px;
      }
      #content h1 {
          /*border-bottom: 1px solid #BBBBBB;*/
          color: #222222;
          font-size: 2em;
          font-weight: normal;
          line-height: 1em;
          text-align: center;
          margin: 0;
          padding-bottom: 14px;
      }
      #content h2 {
          color: #333333;
          font-size: 1.5em;
          font-weight: normal;
          line-height: 18px;
          margin: 30px 0 0;
          border-bottom: 1px solid #BBBBBB;
          padding-bottom: 14px;
      }
      #content h3 {
          color: #9F6227;
          font-size: 1em;
          font-weight: bold;
          margin: 20px 0 10px;
      }
      #content p {
          line-height: 18px;
          margin: 8px 0;
          padding: 0;
      }
      #content ul {
          margin: 0 0 0 30px;
          padding: 0;
      }
      #content li {
          margin: 8px 0;
          padding: 0;
      }
      #content table {
        width: 100%;
        table-layout: fixed;
          border-collapse: collapse;
          margin-top: 15px;
      }
      #content table td, #content table th {
          border-bottom: 1px solid #DBDBDB;
          color: black;
          font-size: 0.9em;
          padding: 4px;
          text-align: left;
      }
      #content table th {
          background: none repeat scroll 0 0 #F1F1F1;
          font-weight: bold;
      }
      #content img {
          border: 0 none;
      }
      #content samp,
      #content pre {
          background-color: #FFFDED;
          border-color: #E2DB8D;
          border-style: solid;
          border-width: 1px 1px 1px 10px;
          padding: 7px;
          width: 97%;
      }
      #content samp{max-width:100%;max-height:300px;display:block; white-space: pre;
      overflow: auto;}
      #content pre span {
          font-family: "Courier New",Courier,monospace;
      }
      #content input, #content select, #content textarea {
          font-family: "LucidaGrande",Arial,Helvetica,sans-serif;
      }
      #content .green {
          color: #008800;
      }
      #content .blue {
          color: #000088;
      }
      #content .hidden {
          display: none;
      }
      #content .left {
          float: left;
      }
      #content .right {
          float: right;
      }
      #content table p.doc {
          margin: 0;
      }
      #content pre.doc, #content pre.uri {
          color: #008800;
      }
      #content table li.doc {
          margin: 2px 0;
      }
      #sidebar {
          background: none repeat scroll 0 0 #FFFFFF;
          border: 1px solid #CCCCCC;
          border-radius: 3px 3px 3px 3px;
          float: left;
          padding: 15px 0;
          text-decoration: none;
          vertical-align: top;
          width: 25%;
      }
      #sidebar ul {
          list-style-type: none;
          margin-bottom: 0;
          margin-left: 20px;
          margin-top: 0;
          padding: 0;
      }
      #sidebar ul a {
          color: #666666;
          display: inline-block;
          font-size: 12px;
          font-weight: bold;
          line-height: 24px;
          padding: 0 10px;
          text-decoration: none;
      }
      #sidebar ul a:hover {
          background-color: #F2F2F2;
          color: #4C94B3;
      }
      #sidebar ul li.active a {
          background-color: #FFFFFF;
          color: #4C94B3;
      }
      #sidebar ul li {
          margin: 0;
      }

      </style>
    </head>
    <body>
      <div id="content">
<h1> Ryft Server </h1>
<h2>Search endpoint /search parameters :</h2>
<table>
  <thead>
    <tr>
        <th style="width:120px">Method</th>
        <th style="width:120px">Input type</th>
        <th style="width:300px">Uri</th>
        <th style="width:400px">Description</th>
    </tr>
  </thead>
    <tbody>
      <tr>
          <td><a>query</a></td>
          <td>string</td>
          <td>GET /search?query={QUERY}</td>
          <td>String that specifying the search criteria. Required file parameter</td>
      </tr>
    <tr>
        <td><a>files</a></td>
        <td>string</td>
        <td>GET /search?query={QUERY}&amp;files={FILE}</td>
        <td>Input data set to be searched. Comma separated list of files or directories.</td>
    </tr>
    <tr>
        <td><a>fuzziness</a></td>
        <td>uint8</td>
        <td>GET /search?query={QUERY}&amp;files={FILE}&amp;fuzziness={VALUE}</td>
        <td>Specify the fuzzy search distance [0..255] .</td>
    </tr>
    <tr>
        <td><a>cs</a></td>
        <td>string</td>
        <td>GET /search?query={QUERY}&amp;files={FILE}&amp;cs=true</td>
        <td>Case sensitive flag. Default 'false'.</td>
    </tr>
    <tr>
        <td><a>format</a></td>
        <td>string</td>
        <td>GET /search?query={QUERY}&amp;files={FILE}&apm;format={FORMAT}</td>
        <td>Parameter for the structed search.  Specify the input data format 'xml' or 'raw'(Default).</td>
    </tr>
    <tr>
        <td><a>surroinding</a></td>
        <td>uint16</td>
        <td>GET /search?query={QUERY}&amp;files={FILE}&amp;surrounding={VALUE}</td>
        <td>Parameter that specifies the number of characters before the match and after the match that will be returned when the input specifier type is raw text</td>
    </tr>
    <tr>
        <td><a>fields</a></td>
        <td>string</td>
        <td>GET /search?query={QUERY}&amp;files={FILE}&amp;format=xml&amp;fields={FIELDS...}</td>
        <td>Parametr that specifies needed keys in result. Required format=xml.</td>
    </tr>
    <tr>
        <td><a>nodes</a></td>
        <td>string</td>
        <td>GET /search?query={QUERY}&amp;files={FILE}&amp;nodes={VALUE}</td>
        <td>Parameter that specifies nodes count [0..4]. Default 4, if nodes=0 system will use default value.</td>
    </tr>
</tbody>
</table>
<!-- <h2>Examples</h2> -->
<h3>Not structed request example</h3>
<p><a href="/search?query=(RAW_TEXT%20CONTAINS%20%2210%22)&files=passengers.txt&surrounding=10&fuzziness=0">
  /search?query=(RAW_TEXT CONTAINS "10")&amp;files=passengers.txt&amp;surrounding=10&amp;fuzziness=0</a ></p>
  <samp>[
    {
        "_index": {
            "file": "/ryftone/passengers.txt",
            "offset": 27,
            "length": 22,
            "fuzziness": 0
        },
        "data": "YWwgU21pdGgsIDEwLTAxLTE5MjgsMA=="
    },
    {
        "_index": {
            "file": "/ryftone/passengers.txt",
            "offset": 43,
            "length": 22,
            "fuzziness": 0
        },
        "data": "MTkyOCwwMTEtMzEwLTU1NS0xMjEyLA=="
    },
    {
        "_index": {
            "file": "/ryftone/passengers.txt",
            "offset": 108,
            "length": 22,
            "fuzziness": 0
        },
        "data": "LTI5LTE5NDUsMzEwLTU1NS0yMzIzLA=="
    },
    {
        "_index": {
            "file": "/ryftone/passengers.txt",
            "offset": 167,
            "length": 22,
            "fuzziness": 0
        },
        "data": "LTMwLTE5MjAsMzEwLTU1NS0zNDM0LA=="
    },
    {
        "_index": {
            "file": "/ryftone/passengers.txt",
            "offset": 234,
            "length": 22,
            "fuzziness": 0
        },
        "data": "MTk1MiwwMTEtMzEwLTU1NS00NTQ1LA=="
    },
    {
        "_index": {
            "file": "/ryftone/passengers.txt",
            "offset": 344,
            "length": 22,
            "fuzziness": 0
        },
        "data": "LTE1LTE5NDQsMzEwLTU1NS01NjU2LA=="
    },
    {
        "_index": {
            "file": "/ryftone/passengers.txt",
            "offset": 478,
            "length": 22,
            "fuzziness": 0
        },
        "data": "LTE0LTE5NDksMzEwLTU1NS02NzY3LA=="
    },
    {
        "_index": {
            "file": "/ryftone/passengers.txt",
            "offset": 569,
            "length": 22,
            "fuzziness": 0
        },
        "data": "LTEyLTE5NTksMzEwLTU1NS0xMjEzLA=="
    },
    {
        "_index": {
            "file": "/ryftone/passengers.txt",
            "offset": 663,
            "length": 22,
            "fuzziness": 0
        },
        "data": "LTEyLTE5NTksMzEwLTU1NS0xMjEzLA=="
    },
    {
        "_index": {
            "file": "/ryftone/passengers.txt",
            "offset": 770,
            "length": 22,
            "fuzziness": 0
        },
        "data": "LTEyLTE5NTksMzEwLTU1NS0xMjEzLA=="
    },
    {
        "_index": {
            "file": "/ryftone/passengers.txt",
            "offset": 890,
            "length": 22,
            "fuzziness": 0
        },
        "data": "LTEyLTE5ODksMzEwLTU1NS05ODc2LA=="
    },
    {
        "_index": {
            "file": "/ryftone/passengers.txt",
            "offset": 966,
            "length": 22,
            "fuzziness": 0
        },
        "data": "LTI1LTE5ODUsMzEwLTU1NS0zNDI1LA=="
    }
]</samp>
<h3>Structed request example</h3>
<p><a href="/search?query=(RECORD.id%20EQUALS%20%2210034183%22)&files=*.pcrime&surrounding=10&fuzziness=0&format=xml">
  /search?query=(RECORD.id EQUALS "10034183")&amp;files=*.pcrime&amp;surrounding=10&amp;fuzziness=0&amp;format=xml</a ></p>
<samp>{
    "Arrest": "false",
    "Beat": "0313",
    "Block": "062XX S ST LAWRENCE AVE",
    "CaseNumber": "HY223673",
    "CommunityArea": "42",
    "Date": "04/15/2015 11:59:00 PM",
    "Description": "DOMESTIC BATTERY SIMPLE",
    "District": "003",
    "Domestic": "true",
    "FBICode": "08B",
    "ID": "10034183",
    "IUCR": "0486",
    "Latitude": "41.781961688",
    "Location": "\"(41.781961688, -87.610984705)\"",
    "LocationDescription": "STREET",
    "Longitude": "-87.610984705",
    "PrimaryType": "BATTERY",
    "UpdatedOn": "04/22/2015 12:47:10 PM",
    "Ward": "20",
    "XCoordinate": "1181263",
    "YCoordinate": "1863965",
    "Year": "2015",
    "_index": {
        "file": "/ryftone/chicago.pcrime",
        "offset": 0,
        "length": 693,
        "fuzziness": 0
    }
}</samp>

<h2>Count endpoint</h2>
<!-- <p>To count elemnts use /count endpoint. Works with all parameters above. </p> -->

<table>
  <thead>
    <tr>
        <th style="width:120px">Method</th>
        <th style="width:120px">Input type</th>
        <th style="width:300px">Uri</th>
        <th style="width:400px">Description</th>
    </tr>
  </thead>
    <tbody>
      <tr>
          <td><a>query</a></td>
          <td>string</td>
          <td>GET /count?query={QUERY}</td>
          <td>String that specifying the search criteria. Required file parameter</td>
      </tr>
    <tr>
        <td><a>files</a></td>
        <td>string</td>
        <td>GET /count?query={QUERY}&amp;files={FILE}</td>
        <td>Input data set to be searched. Comma separated list of files or directories.</td>
    </tr>
    <tr>
        <td><a>fuzziness</a></td>
        <td>uint8</td>
        <td>GET /count?query={QUERY}&amp;files={FILE}&amp;fuzziness={VALUE}</td>
        <td>Specify the fuzzy search distance [0..255] .</td>
    </tr>
    <tr>
        <td><a>cs</a></td>
        <td>string</td>
        <td>GET /count?query={QUERY}&amp;files={FILE}&amp;cs=true</td>
        <td>Case sensitive flag. Default 'false'.</td>
    </tr>
    <tr>
        <td><a>nodes</a></td>
        <td>string</td>
        <td>GET /count?query={QUERY}&amp;files={FILE}&amp;nodes={VALUE}</td>
        <td>Parameter that specifies nodes count [0..4]. Default 4, if nodes=0 system will use default value.</td>
    </tr>
</tbody>
</table>
<h3>Count request example</h3>
<p><a href="/count?query=(RECORD%20CONTAINS%20%22a%22)OR(RECORD%20CONTAINS%20%22b%22)&files=*.pcrime">
  /count?query=(RECORD CONTAINS "a")OR(RECORD CONTAINS "b")&amp;files=*.pcrime</a ></p>
<pre>"Matching: 10000"</pre>
</div>
    </body>
</html>

`
