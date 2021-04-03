Vue.component('action-mongodb', {
  data: function () {
    return {
    }
  },
  props: ['title', 'server', 'database', 'collection', 'command', 'filter', 'document', 'responses', 'mode', "errors", "errors2", "action_index"],
  template: ` 
        <div class="modal overflow-auto" tabindex="-1" role="dialog" id="new_mongodb">
                <div class="modal-dialog modal-dialog-centered modal-lg" role="document">
                        <div class="modal-content">
                                <div class="modal-header">
                                        <h5 class="modal-title">{{ mode }} MONGODB Action</h5>
                                        <button type="button" class="close" data-dismiss="modal" aria-label="Close"><span aria-hidden="true">&times;</span></button>
                                </div>
                                <div class="modal-body">
                                      <p v-if="errors.length">
                                        <div v-for="error in errors" class="alert alert-danger alert-dismissible fade show" role="alert">
                                          {{ error }}
                                        </div>
                                      </p>
                                        <div class="form-group row required">
                                                <label for="title" class="col-sm-3 col-form-label control-label">Title</label>
                                                <input type="text" class="col-sm-8 form-control" id="title" v-bind:value="title" v-on:input="$root.$emit('change_title', $event.target.value)" placeholder="Enter the Title">
                                                &nbsp;<img src="img/info-circle-fill.svg" alt="" width="24" height="24" data-toggle="tooltip" title="mandatory string that qualifies the request - used for the result output and logging">
                                        </div>
                                        <div class="form-group row">
                                                <label for="server" class="col-sm-3 col-form-label control-label">Server</label>
                                                <input type="text" class="col-sm-8 form-control" id="server" v-bind:value="server" v-on:input="$root.$emit('change_server', $event.target.value)" placeholder="Enter the Server name">
                                                &nbsp;<img src="img/info-circle-fill.svg" alt="" width="24" height="24" data-toggle="tooltip" title="(mandatory) If the string does not contain a server (DSN) specification, use the value given by the server key in the default section">
                                        </div>
                                        <div class="form-group row">
                                                <label for="database" class="col-sm-3 col-form-label control-label">Database</label>
                                                <input type="text" class="col-sm-8 form-control" id="database" v-bind:value="database" v-on:input="$root.$emit('change_database', $event.target.value)" placeholder="Enter the Database name">
                                                &nbsp;<img src="img/info-circle-fill.svg" alt="" width="24" height="24" data-toggle="tooltip" title="(mandatory) If the string is empty, use the value given by the database key in the default section">
                                        </div>
                                        <div class="form-group row">
                                                <label for="collection" class="col-sm-3 col-form-label control-label">Collection</label>
                                                <input type="text" class="col-sm-8 form-control" id="database" v-bind:value="collection" v-on:input="$root.$emit('change_collection', $event.target.value)" placeholder="Enter the Collection name">
                                                &nbsp;<img src="img/info-circle-fill.svg" alt="" width="24" height="24" data-toggle="tooltip" title="(mandatory) If the string is empty, use the value given by the collection key in the default section">
                                        </div>
					<div class="form-group row" required>
						<label for="index" class="col-sm-3 col-form-label">Command</label>
						<select id="index" class="col-sm-2 form-control" v-bind:value="command" v-on:input="$root.$emit('change_command', $event.target.value)" placeholder="">
							<option selected>findone</option>
							<option>insertone</option>
							<option>drop</option>
						</select>
						&nbsp;<img src="img/info-circle-fill.svg" alt="" width="24" height="24" data-toggle="tooltip" title="(mandatory) Possible commands are findone, insertone and drop">
					</div>
                                        <div class="form-group row">
                                                <label for="filter" class="col-sm-3 col-form-label control-label">Filter JSON Document</label>
                                                <input type="text" class="col-sm-8 form-control" id="filter" v-bind:value="filter" v-on:input="$root.$emit('change_filter', $event.target.value)" placeholder="Enter the JSON doc for filtering">
                                                &nbsp;<img src="img/info-circle-fill.svg" alt="" width="24" height="24" data-toggle="tooltip" title="If the command is findone, the filter parameter is a JSON document used to filter the search">
                                        </div>
                                        <div class="form-group row">
                                                <label for="document" class="col-sm-3 col-form-label control-label">Insert JSON Document</label>
                                                <input type="text" class="col-sm-8 form-control" id="document" v-bind:value="document" v-on:input="$root.$emit('change_document', $event.target.value)" placeholder="Enter the JSON doc to be inserted">
                                                &nbsp;<img src="img/info-circle-fill.svg" alt="" width="24" height="24" data-toggle="tooltip" title="If the command is insertone, the document parameter is a JSON document that must be inserted in the database collection">
                                        </div>
					<list-responses
						v-bind:mode="mode"
						v-bind:errors="errors2"
						v-bind:action_index="action_index"
						v-bind:responses="responses">
					</list-responses>
                                </div>
                                <div class="modal-footer">
                                        <input type="hidden" v-model:value="action_index">
                                        <button type="button" class="btn btn-secondary" data-dismiss="modal" v-on:click="$root.$emit('clear_action')">Close</button>
                                        <button type="button" class="btn btn-primary" v-on:click="$root.$emit('new_action')">Save changes</button>
                                </div>
                        </div>
                </div>
        </div>
  `
})
