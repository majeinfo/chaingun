Vue.component('action-sql', {
  data: function () {
    return {
    }
  },
  props: ['title', 'db_driver', 'server', 'database', 'statement', 'mode', "errors", "action_index"],
  template: ` 
        <div class="modal" tabindex="-1" role="dialog" id="new_sql">
                <div class="modal-dialog modal-dialog-centered modal-lg" role="document">
                        <div class="modal-content">
                                <div class="modal-header">
                                        <h5 class="modal-title">{{ mode }} SQL Action</h5>
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
                                                <label for="db_driver" class="col-sm-3 col-form-label control-label">DB Driver</label>
                                                <input type="text" class="col-sm-8 form-control" id="db_driver" v-bind:value="db_driver" v-on:input="$root.$emit('change_db_driver', $event.target.value)" placeholder="Enter the DB Driver (mysql)">
                                                &nbsp;<img src="img/info-circle-fill.svg" alt="" width="24" height="24" data-toggle="tooltip" title="(mandatory) If the string is empty, use the value given by the db_driver key in the default section. Only 'mysql' is supported yet">
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
                                        <div class="form-group row required">
                                                <label for="statement" class="col-sm-3 col-form-label control-label">SQL Statement</label>
                                                <input type="text" class="col-sm-8 form-control" id="statement" v-bind:value="statement" v-on:input="$root.$emit('change_statement', $event.target.value)" placeholder="Enter the SQL Statement">
                                                &nbsp;<img src="img/info-circle-fill.svg" alt="" width="24" height="24" data-toggle="tooltip" title="(mandatory) SQL Statement to execute (CREATE, SELECT, INSERT...)">
                                        </div>
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
