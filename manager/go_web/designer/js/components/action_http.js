Vue.component('action-http', {
  data: function () {
    return {
    }
  },
  props: ['title', "url", "method", "use_http2", "store_cookie", "body", "template", "upload_file", "headers", "responses", "errors", "errors2", "action_index", "mode"],
  template: ` 
        <div class="modal overflow-auto" tabindex="-1" role="dialog" id="new_http">
                <div class="modal-dialog modal-lg" role="document">
                        <div class="modal-content">
                                <div class="modal-header">
                                        <h5 class="modal-title">{{ mode }} HTTP Action</h5>
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
                                                <label for="actionURL" class="col-sm-3 col-form-label control-label">URL</label>
                                                <input type="text" class="col-sm-8 form-control" id="url" v-bind:value="url" v-on:input="$root.$emit('change_url', $event.target.value)" placeholder="Enter the URL" required>
                                                &nbsp;<img src="img/info-circle-fill.svg" alt="" width="24" height="24" data-toggle="tooltip" title="(mandatory) If the string does not contain a server specification, use the value given by the server key in the default section">
                                        </div>
                                        <div class="form-group row">
                                                <label for="actionMethod" class="col-sm-3 col-form-label control-label">Method</label>
                                                <select id="actionMethod" class="col-sm-3 form-control" v-bind:value="method" v-on:input="$root.$emit('change_method', $event.target.value)" placeholder="Method">
                                                        <option></option>
                                                        <option>GET</option>
                                                        <option>POST</option>
                                                        <option>PUT</option>
                                                        <option>HEAD</option>
                                                        <option>DELETE</option>
                                                </select>
                                                &nbsp;<img src="img/info-circle-fill.svg" alt="" width="24" height="24" data-toggle="tooltip" title="GET, PUT, POST, HEAD, DELETE. If absent use the value given by the method key in the default section">
                                        </div>
                                        <div class="form-group row">
                                                <label for="actionCookie" class="col-sm-3 col-form-label control-label">Store Cookie</label>
                                                <input type="text" class="col-sm-8 form-control" id="store_ookie" v-bind:value="store_cookie" v-on:input="$root.$emit('change_store_cookie', $event.target.value)" placeholder="Enter the Cookie Name">
                                                &nbsp;<img src="img/info-circle-fill.svg" alt="" width="24" height="24" data-toggle="tooltip" title="if set, indicates which cookies must be stored in the VU session. The predefined value __all__ implies the capture of all possible cookies">
                                        </div>
                                        <div class="form-group row">
                                                <label for="actionBody" class="col-sm-3 col-form-label control-label">Body</label>
                                                <input type="text" class="col-sm-8 form-control" id="actionBody" v-bind:value="body" v-on:input="$root.$emit('change_body', $event.target.value)" placeholder="Enter the Body">
                                                &nbsp;<img src="img/info-circle-fill.svg" alt="" width="24" height="24" data-toggle="tooltip" title="value of HTTP body (works for any HTTP method) (one of body or template is mandatory)">
                                        </div>
                                        <div class="form-group row">
                                                <label for="actionTemplate" class="col-sm-3 col-form-label control-label">Template Filename</label>
                                                <input type="text" class="col-sm-8 form-control" id="actionTemplate" v-bind:value="template" v-on:input="$root.$emit('change_template', $event.target.value)" placeholder="Enter the Template Filename">
                                                &nbsp;<img src="img/info-circle-fill.svg" alt="" width="24" height="24" data-toggle="tooltip" title="a filename which contents will be interpolated and will be used as the request body (one of body or template is mandatory)">
                                        </div>
                                        <div class="form-group row">
                                                <label for="actionUploadFile" class="col-sm-3 col-form-label control-label">Filename to upload</label>
                                                <input type="text" class="col-sm-8 form-control" id="actionUploadFile" v-bind:value="upload_file" v-on:input="$root.$emit('change_upload_file', $event.target.value)" placeholder="Enter the Filename to upload">
                                                &nbsp;<img src="img/info-circle-fill.svg" alt="" width="24" height="24" data-toggle="tooltip" title="when used with the POST or PUT methods, indicates a file which contents will be sent to the server as-is">
                                        </div>
                                        <div class="form-group row">
                                                <label for="actionMethod" class="col-sm-3 col-form-label control-label">Use HTTP/2</label>
                                                <select id="actionMethod" class="col-sm-3 form-control" v-bind:value="use_http2" v-on:input="$root.$emit('change_use_http2', $event.target.value)" placeholder="Use HTTP/2">
                                                        <option></option>
                                                        <option>false</option>
                                                        <option>true</option>
                                                </select>
                                                &nbsp;<img src="img/info-circle-fill.svg" alt="" width="24" height="24" data-toggle="tooltip" title="set to true if you want to use HTTP/2 protocol (default value is false)">
                                        </div>

                                        <!-- Headers -->
                                        <div class="form-group row">
                                                <label for="headers" class="col-sm-3 col-form-label control-label">HTTP Headers</label>
                                                <div>
                                                        <button type="button" class="btn btn-sm btn-primary" data-toggle="modal" xdata-target="#newHeader" onClick="chaingunScript.headerShow()">Add a new Header</button>
                                                        &nbsp;<img src="img/info-circle-fill.svg" alt="" width="24" height="24" data-toggle="tooltip" title="additional HTTP headers to transmit. Each header has the form header_name: value. In case of a POST method, the body is sent with the HTTP Header content-type: application/x-www-form-urlencoded">
                                                </div>
                                        </div>
                                        <div class="form-group row">
                                                <label for="headers" class="col-sm-3 col-form-label control-label"></label>
                                                <div class="col-sm-9">
                                                        <div class="row mb-1" v-for="(header, index) in headers">
                                                                <div class="col">
                                                                        <input disabled type="text" class="form-control form-inline" id="actionHTTPHeaderKey" v-model="header[0]" placeholder="Enter the HTTP Header Name">
                                                                </div>
                                                                <div class="col">
                                                                        <input disabled type="text" class="form-control form-inline" id="actionHTTPHeaderValue" v-model="header[1]" placeholder="Enter the HTTP Header Value">
                                                                </div>
                                                                <div class="col">
                                                                        <img src="img/pencil.svg" height=20 width=20 v-bind:onclick="'chaingunScript.displayForEditHeader(' + parseInt(index, 10) + '); return false'"/>&nbsp;
                                                                        <img src="img/circle-x.svg" height=20 width=20 v-bind:onclick="'chaingunScript.deleteHeader(' + parseInt(index, 10) + '); return false'"/>
                                                                </div>
                                                        </div>
                                                </div>
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
