package strata

import "github.com/jacksonzamorano/strata/core"

type KeychainProvider = core.KeychainProvider
type ContainerKeychainProvider = core.Keychain

var PlatformKeychain KeychainProvider = newPlatformKeychain()
