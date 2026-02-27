package core

import (
	"encoding/json"
	"fmt"
)

type ComponentFunctionDefinition[Request any, Response any] struct {
	Component string
	Function  string
}

func (c ComponentFunctionDefinition[Request, Response]) Must(mod ForeignComponent, req Request) Response {
	resp, err := mod.ExecuteFunction(c.Component, c.Function, req)
	if err != nil {
		panic(fmt.Sprintf("component call failed (%s.%s): %s", c.Component, c.Function, err.Error()))
	}
	var r Response
	err = json.Unmarshal(resp, &r)
	if err != nil {
		panic(fmt.Sprintf("component response decode failed (%s.%s): %s", c.Component, c.Function, err.Error()))
	}
	return r
}
