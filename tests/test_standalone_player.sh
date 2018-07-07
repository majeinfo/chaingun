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

# Test player arguments
Arg_Error "--output-dir /tmp/chaingun/output/ --python-cmd '$PYTHON'" "When not started as a daemon, needs a 'script' file"
Arg_Error "--script dummy --python-cmd $PYTHON" "no such file or directory"
Arg_Error "--script dummy" "You must specify a Python interpreter"
Arg_Error "--script dummy --python-cmd dummy" "Python interpreter .* does not exist"

# Test script syntax
Syn_Error syntax/missing-iterations.yml 'Iterations not set, must be > 0'
Syn_Error syntax/missing-duration.yml 'When Iterations is -1, Duration must be set'
Syn_OK syntax/opt-duration.yml

# EOF
