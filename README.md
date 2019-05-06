# chaingun
An efficient Load Testing Tool for HTTP/MQTT/WS Servers, written in Go Language.

## Table of Contents
1.[What it does](##what-it-does)

2.[Building](#building)
3.[Architecture](##architecture)
4.[How to run it](##how-to-run-it)
5.[Playbook Syntax](##playbook-yaml-syntax)
6.[How to test](##how-to-test)
7.[TODO](##todo)
8.[License](##license)

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

## Architecture

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

## How to run it

#### Run from the command line

a) run a Player in standalone mode :

	$ cd player/bin
	$ ./player --output-dir /path/to/output/ --script /path/to/script.yml

	--output-dir indicates where the results will be stored
	--script sets the name of the script file and is mandatory
	--verbose is optional 
	--no-log disables the 'log actions' (see below for the actions)

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

#### Run from container image

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

## Playbook YAML Syntax

This is a sample script. 
Mandatory parameters are marked with a "# MAND" pseudo-comment at the end of the line.
Please note that wrong parameter names are not detected by the YAML parser !

```
---
iterations: 2		# MAND
duration: 100		# MAND if iterations == -1. Time is in seconds
rampup: 4		# MAND - time is in seconds
users: 2		# MAND - number of VU to launch during the rampup period
timeout: 10		# default value (in seconds)
on_error: continue	# (default) or stop_iteration | stop_vu | stop_test
http_error_codes: 404,403,500	# if set, these HTTP response codes generates errors

default:
  server: www.google.com:80     # port number is optional
  protocol: http                # could be https
  method: GET

variables:		# You can define here variables that can be reused later
  customer: bob
  amount: 1000

feeder:			# Only one Feeder can be defined
  type: csv		# MAND - csv if the only supported type
  filename: data1.csv	# MAND - the first line gives the column names and so the variable names
  separator: ","	# MAND

actions:
  # A simple GET
  - http:
      title: Page 1			# MAND for http action
      method: GET			# MAND for http action (GET/POST/PUT/HEAD/DELETE)
      url: http://server/page1.php	# MAND for http action
      # name of Cookie to store. __all__ catches all cookies !
      storeCookie: __all__

  # Wait 
  - sleep:
      duration: 500			# MAND - time is in milli-seconds

  # GET with variable interpolation - the variable comes from the "feeder" file
  - http:
      title: Page 3
      method: GET
      url: http://server/page3.php?name=${name}

  # POST with application/x-www-form-urlencoded by default
  # Extracts value from response using regexp
  - http:
      title: Page 4
      method: POST
      url: http://server/page4.php              # variables are interpolated in URL
      body: name=${name}&age=${age}	# MAND for POST http action
      headers:
        accept: "text/html,application/json"    # variables are interpolated in Headers
        content-type: text/html
      responses:			# OPT
        - regex: "is: (.*)<br>"		# MAND must be one of regex/jsonpath/xmlpath
          index: first			# OPT must be one of first (default)/last/random
          variable: address		# MAND
          default_value: bob		# used when the regex failed
        - from_header: Via		# OPT HTTP Header name to extract the value from
          regex: "(.*)"			# MAND 
          index: first			# OPT must be one of first (default)/last/random
          variable: proxy_via		# MAND
          default_value: -		# used when the regex failed

  # Simple log... (the customer is defined in the global variables section)
  - log:
      message: Address value is ${address} (customer=${customer})

  # The HTTP_Response variable is always set after a HTTP action
  - log:
      message: HTTP return code=${HTTP_Response}

  # GET with variable interpolation - the variable comes from previous POST response
  - http:
      title: Page 5
      method: GET
      # ${address} is the value extracted from the previous response !
      url: http://server/page5.php?address=${address}

  # POST with variable interpolation in the request
  # Extracts value from response using regexps
  - http:
      title: Page 4bis
      method: POST
      url: http://server/page4.php
      body: name=${name}&age=${age}
      responses:
        - regex: "is: (.*), .*<br>"
          index: first
          variable: address
        - regex: "(?i)is: .*, (.*)<br>"
          index: first
          variable: city

  # Variable interpolation is possible in the URL
  - http:
      title: Page 5bis
      method: GET
      url: http://server/page5.php?address=${address}&city=${city}

  # POST with extraction from response using JSON    
  - http:
      title: Page 6
      method: POST
      url: http://server/page4.php
      body: name=${name}&age=${age}
      responses:
        - jsonpath: $.name+
          index: first
          variable: name
          default_value: bob

  # POST with content specified using a template file       
  - http:
      title: Page 7
      method: POST
      url: /demo/form.php
      template: tpl/mytemplate.tpl	# POST needs body or template
					# template refers to a file which contents
					# will be used as the request body. Variables
					# are interpolated in the file contents.

  # File upload
  - http:
      title: Page 8
      method: POST
      url: http://server/page4.php
      body: name=${name}&age=${age}     # Optional
      upload_file: /path/to/file        # no variable interpolation                    

  # MQTT action is possible (beta)
  - mqtt:
      title: Temperature		# MAND
      url: tcps://endpoint.iot.eu-west-1.amazonaws.com:8883/mqtt	# MAND
      certificatepath: path/to/cert	# OPT needed if auth by certificate
      privatekeypath: path/to/privkey	# OPT needed if auth by certificate
      clientid: basicPubSub		# OPT "chaingun-by-JD" by default
      topic: "sensors/room1"		# MAND
      payload: "{ \"Temp\": \"20°C\" }"	# MAND format depends on your app
      qos: 1				# OPT values can be 0, 1 (defult) or 2
					# Variable interpolation is applied on
					# url, payload and topic

  # Compute formula with variables
  - setvar:
      name: my_var
      expression: "2 * age"

      # notes on expressions:
      # variable interpolation is possible, supported returned types are
      # int, string and bool (floats are converted into ints)
      # supported operators are described here:
      #   https://github.com/Knetic/govaluate/blob/master/MANUAL.md
      # supported functions are:
      # - strlen(string)
      # - substr(string, start, end)

  # Assertion are possible and use the same syntax as "setvar"
  - assert:
      expression: "name == \"bob\""

      # if the assertion fails, the action returns an error

  # Each action can be conditioned by a "when" clause that must be true to trigger the action
  - log:
      message: "something..."
      when: "var1 > 0"
```

The syntax for jsonpath is available at https://github.com/JumboInteractiveLimited/jsonpath.

## How to test

```
$ cd tests
$ docker container run -d -p 8000:80 -v `pwd`/server:/var/www/html php:5.6-apache
$ ./test_standalone_player.sh
```

## TODO
- add a web interface to create/import/export Playbooks
- implements the "connect-to" option to reverse the roles and cross through the firewalls
- add options to handle SSL certificates ?

## License
Licensed under the MIT license.

The golang player (or injector) is originally based on Gotling project available here: 
http://callistaenterprise.se/blogg/teknik/2015/11/22/gotling/
(Thanks to Erik Lupander)

See LICENSE
