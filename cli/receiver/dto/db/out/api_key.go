package out

type ApiKey struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
	Hash string `json:"hash"`
}
