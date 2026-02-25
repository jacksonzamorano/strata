//go:build !(darwin && cgo)

package tasklib

func newPlatformKeychain() KeychainProvider {
	return &memoryKeychainProvider{
		values: map[string]map[string]string{},
	}
}
