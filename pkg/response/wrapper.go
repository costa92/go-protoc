// costa92/go-protoc/go-protoc-acef4f0ceb39155a2d2db028033d358440154368/pkg/response/wrapper.go

package response

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/costa92/go-protoc/pkg/log"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// Wrapper is your desired unified JSON structure
type Wrapper struct {
	Status  string      `json:"status"`            // "success" or "error"
	Data    interface{} `json:"data,omitempty"`    // Business data
	Message string      `json:"message,omitempty"` // Message
	Code    int         `json:"code"`              // Code
}

// CustomMarshaler implements the runtime.Marshaler interface
type CustomMarshaler struct {
	runtime.Marshaler
}

// Marshal wraps the successful proto.Message into the Wrapper struct.
func (c *CustomMarshaler) Marshal(v interface{}) ([]byte, error) {
	// Check if the value is an error, if so, let the error handler deal with it.
	if _, ok := v.(error); ok {
		return c.Marshaler.Marshal(v)
	}

	// Wrap the successful response
	wrappedSuccess := Wrapper{
		Status:  "success",
		Data:    v, // v is the original gRPC response message
		Message: "Request completed successfully",
		Code:    http.StatusOK,
	}

	return json.Marshal(wrappedSuccess)
}

// ForwardResponseMessage is a standard function signature in gRPC-Gateway v2
// for intercepting and customizing the response.
// NOTE: This implementation is kept for reference, but the primary wrapping
// is handled by the Marshal method above for broader compatibility.
func ForwardResponseMessage(ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, req *http.Request, resp proto.Message, opts ...func(context.Context, http.ResponseWriter, proto.Message) error) {
	// This function demonstrates an alternative way to wrap responses.
	// However, using a custom Marshaler is often more straightforward.
	wrappedSuccess := Wrapper{
		Status:  "success",
		Data:    resp,
		Message: "Request completed successfully",
		Code:    http.StatusOK,
	}

	for _, o := range opts {
		if err := o(ctx, w, resp); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", marshaler.ContentType(wrappedSuccess))
	buf, err := marshaler.Marshal(wrappedSuccess)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if _, err := w.Write(buf); err != nil {
		log.Errorf("Failed to write response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// CustomHTTPErrorHandler wraps gRPC errors into the desired uniform JSON format.
func CustomHTTPErrorHandler(ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, req *http.Request, err error) {
	s := status.Convert(err)
	w.Header().Set("Content-Type", "application/json") // Ensure content type is JSON

	wrappedErr := Wrapper{
		Status:  "error",
		Message: s.Message(),
		Data:    nil,
		Code:    runtime.HTTPStatusFromCode(s.Code()),
	}

	buf, _ := json.Marshal(wrappedErr)
	w.WriteHeader(runtime.HTTPStatusFromCode(s.Code()))
	if _, wErr := w.Write(buf); wErr != nil {
		log.Errorf("Failed to write error response: %v", wErr)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
