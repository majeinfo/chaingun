var chaingunScript = new Vue({
	el: '#chaingunApp',

	data: {
		scriptParms: {
			iterations: 1,
			duration: 0,
			rampup: 0,
			users: 1,
			timeout: 10,
			on_error: 'continue',
			http_error_code: '',
			persistent_connections: false,
			'default': {
				server: '',
				protocol: '',
				method: '',
				database: '',
				collection: '',
				db_driver: '',
			},
			variables: {},
			feeder: {
				type: 'csv',
				filename: '',
				separator: ',',
			},
			pre_actions: [],
			actions: [],
		},
		action: {
			type: 'http',
			title: '',
			method: '',
			use_http2: '',
			url: '',
			store_cookie: '',
			body: '',
			template: '',
			upload_file: '',
			headers: [],
			header_name: '',
			header_value: '',
			responses: [],
			from_header: '',
			regex: '',
			jsonpath: '',
			xmlpath: '',
			index: '',
			default_value: '',
			server: '',
			database: '',
			collection: '',
			command: '',
			filter: '',
			document: '',
			db_driver: 'mysql',
			statement: '',
			certificatepath: '',
			privatekeypath: '',
			username: '',
			password: '',
			clientid: '',
			topic: '',
			payload: '',
			qos: 0,
			variables: [],
			variable: '',
			expression: '',
			duration: 100,
			message: '',
			when_clause: '',
		},
		yamlScript: "",
		errors: [],
		errors2: [],
		actionTypes: ["assert", "http", "log", "mongodb", "mqtt", "setvar", "sleep", "sql", "tcp", "udp", "ws"],
		cur_action: '',
		edit_action_mode: '',
		edit_header_mode: '',
		edit_when_mode: '',
                edit_response_mode: '',
		response_index: 0,
		action_index: 0,
		variable_index: 0,
		header_index: 0,
		moving_action: null,
		is_pre_action: false,
	},

	mounted: function() {
		console.log('mounted');
		
		// Emit callback
		this.$on('change_feeder_type', function(value) {
			this.scriptParms.feeder.type = value;
		});
		this.$on('change_feeder_filename', function(value) {
			console.log('change_feeder_filename: ' + value);
			this.scriptParms.feeder.filename = value;
		});
		this.$on('change_feeder_separator', function(value) {
			this.scriptParms.feeder.separator = value;
		});
		this.$on('change_default_server', function(value) {
			this.scriptParms.default.server = value;
		});
		this.$on('change_default_protocol', function(value) {
			this.scriptParms.default.protocol = value;
		});
		this.$on('change_default_method', function(value) {
			this.scriptParms.default.method = value;
		});
		this.$on('change_default_database', function(value) {
			this.scriptParms.default.database = value;
		});
		this.$on('change_default_collection', function(value) {
			this.scriptParms.default.collection = value;
		});
		this.$on('change_default_db_driver', function(value) {
			this.scriptParms.default.db_driver = value;
		});
		this.$on('change_title', function(value) {
			this.action.title = value;
		});
		this.$on('change_address', function(value) {
			this.action.address = value;
		});
		this.$on('change_payload', function(value) {
			this.action.payload = value;
		});
		this.$on('change_log_message', function(value) {
			this.action.message = value;
		});
		this.$on('change_duration', function(value) {
			this.action.duration = parseInt(value, 10);
		});
		this.$on('change_variable', function(value) {
			this.action.variable = value;
		});
		this.$on('change_expression', function(value) {
			this.action.expression = value;
		});
		this.$on('change_url', function(value) {
			this.action.url = value;
		});
		this.$on('change_method', function(value) {
			this.action.method = value;
		});
		this.$on('change_use_http2', function(value) {
			this.action.use_http2 = value;
		});
		this.$on('change_certificatepath', function(value) {
			this.action.certificatepath = value;
		});
		this.$on('change_privatekeypath', function(value) {
			this.action.privatekeypath = value;
		});
		this.$on('change_username', function(value) {
			this.action.username = value;
		});
		this.$on('change_password', function(value) {
			this.action.password = value;
		});
		this.$on('change_clientid', function(value) {
			this.action.clientid = value;
		});
		this.$on('change_topic', function(value) {
			this.action.topic = value;
		});
		this.$on('change_qos', function(value) {
			this.action.qos = parseInt(value, 10);
		});
		this.$on('change_database', function(value) {
			this.action.database = value;
		});
		this.$on('change_db_driver', function(value) {
			this.action.db_driver = value;
		});
		this.$on('change_statement', function(value) {
			this.action.statement = value;
		});
		this.$on('change_server', function(value) {
			this.action.server = value;
		});
		this.$on('change_store_cookie', function(value) {
			this.action.store_cookie = value;
		});
		this.$on('change_body', function(value) {
			this.action.body = value;
		});
		this.$on('change_template', function(value) {
			this.action.template = value;
		});
		this.$on('change_upload_file', function(value) {
			this.action.upload_file = value;
		});
		this.$on('change_responses', function(value) {
			this.action.responses = value;
		});
		this.$on('change_headers', function(value) {
			this.action.headers = value;
		});
		this.$on('change_from_header', function(value) {
			this.action.from_header = value;
		});
		this.$on('change_regex', function(value) {
			this.action.regex = value;
		});
		this.$on('change_jsonpath', function(value) {
			this.action.jsonpath = value;
		});
		this.$on('change_xmlpath', function(value) {
			this.action.xmlpath = value;
		});
		this.$on('change_index', function(value) {
			this.action.index = value;
		});
		this.$on('change_default_value', function(value) {
			this.action.default_value = value;
		});
		this.$on('change_from_header', function(value) {
			this.action.from_header = value;
		});
		this.$on('change_collection', function(value) {
			this.action.collection = value;
		});
		this.$on('change_filter', function(value) {
			this.action.filter = value;
		});
		this.$on('change_document', function(value) {
			this.action.document = value;
		});
		this.$on('change_command', function(value) {
			this.action.command = value;
		});
		this.$on('change_http_header_name', function(value) {
			this.action.header_name = value;
		});
		this.$on('change_http_header_value', function(value) {
			this.action.header_value = value;
		});
		this.$on('change_when_clause', function(value) {
			this.action.when_clause = value;
		});
		this.$on('clear_action', function(value) {
			this.clearAction();
		});
		this.$on('new_action', function(value) {
			this.newAction();
		});
		this.$on('clear_variable', function(value) {
			this.clearVariable();
		});
		this.$on('new_variable', function(value) {
			this.newVariable();
		});
		this.$on('clear_header', function(value) {
			this.clearHeader();
		});
		this.$on('new_header', function(value) {
			this.newHeader();
		});
		this.$on('clear_response', function(value) {
			this.clearResponse();
		});
		this.$on('new_response', function(value) {
			this.newResponse();
		});
		this.$on('clear_when', function(value) {
			this.clearWhen();
		});
		this.$on('new_when', function(value) {
			this.newWhen();
		});
	},

	methods: {
		update: function() {
			this.$forceUpdate();
		},
		// ACTION
                actionShow: function(pre_action) {
			console.log('actionShow: ' + this.action.type);
			this.is_pre_action = pre_action;
			this.cur_action = this.action.type;
			this.edit_action_mode = 'New';
			this.action.title = '';
			this.action.url = '';
			this.action.address = '';
			this.action.payload = '';
			this.action.headers = [];
			this.action.responses = [];
			this.action.variable = '';
			this.action.expression = '';
			this.action.message = '';
			this.action.body = '';
			this.action.template = '';
			this.action.upload_file = '';
			this.action.store_cookie = '';
			this.action.method = '';
			this.action.message = '';
			$('#new_' + this.cur_action).modal('show');
                },
 		clearAction: function() {
			this.errors = [];
		},
                newAction: function() {
			console.log('newAction');
			this.errors = [];
			newAction = _buildNewAction(this.action, this.scriptParms.default, this.errors);
			if (this.errors.length == 0) {
				if (this.edit_action_mode == 'New') {
					console.log('This a new action');
					if (this.is_pre_action) {
						this.scriptParms.pre_actions.push(newAction);
					} else {
						this.scriptParms.actions.push(newAction);
					}
				} else {
					console.log('Should update action ' + parseInt(this.action_index, 10));
					if (this.is_pre_action) {
						this.scriptParms.pre_actions[this.action_index] = newAction;
					} else {
						this.scriptParms.actions[this.action_index] = newAction;
					}
				}
				$('#new_' + this.cur_action).modal('hide');
			}
                },
		displayForEditAction: function(idx, pre_action) {
			console.log('displayForEditAction: ' + parseInt(idx, 10));
			if (pre_action) {
				action_type = _prepareEditAction(this.action, this.scriptParms.pre_actions[idx]);
			} else {
				action_type = _prepareEditAction(this.action, this.scriptParms.actions[idx]);
			}
			this.edit_action_mode = 'Edit';
			this.action_index = idx;
			$('#new_' + action_type).modal('show');
		},
		deleteAction: function(idx, pre_action) {
			console.log('deleteAction' + parseInt(idx, 10));
			if (pre_action) {
				this.scriptParms.pre_actions.splice(idx, 1);
			} else {
				this.scriptParms.actions.splice(idx, 1);
			}
		},
		dragActionStart: function(idx, evt) {
			console.log('dragActionStart: ' + parseInt(idx, 10));
			evt.target.style.opacity = 0.5;
                        evt.dataTransfer.setData('text/plain', 'This Action may be dragged')
			this.moving_action = idx;
		},
		dragActionEnd: function(evt) {
			console.log('dragActionEnd');
			evt.target.style.opacity = 1;
			this.moving_action = null;
		},
                dragActionFinish: function(index, evt, pre_action) {
                        console.log('dragFinish');
			var data = event.dataTransfer.getData('text/plain');
			event.preventDefault();
			console.log('exchange actions ' + parseInt(this.moving_action, 10) + ' with ' + parseInt(index, 10));
			if (index != this.moving_action) {
				if (pre_action) {
					var save_action = this.scriptParms.pre_actions[this.moving_action];
					this.scriptParms.pre_actions[this.moving_action] = this.scriptParms.pre_actions[index];
					this.scriptParms.pre_actions[index] = save_action;
				} else {
					var save_action = this.scriptParms.actions[this.moving_action];
					this.scriptParms.actions[this.moving_action] = this.scriptParms.actions[index];
					this.scriptParms.actions[index] = save_action;
				}
			}
			this.moving_action = null;
			this.update();
                },   
		// HEADER
		headerShow: function() {
			this.edit_header_mode = 'New';
			this.action.header_name = '';
			this.action.header_value = '';
			$('#new_header').modal('show');
		},
 		clearHeader: function() {
			this.errors2 = [];
		},
                newHeader: function() {
			console.log('newHeader');
			this.errors2 = [];

			if (this.action.header_name == '') {
				this.errors2.push("The HTTP Header name must not be empty"); 
			} 
			if (this.action.header_value == '') {
				this.errors2.push("The HTTP Header value must not be empty"); 
			} 
			if (this.errors2.length == 0) {
				var k = this.action.header_name;
				var v = this.action.header_value;
				if (this.edit_header_mode == 'New') {
					// Vue.js v2 cannot iterate on map, so we build an array
					this.action.headers.push([k, v]);
				} else {
					this.action.headers[this.header_index] = [k, v];
				}
				$('#new_header').modal('hide');
			}
			this.update();
                },
		displayForEditHeader: function(idx) {
			this.edit_header_mode = 'Edit';
			this.header_index = idx;
			this.action.header_name = this.action.headers[idx][0];
			this.action.header_value = this.action.headers[idx][1];
			$('#new_header').modal('show');
		},
		deleteHeader: function(idx) {
			this.action.headers.splice(idx, 1);
		},
		// RESPONSE
		responseShow: function() {
			this.edit_response_mode = 'New';
			this.action.from_header = '';
			this.action.regex = '';
			this.action.jsonpath = '';
			this.action.xmlpath = '';
			this.action.index = '';
			this.action.variable = '';
			this.action.default_value = '';
			$('#new_response').modal('show');
		},
 		clearResponse: function() {
			this.errors2 = [];
		},
                newResponse: function() {
			console.log('newResponse');
			this.errors2 = [];

			if (this.action.regex == '' && this.action.jsonpath == '' && this.action.xmlpath == '') {
				this.errors2.push("A regex or jsonpath or xmlpath must be specified"); 
			} 
			if (this.action.variable == '') {
				this.errors2.push("The Variable Name must not be empty"); 
			} 
			if (this.errors2.length == 0) {
				var response = { 
					variable: this.action.variable
				};
				if (this.action.from_header != '') { response['from_header'] = this.action.from_header; }
				if (this.action.regex != '') { response['regex'] = this.action.regex; }
				if (this.action.jsonpath != '') { response['jsonpath'] = this.action.jsonpath; }
				if (this.action.xmlpath != '') { response['xmlpath'] = this.action.xmlpath; }
				if (this.action.default_value != '') { response['default_value'] = this.action.default_value; }
				if (this.action.index != '') { response['index'] = this.action.index; }
				if (this.edit_response_mode == 'New') {
					this.action.responses.push(response);
				} else {
					this.action.responses[this.response_index] = response;
				}
				$('#new_response').modal('hide');
			}
			console.log('length of action.responses=' + parseInt(this.action.responses.length, 10));
			this.update();
                },
		displayForEditResponse: function(idx) {
			console.log('displayForEditResponse: ' + parseInt(idx, 10));
			this.edit_response_mode = 'Edit';
			this.response_index = idx;
			this.action.from_header = this.action.responses[idx]['from_header'];
			this.action.regex = this.action.responses[idx]['regex'];
			this.action.jsonpath = this.action.responses[idx]['jsonpath'];
			this.action.xmlpath = this.action.responses[idx]['xmlpath'];
			this.action.index = this.action.responses[idx]['index'];
			this.action.variable = this.action.responses[idx]['variable'];
			this.action.default_value = this.action.responses[idx]['default_value'];
			$('#new_response').modal('show');
		},
		deleteResponse: function(idx) {
			this.action.responses.splice(idx, 1);
		},
		// VARIABLE
                variableShow: function() {
			console.log('variableShow');
			this.edit_action_mode = 'New';
			this.action.variable = '';
			this.action.expression = '';
			$('#new_variable').modal('show');
                },
 		clearVariable: function() {
			this.errors = [];
		},
                newVariable: function() {
			console.log('newVariable');
			this.errors = [];

			if (this.action.variable == '') {
				this.errors.push("The Variable name must not be empty"); 
			} 
			if (this.action.expression == '') {
				this.errors.push("The Expression value must not be empty"); 
			} 
			if (this.errors.length == 0) {
				var k = this.action.variable;
				var v = this.action.expression;
				if (this.edit_action_mode == 'New') {
					// Vue.js v2 cannot iterate on map, so we build an array
					this.action.variables.push([k, v]);
				} else {
					this.action.variables[this.variable_index] = [k, v];
				}
				$('#new_variable').modal('hide');
			}
			this.update();
                },
		displayForEditVariable: function(idx) {
			console.log('displayForEditVariable: ' + parseInt(idx, 10));
			this.edit_action_mode = 'Edit';
			this.variable_index = idx;
			this.action.variable = this.action.variables[idx][0];
			this.action.expression = this.action.variables[idx][1];
			$('#new_variable').modal('show');
		},
		deleteVariable: function(idx) {
			console.log('deleteVariable ' + parseInt(idx, 10));
			this.action.variables.splice(idx, 1);
			this.update();
		},
		// WHEN
		whenShow: function(idx, pre_action) {
			console.log('whenShow: ' + parseInt(idx));
			this.edit_when_mode = 'New';
			this.action_index = idx;
			this.is_pre_action = pre_action;
			$('#new_when').modal('show');
		},
 		clearWhen: function() {
			this.errors = [];
		},
		newWhen: function() {
			console.log('newWhen');
			if (this.action.when_clause == '') {
				this.errors.push("Expression must not be empty"); 
			} else {
				if (this.is_pre_action) {
					this.scriptParms.pre_actions[this.action_index]['when'] = this.action.when_clause;
				} else {
					this.scriptParms.actions[this.action_index]['when'] = this.action.when_clause;
				}
				$('#new_when').modal('hide');
				this.action.when_clause = '';
			}
			this.update();
		},
		displayForWhen: function(idx, pre_action) {
			console.log('displayForWhen: ' + parseInt(idx));
			this.edit_when_mode = 'Edit';
			this.action_index = idx;
			if (pre_action) {
				this.action.when_clause = this.scriptParms.pre_actions[idx]['when'];
			} else {
				this.action.when_clause = this.scriptParms.actions[idx]['when'];
			}
			$('#new_when').modal('show');
		},
		deleteWhen: function(idx, pre_action) {
			console.log('deleteWhen: ' + parseInt(idx));
			if (pre_action) {
				delete this.scriptParms.pre_actions[idx]['when'];
			} else {
				delete this.scriptParms.actions[idx]['when'];
			}
			this.update();
		},
		// FINAL
		checkForm: function(e) {
			console.log('checkForm called');
			this.errors = [];
			this.yamlScript = "ERROR !";
			validateAll(this.scriptParms, this.errors);
			if (this.errors.length == 0) {
				buildYAML(this.scriptParms, this.action.variables);
			}

			//e.preventDefault();
		},
		getDisplayAction: function(action) {
			return _getDisplayAction(action, this.scriptParms.default);
		},
	},
});

