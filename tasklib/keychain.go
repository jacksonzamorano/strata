package tasklib

type KeychainProvider interface {
	Container(namespace string) ContainerKeychainProvider
}
type ContainerKeychainProvider interface {
	Get(key string) string
	Set(key, value string)
}

var PlatformKeychain KeychainProvider = newPlatformKeychain()
