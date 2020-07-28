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
			responses: '',
			server: '',
			database: '',
			collection: '',
			command: '',
			filter: '',
			document: '',
			db_driver: '',
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
		actionTypes: ["assert", "http", "log", "setvar", "sleep"],
		cur_action: '',
		edit_action_mode: '',
		edit_header_mode: '',
		edit_when_mode: '',
		action_index: 0,
		variable_index: 0,
		header_index: 0,
		moving_action: null,
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
	},

	methods: {
		update: function() {
			this.$forceUpdate();
		},
		// ACTION
                actionShow: function() {
			console.log('actionShow: ' + this.action.type);
			this.cur_action = this.action.type;
			this.edit_action_mode = 'New';
			this.action.title = '';
			this.action.url = '';
			this.action.headers = [];
			this.action.variable = '';
			this.action.expression = '';
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
					this.scriptParms.actions.push(newAction);
				} else {
					console.log('Should update action ' + parseInt(this.action_index, 10));
					this.scriptParms.actions[this.action_index] = newAction;
				}
				$('#new_' + this.cur_action).modal('hide');
			}
                },
		displayForEditAction: function(idx) {
			console.log('displayForEditAction: ' + parseInt(idx, 10));
			var action_type = _prepareEditAction(this.action, this.scriptParms.actions[idx]);
			this.edit_action_mode = 'Edit';
			this.action_index = idx;
			$('#new_' + action_type).modal('show');
		},
		deleteAction: function(idx) {
			console.log('deleteAction' + parseInt(idx, 10));
			this.scriptParms.actions.splice(idx, 1);
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
                dragActionFinish: function(index, evt) {
                        console.log('dragFinish');
			var data = event.dataTransfer.getData('text/plain');
			event.preventDefault();
			console.log('exchange actions ' + parseInt(this.moving_action, 10) + ' with ' + parseInt(index, 10));
			if (index != this.moving_action) {
				var save_action = this.scriptParms.actions[this.moving_action];
				this.scriptParms.actions[this.moving_action] = this.scriptParms.actions[index];
				this.scriptParms.actions[index] = save_action;
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
		whenShow: function(idx) {
			console.log('whenShow: ' + parseInt(idx));
			this.edit_when_mode = 'New';
			this.action_index = idx;
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
				this.scriptParms.actions[this.action_index]['when'] = this.action.when_clause;
				$('#new_when').modal('hide');
			}
			this.update();
		},
		displayForWhen: function(idx) {
			console.log('displayForWhen: ' + parseInt(idx));
			this.edit_when_mode = 'Edit';
			this.action_index = idx;
			this.action.when_clause = this.scriptParms.actions[idx]['when'];
			$('#new_when').modal('show');
		},
		deleteWhen: function(idx) {
			console.log('deleteWhen: ' + parseInt(idx));
			delete this.scriptParms.actions[idx]['when'];
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
		break;
	case 'log':
		if (action.message == '') { errors.push("Log Message cannot be null"); }
		newAction['log'].message = action.message;
		break
	case 'assert':
		if (action.expression == '') { errors.push("Assert Expression cannot be null"); }
		newAction['assert'].expression = action.expression;
		break
	case 'sleep':
		if (action.duration <= 0) { errors.push("Duration cannot be negative or null"); }
		newAction['sleep'].duration = action.duration;
		break;
	case 'setvar':
		if (action.variable == '') { errors.push("Setvar Variable Name cannot be null"); }
		if (action.expression == '') { errors.push("Setvar Expression cannot be null"); }
		newAction['setvar'].variable = action.variable;
		newAction['setvar'].expression = action.expression;
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
		return 'SETVAR"' + action.setvar.variable + '"';
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

