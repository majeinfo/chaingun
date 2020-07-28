Vue.component('action-http-header', {
  data: function () {
    return {
    }
  },
  props: ['header_name', 'header_value', 'mode', "errors", "action_index"],
  template: ` 
        <div class="modal" tabindex="-1" role="dialog" id="new_header">
                <div class="modal-dialog modal-dialog-centered" role="document">
                        <div class="modal-content">
                                <div class="modal-header">
                                        <h5 class="modal-title">{{ mode }} HTTP Header</h5>
                                        <button type="button" class="close" data-dismiss="modal" aria-label="Close"><span aria-hidden="true">&times;</span></button>
                                </div>
                                <div class="modal-body">
                                        <p v-if="errors.length">
                                          <div v-for="error in errors" class="alert alert-danger alert-dismissible fade show" role="alert">
                                            {{ error }}
                                          </div>
                                        </p>
                                        <div class="form-group row required">
                                                <label for="HTTPHeader" class="col-sm-4 col-form-label control-label">HTTP Header Name</label>
                                                <input type="text" class="col-sm-6 form-control" id="httpHeader" v-bind:value="header_name" v-on:input="$root.$emit('change_http_header_name', $event.target.value)" placeholder="Enter the HTTP Header Name">
                                                &nbsp;<img src="img/info-circle-fill.svg" alt="" width="24" height="24" data-toggle="tooltip" title="(mandatory) Enter the HTTP Header full name (like Content-Type)">
                                        </div>
                                        <div class="form-group row required">
                                                <label for="HTTPHeaderValue" class="col-sm-4 col-form-label control-label">HTTP Header Value</label>
                                                <input type="text" class="col-sm-6 form-control" id="httpHeader" v-bind:value="header_value" v-on:input="$root.$emit('change_http_header_value', $event.target.value)" placeholder="Enter the HTTP Header Value">
                                                &nbsp;<img src="img/info-circle-fill.svg" alt="" width="24" height="24" data-toggle="tooltip" title="(mandatory) Enter the HTTP Header value (like application/json)">
                                        </div>
                                </div>
                                <div class="modal-footer">
                                        <input type="hidden" v-model:value="action_index">
                                        <button type="button" class="btn btn-secondary" data-dismiss="modal" v-on:click="$root.$emit('clear_header')">Close</button>
                                        <button type="button" class="btn btn-primary" v-on:click="$root.$emit('new_header')">Save changes</button>
                                </div>
                        </div>
                </div>
        </div>
  `
})
