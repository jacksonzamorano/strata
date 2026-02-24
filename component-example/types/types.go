package types

type SayRequest struct {
	Name string
}
type SayResponse struct {
	CurrentValue string
	LastValue    string
}
