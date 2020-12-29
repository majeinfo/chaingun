#!/bin/sh
#
PLAYER=../player/bin/player 
#VERBOSE=--verbose
ERRORS=0

Arg_Error() {
	if $PLAYER $1 2>&1 | tee $$.out | grep "$2" >/dev/null 2>&1; then
		echo '[OK]' $1
	else
		echo '[FAILED]' $1
		ERRORS=`expr $ERRORS + 1`
		cat $$.out
	fi
}

Syn_Error() {
	if $PLAYER --output-dir output/ $VERBOSE --syntax-check-only --script "$1" 2>&1 | tee $$.out | grep "$2" >/dev/null 2>&1; then
		echo '[OK]' $1
	else
		echo '[FAILED]' $1
		ERRORS=`expr $ERRORS + 1`
		cat $$.out
	fi
}

Syn_OK() {
	if $PLAYER --output-dir output/ $VERBOSE --syntax-check-only --script "$1" >$$.out 2>&1; then
		echo '[OK]' $1
	else
		echo '[FAILED]' $1
		ERRORS=`expr $ERRORS + 1`
		cat $$.out
	fi
}

Req_Error() {
	if $PLAYER --output-dir output/ $VERBOSE --script "$1" 2>&1 | tee $$.out | grep "$2" >/dev/null 2>&1; then
		echo '[OK]' $1
	else
		echo '[FAILED]' $1
		ERRORS=`expr $ERRORS + 1`
		cat $$.out
	fi
}

Req_OK() {
	if $PLAYER --output-dir output/ $VERBOSE --script "$1" >$$.out 2>&1; then
		echo '[OK]' $1
	else
		echo '[FAILED]' $1
		ERRORS=`expr $ERRORS + 1`
		cat $$.out
	fi
}

# Test player arguments
Arg_Error "--output-dir /tmp/chaingun/output/" "When started in standalone mode, needs a script filename (option --script)"
Arg_Error "--output-dir /tmp/chaingun/output/ --script dummy" "no such file or directory"

# Test script syntax
Syn_Error syntax/missing-iterations.yml 'Iterations not set, must be > 0'
Syn_Error syntax/missing-duration.yml 'When Iterations is -1, Duration must be set'
Syn_Error syntax/missing-title.yml 'HttpAction must define a title'
Syn_Error syntax/missing-method.yml 'Action has no Method and no default Method specified'
Syn_Error syntax/bad-method.yml 'HttpAction must specify a valid HTTP method: GET, POST, PUT, HEAD or DELETE: GIT'
Syn_Error syntax/dflt-bad-method.yml 'Default Http Action must specify a valid HTTP method: GET, POST, PUT, HEAD or DELETE: GIT'
Syn_Error syntax/missing-server.yml 'Host missing for URL'
Syn_Error syntax/setvar3.yml 'Undefined function strlenght'
Syn_Error syntax/setvar4.yml 'Unexpected end of expression'
Syn_Error syntax/mongo-findone1.yml 'findone command must define a filter'
Syn_Error syntax/mongo-insertone1.yml 'insertone command must define a document'
Syn_Error syntax/mongo-database-missing.yml 'no Database and no default Database specified'
Syn_Error syntax/mongo-collection-missing.yml 'no Collection and no default Collection specified'
Syn_Error syntax/mongo-bad-command.yml 'must specify a valid command'
Syn_Error syntax/mongo-server-missing.yml 'no Server and no default Server specified'
Syn_Error syntax/sql-db-driver-missing.yml 'no Driver and no default Driver specified'
Syn_Error syntax/sql-bad-db-driver.yml 'DB Driver must specify a valid driver (mysql)'
Syn_Error syntax/sql-server-missing.yml 'no Server and no default Server specified'
Syn_Error syntax/sql-database-missing.yml 'no Database and no default Database specified'
Syn_Error syntax/sql-statement-missing.yml 'no Statement specified'
Syn_OK syntax/opt-duration.yml
Syn_OK syntax/dflt-values.yml
Syn_OK syntax/setvar1.yml 
Syn_OK syntax/setvar2.yml 
Syn_OK syntax/mongo-insert-ok.yml
Syn_OK syntax/sql-insert-ok.yml

# Test JSON request
Req_Error requests/1VU-json-bad.yml 'failed to apply - no default value given'
Req_Error requests/1VU-json-default-value.yml 'Jsonpath failed to apply, uses default value'
Req_OK requests/1VU-json.yml
Req_OK requests/1VU-json-many.yml

# Test Regex request
Req_Error requests/1VU-regex-bad.yml 'failed to apply - no default value given'
Req_Error requests/1VU-regex-default-value.yml 'Regexp failed to apply, uses default value'
Req_OK requests/1VU-regex.yml
Req_OK requests/1VU-regex-random.yml

# Test "on_error" behaviour
Req_Error requests/error-stop-test.yml 'Stop now'
Req_Error requests/error-stop-vu.yml 'Stop VU on error'
Req_Error requests/error-continue.yml 'Continue on error'
Req_Error requests/1VU-empty-variable.yml 'Variable ${name} not set'

# Test Feeder behaviour
Req_Error requests/feeder-bad-file.yml 'Cannot open CSV file'
Req_Error requests/feeder-bad-type.yml 'Unsupported feeder type: xls'
Req_OK requests/2VU-csv.yml

# Test POST body/template
Req_OK requests/2VU-post-template.yml
Req_OK requests/2VU-extract-from-header.yml

# Test Timeout behaviour
Req_Error requests/1VU-http-timeout.yml 'HTTP request failed: net/http: timeout awaiting response headers'
# TODO: test timeout with ws

# Test GRPC
Req_Error syntax/grpc-bad-proto-file.yml 'open unknown.proto: no such file or directory'
Req_Error syntax/grpc-bad-function.yml 'method name must be package.Service.Method or package.Service/Method: "function"'

# Test a pre-action
Req_OK requests/pre_actions1.yml

rm -f $$.out 2>/dev/null
if [ $ERRORS -gt 0 ]; then
	exit 1
fi
# EOF
