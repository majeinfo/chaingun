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

#### Run fro the command line

To be completed...

#### Run from container image

a) run a Player in standalone mode :

`$ docker container run -it -v /path/to/scripts:/scripts \\
			-v /path/to/output/dir:/output \\
			majeinfo/chaingun standalone script.yml`

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

## License
Licensed under the MIT license.

See LICENSE
