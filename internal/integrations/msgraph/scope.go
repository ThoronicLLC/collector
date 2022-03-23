package msgraph

type Scope string

const (
	AtpScope   Scope = "https://api.securitycenter.windows.com/.default"
	GraphScope Scope = "https://graph.microsoft.com/.default"
)

// String makes Scope satisfy the Stringer interface.
func (s Scope) String() string {
	return string(s)
}
