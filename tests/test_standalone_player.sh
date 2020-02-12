#!/bin/sh
#
PLAYER=../player/bin/player
#VERBOSE=--verbose
ERRORS=0

pwd
ls -laR /builds
test -f /build/majeinfo/chaingun/player/bin/player

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
	if $PLAYER --output-dir output/ $VERBOSE --script "$1" 2>&1 | tee $$.out | grep "$2" >/dev/null 2>&1; then
		echo '[OK]' $1
	else
		echo '[FAILED]' $1
		ERRORS=`expr $ERRORS + 1`
		cat $$.out
	fi
}

Syn_OK() {
	if $PLAYER --output-dir output/ $VERBOSE --script "$1" >$$.out 2>&1; then
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
Syn_Error syntax/missing-server.yml 'Host missing for URL'
Syn_Error syntax/setvar3.yml 'Undefined function strlenght'
Syn_Error syntax/setvar4.yml 'Unexpected end of expression'
Syn_OK syntax/opt-duration.yml
Syn_OK syntax/dflt-values.yml
Syn_OK syntax/setvar1.yml 
Syn_OK syntax/setvar2.yml 

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

rm -f $$.out 2>/dev/null
if [ $ERRORS -gt 0 ]; then
	exit 1
fi
# EOF
