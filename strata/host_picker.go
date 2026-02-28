package strata

func UseConsole() *ConfigurationModification {
	return &ConfigurationModification{
		NewHost: newConsoleHost,
	}
}

func UseWeb() *ConfigurationModification {
	return &ConfigurationModification{
		NewHost: newWebHost,
	}
}
