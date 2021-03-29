Vue.component('when-clause', {
  data: function () {
    return {
    }
  },
  props: ['when_clause', 'mode', "errors", "action_index"],
  template: ` 
        <div class="modal" tabindex="-1" role="dialog" id="new_when">
                <div class="modal-dialog modal-dialog-centered" role="document">
                        <div class="modal-content">
                                <div class="modal-header">
                                        <h5 class="modal-title">{{ mode }} 'when' clause</h5>
                                        <button type="button" class="close" data-dismiss="modal" aria-label="Close"><span aria-hidden="true">&times;</span></button>
                                </div>
                                <div class="modal-body">
                                        <p v-if="errors.length">
                                          <div v-for="error in errors" class="alert alert-danger alert-dismissible fade show" role="alert">
                                            {{ error }}
                                          </div>
                                        </p>
                                        <div class="form-group row required">
                                                <label for="whenExpr" class="col-sm-4 col-form-label control-label">Expression</label>
                                                <input type="text" class="col-sm-6 form-control" id="whenExpr" v-bind:value="when_clause" v-on:input="$root.$emit('change_when_clause', $event.target.value)" placeholder="Enter the 'when' Expression">
                                                &nbsp;<img src="img/info-circle-fill.svg" alt="" width="24" height="24" data-toggle="tooltip" title="(mandatory) expression that must be evaluated to True to trigger the Action">
                                        </div>
                                </div>
                                <div class="modal-footer">
                                        <input type="hidden" v-model:value="action_index">
                                        <button type="button" class="btn btn-secondary" data-dismiss="modal" v-on:click="$root.$emit('clear_when')">Close</button>
                                        <button type="button" class="btn btn-primary" v-on:click="$root.$emit('new_when')">Save changes</button>
                                </div>
                        </div>
                </div>
        </div>
  `
})
