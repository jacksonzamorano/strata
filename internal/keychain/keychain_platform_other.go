//go:build !(darwin && cgo)

package keychain

func newPlatformKeychain() KeychainProvider {
	return &memoryKeychainProvider{
		values: map[string]map[string]string{},
	}
}
