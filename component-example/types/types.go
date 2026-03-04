package types

import (
	"github.com/jacksonzamorano/strata/component"
)

type SayRequest struct {
	Name string
}
type SayResponse struct {
	CurrentValue string
	LastValue    string
	TenXValue    int
}
type EmptyRequest struct{}

var SayFeature = component.Define[SayRequest, SayResponse]("say")
var Reset = component.Define[EmptyRequest, string]("reset")
