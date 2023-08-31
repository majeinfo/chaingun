# chaingun
An efficient Load Testing Tool for HTTP/MQTT/WS/MongoDB/MySQL/gRPC/TCP/UDP Servers, written in Go Language.
(The official site is https://chaigun.io)

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
- Provides high-throughput load testing of HTTP/HTTPS/TCP/UDP/WS/MQTT
- Provides limited load testing for MongoDB, MySQL and PostgreSQL database servers
- Provides load testing for gRPC protocol (beta)
- Provides load testing for Kafka service (beta)
- Provides load testing for Elasticsearch service (beta)
- Supports standalone or distributed modes. The distributed mode can be used to play different tests at the same time or to inject stress load from remote injectors
- Supports GET, HEAD, POST, PUT and DELETE HTTP methods
- Supports HTTP/2
- Supports HTTP Basic Authentication
- HTTP Requests and bodies can contain parameters 
- Parameter values can be extracted from HTTP response bodies and bound to a User context. User defined variables are also supported
- Captures Set-Cookie HTTP response headers
- POST data can be inlined or read from template files
- Variables can be fed from an external CSV file
- Embeds a Web server to manage remote injectors but also supports a "batch mode"
- Uses a YAML syntax to describe the stress scenarii
- Embeds a Web Designer to help build the YAML scripts !
- May be run in "proxy mode" to help you create the YAML scripts !

# Building

Requires golang 1.16+ (because the "embed" module is needed).

	$ git clone https://github.com/majeinfo/chaingun
	$ cd chaingun
	$ export GOPATH=`pwd`/player
	$ export GO111MODULE=auto
	$ cd player/src
	$ CGO_ENABLED=0 go install ./player
	$ cd ..
	$ bin/player -h

# Architecture

Chaingun is made of a single binary (named "player") that can serve multi purpose.

The "player" can be started in 7 different ways:

- the standalone mode (which is the default mode): this is the easiest way to proceed and may be
sufficient when the expected test load can be applied by only one Player

- the "ab" mode, so called because it tries to mimick the "ab" command from Apache ! This mode 
can be used to make a quick test...

- the daemon mode: if you need many Players to be coordinated to stress the same server(s) at the same time,
you launch different Players (on different hosts !) in "daemon mode"

- the manager mode: the Player creates a Web interface that lets you manage other remote Players. 
The results will be aggregated by the Web interface.

- the batch mode: like the Manager mode, but you provide the list of the remote Injectors and a script to play.
Everyting is executed from the command line in "batch mode" !

- the designer mode: the "player" offers a Web interface which helps you to create the YAML file !

- the proxy mode: acts as a Web proxy that can intercept your requests and create Playbook skeleton

Note for the daemon mode:
	- Data for feeder can be sent to the Players after sending them the Playbook script.
	- Other files such as Template of files to be uploaded must be sent to the Players before the Playbook script.

# How to run it

### Run from the command line

a) run a Player in standalone mode :

	$ player inject --output-dir /path/to/output/ --script /path/to/script.yml

	--output-dir indicates where the results will be stored
	--script sets the name of the script file and is mandatory
	--verbose is optional 
	--no-log disables the 'log actions' (see below for the actions)
	--trace generates a trace file named traced.out that can be used by 'go tool trace' command
	--display-response displays the full response sent by the remote stressed server
	--syntax-check-only is used to only check the syntax of the script
	--disable-dns-cache can be used to disable the internal DNS cache that reduces the number of DNS Requests
	--trace-requests displays all the HTTP/S requests and their return code
	--store-srv-response-dir indicates where the responses from the servers (mainly HTML files ?) must be stored

b) run a Player in daemon mode :

	$ player daemon --listen-addr 127.0.0.1:12345 

	in daemon mode, the player will listen to the TCP port specified by --listen-addr option
	(default is 127.0.0.1:12345) and will play the orders sent by the manager. This is the normal
	mode in distributed mode.

	--verbose is optional
	--no-log disables the 'log actions' (see below for the actions)
	--disable-dns-cache can be used to disable the internal DNS cache that reduces the number of DNS Requests
	--trace-requests displays all the HTTP/S requests and their return code

c) run the Manager (when Players are started as Daemons) :

	$ player manage --listen-addr 127.0.0.1:8000 --repository-dir ./ 

	in manager mode, the player will listen to the TCP port specified by the --listen-addr option
	(default is 127.0.0.1:8000) and will offer a Web interface that manages the remote players.
	The --repository-dir option gives the location of the results (default is "."). This directory
	must be relative to the location where you launched the manager from.

	--verbose is optional
	--injectors injector1:port1,injector2:port2,... gives the list of already started injectors. In that case,
		the Web Interface will try to automatically add these injectors and connect to them. This is handy
		for batch mode.

	Then open your browser and manage your Players !

d) run in Batch mode (need remote injectors) :

	$ player batch --injectors server1:port1,server2:port2,.... --script /path/to/script.yml

	The local player tries to connect to the remote injectors. Then it sends them the script and the related
	files (feed data, template files) and makes the injectors run in parallel. All the filenames (script and related files)
	must be given with relative names.

	--verbose is optional
	--injectors injector1:port1,injector2:port2,... gives the list of already started injectors. In that case,
		the Web Interface will try to automatically add these injectors and connect to them. This is handy
		for batch mode.
	--repository-dir gives the location of results (default ".")
	--store-srv-response-dir indicates where the responses from the servers (mainly HTML files ?) must be stored

