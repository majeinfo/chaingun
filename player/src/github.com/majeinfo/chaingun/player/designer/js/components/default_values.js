Vue.component('form-dflt-values', {
  data: function () {
    return {
    /*
      feeder_type: 'csv',
      filename: '',
      separator: ',',
    */
    }
  },
  props: ['server', 'protocol', 'method', 'database', 'collection', 'db_driver'],
  template: ` 
      <div class="tab-pane fade m-2" id="v-pills-dv" role="tabpanel" aria-labelledby="v-pills-dv-tab">
		<div class="form-group row">
			<label for="server" class="col-sm-3 col-form-label">Default Server</label>
			<input type="text" class="col-sm-4 form-control" id="dfltServer" v-bind:value="server" v-on:input="$root.$emit('change_default_server', $event.target.value)" placeholder="Enter the default server name">
			&nbsp;<img src="img/info-circle-fill.svg" alt="" width="24" height="24" data-toggle="tooltip" title="name of remoter server - may also specify a port, for SQL this a DSN (examples: www.google.com:80 or www.bing.com or mongodb://localhost:27017)">
		</div>
		<div class="form-group row">
			<label for="protocol" class="col-sm-3 col-form-label">Default Protocol</label>
			<select id="protocol" class="col-sm-2 form-control" v-bind:value="protocol" v-on:input="$root.$emit('change_default_protocol', $event.target.value)" placeholder="Protocol">
				<option selected>http</option>
				<option>https</option>
			</select>
			&nbsp;<img src="img/info-circle-fill.svg" alt="" width="24" height="24" data-toggle="tooltip" title="protocol to be used (http or https)">
		</div>
		<div class="form-group row">
			<label for="method" class="col-sm-3 col-form-label">Default Method</label>
			<select id="method" class="col-sm-2 form-control" v-bind:value="method" v-on:input="$root.$emit('change_default_method', $event.target.value)" placeholder="Method">
				<option></option>
				<option>GET</option>
				<option>POST</option>
				<option>PUT</option>
				<option>HEAD</option>
				<option>DELETE</option>
			</select>
			&nbsp;<img src="img/info-circle-fill.svg" alt="" width="24" height="24" data-toggle="tooltip" title="HTTP method to use">
		</div>
		<div class="form-group row">
			<label for="database" class="col-sm-3 col-form-label">Default Database</label>
			<input type="text" class="col-sm-4 form-control" id="dfltDatabase" v-bind:value="database" v-on:input="$root.$emit('change_default_database', $event.target.value)" placeholder="Enter the default DB Name">
			&nbsp;<img src="img/info-circle-fill.svg" alt="" width="24" height="24" data-toggle="tooltip" title="default database for MongoDB and SQL">
		</div>
		<div class="form-group row">
			<label for="collection" class="col-sm-3 col-form-label">Default Collection</label>
			<input type="text" class="col-sm-4 form-control" id="dfltCollection" v-bind:value="collection" v-on:input="$root.$emit('change_default_collection', $event.target.value)" placeholder="Enter the default DB Name">
			&nbsp;<img src="img/info-circle-fill.svg" alt="" width="24" height="24" data-toggle="tooltip" title="default collection for MongoDB">
		</div>
		<div class="form-group row">
			<label for="db_driver" class="col-sm-3 col-form-label">Default DB Driver</label>
			<input type="text" class="col-sm-4 form-control" id="dfltDBDriver" v-bind:value="db_driver" v-on:input="$root.$emit('change_default_db_driver', $event.target.value)" placeholder="Enter the default DB Driver">
			&nbsp;<img src="img/info-circle-fill.svg" alt="" width="24" height="24" data-toggle="tooltip" title="default SQL Driver - only 'mysql' is supported yet">
		</div>
      </div>
  `
})
