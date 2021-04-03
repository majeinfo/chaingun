Vue.component('action-setvar', {
  data: function () {
    return {
    }
  },
  props: ['variable', 'expression', 'mode', "errors", "action_index"],
  template: ` 
        <div class="modal" tabindex="-1" role="dialog" id="new_setvar">
                <div class="modal-dialog modal-dialog-centered modal-lg" role="document">
                        <div class="modal-content">
                                <div class="modal-header">
                                        <h5 class="modal-title">{{ mode }} SETVAR Action</h5>
                                        <button type="button" class="close" data-dismiss="modal" aria-label="Close"><span aria-hidden="true">&times;</span></button>
                                </div>
                                <div class="modal-body">
                                      <p v-if="errors.length">
                                        <div v-for="error in errors" class="alert alert-danger alert-dismissible fade show" role="alert">
                                          {{ error }}
                                        </div>
                                      </p>
                                        <div class="form-group row required">
                                                <label for="variable" class="col-sm-3 col-form-label control-label">Variable Name</label>
                                                <input type="text" class="col-sm-8 form-control" id="expression" v-bind:value="variable" v-on:input="$root.$emit('change_variable', $event.target.value)" placeholder="Enter the Variable Name">
                                                &nbsp;<img src="img/info-circle-fill.svg" alt="" width="24" height="24" data-toggle="tooltip" title="(mandatory) variable name">
                                        </div>
                                        <div class="form-group row required">
                                                <label for="expression" class="col-sm-3 col-form-label control-label">Expression</label>
                                                <input type="text" class="col-sm-8 form-control" id="expression" v-bind:value="expression" v-on:input="$root.$emit('change_expression', $event.target.value)" placeholder="Enter the Expression">
                                                &nbsp;<img src="img/info-circle-fill.svg" alt="" width="24" height="24" data-toggle="tooltip" title="(mandatory) string that defines an expression to be evaluated. If the expression is an integer or a float, it will be converted into a string">
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
