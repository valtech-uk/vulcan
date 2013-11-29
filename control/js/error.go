package js

import (
	"encoding/json"
	"fmt"
	"github.com/mailgun/vulcan/command"
	"github.com/mailgun/vulcan/netutils"
	"net/http"
)

func errorToJs(inErr error) map[string]interface{} {
	switch err := inErr.(type) {
	case *command.AllUpstreamsDownError:
		return map[string]interface{}{
			"type": "all_upstreams_down",
			"code": http.StatusBadGateway,
			"body": map[string]interface{}{
				"error": http.StatusText(http.StatusBadGateway),
			},
		}
	case *command.RetryError:
		return map[string]interface{}{
			"type":          "retry",
			"retry_seconds": err.Seconds,
			"code":          429,
			"body": map[string]interface{}{
				"error":         "Too Many Requests",
				"retry_seconds": err.Seconds,
			},
		}
	default:
		return map[string]interface{}{
			"type": "internal",
			"code": http.StatusInternalServerError,
			"body": http.StatusText(http.StatusInternalServerError),
		}
	}
}

func errorFromJs(inErr interface{}) (*netutils.HttpError, error) {
	switch err := inErr.(type) {
	case map[string]interface{}:
		return errorFromDict(err)
	default:
		return nil, fmt.Errorf("Unsupported error type")
	}
}

func errorFromDict(in map[string]interface{}) (*netutils.HttpError, error) {
	codeI, ok := in["code"]
	if !ok {
		return nil, fmt.Errorf("Expected 'code' parameter")
	}
	code, ok := codeI.(int)
	if !ok {
		return nil, fmt.Errorf("Parameter 'code' should be integer, got %v", code)
	}
	message := http.StatusText(code)
	messageI, ok := in["message"]
	if ok {
		message, ok = messageI.(string)
		if !ok {
			return nil, fmt.Errorf("Parameter 'message' should be a string")
		}
	}
	bodyI, ok := in["body"]
	if !ok {
		return nil, fmt.Errorf("Expected 'body' parameter")
	}
	bodyBytes, err := json.Marshal(bodyI)
	if err != nil {
		return nil, fmt.Errorf("Failed to serialize body to json: %s", err)
	}

	return &netutils.HttpError{StatusCode: code, Status: message, Body: bodyBytes}, nil
}