function _buildNewAction(action, dflt, errors) {
	let newAction = {};
	newAction[action.type] = {};

	switch (action.type) {
	case 'http':
		if (action.title == '') { errors.push("Title field must not be empty"); }
		if (action.url == '') { errors.push("URL field must not be empty"); }
		if (action.method == '' && dflt.method == '') { errors.push("Method field or Default Method must not be empty"); }
		newAction['http'].title = action.title;
		newAction['http'].url = action.url;
		if (action.method != '') { newAction['http'].method = action.method; }
		if (action.store_cookie != '') { newAction['http'].store_cookie = action.store_cookie; }
		if (action.body != '') { newAction['http'].body = action.body; }
		if (action.template != '') { newAction['http'].template = action.template; }
		if (action.upload_file != '') { newAction['http'].upload_file = action.upload_file; }
		if (action.use_http2 != '') { newAction['http'].use_http2 = (action.use_http2 == 'true'); }
		if (action.headers.length > 0) { 
			newAction['http']['headers'] = {};
			for (var idx = 0 ; idx < action.headers.length ; idx++ ) {
				newAction['http']['headers'][action.headers[idx][0]] = action.headers[idx][1];
			}
		}
		newAction['http'].responses = action.responses;
		break;
	case 'log':
		if (action.message == '') { errors.push("Message cannot be null"); }
		newAction['log'].message = action.message;
		break
	case 'assert':
		if (action.expression == '') { errors.push("Expression cannot be null"); }
		newAction['assert'].expression = action.expression;
		break
	case 'sleep':
		if (action.duration <= 0) { errors.push("Duration cannot be negative or null"); }
		newAction['sleep'].duration = action.duration;
		break;
	case 'setvar':
		if (action.variable == '') { errors.push("Variable Name cannot be null"); }
		if (action.expression == '') { errors.push("Expression cannot be null"); }
		newAction['setvar'].variable = action.variable;
		newAction['setvar'].expression = action.expression;
		break
	case 'tcp':
		if (action.title == '') { errors.push("Title cannot be null"); }
		if (action.address == '') { errors.push("Address cannot be null"); }
		if (action.payload == '') { errors.push("Payload cannot be null"); }
		newAction['tcp'].title = action.title;
		newAction['tcp'].address = action.address;
		newAction['tcp'].payload = action.payload;
		break
	case 'udp':
		if (action.title == '') { errors.push("Title cannot be null"); }
		if (action.address == '') { errors.push("Address cannot be null"); }
		if (action.payload == '') { errors.push("Payload cannot be null"); }
		newAction['udp'].title = action.title;
		newAction['udp'].address = action.address;
		newAction['udp'].payload = action.payload;
		break
	case 'sql':
		if (action.title == '') { errors.push("Title cannot be null"); }
		if (action.statement == '') { errors.push("Statement cannot be null"); }
		if (action.server == '' && dflt.server == '') { errors.push("Server field or default Server cannot be null"); }
		if (action.database == '' && dflt.database == '') { errors.push("Database field or default Database cannot be null"); }
		if (action.db_driver == '' && dflt.db_driver == '') { errors.push("DB Driver field or default DB Driver cannot be null"); }
		newAction['sql'].title = action.title;
		newAction['sql'].statement = action.statement;
		if (action.db_driver != '') { newAction['sql'].db_driver = action.db_driver; }
		if (action.database != '') { newAction['sql'].database = action.database; }
		if (action.server != '') { newAction['sql'].server = action.server; }
		break
	case 'mongodb':
		if (action.title == '') { errors.push("Title cannot be null"); }
		if (action.server == '' && dflt.server == '') { errors.push("Server field or default Server cannot be null"); }
		if (action.database == '' && dflt.database == '') { errors.push("Database field or default Database cannot be null"); }
		if (action.collection == '' && dflt.collection == '') { errors.push("Collection field or default Collection cannot be null"); }
		if (action.command == '') { errors.push("Command cannot be empty (must be findone, insertone or drop)"); }
		if (action.filter != '' && action.command != 'findone') { errors.push("Filter can be applied only with 'findone' command"); }
		if (action.document == '' && action.command == 'insertone') { errors.push("With the 'insertone' command, you need to specify a JSON document"); }

		newAction['mongodb'].title = action.title;
		if (action.server != '') { newAction['mongodb'].server = action.server; }
		if (action.database != '') { newAction['mongodb'].database = action.database; }
		if (action.collection != '') { newAction['mongodb'].collection = action.collection; }
		newAction['mongodb'].command = action.command;
		if (action.filter != '') { newAction['mongodb'].filter = action.filter; }
		if (action.document != '') { newAction['mongodb'].document = action.document; }
		newAction['mongodb'].responses = action.responses;
		break
	case 'ws':
		if (action.title == '') { errors.push("Title cannot be null"); }
		if (action.url == '' && dflt.server == '') { errors.push("URL field or default Server cannot be null"); }
		newAction['ws'].title = action.title;
		newAction['ws'].url = action.url;
		if (action.store_cookie != '') { newAction['ws'].store_cookie = action.store_cookie; }
		if (action.body != '') { newAction['ws'].body = action.body; }
		newAction['ws'].responses = action.responses;
		//if (action.url == '') { newAction['ws'].url = dflt.server; }
		break
	case 'mqtt':
		if (action.title == '') { errors.push("Title cannot be null"); }
		if (action.url == '') { errors.push("URL cannot be null"); }
		if (action.topic == '') { errors.push("Topic cannot be null"); }
		if (action.payload == '') { errors.push("Payload cannot be null"); }
		newAction['mqtt'].title = action.title;
		newAction['mqtt'].url = action.url;
		newAction['mqtt'].certificatepath = action.certificatepath;
		newAction['mqtt'].privatekeypath = action.privatekeypath;
		newAction['mqtt'].username = action.username;
		newAction['mqtt'].password = action.password;
		newAction['mqtt'].clientid = action.clientid;
		newAction['mqtt'].topic = action.topic;
		newAction['mqtt'].payload = action.payload;
		newAction['mqtt'].qos = action.qos;
		break
	}
	
	return newAction;
}

