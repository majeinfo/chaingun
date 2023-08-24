# Playbook  Syntax

The test script are described using the following YAML syntax.
(please note, that wrong parameter names ARE NOT detected by the YAML parser)

First of all, you define some global parameters (the count of iterations, the number of virtual users (VU)
to inject, etc...). Then you can define some default value for common parameters, you can also add your own
variables. At last, you define the list of actions to be performed by `chaingun`.

## Table of Contents
1.[Global Parameters](#global-parameters)

2.[Variables](#variables)

3.[Default values for Actions](#default-value-for-actions)

4.[Actions, Pre-actions and Post-actions](#actions-and-pre-actions)  
4.1.[HTTP/S](#http--https-request)  
4.2.[MongoDB](#mongodb--mongodb-request)  
4.3.[SQL](#sql--sql-request)  
4.4.[WebSocket](#ws--websocket-request)  
4.5.[MQTT](#mqtt--mqtt-request)  
4.6.[gRPC](#grpc--grpc-request-beta)  
4.7.[TCP & UDP](#tcp-or-udp--simple-tcp-or-udp-request)  
4.8.[setvar](#setvar--creates-and-set-variable-values)  
4.9.[sleep](#sleep--wait-action)  
4.10.[log](#log--log-output-action)  
4.11.[assert](#assert--creates-assertion)  
4.12.[timers](#timers--creates-page-timers)

5.[Advanced Topics](#advanced-topics)  
5.1.[Variables usage](#variables-usage)  
5.2.[Expressions](#expressions)  
5.3.[Session variables and Cookies](#session-variables-and-cookies)  
5.4.[The 'when' clause](#the-when-clause-to-trigger-actions)  
5.5.[Import external data](#how-to-import-external-data)  
5.6.[Submit form with multipart/form-data syntax](#submit-form-using-multipartform-data-syntax)  
5.7.[Handle Basic HTTP authentication](#handle-http-basic-authentication)

6.[Full Sample](#full-sample)

# Global Parameters

| Name | Value | Description |
| :--- | :---: | :---        | 
| `iterations` | integer | (mandatory) indicates how many times each VU must play the script. If value is -1, the script is played until the value of `duration` parameter is reached |
| `duration`   | integer | (mandatory if `iteration` equals -1) gives the duration of the script playing in seconds |
| `rampup`     | integer | (mandatory) gives the time in seconds that is use to launch the VU. New VUs are equally launched during this period. |
| `users`      | integer | (mandatory) number of VUs to launch during the `rampup` period. For example, if `users` value equals 100 and `rampup` equals 20, 5 new VUs will be launched every new seconds (because 20*5 = 100) |
| `timeout`    | integer | (default=10) number of seconds before a network timeout occurs |
| `on_error`   | string  | (default=continue,stop_iteration,stop_vu,stop_test) define the behaviour for error handling: just display the error and continue (default), or abort the current iteration, or stop the current VU, or abort the whole test |
| `http_error_codes` | list | (no default value) define the list of what is considered a HTTP error code. For example, `http_error_codes: 404,403,500`. This is only used by HTTP Actions |
| `persistent_http_sessions` | bool | (false) if true and if sessions are stored in Cookies, each VU uses the same session for all its iterations. |
| `persistent_db_connections` | bool | (false) if true, each VU uses the same connection for a script iteration. Note: only work for MongoDB and SQL. Also implies that the script uses only one protocol |
| `grpc_proto` | string | (no default value) if specified, must indicate the path to a ".proto" file. The path to the file is relative to the directory where the player is executed from. This option implies the definition of a default server |

Note : the injector does not support the notion of "keepalive". Connections to the remote servers are opened and closed for each action.

# Variables

Actions can define expressions that may contain variables. Some variables are created by `chaingun` but you can define and use your own variables.
You define your custom variables like this:

## Predefined Variables

| Parameter Name | Description                                                                                                                                                                     |
| :--- |:--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `UID` | integer value which represents the virtual user ID                                                                                                                              |
| `HTTP_Response` | contains the HTTP returned code                                                                                                                                                 |
| `MONGODB_Last_Insert_ID` | contains the value of the "_id" field of the last inserted document (string)                                                                                                    |
| `SQL_Row_Count` | contains the count of rows selected, updated or deleted (for SQL action)                                                                                                        |
| `__cookie__name` | if the option `store_cookie` has been set for `http` actions, Cookies returned by the server can be referenced as variable by prefixing their name with the `__cookie__` string |

## User defined Variables

They are defined in the `variables` section:

```
variables:
  variable_name: value
  other_variable:
    - value1
    - value2
  ...
```

As you can see, variables can be scalar or arrays. In case of arrays, the values are used on after another during the iterations of the same virtual user.

# Default value for Actions

Default values for some parameters of further Actions can be defined like this:

```
default:
  parameter_name: value
  ...
```

The supported parameter_name(s) are:

| Name | Description                                                                                                      | Example values |
| :--- |:-----------------------------------------------------------------------------------------------------------------| :--- |
| `server`   | name of remoter server - may also specify a port, for SQL this a DSN. Mandatory if grpc_proto has been specified | www.google.com:80 or www.bing.com or mongodb://localhost:27017 |
| `protocol` | protocol to be used                                                                                              | http or https |
| `method`   | HTTP method to use                                                                                               | GET, POST, PUT, HEAD or DELETE |
| `database` | default database for MongoDB and SQL                                                                             | my_database |
| `collection` | default collection for MongoDB                                                                                 | my_collection |
| `db_driver` | default SQL Driver - only "mysql" and "postgres" are supported yet                                              | mysql |


# Actions, Pre-Actions and Post-actions

Actions, pre-actions and post-actionsare defined as a list under the `actions`, `pre_actions` and `post_actions` keys :

```
pre_actions:
  - action1 ...
  - action2 ...

actions:
  - action1 ...
  - action2 ...

post_actions:
  - action1 ...
  - action2 ...
```

Pre-Actions are played only once before starting the VUs.
Actions are played by the VUs.
Post-Actions are played only once after the script completion.
A typical usage of pre-action would be to create a database table before injecting data.
A typical usage of post-action would be to clean a database table at the end of the test.
In "batch" mode, only the first injector given on the command line will play the pre-actions and the post-actions.
Pre-actions are also handled in "manager" mode (i.e using the embedded Web Interface).


Here is the list and the description of the implemented Actions :

## http : HTTP/S Request

| Parameter Name | Description |
| :--- | :--- |
| `title` | mandatory string that qualifies the request - used for the result output and logging |
| `method` | GET, PUT, POST, HEAD, DELETE. If absent use the value given by the `method` key in the default section |
| `use_http2` | set to true if you want to use HTTP/2 protocol (default value is false) |
| `url` | mandatory. If the string does not contain a server specification, use the value given by the `server` key in the default section |
| `store_cookie` | if set, indicates which cookies must be stored in the VU session. The predefined value \__all__ implies the capture of all possible cookies |
| `body` | value of HTTP body (works for any HTTP method) (one of `body` or `template` is mandatory) |
| `formdata` | list of key/value pairs used to send data in multipart form (see below) |
| `template` | a filename which contents will be interpolated and will be used as the request body (one of `body` or `template` is mandatory) |
| `upload_file` | when used with the POST or PUT methods, indicates a file which contents will be sent to the server as-is |
| `headers` | additional HTTP headers to transmit. Each header has the form `header_name: value`. In case of a POST method, the body is sent with the HTTP Header `content-type: application/x-www-form-urlencoded` | |
| `responses` | data can be extracted from server responses. The extraction can use the body or a HTTP Header. regex, jsonpath or xmlpath can be used to collect the substrings |


Examples:
```
- http:
    title: Page 1			# MAND for http action
    method: GET				# MAND for http action (GET/POST/PUT/HEAD/DELETE)
    url: http://server/page1.php	# MAND for http action
    # name of Cookie to store. __all__ catches all cookies !
    store_cookie: __all__

# POST with application/x-www-form-urlencoded by default
# Extracts value from response using regexp
- http:
    title: Page 4
    method: POST
    use_http2: true
    url: http://server/page4.php              # variables are interpolated in URL, but only
					      # for the URI part (not the server name)
    body: name=${name}&age=${age}	      # MAND for POST http action
    headers:
      accept: "text/html,application/json"    # variables are interpolated in Headers
      content-type: text/html
    responses:				# OPT
      - regex: "is: (.*)<br>"		# MAND must be one of regex/jsonpath/xmlpath
        index: first			# OPT must be one of first (default)/last/random
        variable: address		# MAND
        default_value: bob		# used when the regex failed (can be a string or an int)
      - from_header: Via		# OPT HTTP Header name to extract the value from
        regex: "(.*)"			# MAND 
        index: first			# OPT must be one of first (default)/last/random
        variable: proxy_via		# MAND
        default_value: -		# used when the regex failed

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

## mongodb : MongoDB Request

| Parameter Name | Description |
| :--- | :--- |
| `title` | mandatory string that qualifies the request - used for the result output and logging |
| `server` | mandatory. If the string does not contain a server specification, use the value given by the `server` key in the default section |
| `database` | mandatory. If the string is empty, use the value given by the `database` key in the default section |
| `collection` | mandatory. If the string is empty, use the value given by the `collection` key in the default section |
| `command` | mandatory. Possible commands are `findone`, `insertone` and `drop` |
| `filter` | If the command is `findone`, the `filter` parameter is a JSON document used to filter the search |
| `document` | If the command is `insertone`, the `document` parameter is a JSON document that must be inserted in the database collection |
| `responses` | data can be extracted from server responses when `findone` command is played. regex, jsonpath or xmlpath can be used to collect the substrings |

Examples:
```
- mongodb:
    title: Insert data
    server: mongodb://localhost:27017
    database: testing
    collection: person
    command: insertone
    document: '{"name": "${name}", "age": 30, "children" : [ {"name": "alice"}, {"name": "bob"} ]}'
- mongodb:
    title: Recherche
    server: mongodb://localhost:27017
    database: testing
    collection: person
    command: findone
    filter: '{"age": { "$eq": 30}}'
    responses:
      - jsonpath: $.name+
        variable: the_name
        index: first
        default_value: alice
- log:
    message: "found name is ${the_name}"
```

## sql : SQL Request

| Parameter Name | Description |
| :--- | :--- |
| `title` | mandatory string that qualifies the request - used for the result output and logging |
| `db_driver` | mandatory. If the string is empty, use the value given by the `db_driver` key in the default section |
| `server` | mandatory. If the string does not contain a server (DSN) specification, use the value given by the `server` key in the default section |
| `database` | mandatory. If the string is empty, use the value given by the `database` key in the default section |
| `statement` | mandatory. SQL Statement to execute (CREATE, SELECT, INSERT...) |

Examples:
```
# the server value is a connection string which depends on the SQL driver. Only mysql and postgres are supported
- sql:
    title: Clean Table
    # MySQL connection string
    server: "user:password@tcp(localhost:3306)"
    # PostgreSQL connection string
    # server: user:password@localhost:5432
    database: testing
    statament: 'DROP TABLE IF EXISTS my_table'
- sql:
    title: Create Table
    server: "user:password@tcp(localhost:3306)"
    database: testing
    statement: 'CREATE TABLE my_table (name CHAR(32), age INT)'
- sql:
    title: Select
    server: "user:password@tcp(localhost:3306)"
    database: testing
    statement: 'INSERT INTO my_table (name, age) VALUES("bob", 30)'
- log:
    message: "Row count=${SQL_Row_Count}"
```

Remember that `SQL_Row_Count` is a [predefined variable](#predefined-variables)...

## ws : WebSocket Request

| Parameter Name | Description |
| :--- | :--- |
| `title` | mandatory string that qualifies the request - used for the result output and logging |
| `url` | mandatory. If the string does not contain a server specification, use the value given by the `server` key in the default section |
| `store_cookie` | if set, indicates which cookies must be stored in the VU session. The predefined value \__all__ implies the capture of all possible cookies |
| `body` | value of body to send |
| `responses` | data can be extracted from server responses. The extraction can use the body or a HTTP Header. regex, jsonpath or xmlpath can be used to collect the substrings |

Examples:
```
- ws:
      title: Page 1
      url: wss://echo.websocket.org/echo
      body: hello
      responses:
        - regex: hello
```

## mqtt : MQTT Request

| Parameter Name | Description |
| :--- | :--- |
| `title` | mandatory string that qualifies the request - used for the result output and logging |
| `url` | mandatory |
| `certificatepath` | optional path to the certificate to offer to the server |
| `privatekeypath` | optional path to the private key to be used with the certificate to offer to the server |
| `username` | optional username |
| `password` | optional password |
| `clientid` | client name (chaingun-by-JD by default) |
| `topic` | mandatory MQTT topic |
| `payload` | mandatory MQTT paylaod, the format depends on the application |
| `qos` | MQTT wanted QoS (default=0, 1 or 2) |

Variable interpolation applies to url, payload and topic.

Example:
```
- mqtt:
    title: Temperature			# MAND
    url: tcps://endpoint.iot.eu-west-1.amazonaws.com:8883/mqtt	# MAND
    certificatepath: path/to/cert	# OPT needed if auth by certificate
    privatekeypath: path/to/privkey	# OPT needed if auth by certificate
    clientid: basicPubSub		# OPT "chaingun-by-JD" by default
    topic: "sensors/room1"		# MAND
    payload: "{ \"Temp\": \"20°C\" }"	# MAND format depends on your app
    qos: 1				# OPT values can be 0, 1 (default) or 2
					# Variable interpolation is applied on
					# url, payload and topic
```

## grpc : gRPC Request (beta)

Note : streaming requests and/or responses are not yet supported.

| Parameter Name | Description |
| :--- | :--- |
| `title` | mandatory string that qualifies the request - used for the result output and logging |
| `call` | mandatory string that indicates the function to call (ex: package.service.function) |
| `data` | mandatory JSON string to send as the payload |
| `responses` | data can be extracted from server responses. The response is considered as a JSON document, so you must use jsonpath |

Variable interpolation applies to requests and responses.

Example:
```
- grpc:
    title: Hello
    call: chat.ChatSerice.SayHello
    data: '{"body": "hello !"}'
    responses:
      - jsonpath: $.body+
        variable: name
        index: first
        default_value: alice
```

## tcp or udp : simple TCP or UDP Request

| Parameter Name | Description |
| :--- | :--- |
| `title` | mandatory string that qualifies the request - used for the result output and logging |
| `address` | mandatory string that indicates the server address and the port to connect to |
| `payload` | mandatory string to send as the payload. If you want to specify special characters (like \n), do not forget to enclose the string in double-quotes |
| `payload64` | mandatory base64-encoded string to send as the payload. Use it to send binary data. One of either payload either payload64 can be declared |

Variable interpolation applies to url and payload, not to payload64.

Example:
```
- tcp:
    address: 127.0.0.1:8081
    payload: "ACT|LOGIN|${person}|${name}\n"
```

## setvar : creates and set variable values

| Parameter Name | Description |
| :--- | :--- |
| `variable` | mandatory variable name |
| `expression` | mandatory string that defines an expression to be evaluated. If the expression is an integer or a float, it will be converted into a string |

Example :

```
- setvar:
    variable: my_var
    expression: "2 * age"
```

## sleep : wait Action

| Parameter Name | Description |
| :--- | :--- |
| `duration` | mandatory integer that gives the sleep time in milliseconds |

Example :

```
- sleep:
    duration: 2000
```

## log : log output Action

| Parameter Name | Description |
| :--- | :--- |
| `message` | mandatory string that will be displayed on the output or gathered in the logs if the Player is launched in daemon mode. The message can reference variables. |

Example:
```
  - log:
      message: "Variable interpolation is possible : ${name}"
```

## assert : creates assertion

| Parameter Name | Description |
| :--- | :--- |
| `expression` | mandatory string that defines an expression to be evaluated |

Example:
```
  - assert:
      expression: "name == \"bob\""
```

## timers : creates page timers

Timers can be used to measure the latency time for a set of requests.
You define a Timer with the `start_timer` action which contains the Timer name.
You stop the Timer by repeating the name in the `end_action`.
Timer values will be displayed in the results with a name like `__timer__name`.

| Parameter Name | Description |
| :--- | :--- |
| `name` | mandatory string that defines the Timer name |

Example:
```
  - start_timer:
      name: HomePage

  # other actions
  #- http:
  #    ...

  - end_timer:
      name: HomePage
```

# Advanced Topics

## Variables usage

Variables can be used in the following contexts :

in the following Action parameters, `http.url`, `http.body`, `mqtt.url`, `log.message`, `mqtt.url`, `mqtt.topic`, `mqtt.payload`, `mongodb.server`, `mongodb.document`, `mongodb.filter`, `sql.statement`. 
In these cases, the variable names must be enclosed between `${....}`.

For example:

```
- http:
    title: Page 3
    method: GET
    url: http://server/page3.php?name=${name}

- log:
    message: Address value is ${address} (customer=${customer})

# The HTTP_Response variable is always set after a HTTP action
- log:
    message: HTTP return code=${HTTP_Response}
```

## Expressions

Expressions are strings that can contain scalar values (int, float, string, bool), standard operators and variables.
Variables are not surrounded by a `${...}`, they are named as is.

The evaluation of the expression must return an int, a string or a bool (floats are converted to ints)

The supported operators are described here:
   https://github.com/Knetic/govaluate/blob/master/MANUAL.md

Summary:
 - allowed types are: `float64`, `int`, `bool`, `string` and arrays
 - strings that matches date format are converted into a `float64`
 - __+__ operator can be used with numbers and `string` (concatenation)
 - __-__, __*__, __/__, __**__ and __%__ only work with numbers
 - __>>__, __<<__, __|__, __&__ and __^__ use `int64` (`float64` will be converted)
 - __-__ as unary operator works with number
 - __!__ works with `bool`
 - __~__ (bitwise not) works with numbers
 - __||__, __&&__ work with `bool`
 - __?__ (ternary true) uses a `bool`, any type and returns the 3rd op or nil
 - __:__ (ternary false) uses any type, any type and returns the 3rd op or nil
 - __??__ (null coalescence) returns the left-sie if non-nil otherwise returns the right-side
 - __>__, __<__, __>=__, __<=__ are comparators, both ops must have the same type (number or `string`)
 - __=~__, __!~__ are used for regexp matching, both ops are `string`s

The supported functions are:
 - strlen(string)
 - substr(string, start, end)
 - random(start, end) which returns an integer between start and end, included

Examples:
```
  expression: "var1 + 3 > 4 * var2"
  expression: "strlen(var3) > 0"
  expression: "random(1990, 2020)"
```

## Session variables and Cookies

The session variables and the Cookies are deleted at the end of each script iteration (if there are many) played by a VU.
This behaviour can be changed for HTTP/S actions by setting the global parameter `persistent_http_sessions` to true.


## The `when` clause to trigger Actions

Each Action can be triggered by a `when` clause which defines an expression that must be evaluated to True to trigger the Action.

Example:
```
- setvar:
    variable: xyz
    expression: "10 * delay"
  when: "delay > 2"
```

## How to import external data

The `feeder` global section can be used to define an single source of external data. The following keys are mandatory :

| Key Name | Description |
| :--- | :--- |
| `type` | mandatory file type (only "csv" is supported) |
| `filename` | mandatory string that gives the filename |
| `separator` | mandatory string that gives the column separator |

The first line of the file must contain the column names. These names will be used to name the feeded variables !

Example :

```
feeder:
  type: csv
  filename: /path/to/data1.csv
  separator: ","
```

## Submit form using "multipart/form-data" syntax

You can use the `formdata` key to submit form fields encoded with `multipart/form-data`. This is necessary if
your form embeds a field of type `file`.

```
- http:
    title: Page upload
    method: POST		# only for POST methods
    url: http://yourserver/a_page.php
    formdata:
      - name: name
        value: ${a_variable} Doe
      - name: fileToUpload
        value: a_filename.txt
        type: file		# mandatory for files
      - name: submit
```

## Handle HTTP Basic authentication

This method should not be used on unencrypted channel (HTTP)... If the web server requires a basic authentication
you just have to specify the username and the password in the URL.
Since variable interpolation is possible in the "server" part of the URL, you can reference variables defined
in the 'variable' section. Here are some examples :

```
variable:
  user: bob
  pwd: secret

actions:
  - http:
      title: private page1
      url: https://${user}:${pwd}@myserver/private/page1.html

  - http:
      title: private page2
      url: http://alice:terces@myserver/private/page2.html
```


# Full sample

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

pre_actions:		# some kind of inits just played once before the actions
  - http:
      title: init
      method: DELETE
      url: http://server/index

actions:
  # A simple GET
  - http:
      title: Page 1			# MAND for http action
      method: GET			# MAND for http action (GET/POST/PUT/HEAD/DELETE)
      url: http://server/page1.php	# MAND for http action
      # name of Cookie to store. __all__ catches all cookies !
      store_cookie: __all__

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
      variable: my_var
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
