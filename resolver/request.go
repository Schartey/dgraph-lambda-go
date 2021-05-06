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
}
