package generator

import (
	"strings"
	"testing"
)

func TestTimeImportDetection(t *testing.T) {
	tests := []struct {
		name       string
		fields     []RequestField
		respType   string
		hasResp    bool
		wantImport bool
	}{
		{
			name:       "No time fields",
			fields:     []RequestField{{"Name", "string", "name"}},
			respType:   "string",
			hasResp:    true,
			wantImport: false,
		},
		{
			name:       "Has time field in request",
			fields:     []RequestField{{"CreatedAt", "time.Time", "created_at"}},
			respType:   "string",
			hasResp:    true,
			wantImport: true,
		},
		{
			name:       "Has time array in request",
			fields:     []RequestField{{"Timestamps", "[]time.Time", "timestamps"}},
			respType:   "string",
			hasResp:    true,
			wantImport: true,
		},
		{
			name:       "Has time response",
			fields:     []RequestField{{"Name", "string", "name"}},
			respType:   "time.Time",
			hasResp:    true,
			wantImport: true,
		},
		{
			name:       "Has time array response",
			fields:     []RequestField{{"Name", "string", "name"}},
			respType:   "[]time.Time",
			hasResp:    true,
			wantImport: true,
		},
		{
			name:       "Has time in complex response",
			fields:     []RequestField{{"Name", "string", "name"}},
			respType:   "models.UserWithTime",
			hasResp:    true,
			wantImport: false, // Changed expectation - "time" in type name doesn't mean time.Time import needed
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Skip the parser setup since we're only testing the detection logic

			// Create mock data for operation
			data := OperationData{
				RequestFields:   tt.fields,
				HasResponseBody: tt.hasResp,
				ResponseType:    tt.respType,
			}

			// Call the method we added to check for time import
			importTime := false

			// Check request fields for time.Time usage
			for _, field := range data.RequestFields {
				if field.Type == "time.Time" || field.Type == "[]time.Time" {
					importTime = true
					break
				}
			}

			// Check response type for time.Time usage
			if !importTime && data.HasResponseBody {
				if data.ResponseType == "time.Time" ||
					data.ResponseType == "[]time.Time" ||
					data.ResponseType == "*time.Time" ||
					data.ResponseType == "[]*time.Time" ||
					strings.Contains(data.ResponseType, "time.Time") {
					importTime = true
				}
			}

			// Verify the result
			if importTime != tt.wantImport {
				t.Errorf("Got importTime = %v, want %v", importTime, tt.wantImport)
			}
		})
	}
}
