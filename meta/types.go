package meta

type Material struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
	Tags        string `json:"tags,omitempty"`
	Alias       string `json:"alias,omitempty"`
}

type MaterialType struct {
	Name string `json:"name"`
}

type MaterialStore struct {
	Materials []*Material `json:"materials,omitempty"`
}
