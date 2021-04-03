Vue.component('action-ws', {
  data: function () {
    return {
    }
  },
  props: ['title', 'url', 'store_cookie', 'body', 'responses', 'mode', "errors", "errors2", "action_index"],
  template: ` 
        <div class="modal overflow-auto" tabindex="-1" role="dialog" id="new_ws">
                <div class="modal-dialog modal-dialog-centered modal-lg" role="document">
                        <div class="modal-content">
                                <div class="modal-header">
                                        <h5 class="modal-title">{{ mode }} WS Action</h5>
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
                                                <label for="url" class="col-sm-3 col-form-label control-label">URL</label>
                                                <input type="text" class="col-sm-8 form-control" id="server" v-bind:value="url" v-on:input="$root.$emit('change_url', $event.target.value)" placeholder="Enter the URL">
                                                &nbsp;<img src="img/info-circle-fill.svg" alt="" width="24" height="24" data-toggle="tooltip" title="(mandatory) If the string does not contain a server specification, use the value given by the server key in the default section">
                                        </div>
                                        <div class="form-group row">
                                                <label for="store_cookie" class="col-sm-3 col-form-label control-label">Store Cookie</label>
                                                <input type="text" class="col-sm-8 form-control" id="store_cookie" v-bind:value="store_cookie" v-on:input="$root.$emit('change_store_cookie', $event.target.value)" placeholder="Enter the Cookie name">
                                                &nbsp;<img src="img/info-circle-fill.svg" alt="" width="24" height="24" data-toggle="tooltip" title="if set, indicates which cookies must be stored in the VU session. The predefined value __all__ implies the capture of all possible cookies">
                                        </div>
                                        <div class="form-group row">
                                                <label for="body" class="col-sm-3 col-form-label control-label">Body</label>
                                                <input type="text" class="col-sm-8 form-control" id="body" v-bind:value="body" v-on:input="$root.$emit('change_body', $event.target.value)" placeholder="Enter the Body">
                                                &nbsp;<img src="img/info-circle-fill.svg" alt="" width="24" height="24" data-toggle="tooltip" title="value of body to send">
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
