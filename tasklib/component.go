package tasklib

type Component interface {
	Initialize(appState *AppState)
}
