package main

const IndexHTML = `
<!DOCTYPE html>
<html>
    <head>
        <title>index</title>
    </head>
    <body>
<h2>Search Endpoint /search </h2>
<h3>Parameters</h3>
<p><b>query - </b> is the string specifying the search criteria.</p>
<p><b>files - </b> is the input data set to be searched. Comma separated list of files or directories. </p>
<p><b>fuzziness - </b> Specify the fuzzy search distance [0..255] . </p>
<p><b>cs - </b>  Case sensitive flag. Default 'false' </p>
<p><b>format - </b>  is the parameter for the structed search.  Specify the input data format 'xml' or 'raw'(Default). </p>
<p><b>surrounding - </b> specifies the number of characters before the match and after the match that will be returned when the input specifier type is raw text </p>
<p><b>fields - </b> specifies needed keys in result. Required format=xml. </p>
<p><b>nodes - </b> specifies nodes count [0..4]. Default 4, if nodes=0 system will use default value. </p>

<h2>Examples</h2>
<h3>Not structed request example</h3>
<p><a href="/search?query=(RECORD.id%20EQUALS%20%2210034183%22)&amp;files=*.pcrime&amp;surrounding=10&amp;fuzziness=0">
  /search?query=(RECORD.id%20EQUALS%20%2210034183%22)&amp;files=*.pcrime&amp;surrounding=10&amp;fuzziness=0</a ></p>
<h3>Structed request example</h3>
<p><a href="/search?query=(RECORD.id%20EQUALS%20%2210034183%22)&amp;files=*.pcrime&amp;surrounding=10&amp;fuzziness=0&amp;format=xml">
  /search?query=(RECORD.id%20EQUALS%20%2210034183%22)&amp;files=*.pcrime&amp;surrounding=10&amp;fuzziness=0&amp;format=xml</a ></p>
<h3>Count request example</h3>
<p><a href="/count?query=(RECORD%20CONTAINS%20%22a%22)OR(RECORD%20CONTAINS%20%22b%22)&files=*.pcrime&format=xml">
  /count?query=(RECORD%20CONTAINS%20%22a%22)OR(RECORD%20CONTAINS%20%22b%22)&files=*.pcrime&format=xml</a ></p>

    </body>
</html>
`
