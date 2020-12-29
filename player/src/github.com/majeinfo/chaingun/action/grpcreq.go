package action

import (
	_ "bytes"
	"context"
	"encoding/json"
	"fmt"
	//"google.golang.org/grpc/encoding/gzip"   # pour la compression
	"google.golang.org/grpc/metadata"
	"math"
	_ "strconv"
	"strings"
	_ "strings"
	"time"

	"github.com/majeinfo/chaingun/config"
	"github.com/majeinfo/chaingun/reporter"
	"github.com/majeinfo/chaingun/utils"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"github.com/golang/protobuf/jsonpb"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/jhump/protoreflect/dynamic/grpcdynamic"
	"google.golang.org/grpc/keepalive"
	"github.com/gogo/protobuf/proto"
)

const (
	REPORTER_GRPC string = "gRPC"
)

// GRPCrequest describes a GRPC Request
type GRPCRequest struct {
	Title string
	Stub  grpcdynamic.Stub
	Call  string
	Data  string
	Func  *desc.MethodDescriptor
}

// DoGRPCRequest accepts a GRPCAction and a one-way channel to write the results to.
func DoGRPCRequest(grpcAction GRPCAction, resultsChannel chan reporter.SampleReqResult, sessionMap map[string]string, vucontext *config.VUContext, vulog *log.Entry, playbook *config.TestDef) bool {
	var trace_req string
	sampleReqResult := buildSampleResult(REPORTER_GRPC, sessionMap["UID"], 0, reporter.NETWORK_ERROR, 0, grpcAction.Title, "")

	req, err := buildGRPCRequest(grpcAction, sessionMap, vulog)
	if err != nil {
		vulog.Error(err)
		return false
	}

	if must_trace_request {
		trace_req = fmt.Sprintf("%s %s", req.Call, req.Data)
	} else {
		vulog.Debugf("New Request: Call: %s, Data: %s", req.Call, req.Data)
	}

	// Try to substitute the server name by an IP address
	server := playbook.DfltValues.Server
	if !disable_dns_cache {
		if addr, status := utils.GetServerAddress(server); status == true {
			server = addr
		}
	}

	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())

	/*
	sh := &statsHandler{
		id:      len(b.handlers),
		results: b.results,
		hasLog:  b.config.hasLog,
		log:     b.config.log,
	}

	b.handlers = append(b.handlers, sh)
	opts = append(opts, grpc.WithStatsHandler(sh))
	 */

	opts = append(opts, grpc.WithKeepaliveParams(keepalive.ClientParameters{
		Time:    time.Duration(playbook.Timeout) * time.Second,
		Timeout: time.Duration(playbook.Timeout) * time.Second,
	}))


	// increase max receive and send message sizes
	opts = append(opts,
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(math.MaxInt32),
			grpc.MaxCallSendMsgSize(math.MaxInt32),
		))

	ctx := context.Background()
	ctx, _ = context.WithTimeout(ctx, time.Duration(playbook.Timeout) * time.Second)
	// cancel is ignored here as connection.Close() is used.
	// See https://godoc.org/google.golang.org/grpc#DialContext

	start := time.Now()

	// create client connection
	conn, err := grpc.DialContext(ctx, server, opts...)
	defer conn.Close()
	if err != nil {
		if must_trace_request {
			vulog.Infof("%s: FAILED (%s)", trace_req, err)
		}
		vulog.Errorf("GRPC connection failed: %s", err)
		buildGRPCSampleResult(&sampleReqResult, 0, reporter.NETWORK_ERROR, 0, err.Error())
		resultsChannel <- sampleReqResult
		return false
	}

	req.Stub = grpcdynamic.NewStub(conn)

	// Unary request
	var inputs []*dynamic.Message
	if inputs, err = getMessages(req, []byte(req.Data)); err != nil {
		vulog.Error(err)
		return false
	}
	_ = makeUnaryRequest(&ctx, req, nil, inputs[0], vulog)

	elapsed := time.Since(start)
	/*
	sessionMap[config.HTTP_RESPONSE] = strconv.Itoa(resp.StatusCode)
	elapsed := time.Since(start)
	responseBody, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		if must_trace_request {
			vulog.Infof("%s: FAILED (%s)", trace_req, err)
		}
		vulog.Printf("Reading HTTP response failed: %s", err)
		buildHTTPSampleResult(&sampleReqResult, 0, resp.StatusCode, elapsed.Nanoseconds(), req.URL.String())
		resultsChannel <- sampleReqResult
		return false
	}

	if must_trace_request {
		vulog.Infof("%s; RetCode=%d; RcvdBytes=%d", trace_req, resp.StatusCode, len(responseBody))
	}
	if must_display_srv_resp {
		vulog.Debugf("[GRPC Response=%d] Received data: %s", resp.StatusCode, responseBody)
	}
	*/
	valid := true

	/*
	// If the HTTP response code is listed in "http_error_codes" (404, 403, 500...),
	// the result is not processed and a false value is returned
	if strings.Contains(playbook.HttpErrorCodes, strconv.FormatInt(int64(resp.StatusCode), 10)) {
		vulog.Errorf("HTTP response code is considered as an error: %d", resp.StatusCode)
		valid = false
	}

	// if action specifies response action, parse using regexp/jsonpath
	if valid && !processResult(grpcAction.ResponseHandlers, sessionMap, vulog, responseBody, resp.Header) {
		valid = false
	}
	 */

	//buildHTTPSampleResult(&sampleReqResult, len(responseBody), resp.StatusCode, elapsed.Nanoseconds(), req.Call)
	buildHTTPSampleResult(&sampleReqResult, 9999, 999, elapsed.Nanoseconds(), req.Call)
	resultsChannel <- sampleReqResult

	return valid
}

