Vue.component('action-tcp', {
  data: function () {
    return {
    }
  },
  props: ['title', 'address', 'payload', 'mode', "errors", "action_index"],
  template: ` 
        <div class="modal" tabindex="-1" role="dialog" id="new_tcp">
                <div class="modal-dialog modal-dialog-centered" role="document">
                        <div class="modal-content">
                                <div class="modal-header">
                                        <h5 class="modal-title">{{ mode }} TCP Action</h5>
                                        <button type="button" class="close" data-dismiss="modal" aria-label="Close"><span aria-hidden="true">&times;</span></button>
                                </div>
                                <div class="modal-body">
                                      <p v-if="errors.length">
                                        <div v-for="error in errors" class="alert alert-danger alert-dismissible fade show" role="alert">
                                          {{ error }}
                                        </div>
                                      </p>
                                        <div class="form-group row required">
                                                <label for="variable" class="col-sm-3 col-form-label control-label">Title</label>
                                                <input type="text" class="col-sm-8 form-control" id="title" v-bind:value="title" v-on:input="$root.$emit('change_title', $event.target.value)" placeholder="Enter the Title">
                                                &nbsp;<img src="img/info-circle-fill.svg" alt="" width="24" height="24" data-toggle="tooltip" title="(mandatory) string that qualifies the request - used for the result output and logging">
                                        </div>
                                        <div class="form-group row required">
                                                <label for="address" class="col-sm-3 col-form-label control-label">Address</label>
                                                <input type="text" class="col-sm-8 form-control" id="address" v-bind:value="address" v-on:input="$root.$emit('change_address', $event.target.value)" placeholder="Enter the Address">
                                                &nbsp;<img src="img/info-circle-fill.svg" alt="" width="24" height="24" data-toggle="tooltip" title="(mandatory) string that indicates the server address and the port to connect to">
                                        </div>
                                        <div class="form-group row required">
                                                <label for="payload" class="col-sm-3 col-form-label control-label">Payload</label>
                                                <input type="text" class="col-sm-8 form-control" id="payload" v-bind:value="payload" v-on:input="$root.$emit('change_payload', $event.target.value)" placeholder="Enter the Payload">
                                                &nbsp;<img src="img/info-circle-fill.svg" alt="" width="24" height="24" data-toggle="tooltip" title="(mandatory) string to send as the payload. If you want to specify special characters (like \n), do not forget to enclose the string in double-quotes">
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
