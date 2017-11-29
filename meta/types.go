package meta

type Material struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
	Tags        string `json:"tags,omitempty"`
}

type MaterialStore struct {
	Materials []*Material `json:"materials,omitempty"`
}
