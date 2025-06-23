package server

import (
	"encoding/json"
	"net/http"

	"github.com/costa92/go-protoc/v2/pkg/api/errno"
	kratoserrors "github.com/go-kratos/kratos/v2/errors"
)

// UnifiedResponse defines the standard JSON response structure for HTTP APIs.
type UnifiedResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// DefaultSuccessMessage is the default message for successful responses.
const DefaultSuccessMessage = "success"

// ErrorCodeMapper defines the interface for mapping error reasons to response codes.
type ErrorCodeMapper interface {
	// MapErrorCode maps a Kratos error reason to a response code.
	// Returns the mapped code and true if mapping exists, otherwise returns 0 and false.
	MapErrorCode(reason string, httpCode int32) (int, bool)
}

// DefaultErrorCodeMapper implements ErrorCodeMapper using errno.ErrorReason_value.
type DefaultErrorCodeMapper struct{}

// MapErrorCode maps error reasons using errno.ErrorReason_value.
func (m *DefaultErrorCodeMapper) MapErrorCode(reason string, httpCode int32) (int, bool) {
	if businessCodeInt32, ok := errno.ErrorReason_value[reason]; ok {
		if businessCodeInt32 != 0 {
			return int(businessCodeInt32), true
		}
	}
	return 0, false
}

// HTTPStatusCodeMapper implements ErrorCodeMapper by using HTTP status codes directly.
type HTTPStatusCodeMapper struct {
	errorReasons map[string]bool // Set of supported error reasons
}

// NewHTTPStatusCodeMapper creates a new HTTPStatusCodeMapper with given error reasons.
func NewHTTPStatusCodeMapper(errorReasons map[string]int32) *HTTPStatusCodeMapper {
	reasons := make(map[string]bool)
	for reason := range errorReasons {
		reasons[reason] = true
	}
	return &HTTPStatusCodeMapper{errorReasons: reasons}
}

// MapErrorCode maps error reasons by using HTTP status codes directly.
func (m *HTTPStatusCodeMapper) MapErrorCode(reason string, httpCode int32) (int, bool) {
	if m.errorReasons[reason] {
		return int(httpCode), true
	}
	return 0, false
}

// Global error code mappers registry
var errorCodeMappers []ErrorCodeMapper

// RegisterErrorCodeMapper registers an error code mapper.
func RegisterErrorCodeMapper(mapper ErrorCodeMapper) {
	errorCodeMappers = append(errorCodeMappers, mapper)
}

// init registers the default error code mapper.
func init() {
	RegisterErrorCodeMapper(&DefaultErrorCodeMapper{})
}

// EncodeResponseFunc is a function that encodes the object v to the ResponseWriter.
// It wraps the original data `v` in a UnifiedResponse structure.
func EncodeResponseFunc(w http.ResponseWriter, r *http.Request, v interface{}) error {
	// Ensure Content-Type is application/json
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	resp := UnifiedResponse{
		Code:    0, // 0 for success as per user requirement
		Message: DefaultSuccessMessage,
		Data:    v,
	}
	return json.NewEncoder(w).Encode(resp)
}

// EncodeErrorFunc is a function that encodes errors to the ResponseWriter.
// It formats errors into the UnifiedResponse structure, ensuring non-zero error codes.
func EncodeErrorFunc(w http.ResponseWriter, r *http.Request, err error) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	kErr := kratoserrors.FromError(err) // Guarantees we have a Kratos error

	responseCode := DefaultErrorCodeForMismatch // Start with a non-zero default for UnifiedResponse.Code
	responseMessage := kErr.Message

	// Try to map the Kratos error reason using registered mappers
	mapped := false
	for _, mapper := range errorCodeMappers {
		if code, ok := mapper.MapErrorCode(kErr.Reason, kErr.Code); ok {
			responseCode = code
			mapped = true
			break
		}
	}

	// Handle special case for errno.Unknown mapping to 0
	if !mapped && kErr.Reason == errno.ErrorReason_Unknown.String() {
		if responseMessage == "" {
			responseMessage = DefaultErrorMessageInternal
		}
	} else if !mapped {
		// The Kratos error reason is not in any mapper (e.g. a raw error).
		// UnifiedResponse.Code remains DefaultErrorCodeForMismatch.
		if responseMessage == "" {
			responseMessage = DefaultErrorMessageInternal
		}
	}

	// Ensure final message is not empty for the client
	if responseMessage == "" {
		responseMessage = DefaultErrorMessageInternal
	}

	resp := UnifiedResponse{
		Code:    responseCode,
		Message: responseMessage,
		Data:    nil, // Or consider kErr.Metadata if useful
	}

	// Set HTTP status code based on Kratos error's Code field (which is the HTTP status)
	httpStatusCode := int(kErr.Code)
	// Basic validation for HTTP status code to ensure it's a typical error code.
	if httpStatusCode < 400 || httpStatusCode > 599 {
		// If kErr.Code is not a standard client/server error HTTP status (e.g. 0 or 200 for some reason),
		// default to 500 Internal Server Error for the HTTP transport.
		httpStatusCode = http.StatusInternalServerError
	}
	w.WriteHeader(httpStatusCode)

	if encErr := json.NewEncoder(w).Encode(resp); encErr != nil {
		// Log this encoding error, as the response could not be sent.
		// Depending on project's logger: log.Error("Failed to encode error response", "error", encErr)
	}
}

const (
	// DefaultErrorCodeForMismatch is used when a Kratos error reason maps to 0 (success code)
	// or when the reason is not found in the business error code map.
	// This ensures errors always have a non-zero UnifiedResponse.Code.
	// This value should ideally be a defined enum in your .proto for generic errors.
	DefaultErrorCodeForMismatch = 50000
	// DefaultErrorMessageInternal is a generic message for internal or unclassified errors.
	DefaultErrorMessageInternal = "Internal server error"
)
