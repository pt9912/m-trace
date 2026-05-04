// Package http is the inbound HTTP adapter for the spike. It parses
// the JSON wire format into driving.BatchInput, calls the use case,
// and maps domain errors to HTTP status codes per
// spec/backend-api-contract.md §5.
package http

import "github.com/pt9912/m-trace/apps/api/hexagon/port/driving"

// wireEvent is the JSON shape of a single event in the request body.
// The use case never sees these tags; mapping happens in
// toEventInputs.
type wireEvent struct {
	EventName       string         `json:"event_name"`
	ProjectID       string         `json:"project_id"`
	SessionID       string         `json:"session_id"`
	ClientTimestamp string         `json:"client_timestamp"`
	SequenceNumber  *int64         `json:"sequence_number,omitempty"`
	SDK             wireSDK        `json:"sdk"`
	Meta            map[string]any `json:"meta,omitempty"`
}

type wireSDK struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// wireBoundary spiegelt einen Eintrag aus dem optionalen
// `session_boundaries[]`-Block (API-Kontrakt §3.4). Use-Case validiert
// alle Felder; HTTP-Adapter dekodiert nur das Wire-Format.
type wireBoundary struct {
	Kind            string `json:"kind"`
	ProjectID       string `json:"project_id"`
	SessionID       string `json:"session_id"`
	NetworkKind     string `json:"network_kind"`
	Adapter         string `json:"adapter"`
	Reason          string `json:"reason"`
	ClientTimestamp string `json:"client_timestamp"`
}

type wireBatch struct {
	SchemaVersion     string         `json:"schema_version"`
	Events            []wireEvent    `json:"events"`
	SessionBoundaries []wireBoundary `json:"session_boundaries,omitempty"`
}

func toBoundaryInputs(boundaries []wireBoundary) []driving.BoundaryInput {
	if len(boundaries) == 0 {
		return nil
	}
	out := make([]driving.BoundaryInput, len(boundaries))
	for i, b := range boundaries {
		out[i] = driving.BoundaryInput{
			Kind:            b.Kind,
			ProjectID:       b.ProjectID,
			SessionID:       b.SessionID,
			NetworkKind:     b.NetworkKind,
			Adapter:         b.Adapter,
			Reason:          b.Reason,
			ClientTimestamp: b.ClientTimestamp,
		}
	}
	return out
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
			Meta: copyMeta(e.Meta),
		}
	}
	return out
}

func copyMeta(in map[string]any) map[string]any {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]any, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
