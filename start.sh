#!/bin/sh
#
# $1: daemon|manager|standalone
#
# verbose is defined using VERBOSE environment variable
#
if [ "${VERBOSE}" != "" ]; then
	VERBOSE_MODE="--verbose"
fi

case "$1" in
daemon)
	LISTEN_ADDR=${2:-0.0.0.0:12345}
	cd player
	bin/player --mode daemon --listen-addr ${LISTEN_ADDR} ${VERBOSE_MODE}
;;

standalone)
	if [ "$2" = "" ]; then
		echo "The YML Playbook is missing"
		exit 1
	fi
	cd player
	bin/player --mode standalone --script "$2" --output-dir /tmp/output ${VERBOSE_MODE}
;;

manager)
	shift
	if [ "$1" = "" -o "$1" = "-" ]; then
		LISTEN_ADDR="0.0.0.0:8000"
		shift
	fi
	#LISTEN_ADDR=${1:-0.0.0.0:8000}
	cd player
	bin/player --mode manager --manager-listen-addr ${LISTEN_ADDR} ${VERBOSE_MODE} -v /data $*
;;

batch)
	if [ "$2" = "" ]; then
		echo "The YML Playbook is missing"
		exit 1
	fi
	if [ "$3" = "" ]; then
		echo "Injectors are mmissing"
		exit 1
	fi
	cd player
	bin/player --mode batch --script "$2" --injectors "$3"
;;

*)
	echo "Usage:"
	echo "$0 daemon [<IP>:<Port>] (default is 0.0.0.0:12345)"
	echo "$0 standalone <path_to_playbook.yml> (normally something like: /scripts/myscript.yml)"
	echo "$0 manager [<IP>:<Port>] (default is 0.0.0.0:8000)"
	echo "$0 batch <path_to_playbook.yml> <injector_list>"
	exit 1
esac

# EOF
