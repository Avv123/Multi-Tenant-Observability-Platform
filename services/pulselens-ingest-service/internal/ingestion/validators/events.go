package validators

import (
	"strings"

	pulsetelemetry "github.com/omniful/pulselens-common/telemetry"
	pulselens_error "github.com/omniful/pulselens-ingest-service/pkg/error"
	"github.com/omniful/pulselens-platform/errs"
)

func ValidateBatch(request *pulsetelemetry.BatchIngestRequest) errs.CustomError {
	if request == nil || len(request.Events) == 0 {
		return errs.New(pulselens_error.BadRequest, "events are required")
	}

	for _, event := range request.Events {
		if customError := ValidateEvent(event); customError.Exists() {
			return customError
		}
	}

	return errs.CustomError{}
}

func ValidateEvent(event pulsetelemetry.ClientEvent) errs.CustomError {
	if len(event.Payload) == 0 {
		return errs.New(pulselens_error.BadRequest, "payload is required")
	}

	switch event.EventType {
	case pulsetelemetry.EventTypeMetric:
		if strings.TrimSpace(stringValue(event.Payload["metric_name"])) == "" {
			return errs.New(pulselens_error.BadRequest, "metric payload requires metric_name")
		}
		if _, ok := numericValue(event.Payload["value"]); !ok {
			return errs.New(pulselens_error.BadRequest, "metric payload requires numeric value")
		}
	case pulsetelemetry.EventTypeTrace:
		if strings.TrimSpace(stringValue(event.Payload["span_id"])) == "" {
			return errs.New(pulselens_error.BadRequest, "trace payload requires span_id")
		}
		if strings.TrimSpace(stringValue(event.Payload["operation"])) == "" {
			return errs.New(pulselens_error.BadRequest, "trace payload requires operation")
		}
	case pulsetelemetry.EventTypeCustom:
		return errs.CustomError{}
	default:
		if strings.TrimSpace(stringValue(event.Payload["message"])) == "" {
			return errs.New(pulselens_error.BadRequest, "log payload requires message")
		}
	}

	return errs.CustomError{}
}

func stringValue(value interface{}) string {
	typed, _ := value.(string)
	return typed
}

func numericValue(value interface{}) (float64, bool) {
	switch typed := value.(type) {
	case int:
		return float64(typed), true
	case int64:
		return float64(typed), true
	case float64:
		return typed, true
	case float32:
		return float64(typed), true
	default:
		return 0, false
	}
}
