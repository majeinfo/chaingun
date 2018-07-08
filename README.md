# chaingun
golang & Python-based load test application using YAML documents as specification.

The golang player (or injector) is base on Gotling project available here: 
http://callistaenterprise.se/blogg/teknik/2015/11/22/gotling/
(Thanks to Erik Lupander)

## What it does
- Provides high-throughput load testing of HTTP services
    - Supports GET, POST, PUT and DELETE
    - Request URLs and bodies can contain ${paramName} parameters
    - ${paramName} values can be extracted from HTTP response bodies and bound to a User context
    - Capturing Set-Cookie response headers
    - POST data can be inlined or read from template files

## Building

To be completed...

## Architecture

Chaingun is made of 2 parts :

- a Player which role is to inject requests to the tested server(s)
- a Manager that provides a Web interface to manage the Players

Players can be run in standalone mode : this is the easiest way to proceed and may be
sufficient when the expected test load can be ensured by only one Player. In such a case
the Manager is not needed.

If you need many Players to be coordinated to stress the same server in the same time,
you launch different Players (on different hosts !) in "daemon mode". Then you start the Web
interface of the Manager and you can drive the Players remotely. The results will be aggregated by
the Manager.

#### Run from the command line

a) run a Player in standalone mode :

	$ cd player/bin
	$ ./player --output-dir /path/to/output/ --python-cmd /path/to/python3.6 --script /path/to/script.yml --verbose

	--python-cmd is optional if PYTHON environment variable is set and points to at least a Python 3.6
	--output-dir indicates where the results will be stored
	--script is mandatory
	--verbose is optional 

b) run a Player in daemon mode :

	$ cd player/bin
	$ ./player --daemon --listen-addr 127.0.0.1:12345 --verbose

	in daemon mode, the player will listen to the TCP port specified by --listen-addr option
	(default is 127.0.0.1:12345) and will play the orders sent by the manager. This is the normal
	mode in distributed mode.
	--verbose is optional

c) run the Manager (when Players are started as Daemons) :

	You must have a valid Python 3.6+ or prepare a virtual environment for the Manager, then run it :

	$ cd manager/server
	$ python manage.py runserver 127.0.0.1:8000

	Then open your browser and manage your Players !

#### Run from container image

a) run a Player in standalone mode :

	$ docker container run -it -v /path/to/scripts:/scripts \
				   -v /path/to/output/dir:/output \
				   majeinfo/chaingun standalone script.yml

b) run a Player in daemon mode :

	$ docker container run -it majeinfo/chaingun daemon [<IP>:<Listen_Port>]

	- default IP is 0.0.0.0 
	- default port is 12345

c) run the Manager (when Players are started as Daemons) :

	$ docker container run -it -v /path/to/scripts:/scripts \
				   -v /path/to/output/dir:/output \
				   -v /path/to/data_and_graphs:/data \
				   majeinfo/chaingun manager [<IP>:<Listen_Port>]

	- default IP is 0.0.0.0 
	- default port is 8000

Then connect with a Web Browser to the specified port on localhost by default.

The verbose mode can be specified using the VERBOSE environment variable :

	-e VERBOSE=1

## YAML Script (=Playbook)

This is a sample script. 
Mandatory paremeters are marked with a "# MAND" pseudo-comment at the end of the line

```
---
iterations: 2		# MAND
duration: 100		# MAND if iterations == -1. Time is in seconds
rampup: 4		# MAND - time is in seconds
users: 2		# MAND - number of VU to launch during the rampup period
timeout: 10		# default value (in seconds)
feeder:			# Only one Feeder can be defined
  type: csv		# MAND - csv if the only supported type
  filename: data1.csv	# MAND - the first line gives the columns and so the variable names
  separator: ","	# MAND
actions:
  - http:
      title: Page 1			# MAND for http action
      method: GET			# MAND for http action (GET/POST/PUT/HEAD/DELETE)
      url: http://server/page1.php	# MAND for http action
      # name of Cookie to store. __all__ catches all cookies !
      storeCookie: __all__
  - sleep:
      duration: 5			# MAND - time is in seconds
  - http:
      title: Page 3
      method: GET
      url: http://server/page3.php?name=${name}
  - http:
      title: Page 4
      method: POST
      url: http://server/page4.php
      body: name=${name}&age=${age}	# MAND for POST http action
      accept: "text/html,application/json"
      contentType: text/html
      responses:				# OPT
        - regex: "is: (.*)<br>"		# MAND must be one of regex/jsonpath/xmlpath
          index: first			# MAND must be one of first/last/random
          variable: address		# MAND
          default_value: bob		# used when the regex failed
  - http:
      title: Page 5
      method: GET
      # ${address} is the value extracted from the previous response !
      url: http://server/page5.php?address=${address}
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
  - http:
      title: Page 5bis
      method: GET
      url: http://server/page5.php?address=${address}&city=${city}
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
```

The syntax for jsonpath is available at https://github.com/JumboInteractiveLimited/jsonpath.

## How to test

$ cd tests
$ docker container run -d -p 8000:80 -v `pwd`/server:/var/www/html php:5.6-apache
$ ./test_standalone_player.sh

## License
Licensed under the MIT license.

See LICENSE