function _prepareEditAction(target, action) {
	if ('http' in action) {
		target.title = action.http.title;
		target.url = action.http.url;
		if ('method' in action.http) { target.method = action.http.method; }
		if ('store_cookie' in action.http) { target.store_cookie = action.http.store_cookie; }
		if ('body' in action.http) { target.body = action.http.body; }
		if ('template' in action.http) { target.template = action.http.template; }
		if ('upload_file' in action.http) { target.upload_file = action.http.upload_file; }
		if ('use_http2' in action.http) { target.use_http2 = action.http.use_http2; }
		target.headers = [];
		if ('headers' in action.http) {
			for (var entry in action.http.headers) {
				target.headers.push([entry, action.http.headers[entry]]);
			}
		}
		target.responses = action.http.responses;
		return 'http';
	}
	else if ('log' in action) {
		target.message = action.log.message;
		return 'log';
	}
	else if ('sleep' in action) {
		target.duration = action.sleep.duration;
		return 'sleep';
	}
	else if ('assert' in action) {
		target.expression = action.assert.expression;
		return 'assert';
	}
	else if ('setvar' in action) {
		target.variable = action.setvar.variable;
		target.expression = action.setvar.expression;
		return 'setvar';
	}
	else if ('tcp' in action) {
		target.title = action.tcp.title;
		target.address = action.tcp.address;
		target.payload = action.tcp.payload;
		return 'tcp';
	}
	else if ('udp' in action) {
		target.title = action.udp.title;
		target.address = action.udp.address;
		target.payload = action.udp.payload;
		return 'udp';
	}
	else if ('mqtt' in action) {
		target.title = action.mqtt.title;
		target.url = action.mqtt.url;
		target.certificatepath = action.mqtt.certificatepath;
		target.privatekeypath = action.mqtt.privatekeypath;
		target.username = action.mqtt.username;
		target.password = action.mqtt.password;
		target.clientid = action.mqtt.clientid;
		target.topic = action.mqtt.topic;
		target.payload = action.mqtt.payload;
		target.qos = action.mqtt.qos;
		return 'mqtt';
	}
	else if ('sql' in action) {
		target.title = action.sql.title;
		target.server = action.sql.server;
		target.db_driver = action.sql.db_driver;
		target.database = action.sql.database;
		target.statement = action.sql.statement;
		return 'sql';
	}
	else if ('ws' in action) {
		target.title = action.ws.title;
		target.url = action.ws.url;
		target.store_cookie = action.ws.store_cookie;
		target.body = action.ws.body;
		target.responses = action.ws.responses;
		return 'ws';
	}
	else if ('mongodb' in action) {
		target.title = action.mongodb.title;
		target.server = action.mongodb.server;
		target.database = action.mongodb.database;
		target.collection = action.mongodb.collection;
		target.command = action.mongodb.command;
		target.filter = action.mongodb.filter;
		target.document = action.mongodb.document;
		target.responses = action.mongodb.responses;
		return 'mongodb';
	}

	alert('Action has unknown Type !');
}

