package core

import "fmt"

type ApprovedComponentPermission struct {
	Container string
	Action    PermissionAction
	Scopes    []string
}

func (p *Permission) Hash() string {
	return fmt.Sprintf("%s.%s.%s", p.Container, p.Action, *p.Scope)
}
