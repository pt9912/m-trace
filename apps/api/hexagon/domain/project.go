package domain

// Project is a tenant-like principal that owns events and tokens.
//
// AllowedOrigins listet die exakten Origin-Werte (Schema+Host+Port),
// gegen die der HTTP-Adapter den `Origin`-Header eines POST-Requests
// validiert (CORS Variante B, plan-0.1.0.md §5.1). Die Liste enthält
// **konkrete** Origins; Wildcards `*` sind nicht zulässig.
type Project struct {
	ID             string
	Token          ProjectToken
	AllowedOrigins []string
}

// IsOriginAllowed gibt true zurück, wenn der gegebene Origin exakt in
// der Allowed-Origins-Liste des Projects steht. Leerer Origin gilt als
// erlaubt — so bleibt der CLI/curl-Pfad ohne `Origin`-Header offen
// (plan-0.1.0.md §5.1, CORS Variante B).
func (p Project) IsOriginAllowed(origin string) bool {
	if origin == "" {
		return true
	}
	for _, o := range p.AllowedOrigins {
		if o == origin {
			return true
		}
	}
	return false
}
