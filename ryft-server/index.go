package main

const IndexHTML = `
<html><body>
Examples:
<ul>
<li>
	<a href='/search?query=(RAW_TEXT CONTAINS "night")&files=passengers.txt&surrounding=10'>
		/search?query=(RAW_TEXT CONTAINS "night")&files=passengers.txt&surrounding=10
	</a>
</li>
<li>
	<a href='/search?query=(RAW_TEXT CONTAINS "night")&files=passengers.txt&surrounding=10&fuzziness=2'>
		/search?query=(RAW_TEXT CONTAINS "night")&files=passengers.txt&surrounding=10&fuzziness=2
	</a>
</li>
</ul>
</body></html>
`
