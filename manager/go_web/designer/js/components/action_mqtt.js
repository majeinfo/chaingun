Vue.component('action-mqtt', {
  data: function () {
    return {
    }
  },
  props: ['title', 'url', 'certificatepath', 'privatekeypath', 'username', 'password', 'clientid', 'topic', 'payload', 'qos', 'mode', "errors", "action_index"],
  template: ` 
        <div class="modal" tabindex="-1" role="dialog" id="new_mqtt">
                <div class="modal-dialog modal-dialog-centered modal-lg" role="document">
                        <div class="modal-content">
                                <div class="modal-header">
                                        <h5 class="modal-title">{{ mode }} MQTT Action</h5>
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
                                                <label for="url" class="col-sm-3 col-form-label control-label">URL</label>
                                                <input type="text" class="col-sm-8 form-control" id="url" v-bind:value="url" v-on:input="$root.$emit('change_url', $event.target.value)" placeholder="Enter the URL">
                                                &nbsp;<img src="img/info-circle-fill.svg" alt="" width="24" height="24" data-toggle="tooltip" title="(mandatory) endpoint to contact">
                                        </div>
                                        <div class="form-group row">
                                                <label for="certpath" class="col-sm-3 col-form-label control-label">Certificate Path</label>
                                                <input type="text" class="col-sm-8 form-control" id="certificatepath" v-bind:value="certificatepath" v-on:input="$root.$emit('change_certificatepath', $event.target.value)" placeholder="Enter the Certificate Path">
                                                &nbsp;<img src="img/info-circle-fill.svg" alt="" width="24" height="24" data-toggle="tooltip" title="optional path to the certificate to offer to the server">
                                        </div>
                                        <div class="form-group row">
                                                <label for="privkeypath" class="col-sm-3 col-form-label control-label">Private Key Path</label>
                                                <input type="text" class="col-sm-8 form-control" id="privatekeypath" v-bind:value="privatekeypath" v-on:input="$root.$emit('change_privatekeypath', $event.target.value)" placeholder="Enter the Private Key Path">
                                                &nbsp;<img src="img/info-circle-fill.svg" alt="" width="24" height="24" data-toggle="tooltip" title="optional path to the certificate to offer to the server">
                                        </div>
                                        <div class="form-group row">
                                                <label for="username" class="col-sm-3 col-form-label control-label">Username</label>
                                                <input type="text" class="col-sm-8 form-control" id="username" v-bind:value="username" v-on:input="$root.$emit('change_username', $event.target.value)" placeholder="Enter the Username">
                                                &nbsp;<img src="img/info-circle-fill.svg" alt="" width="24" height="24" data-toggle="tooltip" title="optional username">
                                        </div>
                                        <div class="form-group row">
                                                <label for="password" class="col-sm-3 col-form-label control-label">Password</label>
                                                <input type="text" class="col-sm-8 form-control" id="password" v-bind:value="password" v-on:input="$root.$emit('change_password', $event.target.value)" placeholder="Enter the Password">
                                                &nbsp;<img src="img/info-circle-fill.svg" alt="" width="24" height="24" data-toggle="tooltip" title="optional password">
                                        </div>
                                        <div class="form-group row">
                                                <label for="clientid" class="col-sm-3 col-form-label control-label">Client ID</label>
                                                <input type="text" class="col-sm-8 form-control" id="clientid" v-bind:value="clientid" v-on:input="$root.$emit('change_clientid', $event.target.value)" placeholder="Enter the Client ID">
                                                &nbsp;<img src="img/info-circle-fill.svg" alt="" width="24" height="24" data-toggle="tooltip" title="client name (chaingun-by-JD by default)">
                                        </div>
                                        <div class="form-group row required">
                                                <label for="topic" class="col-sm-3 col-form-label control-label">Topic</label>
                                                <input type="text" class="col-sm-8 form-control" id="topic" v-bind:value="topic" v-on:input="$root.$emit('change_topic', $event.target.value)" placeholder="Enter the Topic Name">
                                                &nbsp;<img src="img/info-circle-fill.svg" alt="" width="24" height="24" data-toggle="tooltip" title="(mandatory) MQTT topic">
                                        </div>
                                        <div class="form-group row required">
                                                <label for="payload" class="col-sm-3 col-form-label control-label">Payload</label>
                                                <input type="text" class="col-sm-8 form-control" id="payload" v-bind:value="payload" v-on:input="$root.$emit('change_payload', $event.target.value)" placeholder="Enter the Payload">
                                                &nbsp;<img src="img/info-circle-fill.svg" alt="" width="24" height="24" data-toggle="tooltip" title="(mandatory) MQTT paylaod, the format depends on the application">
                                        </div>
                                        <div class="form-group row">
                                                <label for="qos" class="col-sm-3 col-form-label control-label">QoS</label>
                                                <input type="number" class="col-sm-2 form-control" id="qos" v-bind:value="qos" v-on:input="$root.$emit('change_qos', $event.target.value)" placeholder="Enter the QoS">
                                                &nbsp;<img src="img/info-circle-fill.svg" alt="" width="24" height="24" data-toggle="tooltip" title="MQTT wanted QoS (default=0, 1 or 2)">
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
