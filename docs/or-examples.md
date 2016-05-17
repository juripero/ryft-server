# Demo April 28, 2015 (OR examples)

## Input file

As an input we will use special `or.pcrime` file:

```{.xml}
<rec><ID>10034183</ID><CaseNumber>HY223673</CaseNumber><Date>04/10/2015 10:15:00 PM</Date><Block>002XX</Block><IUCR>0486</IUCR><PrimaryType>BATTERY</PrimaryType><Description>John</Description><LocationDescription>STREET</LocationDescription><Arrest>false</Arrest><Domestic>true</Domestic><Beat>0313</Beat><District>003</District><Ward>20</Ward><CommunityArea>42</CommunityArea><FBICode>08B</FBICode><XCoordinate>1181263</XCoordinate><YCoordinate>1863965</YCoordinate><Year>2015</Year><UpdatedOn>04/22/2015 12:47:10 PM</UpdatedOn><Latitude>41.781961688</Latitude><Longitude>-87.610984705</Longitude><Location>"(41.781961688, -87.610984705)"</Location></rec>
<rec><ID>10034188</ID><CaseNumber>HY223687</CaseNumber><Date>04/10/2015 10:30:00 PM</Date><Block>003XX</Block><IUCR>0820</IUCR><PrimaryType>THEFT</PrimaryType><Description>Jonny</Description><LocationDescription>SIDEWALK</LocationDescription><Arrest>false</Arrest><Domestic>true</Domestic><Beat>1123</Beat><District>011</District><Ward>28</Ward><CommunityArea>27</CommunityArea><FBICode>06</FBICode><XCoordinate>1152292</XCoordinate><YCoordinate>1901795</YCoordinate><Year>2015</Year><UpdatedOn>04/22/2015 12:47:10 PM</UpdatedOn><Latitude>41.886390821</Latitude><Longitude>-87.716204071</Longitude><Location>"(41.886390821, -87.716204071)"</Location></rec>
<rec><ID>10034213</ID><CaseNumber>HY223716</CaseNumber><Date>04/11/2015 10:45:00 PM</Date><Block>001XX</Block><IUCR>0470</IUCR><PrimaryType>PUBLIC PEACE VIOLATION</PrimaryType><Description>Jenny</Description><LocationDescription>ALLEY</LocationDescription><Arrest>true</Arrest><Domestic>false</Domestic><Beat>0522</Beat><District>005</District><Ward>9</Ward><CommunityArea>53</CommunityArea><FBICode>24</FBICode><XCoordinate>1177304</XCoordinate><YCoordinate>1825999</YCoordinate><Year>2015</Year><UpdatedOn>04/22/2015 12:47:10 PM</UpdatedOn><Latitude>41.67786846</Latitude><Longitude>-87.626642113</Longitude><Location>"(41.67786846, -87.626642113)"</Location></rec>
<rec><ID>10034327</ID><CaseNumber>HY223684</CaseNumber><Date>04/11/2015 10:53:00 PM</Date><Block>075XX</Block><IUCR>0486</IUCR><PrimaryType>BATTERY</PrimaryType><Description>Lenny</Description><LocationDescription>RESIDENTIAL YARD (FRONT/BACK)</LocationDescription><Arrest>false</Arrest><Domestic>true</Domestic><Beat>0623</Beat><District>006</District><Ward>6</Ward><CommunityArea>69</CommunityArea><FBICode>08B</FBICode><XCoordinate>1178866</XCoordinate><YCoordinate>1854896</YCoordinate><Year>2015</Year><UpdatedOn>04/22/2015 12:47:10 PM</UpdatedOn><Latitude>41.757130314</Latitude><Longitude>-87.620048394</Longitude><Location>"(41.757130314, -87.620048394)"</Location></rec>
<rec><ID>10034247</ID><CaseNumber>HY223708</CaseNumber><Date>04/12/2015 10:52:00 PM</Date><Block>081XX</Block><IUCR>0486</IUCR><PrimaryType>BATTERY</PrimaryType><Description>Manny</Description><LocationDescription>VEHICLE NON-COMMERCIAL</LocationDescription><Arrest>false</Arrest><Domestic>true</Domestic><Beat>0414</Beat><District>004</District><Ward>8</Ward><CommunityArea>46</CommunityArea><FBICode>08B</FBICode><XCoordinate>1190898</XCoordinate><YCoordinate>1851594</YCoordinate><Year>2015</Year><UpdatedOn>04/22/2015 12:47:10 PM</UpdatedOn><Latitude>41.747787125</Latitude><Longitude>-87.57606016</Longitude><Location>"(41.747787125, -87.57606016)"</Location></rec>
<rec><ID>10034197</ID><CaseNumber>HY223707</CaseNumber><Date>04/12/2015 11:18:00 PM</Date><Block>001XX</Block><IUCR>1811</IUCR><PrimaryType>NARCOTICS</PrimaryType><Description>More</Description><LocationDescription>STREET</LocationDescription><Arrest>true</Arrest><Domestic>false</Domestic><Beat>1523</Beat><District>015</District><Ward>28</Ward><CommunityArea>25</CommunityArea><FBICode>18</FBICode><XCoordinate>1141655</XCoordinate><YCoordinate>1900379</YCoordinate><Year>2015</Year><UpdatedOn>04/22/2015 12:47:10 PM</UpdatedOn><Latitude>41.882708414</Latitude><Longitude>-87.75530118</Longitude><Location>"(41.882708414, -87.75530118)"</Location></rec>
<rec><ID>10034248</ID><CaseNumber>HY223738</CaseNumber><Date>04/13/2015 11:48:00 PM</Date><Block>015XX</Block><IUCR>1121</IUCR><PrimaryType>DECEPTIVE PRACTICE</PrimaryType><Description>Less</Description><LocationDescription>RESTAURANT</LocationDescription><Arrest>false</Arrest><Domestic>false</Domestic><Beat>1012</Beat><District>010</District><Ward>24</Ward><CommunityArea>29</CommunityArea><FBICode>10</FBICode><XCoordinate>1149938</XCoordinate><YCoordinate>1891833</YCoordinate><Year>2015</Year><UpdatedOn>04/22/2015 12:47:10 PM</UpdatedOn><Latitude>41.859100084</Latitude><Longitude>-87.725107817</Longitude><Location>"(41.859100084, -87.725107817)"</Location></rec>
<rec><ID>10037110</ID><CaseNumber>HY224060</CaseNumber><Date>04/13/2015 11:35:00 PM</Date><Block>011XX</Block><IUCR>0910</IUCR><PrimaryType>MOTOR VEHICLE THEFT</PrimaryType><Description>No</Description><LocationDescription>STREET</LocationDescription><Arrest>false</Arrest><Domestic>false</Domestic><Beat>1211</Beat><District>012</District><Ward>26</Ward><CommunityArea>24</CommunityArea><FBICode>07</FBICode><XCoordinate>1156128</XCoordinate><YCoordinate>1907708</YCoordinate><Year>2015</Year><UpdatedOn>04/22/2015 12:47:10 PM</UpdatedOn><Latitude>41.902540073</Latitude><Longitude>-87.701957503</Longitude><Location>"(41.902540073, -87.701957503)"</Location></rec>
<rec><ID>10034200</ID><CaseNumber>HY223668</CaseNumber><Date>04/14/2015 11:40:00 PM</Date><Block>054XX</Block><IUCR>0486</IUCR><PrimaryType>BATTERY</PrimaryType><Description>John</Description><LocationDescription>RESIDENCE</LocationDescription><Arrest>false</Arrest><Domestic>true</Domestic><Beat>1522</Beat><District>015</District><Ward>29</Ward><CommunityArea>25</CommunityArea><FBICode>08B</FBICode><XCoordinate>1140152</XCoordinate><YCoordinate>1897108</YCoordinate><Year>2015</Year><UpdatedOn>04/22/2015 12:47:10 PM</UpdatedOn><Latitude>41.873760014</Latitude><Longitude>-87.760900431</Longitude><Location>"(41.873760014, -87.760900431)"</Location></rec>
<rec><ID>10034234</ID><CaseNumber>HY223685</CaseNumber><Date>04/15/2015 11:30:00 PM</Date><Block>011XX</Block><IUCR>0320</IUCR><PrimaryType>ROBBERY</PrimaryType><Description>Job</Description><LocationDescription>SIDEWALK</LocationDescription><Arrest>false</Arrest><Domestic>false</Domestic><Beat>1824</Beat><District>018</District><Ward>42</Ward><CommunityArea>8</CommunityArea><FBICode>03</FBICode><XCoordinate>1175283</XCoordinate><YCoordinate>1908223</YCoordinate><Year>2015</Year><UpdatedOn>04/22/2015 12:47:10 PM</UpdatedOn><Latitude>41.903544846</Latitude><Longitude>-87.631582982</Longitude><Location>"(41.903544846, -87.631582982)"</Location></rec>
```

