#!/bin/bash
#
PLAYER=../player/bin/player
PYTHON=/usr/local/bin/python3.6
#VERBOSE=--verbose

function Arg_Error {
	if "$PLAYER" $1 2>&1 | grep "$2" >/dev/null 2>&1; then
		echo '[OK]' $1
	else
		echo '[FAILED]' $1
	fi
}

function Syn_Error {
	if "$PLAYER" --output-dir output/ --python-cmd "$PYTHON" $VERBOSE --script "$1" 2>&1 | grep "$2" >/dev/null 2>&1; then
		echo '[OK]' $1
	else
		echo '[FAILED]' $1
	fi
}

function Syn_OK {
	if "$PLAYER" --output-dir output/ --python-cmd "$PYTHON" $VERBOSE --script "$1" >/dev/null 2>&1; then
		echo '[OK]' $1
	else
		echo '[FAILED]' $1
	fi
}

function Req_Error {
	if "$PLAYER" --output-dir output/ --python-cmd "$PYTHON" $VERBOSE --script "$1" 2>&1 | grep "$2" >/dev/null 2>&1; then
		echo '[OK]' $1
	else
		echo '[FAILED]' $1
	fi
}

function Req_OK {
	if "$PLAYER" --output-dir output/ --python-cmd "$PYTHON" $VERBOSE --script "$1" >/dev/null 2>&1; then
		echo '[OK]' $1
	else
		echo '[FAILED]' $1
	fi
}
# Test player arguments
Arg_Error "--output-dir /tmp/chaingun/output/ --python-cmd '$PYTHON'" "When not started as a daemon, needs a 'script' file"
Arg_Error "--script dummy --python-cmd $PYTHON" "no such file or directory"
Arg_Error "--script dummy" "You must specify a Python interpreter"
Arg_Error "--script dummy --python-cmd dummy" "Python interpreter .* does not exist"

# Test script syntax
Syn_Error syntax/missing-iterations.yml 'Iterations not set, must be > 0'
Syn_Error syntax/missing-duration.yml 'When Iterations is -1, Duration must be set'
Syn_Error syntax/missing-method.yml 'Action has no Method and no default Method specified'
Syn_Error syntax/missing-server.yml 'Host missing for URL'
Syn_OK syntax/opt-duration.yml
Syn_OK syntax/dflt-values.yml

# Test JSON request
Req_Error requests/1VU-json-bad.yml 'failed to apply - no default value given'
Req_Error requests/1VU-json-default-value.yml 'Jsonpath failed to apply, uses default value'
Req_OK requests/1VU-json.yml

# Test Regex request
Req_Error requests/1VU-regex-bad.yml 'failed to apply - no default value given'
Req_Error requests/1VU-regex-default-value.yml 'Regexp failed to apply, uses default value'
Req_OK requests/1VU-regex.yml

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

# Test Timeout behaviour
Req_Error requests/1VU-http-timeout.yml 'HTTP request failed: net/http: timeout awaiting response headers'
# TODO: test timeout with ws

# EOF
