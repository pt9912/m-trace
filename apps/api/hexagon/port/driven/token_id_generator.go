package driven

// TokenIDGenerator vergibt opake Token-IDs (z. B. ULIDs) für neue
// Session Tokens. Der Wert landet als `jti` im Claim-Set und als
// `token_id` in der Wire-Antwort sowie in Logs und Audit-Records.
//
// Die Eindeutigkeit ist Aufrufer-Verantwortung: doppelte IDs würden
// die `jti`-Semantik (`one-time identifier`) verletzen. Der
// Adapter nutzt dafür `crypto/rand` plus `st_`-Prefix.
type TokenIDGenerator interface {
	NewTokenID() (string, error)
}
