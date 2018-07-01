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
	bin/player --daemon --listen-addr ${LISTEN_ADDR} ${VERBOSE_MODE}
;;

standalone)
	if [ "$2" != "" ]; then
		echo "The YML Playbook is missing"
		exit 1
	fi
	cd player
	bin/player "$2" --output-dir /output --python-cmd /usr/bin/python3 ${VERBOSE_MODE}
;;

manager)
	cd manager/server
	python manage.py runserver "$2" 
;;

*)
	echo "Usage:"
	echo "$0 daemon [<IP>:<Port>] (default is 0.0.0.0:12345)"
	echo "$0 standalone <path_to_playbook.yml> (normally something like: /scripts/myscript.yml)"
	echo "$0 manager [<IP>:<Port>] (default is 0.0.0.0:8000)"
	exit 1
esac

# EOF
