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

<h2>Examples</h2>
<h3>Not structed request example</h3>
<p><a href="http://52.20.99.136:8765/search?query=(RECORD.id%20EQUALS%20%2210034183%22)&amp;files=*.pcrime&amp;surrounding=10&amp;fuzziness=0">
    http://52.20.99.136:8765/search?query=(RECORD.id%20EQUALS%20%2210034183%22)&amp;files=*.pcrime&amp;surrounding=10&amp;fuzziness=0</a ></p>
<h3>Structed request example</h3>
<p><a href="http://52.20.99.136:8765/search?query=(RECORD.id%20EQUALS%20%2210034183%22)&amp;files=*.pcrime&amp;surrounding=10&amp;fuzziness=0&amp;format=xml">
  http://52.20.99.136:8765/search?query=(RECORD.id%20EQUALS%20%2210034183%22)&amp;files=*.pcrime&amp;surrounding=10&amp;fuzziness=0&amp;format=xml</a ></p>

    </body>
</html>
`