func buildGRPCRequest(grpcAction GRPCAction, sessionMap map[string]string, vulog *log.Entry) (*GRPCRequest, error) {
	req := &GRPCRequest{
		Title: grpcAction.Title,
		//Stub: nil,
		Call: grpcAction.Call,
		Data: grpcAction.Data,
		Func: grpcAction.Func,
	}
	//var err error
	vulog.Debug("buildGRPCRequest")

	/*
	// Hack: the Path has been concatened with EscapedPath() (from net/url.go)
	// We must re-convert strings like $%7Bxyz%7D into ${xyz} to make variable substitution work !
	unescapedURL := RedecodeEscapedPath(httpAction.URL)

	if err != nil {
		err := fmt.Errorf("http.newRequest failed in buildHttpRequest: %s", err)
		return nil, err
	}
*/
	return req, nil
}

func getMessages(req *GRPCRequest, inputData []byte) ([]*dynamic.Message, error) {
	var inputs []*dynamic.Message

	inputs, err := createPayloadsFromJSON(string(inputData), req.Func)
	if err != nil {
		return nil, err
	}

	return inputs, nil
}

func createPayloadsFromJSON(data string, mtd *desc.MethodDescriptor) ([]*dynamic.Message, error) {
	md := mtd.GetInputType()
	var inputs []*dynamic.Message

	if len(data) > 0 {
		if strings.IndexRune(data, '[') == 0 {
			dataArray := make([]map[string]interface{}, 5)
			err := json.Unmarshal([]byte(data), &dataArray)
			if err != nil {
				return nil, fmt.Errorf("Error unmarshalling gRPC payload. Data: '%v' Error: %v", data, err.Error())
			}

			elems := len(dataArray)
			if elems > 0 {
				inputs = make([]*dynamic.Message, elems)
			}

			for i, elem := range dataArray {
				elemMsg := dynamic.NewMessage(md)
				err := messageFromMap(elemMsg, &elem)
				if err != nil {
					return nil, fmt.Errorf("Error creating gRPC message: %v", err.Error())
				}

				inputs[i] = elemMsg
			}
		} else {
			inputs = make([]*dynamic.Message, 1)
			inputs[0] = dynamic.NewMessage(md)
			err := jsonpb.UnmarshalString(data, inputs[0])
			if err != nil {
				return nil, fmt.Errorf("Error creating gRPC message from data. Data: '%v' Error: %v", data, err.Error())
			}
		}
	}

	return inputs, nil
}

// creates a message from a map
// marshal to JSON then use jsonpb to marshal to message
// this way we follow protobuf more closely and allow camelCase properties.
func messageFromMap(input *dynamic.Message, data *map[string]interface{}) error {
	strData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	err = jsonpb.UnmarshalString(string(strData), input)
	if err != nil {
		return err
	}

	return nil
}

func makeUnaryRequest(ctx *context.Context, req *GRPCRequest, reqMD *metadata.MD, input *dynamic.Message, vulog *log.Entry) error {
	var res proto.Message
	var resErr error
	var callOptions = []grpc.CallOption{}

	/* TODO: enable compression
	if w.config.enableCompression {
		callOptions = append(callOptions, grpc.UseCompressor(gzip.Name))
	}
	*/

	res, resErr = req.Stub.InvokeRpc(*ctx, req.Func, input, callOptions...)

	vulog.Debug("Received response from call type: unary",
		", call", req.Func.GetFullyQualifiedName(),
		", input", input, "metadata", reqMD,
		", response", res, "error", resErr)

	return resErr
}

func buildGRPCSampleResult(sample *reporter.SampleReqResult, contentLength int, status int, elapsed int64, fullreq string) {
	sample.Status = status
	sample.Size = contentLength
	sample.Latency = elapsed
	sample.FullRequest = fullreq
}
