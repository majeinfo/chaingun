Vue.component('action-sleep', {
  data: function () {
    return {
    }
  },
  props: ['duration', 'mode', "errors", "action_index"],
  template: ` 
        <div class="modal" tabindex="-1" role="dialog" id="new_sleep">
                <div class="modal-dialog modal-dialog-centered" role="document">
                        <div class="modal-content">
                                <div class="modal-header">
                                        <h5 class="modal-title">{{ mode }} SLEEP Action</h5>
                                        <button type="button" class="close" data-dismiss="modal" aria-label="Close"><span aria-hidden="true">&times;</span></button>
                                </div>
                                <div class="modal-body">
                                      <p v-if="errors.length">
                                        <div v-for="error in errors" class="alert alert-danger alert-dismissible fade show" role="alert">
                                          {{ error }}
                                        </div>
                                      </p>
					<div class="form-group row required">
                                                <label for="actionDuration" class="col-sm-4 col-form-label control-label">Duration (in ms)</label>
                                                <input type="number" class="col-sm-4 form-control" id="actionDuration" v-bind:value="duration" v-on:input="$root.$emit('change_duration', $event.target.value)" placeholder="Enter the Action Duration">
                                                &nbsp;<img src="img/info-circle-fill.svg" alt="" width="24" height="24" data-toggle="tooltip" title="(mandatory) integer that gives the sleep time in milliseconds">
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