function _getDisplayAction(action, dflt) {
	if ('http' in action) {
		return 'HTTP ' + ((action.http.method != '')?action.http.method:dflt.method) + ': "' + action.http.title + '"';
	}
	else if ('sleep' in action) {
		return 'SLEEP for ' + parseInt(action.sleep.duration) + ' ms';
	} 
	else if ('log' in action) {
		return 'LOG "' + action.log.message + '"';
	}
	else if ('assert' in action) {
		return 'ASSERT "' + action.assert.expression + '"';
	}
	else if ('setvar' in action) {
		return 'SETVAR "' + action.setvar.variable + '"';
	}
	else if ('tcp' in action) {
		return 'TCP "' + action.tcp.title + '"';
	}
	else if ('udp' in action) {
		return 'UDP "' + action.udp.title + '"';
	}
	else if ('mqtt' in action) {
		return 'MQTT "' + action.mqtt.title + '"';
	}
	else if ('sql' in action) {
		return 'SQL "' + action.sql.statement + '"';
	}
	else if ('ws' in action) {
		return 'WS "' + action.ws.title + '"';
	}
	else if ('mongodb' in action) {
		return 'MONGODB "' + action.mongodb.title + '"';
	}

	alert('Action has unknown Type !');
	return 'Unknown Action Type'; 
}

