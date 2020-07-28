Vue.component('form-feeder', {
  data: function () {
    return {
    /*
      feeder_type: 'csv',
      filename: '',
      separator: ',',
    */
    }
  },
  props: ['feeder_type', 'filename', 'separator'],
  template: ` 
    <div class="tab-pane fade m-2" id="v-pills-fd" role="tabpanel" aria-labelledby="v-pills-fd-tab">
	<div class="form-group row">
		<label for="feederType" class="col-sm-3 col-form-label">Feeder Type</label>
		<input disabled type="text" class="col-sm-4 form-control" id="feederType" v-bind:value="feeder_type" v-on:input="$root.$emit('change_feeder_type', $event.target.value)" placeholder="Choose the Feeder Type" value="csv">
		&nbsp;<img src="img/info-circle-fill.svg" alt="" width="24" height="24" data-toggle="tooltip" title="(mandatory) file type (only 'csv' is supported)">
	</div>
	<div class="form-group row">
		<label for="filename" class="col-sm-3 col-form-label">Filename</label>
		<input type="text" class="col-sm-4 form-control" id="filename" v-bind:value="filename" v-on:input="$root.$emit('change_feeder_filename', $event.target.value)" placeholder="Enter the Filename">
		&nbsp;<img src="img/info-circle-fill.svg" alt="" width="24" height="24" data-toggle="tooltip" title="(mandatory) string that gives the filename">
	</div>
	<div class="form-group row">
		<label for="separator" class="col-sm-3 col-form-label">Field Separator</label>
		<input type="text" class="col-sm-4 form-control" id="separator" v-bind:value="separator" v-on:input="$root.$emit('change_feeder_separator', $event.target.value)" placeholder="Enter the Field Separator">
		&nbsp;<img src="img/info-circle-fill.svg" alt="" width="24" height="24" data-toggle="tooltip" title="(mandatory) string that gives the column separator">
	</div>
    </div>
  `
})
