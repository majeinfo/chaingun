Vue.component('list-responses', {
  data: function () {
    return {
    }
  },
  props: ["responses", "action_index", "mode", "errors"],
  template: ` 
<div>
	<div class="form-group row">
		<label for="responses" class="col-sm-3 col-form-label control-label">Responses</label>
		<div>
			<button type="button" class="btn btn-sm btn-primary" data-toggle="modal" xdata-target="#newResponse" onClick="chaingunScript.responseShow()">Add a new Response</button>
			&nbsp;<img src="img/info-circle-fill.svg" alt="" width="24" height="24" data-toggle="tooltip" title="data can be extracted from server responses. The extraction can use the body or a HTTP Header. regex, jsonpath or xmlpath can be used to collect the substrings">
		</div>
	</div>
	<div class="form-group row">
		<label for="responses" class="col-sm-3 col-form-label control-label"></label>
		<div class="col-sm-8">
			<div class="border row mb-1" v-for="(resp, index) in responses">
				<div v-if="resp.regex != '' && resp.regex != undefined" class="col">
					Use a Regex to set value in variable {{ resp.variable }}
				</div>
				<div v-else-if="resp.jsonpath != '' && resp.jsonpath != undefined" class="col">
					Use a JSON Path to set value in variable {{ resp.variable }}
				</div>
				<div v-else-if="resp.xmlpath != '' && resp.xmlpath != undefined" class="col">
					Use a XML Path to set value in variable {{ resp.variable }}
				</div>
				<div>
					<img src="img/pencil.svg" height=20 width=20 v-bind:onclick="'chaingunScript.displayForEditResponse(' + parseInt(index, 10) + '); return false'"/>&nbsp;
					<img src="img/circle-x.svg" height=20 width=20 v-bind:onclick="'chaingunScript.deleteResponse(' + parseInt(index, 10) + '); return false'"/>
				</div>
			</div>
		</div>
	</div>
</div>
  `
})