function validateAll(data, errors) {
	/*
	data.iterations = parseInt(data.iterations, 10);
	if (isNaN(data.iterations)) {
		errors.push("Iteration Count must be an Integer");
	}
	*/
	if (data.iterations == 0) {
		errors.push("Iteration Count cannot be 0");
	}
	if (data.iterations < 0 && data.iterations != -1) {
		errors.push("Iteration Count cannot be negative and not equal to -1");
	}
	if (data.users <= 0) {
		errors.push("Users Count must be > 0");
	}
	data.persistent_connections = (data.persistent_connections == 'true') ? true : false;
}

function buildYAML(data, variables) {
	// var doc = jsyaml.load('greeting: hello\nname: world');
	// Strip default values:
	
	// transform action.variables into scriptParms.variables
	for (var idx = 0; idx < variables.length ; idx++) {
		var k = variables[idx][0];
		var v = variables[idx][1];
		data.variables[k] = v;
	}

	var copy = JSON.parse(JSON.stringify(data));
	
	//if (!data.variables.size) { delete copy.variables; }
	if (copy.http_error_code == '') { delete copy['http_error_code']; }
	if (copy.on_error == 'continue') { delete copy['on_error']; }
	if (copy.timeout == 10 ) { delete copy['timeout']; }
	if (!copy.persistent_connections) { delete copy['persistent_connections']; }
	if (copy.default.server == '') { delete copy.default['server']; }
	if (copy.default.protocol == '') { delete copy.default['protocol']; }
	if (copy.default.method == '') { delete copy.default['method']; }
	if (copy.default.database == '') { delete copy.default['database']; }
	if (copy.default.collection == '') { delete copy.default['collection']; }
	if (copy.default.db_driver == '') { delete copy.default['db_driver']; }
	if (copy.feeder.filename == '') { delete copy.feeder; }
	//if (!data.default.size) { delete copy.default; }

	var text = jsyaml.dump(copy);
	text = text.replace(/\n/g, "<br/>");
	chaingunScript.yamlScript = text.replace(/ /g, "&nbsp;");
}

function allowDrop(evt) {
        console.log('allowDrop');
        evt.preventDefault();
}

