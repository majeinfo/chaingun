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

If you need many Players to be coordinated to stress the same server(s) in the same time,
you launch different Players (on different hosts !) in "daemon mode". Then you start the Web
interface of the Manager and you can drive the Players remotely. The results will be aggregated by
the Manager.

Data for feeder can be sent to the Players after sending them the Playbook script.
Other fils such as Template of files to be uploaded must be sent to the Players before the Playbook script.

#### Run from the command line

a) run a Player in standalone mode :

	$ cd player/bin
	$ ./player --output-dir /path/to/output/ --python-cmd /path/to/python3.6 --script /path/to/script.yml \
               --viewer /path/to/viewer.py --verbose

	--python-cmd is optional if PYTHON environment variable is set and points to at least a Python 3.6
	--viewer indicates the path to the viewer.py script that build the HTML page with results
	--output-dir indicates where the results will be stored
	--script is mandatory
	--verbose is optional 
    --no-log disabled the 'log actions' (see below for the actions)

b) run a Player in daemon mode :

	$ cd player/bin
	$ ./player --daemon --listen-addr 127.0.0.1:12345 --verbose

	in daemon mode, the player will listen to the TCP port specified by --listen-addr option
	(default is 127.0.0.1:12345) and will play the orders sent by the manager. This is the normal
	mode in distributed mode.
	--verbose is optional
    --no-log disabled the 'log actions' (see below for the actions)

c) run the Manager (when Players are started as Daemons) :

	You must have a valid Python 3.6+ or prepare a virtual environment for the Manager, then run it :

	$ cd manager/server
	$ python manage.py runserver 127.0.0.1:8000

	Then open your browser and manage your Players !

#### Run from container image

a) run a Player in standalone mode :

	$ docker container run -it -v /path/to/scripts:/scripts \
				   -v /path/to/output/dir:/output \
				   majeinfo/chaingun standalone /scripts/script.yml

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

## YAML Script (Playbook)

This is a sample script. 
Mandatory parameters are marked with a "# MAND" pseudo-comment at the end of the line

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
      responses:				# OPT
        - regex: "is: (.*)<br>"		# MAND must be one of regex/jsonpath/xmlpath
          index: first			# OPT must be one of first (default)/last/random
          variable: address		# MAND
          default_value: bob		# used when the regex failed

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
```

The syntax for jsonpath is available at https://github.com/JumboInteractiveLimited/jsonpath.

## How to test

```
$ cd tests
$ docker container run -d -p 8000:80 -v `pwd`/server:/var/www/html php:5.6-apache
$ ./test_standalone_player.sh
```

## TODO
- add a way to define global variables
- add a way to add an action that can compute values for variabls (var = substr(var, 1, 5))
- add a way to extract HTTP Headers values from responses
- add a way to get the HTTP return code from responses
- sleep action should take its time in seconds or milliseconds
- add a web interface to create/import/export Playbooks

## License
Licensed under the MIT license.

See LICENSE
