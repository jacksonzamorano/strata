package strata

import "github.com/jacksonzamorano/tasks/strata/core"

type KeychainProvider = core.KeychainProvider
type ContainerKeychainProvider = core.Keychain

var PlatformKeychain KeychainProvider = newPlatformKeychain()
