//go:build !(darwin && cgo)

package strata

func newPlatformKeychain() KeychainProvider {
	return &memoryKeychainProvider{
		values: map[string]map[string]string{},
	}
}
