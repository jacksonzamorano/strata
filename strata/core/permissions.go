package core

type ApprovedComponentPermission struct {
	Container string
	Action    PermissionAction
	Scopes    []string
}