e) run in Designer mode (the Web Interface for creating YAML files) :

	$ player design --listen-addr 127.0.0.1:12345

	The you can browse to the specified address ans starts creating your Playbook...

f) run in Proxy mode :

	$ player proxy --listen-addr 127.0.0.1:12345 --proxy-domain example.com

	If you plan to use HTTPS instead of HTTP, make sure the CA of the certificate used by the Proxy is installed
	in your Browser store. One the Proxy os activated, configure your Browser and navigate to the site that must
	be load-tested.
	In order to display the Playbook, type <Ctrl-C> in the terminal where you launched the Proxy. A prompt will ask you
	if you want to exit(e), to reset(r) the capture or to display the playbook(p).

	--listen-addr <IP>:<port> (default is 127.0.0.1:12345) gives the IP and Port the Proxy must be listening to
		(you have to configure your browser with these values)
	--proxy-domain <domain> is mandatory and sets the domain name which requests must be captured to build the Playbook
	--proxy-ignore-suffixes <suffix,suffix...> gives a list of suffixes that must be ignored
		(default value is: .gif,.png,.jpg,.jpeg,.css,.js,.ico,.ttf,.woff,.pdf)
	--verbose is optional

g) run in "ab" mode :

	$ player ab --request http://mysite.com --user 100 --rampup 10 --duration 10

	This mode does not need any Playbook but it can only inject one request in the remote server.
	The option are similar to the "standalone" mode, plus some options to describe the request and
	the load to inject.

	--body string                     Request body for POST requests
	--disable-dns-cache               Disable the embedded DNS cache used to reduce the number of DNS requests
	--display-response                Used with verbose mode to display the Server Responses
	--duration int                    Total duration (in seconds) of the stress - mandatory if 'iterations' is set to -1
	--iterations int                  Count of iterations for each VU (default value is -1 which means aonly the 'duration' parameter value is used (default -1)
	--listen-addr string              Address and port to listen to (ex: 127.0.0.1:8080) (default "127.0.0.1:12345")
	--method string                   HTTP method to use (GET=default, POST, PUT, HEAD) (default "GET")
	--no-log                          Disable the 'log' actions from the Script
	--output-dir string               Set the output directory - where to put the data.csv file and the results (default "./results")
	--rampup int                      (mandatory) Gives the time in seconds that is use to launch the VU. New VUs are equally launched during this period
	--request string                  URL of the request to be player
	--store-srv-response-dir string   Set the directory where to store the whole server response (often HTML)
	--trace                           Generate a trace.out file useable by 'go tool trace' command
	--trace-requests                  Displays the requests and their return code
	--users int                       (mandatory) Count of VU to simulate

h) run in "graph" mode :

        $ player graph --output-dir /path/to/output/ --script /path/to/script.yml --output-type json

        This mode is used to regenerate the results from a previous tests. For example, if you fix
        or make some modification in the already computed CSV files, you can re-generate the graphs.
        If you specify the "--output-type json" option, a JSON document containing all the results
	will be displayed on STDOUT from previously obtained CSV file. This is a convenient way
	to insert the player in a CI/CD pipeline !

        --output-dir indicates where the results will be stored
        --script sets the name of the script file and is mandatory
        --output-type csv|json
        --verbose is optional

### Run from container image

a) run a Player in standalone mode :

	$ docker container run -it -v /path/to/scripts:/scripts \
				   -v /path/to/output/dir:/output \
				   majetraining/chaingun inject /scripts/script.yml

b) run a Player in daemon mode :

	$ docker container run -it -d majetraining/chaingun daemon [<IP>:<Listen_Port>]

	- default IP is 0.0.0.0 
	- default port is 12345

c) run the Manager (when Players are started as Daemons) :

	$ docker container run -it -d -v /path/to/scripts:/scripts \
				      -v /path/to/output/dir:/output \
				      -v /path/to/data_and_graphs:/data \
				      majetraining/chaingun manage [<IP>:<Listen_Port>]

	- default IP is 0.0.0.0 
	- default port is 8000

Then connect with a Web Browser to the specified port on localhost by default.

d) run in batch mode :

	$ docker container run -it -d -v /path/to/scripts:/scripts \
				      -v /path/to/output/dir:/output \
				      -v /path/to/data_and_graphs:/data \
				      majetraining/chaingun batch /path/to/script.yml injector_list

In all cases, the verbose mode can be specified using the VERBOSE environment variable :

	-e VERBOSE=1


# How to test

```
$ cd tests
$ docker container run -d -p 8000:80 -v `pwd`/server:/var/www/html php:5.6-apache
$ ./test_standalone_player.sh
```

# TODO

# License
Licensed under the MIT license.

The golang player (or injector) is originally based on Gotling project available here: 
http://callistaenterprise.se/blogg/teknik/2015/11/22/gotling/
(Thanks to Erik Lupander)

See LICENSE
