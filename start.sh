#!/bin/sh
#
# $1: daemon|manager|standalone|batch|proxy
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
	bin/player --mode manager --manager-listen-addr ${LISTEN_ADDR} ${VERBOSE_MODE} $*
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

proxy)
	if [ "$2" = "" ]; then
		echo "The Proxied Domain Name is missing"
		exit 1
	fi
	LISTEN_ADDR=${3:-127.0.0.1:12345}
	cd player
	bin/player --mode proxy --listen-addr ${LISTEN_ADDR} ${VERBOSE_MODE}
;;

*)
	echo "Usage:"
	echo "$0 daemon [<IP>:<Port>] (default is 0.0.0.0:12345)"
	echo "$0 standalone <path_to_playbook.yml> (normally something like: /scripts/myscript.yml)"
	echo "$0 manager [<IP>:<Port>] (default is 0.0.0.0:8000)"
	echo "$0 batch <path_to_playbook.yml> <injector_list>"
	echo "$0 proxy proxied_domain [<IP>:<Port>] (default is 0.0.0.0:12345)"
	exit 1
esac

# EOF
