package types

import (
	"time"

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

type TriggerTest struct {
	Time time.Time `json:"time"`
}

var SayFeature = component.Define[SayRequest, SayResponse]("say")
var Reset = component.Define[EmptyRequest, string]("reset")
var TestTrigger = component.NewComponentTrigger[TriggerTest]("test")
