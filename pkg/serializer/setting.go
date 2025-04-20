package serializer

// VolResponse VOL query response
type VolResponse struct {
	Signature string `json:"signature"`
	Content   string `json:"content"`
}
