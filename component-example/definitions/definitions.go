package definitions

import (
	"time"

	"github.com/jacksonzamorano/strata/component"
)

var Manifest = component.ComponentManifest{
	Name:    "example",
	Version: "1.1.0",
}

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

var SayFeature = component.Define[SayRequest, SayResponse](Manifest, "say")
var Reset = component.Define[EmptyRequest, string](Manifest, "reset")
var GetSecret = component.Define[EmptyRequest, string](Manifest, "get-secret")
var TestTrigger = component.NewComponentTrigger[TriggerTest](Manifest, "test")
