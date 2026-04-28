// Package http is the inbound HTTP adapter for the spike. It parses
// the JSON wire format into driving.BatchInput, calls the use case,
// and maps domain errors to HTTP status codes per
// docs/spike/backend-api-contract.md §5.
package http

import "github.com/example/m-trace/apps/api/hexagon/port/driving"

// wireEvent is the JSON shape of a single event in the request body.
// The use case never sees these tags; mapping happens in
// toEventInputs.
type wireEvent struct {
	EventName       string  `json:"event_name"`
	ProjectID       string  `json:"project_id"`
	SessionID       string  `json:"session_id"`
	ClientTimestamp string  `json:"client_timestamp"`
	SequenceNumber  *int64  `json:"sequence_number,omitempty"`
	SDK             wireSDK `json:"sdk"`
}

type wireSDK struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type wireBatch struct {
	SchemaVersion string      `json:"schema_version"`
	Events        []wireEvent `json:"events"`
}

func toEventInputs(events []wireEvent) []driving.EventInput {
	out := make([]driving.EventInput, len(events))
	for i, e := range events {
		out[i] = driving.EventInput{
			EventName:       e.EventName,
			ProjectID:       e.ProjectID,
			SessionID:       e.SessionID,
			ClientTimestamp: e.ClientTimestamp,
			SequenceNumber:  e.SequenceNumber,
			SDK: driving.SDKInput{
				Name:    e.SDK.Name,
				Version: e.SDK.Version,
			},
		}
	}
	return out
}
