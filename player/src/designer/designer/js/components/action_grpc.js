Vue.component('action-grpc', {
    data: function () {
        return {
        }
    },
    props: ['title', "data", "call", "responses", "errors", "errors2", "action_index", "mode"],
    template: ` 
        <div class="modal overflow-auto" tabindex="-1" role="dialog" id="new_grpc">
                <div class="modal-dialog modal-lg" role="document">
                        <div class="modal-content">
                                <div class="modal-header">
                                        <h5 class="modal-title">{{ mode }} gRPC Action</h5>
                                        <button type="button" class="close" data-dismiss="modal" aria-label="Close"><span aria-hidden="true">&times;</span></button>
                                </div>
                                <div class="modal-body">
                                      <p v-if="errors.length">
                                        <div v-for="error in errors" class="alert alert-danger alert-dismissible fade show" role="alert">
                                          {{ error }}
                                          <!--
                                          <button type="button" class="close" data-dismiss="alert" aria-label="Close">
                                            <span aria-hidden="true">&times;</span>
                                          </button>
                                          -->
                                        </div>
                                      </p>
                                      <div class="form-group required row">
                                                <label for="actionTitle" class="col-sm-3 col-form-label control-label">Title</label>
                                                <input type="text" class="col-sm-8 form-control" id="title" v-bind:value="title" v-on:input="$root.$emit('change_title', $event.target.value)" placeholder="Enter the Title" required>
                                                &nbsp;<img src="img/info-circle-fill.svg" alt="" width="24" height="24" data-toggle="tooltip" title="(mandatory) string that qualifies the action - used for the result output and logging">
                                      </div>
                                      <div class="form-group required row">
                                                <label for="actionCall" class="col-sm-3 col-form-label control-label">Function to call</label>
                                                <input type="text" class="col-sm-8 form-control" id="url" v-bind:value="call" v-on:input="$root.$emit('change_call', $event.target.value)" placeholder="Enter the name of the function to call" required>
                                                &nbsp;<img src="img/info-circle-fill.svg" alt="" width="24" height="24" data-toggle="tooltip" title="(mandatory) name of the function to call">
                                      </div>
                                      <div class="form-group row">
                                                <label for="actionData" class="col-sm-3 col-form-label control-label">Data</label>
                                                <input type="text" class="col-sm-8 form-control" id="actionData" v-bind:value="data" v-on:input="$root.$emit('change_data', $event.target.value)" placeholder="Enter the JSON string to send as the payload">
                                                &nbsp;<img src="img/info-circle-fill.svg" alt="" width="24" height="24" data-toggle="tooltip" title="JSON string to send as the payload">
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
