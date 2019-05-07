# Playbook  Syntax

The test script are described using the following YAML syntax.
(please note, that wrong parameter names ARE NOT detected by the YAML parser)

First of all, you define some global parameters (the count of iterations, the number of virtual users (VU)
to inject, etc...). Then you can define some default value for common parameters, you can also add your own
variables. At last, you define the list of actions to be performed by `chaingun`.

## Global Parameters

| Name | Value | Description |
| :--- | :---: | :---        | 
| `iterations` | integer | (mandatory) indicates how many times each VU must play the script. If value is -1, the script is played until the value of `duration` parameter is reached |
| `duration`   | integer | (mandatory if `iteration` equals -1) gives the duration of the script playing in seconds |
| `rampup`     | integer | (mandatory) gives the time in seconds that is use to launch the VU. New VUs are equally launched during this period. |
| `users`      | integer | (mandatory) number of VUs to launch during the `rampup` period. For example, if `users` value equals 100 and `rampup` equals 20, 5 new VUs will be launched every new seconds (because 20*5 = 100) |
| `timeout`    | integer | (default=10) number of seconds before a network timeout occurs |
| `on_error`   | string  | (default=continue,stop_iteration,stop_vu,stop_test) define the behaviour for error handling: just display the error and continue (default), or abort the current iteration, or stop the current VU, or abort the whole test |
| `http_error_code` | list | (no default value) define the list of what is considered an HTTP error code. For example, `http_error_code: 404,403,500`. This is only used by HTTP Actions |


## Variables

Actions can define expressions that may contain variables. Some variables are created by `chaingun` but you can define and use your own variables.
You define your custom variables like this:

```
variables:
  variable_name: value
  ...
```

## Default value for Actions

Default values for some parameters of further Actions can be defined like this:

```
default:
  parameter_name: value
  ...
```

The supported parameter_name(s) are:

| Name | Description | Example values |
| :--- | :---:       | :--- |
| `server`   | name of remoter server - may also specify a port | www.google.com:80 or www.bing.com |
| `protocol` | protocol to be used | http or https |
| `method`   [ HTTP method to use | GET or POST |



## Actions


## Advanced Topics

### Variables usage

### Expressions

### How to import data from outside


## Full sample

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
