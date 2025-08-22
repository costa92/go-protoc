package logger

import (
	"fmt"
	"strings"

	krtlog "github.com/go-kratos/kratos/v2/log"
)

// NewKratosLogger creates a Kratos-compatible logger using the provided logger.
func NewKratosLogger(l Logger, id, name, version string) krtlog.Logger {
	kratosLogger := &kratosLoggerAdapter{
		logger: l.With(
			"service.id", id,
			"service.name", name,
			"service.version", version,
		),
	}
	
	return krtlog.With(kratosLogger,
		"ts", krtlog.DefaultTimestamp,
		"caller", krtlog.DefaultCaller,
	)
}

// kratosLoggerAdapter adapts the generic logger.Logger to implement Kratos's log.Logger interface.
type kratosLoggerAdapter struct {
	logger Logger
}

// Log implements the Kratos Logger interface.
func (l *kratosLoggerAdapter) Log(level krtlog.Level, keyvals ...interface{}) error {
	// Extract message from keyvals if present
	msg := extractMessage(keyvals)
	
	switch level {
	case krtlog.LevelDebug:
		l.logger.Debugw(msg, keyvals...)
	case krtlog.LevelInfo:
		l.logger.Infow(msg, keyvals...)
	case krtlog.LevelWarn:
		l.logger.Warnw(msg, keyvals...)
	case krtlog.LevelError:
		l.logger.Errorw(msg, keyvals...)
	case krtlog.LevelFatal:
		l.logger.Fatalw(msg, keyvals...)
	default:
		l.logger.Infow(msg, keyvals...)
	}
	return nil
}

// extractMessage extracts a meaningful message from keyvals
func extractMessage(keyvals ...interface{}) string {
	if len(keyvals) == 0 {
		return "Kratos log"
	}
	
	// Flatten keyvals if it's nested (Kratos may pass keyvals as nested slice)
	var flatKeyvals []interface{}
	for _, kv := range keyvals {
		if slice, ok := kv.([]interface{}); ok {
			// If it's a slice, append all its elements
			flatKeyvals = append(flatKeyvals, slice...)
		} else {
			// If it's not a slice, append as-is
			flatKeyvals = append(flatKeyvals, kv)
		}
	}
	
	// Collect all relevant fields
	msg := ""
	operation := ""
	component := ""
	kind := ""
	caller := ""
	code := ""
	
	// Parse keyvals pairs - we might have multiple 'msg' fields, prefer meaningful ones
	var allMsgs []string
	
	for i := 0; i < len(flatKeyvals)-1; i += 2 {
		if key, ok := flatKeyvals[i].(string); ok {
			switch key {
			case "msg":
				if value, ok := flatKeyvals[i+1].(string); ok && value != "" {
					allMsgs = append(allMsgs, value)
				}
			case "operation":
				if value, ok := flatKeyvals[i+1].(string); ok {
					operation = value
				}
			case "component":
				if value, ok := flatKeyvals[i+1].(string); ok {
					component = value
				}
			case "kind":
				if value, ok := flatKeyvals[i+1].(string); ok {
					kind = value
				}
			case "caller":
				if value, ok := flatKeyvals[i+1].(string); ok {
					caller = value
				}
			case "code":
				// Handle both string and integer codes
				if value, ok := flatKeyvals[i+1].(string); ok {
					code = value
				} else if value, ok := flatKeyvals[i+1].(int); ok {
					code = fmt.Sprintf("%d", value)
				}
			}
		}
	}
	
	// Prioritize messages: prefer Kratos framework messages first
	for _, m := range allMsgs {
		if strings.HasPrefix(m, "[HTTP]") || strings.HasPrefix(m, "[gRPC]") {
			msg = m
			break
		}
	}
	
	// If no framework message found, use the last meaningful message
	if msg == "" && len(allMsgs) > 0 {
		msg = allMsgs[len(allMsgs)-1]
	}
	
	// Priority 1: Use original Kratos message if available and meaningful
	if msg != "" {
		// Clean up common Kratos message patterns
		if strings.HasPrefix(msg, "[HTTP]") || strings.HasPrefix(msg, "[gRPC]") {
			return strings.TrimSpace(msg)
		}
		// For other messages, return as-is
		return msg
	}
	
	// Priority 2: Construct from operation (API calls)
	if operation != "" {
		if strings.Contains(operation, "/") {
			// Extract service and method from operation like "/apiserver.v1.ApiServer/GetUser"
			parts := strings.Split(operation, "/")
			if len(parts) >= 3 {
				service := parts[1]
				method := parts[2]
				if code != "" && code != "0" {
					return fmt.Sprintf("%s.%s (code:%s)", service, method, code)
				}
				return fmt.Sprintf("%s.%s", service, method)
			}
		}
		return "API: " + operation
	}
	
	// Priority 3: Construct from component and kind
	if component != "" && kind != "" {
		switch component {
		case "http":
			return "HTTP " + kind
		case "grpc":
			return "gRPC " + kind
		default:
			return component + " " + kind
		}
	}
	
	// Priority 4: Use component only
	if component != "" {
		switch component {
		case "http":
			return "HTTP event"
		case "grpc":
			return "gRPC event"
		default:
			return component + " event"
		}
	}
	
	// Priority 5: Use kind only
	if kind != "" {
		return "Server " + kind
	}
	
	// Priority 6: Extract info from caller if available
	if caller != "" {
		// Extract meaningful part from caller like "http/server.go:330"
		if strings.Contains(caller, "/") {
			parts := strings.Split(caller, "/")
			lastPart := parts[len(parts)-1]
			if strings.Contains(lastPart, ":") {
				fileName := strings.Split(lastPart, ":")[0]
				return "Kratos " + strings.TrimSuffix(fileName, ".go")
			}
		}
	}
	
	// Fallback: Generic message with context hint
	return "Kratos event"
}