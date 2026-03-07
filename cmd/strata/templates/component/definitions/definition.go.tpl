package definitions

import "github.com/jacksonzamorano/strata/component"

var Manifest = component.ComponentManifest{
	Name:    "component-template",
	Version: "0.1.0",
}

type EchoRequest struct {
	Message string
}

type EchoResponse struct {
	Message string
}

var Echo = component.Define[EchoRequest, EchoResponse](Manifest, "echo")
