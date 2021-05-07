package resolver

type AuthHeader struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type DBody struct {
	AccessToken string                   `json:"X-Dgraph-AccessToken"`
	Args        map[string]interface{}   `json:"args"`
	AuthHeader  AuthHeader               `json:"authHeader"`
	Resolver    string                   `json:"resolver"`
	Parents     []map[string]interface{} `json:"parents"`
	Event       Event                    `json:"event"`
}

type Event struct {
	TypeName  string          `json:"__typename"`
	CommitTs  uint64          `json:"commitTs"`
	Operation string          `json:"operation"`
	Add       AddEventInfo    `json:"add"`
	Update    UpdateEventInfo `json:"update"`
	Delete    DeleteEventInfo `json:"delete"`
}

type AddEventInfo struct {
	RootUIDs []string                 `json:"rootUIDs"`
	Input    []map[string]interface{} `json:"input"`
}

type UpdateEventInfo struct {
	RootUIDs    []string               `json:"rootUIDs"`
	SetPatch    map[string]interface{} `json:"setPatch"`
	RemovePatch map[string]interface{} `json:"removePatch"`
}

type DeleteEventInfo struct {
	RootUIDs []string `json:"rootUIDs"`
}
