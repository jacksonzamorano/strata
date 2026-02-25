package tasklib

import "github.com/jacksonzamorano/tasks/tasklib/core"

type KeychainProvider = core.KeychainProvider
type ContainerKeychainProvider = core.Keychain

var PlatformKeychain KeychainProvider = newPlatformKeychain()
