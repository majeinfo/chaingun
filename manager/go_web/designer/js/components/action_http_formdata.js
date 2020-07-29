Vue.component('action-http-formdata', {
  data: function () {
    return {
    }
  },
  props: ['formdata_name', 'formdata_value', 'formdata_type', 'mode', "errors", "action_index"],
  template: ` 
        <div class="modal" tabindex="-1" role="dialog" id="new_formdata">
                <div class="modal-dialog modal-dialog-centered" role="document">
                        <div class="modal-content">
                                <div class="modal-header">
                                        <h5 class="modal-title">{{ mode }} Form Data </h5>
                                        <button type="button" class="close" data-dismiss="modal" aria-label="Close"><span aria-hidden="true">&times;</span></button>
                                </div>
                                <div class="modal-body">
                                        <p v-if="errors.length">
                                          <div v-for="error in errors" class="alert alert-danger alert-dismissible fade show" role="alert">
                                            {{ error }}
                                          </div>
                                        </p>
                                        <div class="form-group row required">
                                                <label for="dataname" class="col-sm-4 col-form-label control-label">Data Name</label>
                                                <input type="text" class="col-sm-6 form-control" id="dataname" v-bind:value="formdata_name" v-on:input="$root.$emit('change_http_formdata_name', $event.target.value)" placeholder="Enter the Data Name">
                                                &nbsp;<img src="img/info-circle-fill.svg" alt="" width="24" height="24" data-toggle="tooltip" title="(mandatory) Enter the Data Name">
                                        </div>
                                        <div class="form-group row required">
                                                <label for="datavalue" class="col-sm-4 col-form-label control-label">Data Value</label>
                                                <input type="text" class="col-sm-6 form-control" id="datavalue" v-bind:value="formdata_value" v-on:input="$root.$emit('change_http_formdata_value', $event.target.value)" placeholder="Enter the HTTP Header Value">
                                                &nbsp;<img src="img/info-circle-fill.svg" alt="" width="24" height="24" data-toggle="tooltip" title="(mandatory) Enter the Data Value">
                                        </div>
                                        <div class="form-group row">
                                                <label for="datavalue" class="col-sm-4 col-form-label control-label">Data Type</label>
						<select id="index" class="col-sm-2 form-control" v-bind:value="formdata_type" v-on:input="$root.$emit('change_http_formdata_type', $event.target.value)" placeholder="">
							<option selected></option>
							<option>file</option>
						</select>
                                                &nbsp;<img src="img/info-circle-fill.svg" alt="" width="24" height="24" data-toggle="tooltip" title="Must be set to 'file' if the data is a file !">
                                        </div>
                                </div>
                                <div class="modal-footer">
                                        <input type="hidden" v-model:value="action_index">
                                        <button type="button" class="btn btn-secondary" data-dismiss="modal" v-on:click="$root.$emit('clear_formdata')">Close</button>
                                        <button type="button" class="btn btn-primary" v-on:click="$root.$emit('new_formdata')">Save changes</button>
                                </div>
                        </div>
                </div>
        </div>
  `
})
