
# Installing Dependencies & Building

1. See installation instructions for golang environment — https://golang.org/doc/install
2. ``go get github.com/getryft/ryft-server``
3. ``go install``
4. Get ``ryft-server`` binary from ``$GOPATH/bin``

# Running & Command Line Parameters

```
usage: ryft-server [<flags>] [<address>]

Flags:
  --help           Show context-sensitive help (also try --help-long and --help-man).
  -k, --keep       Keep search results temporary files.
  -d, --debug      Run http server in debug mode.
  -a, --auth=AUTH  Authentication type: none, file, ldap.
  --users-file=USERS-FILE
                   File with user credentials. Required for --auth=file.
  --ldap-server=LDAP-SERVER
                   LDAP Server address:port. Required for --auth=ldap.
  --ldap-user=LDAP-USER
                   LDAP username for binding. Required for --auth=ldap.
  --ldap-pass=LDAP-PASS
                   LDAP password for binding. Required for --auth=ldap.
  --ldap-query="(&(uid=%s))"
                   LDAP user lookup query. Defauls is '(&(uid=%s))'. Required for --auth=ldap.
  --ldap-basedn=LDAP-BASEDN
                   LDAP BaseDN for lookups.'. Required for --auth=ldap.

  -t, --tls          
                    Enable TLS/SSL. Default 'false'.
  --tls-crt=TLS-CRT  
                    Certificate file. Required for --tls=true.
  --tls-key=TLS-KEY  
                    Key-file. Required for --tls=true.
  --tls-address=0.0.0.0:8766  
                     Address:port to listen on HTTPS. Default is 0.0.0.0:8766

Args:
  [<address>]  Address:port to listen on. Default is 0.0.0.0:8765.

```
Default value ``port`` is ``8765``
# Packaging into deb file

https://github.com/getryft/ryft-server/blob/master/ryft-server-make-deb/README.md

# Keeping search results

By default REST-server removes search results from ``/ryftone/RyftServer-PORT/``. But it behaviour may be prevented:

```
ryft-server --keep
```
Please pay attention that REST-server removes ``/ryftone/RyftServer-PORT`` when it starts.

# How to do a search?
Do request in browser:

```
http://localhost:8765/search?query=( RAW_TEXT CONTAINS "Johm" )&files=passengers.txt&surrounding=10&fuzziness=2&cs=true

```

# How to do search by field's value?

```
http://localhost:8765/search?query=(RECORD.id%20EQUALS%20%2210034183%22)&files=*.pcrime&surrounding=10&fuzziness=0&format=xml

```
# How to count matches?

```
http://localhost:8765//count?query=(RECORD CONTAINS "a")OR(RECORD CONTAINS "b")&files=*.pcrime&format=xml

```

# Parameters (TODO)
* ``query`` is the string specifying the search criteria.
* ``files``  is the input data set to be searched
* ``fuzziness`` Specify the fuzzy search distance [0..255]
* ``cs`` Case sensitive flag. Default 'false'.
* ``format`` is the parameter for the structed search. Specify the search format.
* ``surrounding`` width when generating results. For example, a value of 2 means that 2 + * characters before and after a search match will be included with data result
* ``fields`` specifies needed keys in result. Required format=xml.
* ``nodes`` specifies nodes count [0..4]. Default 4, if nodes=0 system will use default value .
