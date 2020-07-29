Vue.component('form-response', {
  data: function () {
    return {
    }
  },
  props: ["response_index", "mode", "errors", "from_header", "regex", "jsonpath", "xmlpath", "variable", "index", "default_value"],
  template: ` 
        <div class="modal" tabindex="-1" role="dialog" id="new_response">
                <div class="modal-dialog modal-dialog-centered modal-lg" role="document">
                        <div class="modal-content">
                                <div class="modal-header">
                                        <h5 class="modal-title">{{ mode }} Response </h5>
                                        <button type="button" class="close" data-dismiss="modal" aria-label="Close"><span aria-hidden="true">&times;</span></button>
                                </div>
                                <div class="modal-body">
                                      <p v-if="errors.length">
                                        <div v-for="error in errors" class="alert alert-danger alert-dismissible fade show" role="alert">
                                          {{ error }}
                                        </div>
                                      </p>
                                        <div class="form-group row">
                                                <label for="from_header" class="col-sm-3 col-form-label control-label">From Header ?</label>
                                                <input type="text" class="col-sm-8 form-control" id="from_header" v-bind:value="from_header" v-on:input="$root.$emit('change_from_header', $event.target.value)" placeholder="Enter the Header Name">
                                                &nbsp;<img src="img/info-circle-fill.svg" alt="" width="24" height="24" data-toggle="tooltip" title="Gives the name of the HTTP Header to get value from, if empty, values are extracted from the Body">
                                        </div>
                                        <div class="form-group row">
                                                <label for="regex" class="col-sm-3 col-form-label control-label">Regex</label>
                                                <input type="text" class="col-sm-8 form-control" id="regex" v-bind:value="regex" v-on:input="$root.$emit('change_regex', $event.target.value)" placeholder="Enter the Regex">
                                                &nbsp;<img src="img/info-circle-fill.svg" alt="" width="24" height="24" data-toggle="tooltip" title="Gives the value of the Regex that will be applied to extract the value">
                                        </div>
                                        <div class="form-group row">
                                                <label for="jsonpath" class="col-sm-3 col-form-label control-label">or JSON Path</label>
                                                <input type="text" class="col-sm-8 form-control" id="regex" v-bind:value="jsonpath" v-on:input="$root.$emit('change_jsonpath', $event.target.value)" placeholder="Enter the JSON Path">
                                                &nbsp;<img src="img/info-circle-fill.svg" alt="" width="24" height="24" data-toggle="tooltip" title="Gives the value of the JSON Path that will be applied to extract the value">
                                        </div>
                                        <div class="form-group row">
                                                <label for="xmlpath" class="col-sm-3 col-form-label control-label">or XML Path</label>
                                                <input type="text" class="col-sm-8 form-control" id="xmlpath" v-bind:value="xmlpath" v-on:input="$root.$emit('change_xmlpath', $event.target.value)" placeholder="Enter the XML Path">
                                                &nbsp;<img src="img/info-circle-fill.svg" alt="" width="24" height="24" data-toggle="tooltip" title="Gives the value of the XML Path that will be applied to extract the value">
                                        </div>
                                        <div class="form-group row required">
                                                <label for="regex" class="col-sm-3 col-form-label control-label">Variable</label>
                                                <input type="text" class="col-sm-8 form-control" id="variable" v-bind:value="variable" v-on:input="$root.$emit('change_variable', $event.target.value)" placeholder="Enter the Variable Name">
                                                &nbsp;<img src="img/info-circle-fill.svg" alt="" width="24" height="24" data-toggle="tooltip" title="Gives the name of the Variable that will receive the result">
                                        </div>
					<div class="form-group row">
						<label for="index" class="col-sm-3 col-form-label">Matching Selection</label>
						<select id="index" class="col-sm-2 form-control" v-bind:value="index" v-on:input="$root.$emit('change_index', $event.target.value)" placeholder="">
							<option selected>first</option>
							<option>last</option>
							<option>random</option>
						</select>
						&nbsp;<img src="img/info-circle-fill.svg" alt="" width="24" height="24" data-toggle="tooltip" title="Select the good result when many values match">
					</div>
                                        <div class="form-group row">
                                                <label for="dflt_value" class="col-sm-3 col-form-label control-label">Default Value</label>
                                                <input type="text" class="col-sm-8 form-control" id="dflt_value" v-bind:value="default_value" v-on:input="$root.$emit('change_default_value', $event.target.value)" placeholder="Enter the Default Value">
                                                &nbsp;<img src="img/info-circle-fill.svg" alt="" width="24" height="24" data-toggle="tooltip" title="Give the Default Value in case the expression does not match - the default value is 'first'">
                                        </div>
                                </div>
                                <div class="modal-footer">
                                        <input type="hidden" v-model:value="response_index">
                                        <button type="button" class="btn btn-secondary" data-dismiss="modal" v-on:click="$root.$emit('clear_response')">Close</button>
                                        <button type="button" class="btn btn-primary" v-on:click="$root.$emit('new_response')">Save changes</button>
                                </div>
                        </div>
                </div>
        </div>
  `
})
