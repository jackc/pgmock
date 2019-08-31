# pgmock

pgmock provides the ability to mock a PostgreSQL server.

See pgmock_test.go for example usage.

## pgmockproxy

pgmockproxy is a PostgreSQL proxy that logs the messages back and forth between the PostgreSQL client and server. This
can aid in building a mocking script by running commands against a real server to observe the results. It can also be
used to debug applications that speak the PostgreSQL wire protocol without needing to use a tool like Wireshark.

Example usage:

```
$ pgmockproxy -remote "/private/tmp/.s.PGSQL.5432"
F {"Type":"StartupMessage","ProtocolVersion":196608,"Parameters":{"application_name":"psql","client_encoding":"UTF8","database":"jack","user":"jack"}}
B {"Type":0,"Salt":[0,0,0,0],"SASLAuthMechanisms":null,"SASLData":null}
B {"Type":"ParameterStatus","Name":"application_name","Value":"psql"}
B {"Type":"ParameterStatus","Name":"client_encoding","Value":"UTF8"}
B {"Type":"ParameterStatus","Name":"DateStyle","Value":"ISO, MDY"}
B {"Type":"ParameterStatus","Name":"integer_datetimes","Value":"on"}
B {"Type":"ParameterStatus","Name":"IntervalStyle","Value":"postgres"}
B {"Type":"ParameterStatus","Name":"is_superuser","Value":"on"}
B {"Type":"ParameterStatus","Name":"server_encoding","Value":"UTF8"}
B {"Type":"ParameterStatus","Name":"server_version","Value":"11.5"}
B {"Type":"ParameterStatus","Name":"session_authorization","Value":"jack"}
B {"Type":"ParameterStatus","Name":"standard_conforming_strings","Value":"on"}
B {"Type":"ParameterStatus","Name":"TimeZone","Value":"US/Central"}
B {"Type":"BackendKeyData","ProcessID":31007,"SecretKey":1013083042}
B {"Type":"ReadyForQuery","TxStatus":"I"}
F {"Type":"Query","String":"select generate_series(1,5);"}
B {"Type":"RowDescription","Fields":[{"Name":"generate_series","TableOID":0,"TableAttributeNumber":0,"DataTypeOID":23,"DataTypeSize":4,"TypeModifier":-1,"Format":0}]}
B {"Type":"DataRow","Values":[{"text":"1"}]}
B {"Type":"DataRow","Values":[{"text":"2"}]}
B {"Type":"DataRow","Values":[{"text":"3"}]}
B {"Type":"DataRow","Values":[{"text":"4"}]}
B {"Type":"DataRow","Values":[{"text":"5"}]}
B {"Type":"CommandComplete","CommandTag":"SELECT 5"}
B {"Type":"ReadyForQuery","TxStatus":"I"}
F {"Type":"Terminate"}
```