This file is based on `chicago.pcrime` with reduced number of records (10) and modified content.

## Queries

1. To get all records: [(RECORD.id CONTAINS "1003")](http://localhost:8765/search?local=true&query=%28RECORD.id%20CONTAINS%20%221003%22%29&files=or.pcrime&fields=ID,Date,Description&format=xml&stats=true)
2. To get first 6 records: [(RECORD.date CONTAINS DATE(MM/DD/YYYY <= 04/12/2015))](http://localhost:8765/search?local=true&query=%28RECORD.date%20CONTAINS%20DATE%28MM/DD/YYYY%20%3C=%2004/12/2015%29%29&files=or.pcrime&fields=ID,Date,Description&format=xml&stats=true)
3. To get last 5 records: [(RECORD.date CONTAINS TIME(HH:MM:SS > 11:15:00))](localhost:8765/search?local=true&query=(RECORD.date CONTAINS TIME(HH:MM:SS > 11:15:00))&files=or.pcrime&fields=ID,Date,Description&format=xml&fuzziness=0&stats=true)

Combining 2 and 3 queries with OR operator we should get 11 records (ID=10034197 will be included twice):
[((RECORD.date CONTAINS DATE(MM/DD/YYYY <= 04/12/2015)) OR (RECORD.date CONTAINS TIME(HH:MM:SS > 11:15:00)))](http://localhost:8765/search?local=true&query=%28%28RECORD.date%20CONTAINS%20DATE%28MM/DD/YYYY%20%3C=%2004/12/2015%29%29%20OR%20%28RECORD.date%20CONTAINS%20TIME%28HH:MM:SS%20%3E%2011:15:00%29%29%29&files=or.pcrime&fields=ID,Date,Description&format=xml&fuzziness=0&stats=true)

```{.json}
{"results":[{"Date":"04/10/2015 10:15:00 PM","Description":"John","ID":"10034183","_index":{"file":"/or.pcrime","offset":0,"length":656,"fuzziness":0,"host":"ryftone-vm-selaptop"}}
,{"Date":"04/10/2015 10:30:00 PM","Description":"Jonny","ID":"10034188","_index":{"file":"/or.pcrime","offset":657,"length":656,"fuzziness":0,"host":"ryftone-vm-selaptop"}}
,{"Date":"04/11/2015 10:45:00 PM","Description":"Jenny","ID":"10034213","_index":{"file":"/or.pcrime","offset":1314,"length":667,"fuzziness":0,"host":"ryftone-vm-selaptop"}}
,{"Date":"04/11/2015 10:53:00 PM","Description":"Lenny","ID":"10034327","_index":{"file":"/or.pcrime","offset":1982,"length":679,"fuzziness":0,"host":"ryftone-vm-selaptop"}}
,{"Date":"04/12/2015 10:52:00 PM","Description":"Manny","ID":"10034247","_index":{"file":"/or.pcrime","offset":2662,"length":670,"fuzziness":0,"host":"ryftone-vm-selaptop"}}
,{"Date":"04/12/2015 11:18:00 PM","Description":"More","ID":"10034197","_index":{"file":"/or.pcrime","offset":3333,"length":655,"fuzziness":0,"host":"ryftone-vm-selaptop"}}
,{"Date":"04/12/2015 11:18:00 PM","Description":"More","ID":"10034197","_index":{"file":"/or.pcrime","offset":3333,"length":655,"fuzziness":0,"host":"ryftone-vm-selaptop"}}
,{"Date":"04/13/2015 11:48:00 PM","Description":"Less","ID":"10034248","_index":{"file":"/or.pcrime","offset":3989,"length":671,"fuzziness":0,"host":"ryftone-vm-selaptop"}}
,{"Date":"04/13/2015 11:35:00 PM","Description":"No","ID":"10037110","_index":{"file":"/or.pcrime","offset":4661,"length":666,"fuzziness":0,"host":"ryftone-vm-selaptop"}}
,{"Date":"04/14/2015 11:40:00 PM","Description":"John","ID":"10034200","_index":{"file":"/or.pcrime","offset":5328,"length":659,"fuzziness":0,"host":"ryftone-vm-selaptop"}}
,{"Date":"04/15/2015 11:30:00 PM","Description":"Job","ID":"10034234","_index":{"file":"/or.pcrime","offset":5988,"length":656,"fuzziness":0,"host":"ryftone-vm-selaptop"}}
],"stats":{"matches":11,"totalBytes":13290,"duration":1027,"dataRate":0.17293808260219598,"fabricDataRate":0.172939}
}
```

Combining 1 and 3 queries we should get 15 records (the last 5 will have duplicates):
[(RECORD.id CONTAINS "1003") OR (RECORD.date CONTAINS TIME(HH:MM:SS > 11:15:00))](http://localhost:8765/search?local=true&query=%28RECORD.id%20CONTAINS%20%221003%22%29%20OR%20%28RECORD.date%20CONTAINS%20TIME%28HH:MM:SS%20%3E%2011:15:00%29%29&files=or.pcrime&fields=ID,Date,Description&format=xml&fuzziness=0&stats=true)

```{.json}
{"results":[{"Date":"04/10/2015 10:15:00 PM","Description":"John","ID":"10034183","_index":{"file":"/or.pcrime","offset":0,"length":656,"fuzziness":0,"host":"ryftone-vm-selaptop"}}
,{"Date":"04/10/2015 10:30:00 PM","Description":"Jonny","ID":"10034188","_index":{"file":"/or.pcrime","offset":657,"length":656,"fuzziness":0,"host":"ryftone-vm-selaptop"}}
,{"Date":"04/11/2015 10:45:00 PM","Description":"Jenny","ID":"10034213","_index":{"file":"/or.pcrime","offset":1314,"length":667,"fuzziness":0,"host":"ryftone-vm-selaptop"}}
,{"Date":"04/11/2015 10:53:00 PM","Description":"Lenny","ID":"10034327","_index":{"file":"/or.pcrime","offset":1982,"length":679,"fuzziness":0,"host":"ryftone-vm-selaptop"}}
,{"Date":"04/12/2015 10:52:00 PM","Description":"Manny","ID":"10034247","_index":{"file":"/or.pcrime","offset":2662,"length":670,"fuzziness":0,"host":"ryftone-vm-selaptop"}}
,{"Date":"04/12/2015 11:18:00 PM","Description":"More","ID":"10034197","_index":{"file":"/or.pcrime","offset":3333,"length":655,"fuzziness":0,"host":"ryftone-vm-selaptop"}}
,{"Date":"04/13/2015 11:48:00 PM","Description":"Less","ID":"10034248","_index":{"file":"/or.pcrime","offset":3989,"length":671,"fuzziness":0,"host":"ryftone-vm-selaptop"}}
,{"Date":"04/13/2015 11:35:00 PM","Description":"No","ID":"10037110","_index":{"file":"/or.pcrime","offset":4661,"length":666,"fuzziness":0,"host":"ryftone-vm-selaptop"}}
,{"Date":"04/14/2015 11:40:00 PM","Description":"John","ID":"10034200","_index":{"file":"/or.pcrime","offset":5328,"length":659,"fuzziness":0,"host":"ryftone-vm-selaptop"}}
,{"Date":"04/15/2015 11:30:00 PM","Description":"Job","ID":"10034234","_index":{"file":"/or.pcrime","offset":5988,"length":656,"fuzziness":0,"host":"ryftone-vm-selaptop"}}
,{"Date":"04/12/2015 11:18:00 PM","Description":"More","ID":"10034197","_index":{"file":"/or.pcrime","offset":3333,"length":655,"fuzziness":0,"host":"ryftone-vm-selaptop"}}
,{"Date":"04/13/2015 11:48:00 PM","Description":"Less","ID":"10034248","_index":{"file":"/or.pcrime","offset":3989,"length":671,"fuzziness":0,"host":"ryftone-vm-selaptop"}}
,{"Date":"04/13/2015 11:35:00 PM","Description":"No","ID":"10037110","_index":{"file":"/or.pcrime","offset":4661,"length":666,"fuzziness":0,"host":"ryftone-vm-selaptop"}}
,{"Date":"04/14/2015 11:40:00 PM","Description":"John","ID":"10034200","_index":{"file":"/or.pcrime","offset":5328,"length":659,"fuzziness":0,"host":"ryftone-vm-selaptop"}}
,{"Date":"04/15/2015 11:30:00 PM","Description":"Job","ID":"10034234","_index":{"file":"/or.pcrime","offset":5988,"length":656,"fuzziness":0,"host":"ryftone-vm-selaptop"}}
],"stats":{"matches":15,"totalBytes":13290,"duration":1038,"dataRate":0.25959180269627213,"fabricDataRate":0.25959200000000004}
}
```

We also could use AND and OR together:
[(RECORD.id CONTAINS "1003") AND ((RECORD.date CONTAINS DATE(MM/DD/YYYY <= 04/12/2015)) OR (RECORD.date CONTAINS TIME(HH:MM:SS > 11:15:00)))](http://localhost:8765/search?local=true&query=%28RECORD.id%20CONTAINS%20%221003%22%29%20AND%20%28%28RECORD.date%20CONTAINS%20DATE%28MM/DD/YYYY%20%3C=%2004/12/2015%29%29%20OR%20%28RECORD.date%20CONTAINS%20TIME%28HH:MM:SS%20%3E%2011:15:00%29%29%29&files=or.pcrime&fields=ID,Date,Description&format=xml&fuzziness=0&stats=true)

```{.json}
{"results":[{"Date":"04/10/2015 10:15:00 PM","Description":"John","ID":"10034183","_index":{"file":"/RyftServer-8765/.temp-dec-0000000e-1-and.pcrime","offset":0,"length":656,"fuzziness":0,"host":"ryftone-vm-selaptop"}}
,{"Date":"04/10/2015 10:30:00 PM","Description":"Jonny","ID":"10034188","_index":{"file":"/RyftServer-8765/.temp-dec-0000000e-1-and.pcrime","offset":656,"length":656,"fuzziness":0,"host":"ryftone-vm-selaptop"}}
,{"Date":"04/11/2015 10:45:00 PM","Description":"Jenny","ID":"10034213","_index":{"file":"/RyftServer-8765/.temp-dec-0000000e-1-and.pcrime","offset":1312,"length":667,"fuzziness":0,"host":"ryftone-vm-selaptop"}}
,{"Date":"04/11/2015 10:53:00 PM","Description":"Lenny","ID":"10034327","_index":{"file":"/RyftServer-8765/.temp-dec-0000000e-1-and.pcrime","offset":1979,"length":679,"fuzziness":0,"host":"ryftone-vm-selaptop"}}
,{"Date":"04/12/2015 10:52:00 PM","Description":"Manny","ID":"10034247","_index":{"file":"/RyftServer-8765/.temp-dec-0000000e-1-and.pcrime","offset":2658,"length":670,"fuzziness":0,"host":"ryftone-vm-selaptop"}}
,{"Date":"04/12/2015 11:18:00 PM","Description":"More","ID":"10034197","_index":{"file":"/RyftServer-8765/.temp-dec-0000000e-1-and.pcrime","offset":3328,"length":655,"fuzziness":0,"host":"ryftone-vm-selaptop"}}
,{"Date":"04/12/2015 11:18:00 PM","Description":"More","ID":"10034197","_index":{"file":"/RyftServer-8765/.temp-dec-0000000e-1-and.pcrime","offset":3328,"length":655,"fuzziness":0,"host":"ryftone-vm-selaptop"}}
,{"Date":"04/13/2015 11:48:00 PM","Description":"Less","ID":"10034248","_index":{"file":"/RyftServer-8765/.temp-dec-0000000e-1-and.pcrime","offset":3983,"length":671,"fuzziness":0,"host":"ryftone-vm-selaptop"}}
,{"Date":"04/13/2015 11:35:00 PM","Description":"No","ID":"10037110","_index":{"file":"/RyftServer-8765/.temp-dec-0000000e-1-and.pcrime","offset":4654,"length":666,"fuzziness":0,"host":"ryftone-vm-selaptop"}}
,{"Date":"04/14/2015 11:40:00 PM","Description":"John","ID":"10034200","_index":{"file":"/RyftServer-8765/.temp-dec-0000000e-1-and.pcrime","offset":5320,"length":659,"fuzziness":0,"host":"ryftone-vm-selaptop"}}
,{"Date":"04/15/2015 11:30:00 PM","Description":"Job","ID":"10034234","_index":{"file":"/RyftServer-8765/.temp-dec-0000000e-1-and.pcrime","offset":5979,"length":656,"fuzziness":0,"host":"ryftone-vm-selaptop"}}
],"stats":{"matches":11,"totalBytes":13270,"duration":1042,"dataRate":0.1436297348163805,"fabricDataRate":0.14363}
}
```

Note the index of the resulting records. The file contains reference to temporary generated file - result of first AND operand.
The number of records is still 11.

OR operator has lower priority, so `a OR b AND c` is equivalent to `a OR (b AND c)`:
[(RECORD.date CONTAINS DATE(MM/DD/YYYY <= 04/12/2015)) OR (RECORD.date CONTAINS TIME(HH:MM:SS > 11:15:00)) AND (RECORD.id CONTAINS "1003")](http://localhost:8765/search?local=true&query=%28RECORD.date%20CONTAINS%20DATE%28MM/DD/YYYY%20%3C=%2004/12/2015%29%29%20OR%20%28RECORD.date%20CONTAINS%20TIME%28HH:MM:SS%20%3E%2011:15:00%29%29%20AND%20%28RECORD.id%20CONTAINS%20%221003%22%29&files=or.pcrime&fields=ID,Date,Description&format=xml&fuzziness=0&stats=true)
