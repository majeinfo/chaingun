# chaingun
An efficient Load Testing Tool for HTTP/MQTT/WS Servers, written in Go Language.

## Table of Contents
1.[What it does](#what-it-does)

2.[Building](#building)

3.[Architecture](#architecture)

4.[How to run it](#how-to-run-it)

5.[Playbook Syntax](SYNTAX.md)

6.[How to test](#how-to-test)

7.[TODO](#todo)

8.[License](#license)

# What it does
- Provides high-throughput load testing of HTTP/TCP/UDP/WS/MQTT services
- Supports standalone or distributed modes
- Supports GET, HEAD, POST, PUT and DELETE HTTP methods
- Requests and bodies can contain parameters 
- Parameter values can be extracted from HTTP response bodies and bound to a User context. User defined variables are also supported
- Captures Set-Cookie HTTP response headers
- POST data can be inlined or read from template files
- Variables can be fed from an external CSV file
- The distributed mode can be used to play different tests in the same time or to inject stress load from remote injectors
- Use a YAML syntax to describe the stress scenarii

# Building

	$ git clone https://github.com/majeinfo/chaingun
	$ cd chaingun
	$ export GOPATH=`pwd`/player
	$ go get ./...
	$ cd player/src
	$ ../bin/statik -f -src=../../manager/go_web
	$ go install github.com/majeinfo/chaingun/player
	$ player/bin/player -h

# Architecture

Chaingun is made of a single binary (named "player") that can serve multi purpose.

The "player" can be started in 3 different ways:

- the standalone mode (which is the default mode): this is the easiest way to proceed and may be
sufficient when the expected test load can be applied by only one Player

- the daemon mode: if you need many Players to be coordinated to stress the same server(s) at the same time,
you launch different Players (on different hosts !) in "daemon mode"

- the manager mode: the Player creates a Web interface that lets you manage other remote Players. 
The results will be aggregated by the Web interface.

Note for the daemon mode:
	- Data for feeder can be sent to the Players after sending them the Playbook script.
	- Other files such as Template of files to be uploaded must be sent to the Players before the Playbook script.

# How to run it

### Run from the command line

a) run a Player in standalone mode :

	$ cd player/bin
	$ ./player --output-dir /path/to/output/ --script /path/to/script.yml

	--output-dir indicates where the results will be stored
	--script sets the name of the script file and is mandatory
	--verbose is optional 
	--no-log disables the 'log actions' (see below for the actions)
        --trace generates a trace file named traced.out that can be used by 'go tool trace' command
        --display-response displays the full response sent by the remote stressed server

b) run a Player in daemon mode :

	$ cd player/bin
	$ ./player --mode daemon --listen-addr 127.0.0.1:12345 

	in daemon mode, the player will listen to the TCP port specified by --listen-addr option
	(default is 127.0.0.1:12345) and will play the orders sent by the manager. This is the normal
	mode in distributed mode.

	--verbose is optional
	--no-log disables the 'log actions' (see below for the actions)

c) run the Manager (when Players are started as Daemons) :

	$ cd player/bin
	$ ./player --mode manager --manager-listen-addr 127.0.0.1:8000 --repository-dir /tmp/chaingun

	in manager mode, the player will listen to the TCP port specified by --manager-listen-addr option
	(default is 127.0.0.1:8000) and will offer a Web interface that manages the remote players.
	The --repository-dir option gives the location of the results (default is ".")

	--verbose is optional

	Then open your browser and manage your Players !

### Run from container image

a) run a Player in standalone mode :

	$ docker container run -it -v /path/to/scripts:/scripts \
				   -v /path/to/output/dir:/output \
				   majetraining/chaingun standalone /scripts/script.yml

b) run a Player in daemon mode :

	$ docker container run -it -d majetraining/chaingun daemon [<IP>:<Listen_Port>]

	- default IP is 0.0.0.0 
	- default port is 12345

c) run the Manager (when Players are started as Daemons) :

	$ docker container run -it -d -v /path/to/scripts:/scripts \
				      -v /path/to/output/dir:/output \
				      -v /path/to/data_and_graphs:/data \
				      majetraining/chaingun manager [<IP>:<Listen_Port>]

	- default IP is 0.0.0.0 
	- default port is 8000

Then connect with a Web Browser to the specified port on localhost by default.

The verbose mode can be specified using the VERBOSE environment variable :

	-e VERBOSE=1


# How to test

```
$ cd tests
$ docker container run -d -p 8000:80 -v `pwd`/server:/var/www/html php:5.6-apache
$ ./test_standalone_player.sh
```

# TODO
- add a web interface to create/import/export Playbooks
- implements the "connect-to" option to reverse the roles and cross through the firewalls
- add options to handle SSL certificates ?

# License
Licensed under the MIT license.

The golang player (or injector) is originally based on Gotling project available here: 
http://callistaenterprise.se/blogg/teknik/2015/11/22/gotling/
(Thanks to Erik Lupander)

See LICENSE
