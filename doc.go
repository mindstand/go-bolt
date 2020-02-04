/*Package goBolt implements drivers for the Neo4J Bolt Protocol Versions 1-4.

There are some limitations to the types of collections the internalDriver
supports.  Specifically, maps should always be of type map[string]interface{}
and lists should always be of type []interface{}.  It doesn't seem that
the Bolt protocol supports uint64 either, so the biggest number it can send
right now is the int64 max.

The URL format is: `bolt://(user):(password)@(host):(port)`
Schema must be `bolt`. User and password is only necessary if you are authenticating.
TLS is supported by using query parameters on the connection string, like so:
`bolt://host:port?tls=true&tls_no_verify=false`

The supported query params are:

* timeout - the number of seconds to set the connection timeout to. Defaults to 60 seconds.
* tls - Set to 'true' or '1' if you want to use TLS encryption
* tls_no_verify - Set to 'true' or '1' if you want to accept any server certificate (for testing, not secure)
* tls_ca_cert_file - path to a custom ca cert for a self-signed TLS cert
* tls_cert_file - path to a cert file for this client (need to verify this is processed by Neo4j)
* tls_key_file - path to a key file for this client (need to verify this is processed by Neo4j)

Errors returned from the API support wrapping, so if you receive an error
from the library, it might be wrapping other errors.  You can get the innermost
error by using the `InnerMost` method.  Failure messages from Neo4J are reported,
along with their metadata, as an error.  In order to get the failure message metadata
from a wrapped error, you can do so by calling
`err.(*errors.Error).InnerMost().(messages.FailureMessage).Metadata`

If there is an error with the database connection, you should get a sql/internalDriver ErrBadConn
as per the best practice recommendations of the Golang SQL Driver. However, this error
may be wrapped, so you might have to call `InnerMost` to get it, as specified above.
*/
package goBolt
