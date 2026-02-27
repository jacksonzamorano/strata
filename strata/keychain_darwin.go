//go:build darwin && cgo

package strata

/*
#cgo darwin LDFLAGS: -framework Security
#include <CoreFoundation/CoreFoundation.h>
#include <Security/Security.h>
#include <stdlib.h>
#include <string.h>

static CFStringRef strata_cfstring_from_bytes(const char* bytes, size_t len) {
	if (len == 0) {
		return CFStringCreateWithCString(NULL, "", kCFStringEncodingUTF8);
	}

	return CFStringCreateWithBytes(
		NULL,
		(const UInt8*)bytes,
		(CFIndex)len,
		kCFStringEncodingUTF8,
		false
	);
}

static int strata_keychain_get(
	const char* service,
	size_t serviceLen,
	const char* account,
	size_t accountLen,
	char** outData,
	size_t* outLen
) {
	CFStringRef serviceRef = strata_cfstring_from_bytes(service, serviceLen);
	CFStringRef accountRef = strata_cfstring_from_bytes(account, accountLen);
	if (serviceRef == NULL || accountRef == NULL) {
		if (serviceRef != NULL) {
			CFRelease(serviceRef);
		}
		if (accountRef != NULL) {
			CFRelease(accountRef);
		}
		return -1;
	}

	const void* queryKeys[] = {
		kSecClass,
		kSecAttrService,
		kSecAttrAccount,
		kSecReturnData,
		kSecMatchLimit,
	};
	const void* queryValues[] = {
		kSecClassGenericPassword,
		serviceRef,
		accountRef,
		kCFBooleanTrue,
		kSecMatchLimitOne,
	};
	CFDictionaryRef query = CFDictionaryCreate(
		NULL,
		queryKeys,
		queryValues,
		5,
		&kCFTypeDictionaryKeyCallBacks,
		&kCFTypeDictionaryValueCallBacks
	);

	CFRelease(serviceRef);
	CFRelease(accountRef);

	if (query == NULL) {
		return -1;
	}

	CFTypeRef result = NULL;
	OSStatus status = SecItemCopyMatching(query, &result);
	CFRelease(query);

	if (status != errSecSuccess) {
		return (int)status;
	}

	if (result == NULL || CFGetTypeID(result) != CFDataGetTypeID()) {
		if (result != NULL) {
			CFRelease(result);
		}
		return -1;
	}

	CFDataRef passwordData = (CFDataRef)result;
	CFIndex passwordLen = CFDataGetLength(passwordData);
	const UInt8* passwordBytes = CFDataGetBytePtr(passwordData);

	char* copied = NULL;
	if (passwordLen > 0) {
		copied = (char*)malloc((size_t)passwordLen);
		if (copied == NULL) {
			CFRelease(passwordData);
			return -1;
		}
		memcpy(copied, passwordBytes, (size_t)passwordLen);
	}

	CFRelease(passwordData);
	*outData = copied;
	*outLen = (size_t)passwordLen;
	return 0;
}

static int strata_keychain_set(
	const char* service,
	size_t serviceLen,
	const char* account,
	size_t accountLen,
	const char* password,
	size_t passwordLen
) {
	CFStringRef serviceRef = strata_cfstring_from_bytes(service, serviceLen);
	CFStringRef accountRef = strata_cfstring_from_bytes(account, accountLen);
	CFDataRef passwordRef = CFDataCreate(NULL, (const UInt8*)password, (CFIndex)passwordLen);
	if (serviceRef == NULL || accountRef == NULL || passwordRef == NULL) {
		if (serviceRef != NULL) {
			CFRelease(serviceRef);
		}
		if (accountRef != NULL) {
			CFRelease(accountRef);
		}
		if (passwordRef != NULL) {
			CFRelease(passwordRef);
		}
		return -1;
	}

	const void* queryKeys[] = {
		kSecClass,
		kSecAttrService,
		kSecAttrAccount,
	};
	const void* queryValues[] = {
		kSecClassGenericPassword,
		serviceRef,
		accountRef,
	};
	CFDictionaryRef query = CFDictionaryCreate(
		NULL,
		queryKeys,
		queryValues,
		3,
		&kCFTypeDictionaryKeyCallBacks,
		&kCFTypeDictionaryValueCallBacks
	);

	const void* attrsKeys[] = {
		kSecValueData,
	};
	const void* attrsValues[] = {
		passwordRef,
	};
	CFDictionaryRef attrs = CFDictionaryCreate(
		NULL,
		attrsKeys,
		attrsValues,
		1,
		&kCFTypeDictionaryKeyCallBacks,
		&kCFTypeDictionaryValueCallBacks
	);

	if (query == NULL || attrs == NULL) {
		if (query != NULL) {
			CFRelease(query);
		}
		if (attrs != NULL) {
			CFRelease(attrs);
		}
		CFRelease(serviceRef);
		CFRelease(accountRef);
		CFRelease(passwordRef);
		return -1;
	}

	OSStatus updateStatus = SecItemUpdate(query, attrs);
	CFRelease(attrs);

	if (updateStatus == errSecItemNotFound) {
		const void* addKeys[] = {
			kSecClass,
			kSecAttrService,
			kSecAttrAccount,
			kSecValueData,
		};
		const void* addValues[] = {
			kSecClassGenericPassword,
			serviceRef,
			accountRef,
			passwordRef,
		};
		CFDictionaryRef addItem = CFDictionaryCreate(
			NULL,
			addKeys,
			addValues,
			4,
			&kCFTypeDictionaryKeyCallBacks,
			&kCFTypeDictionaryValueCallBacks
		);
		if (addItem == NULL) {
			CFRelease(query);
			CFRelease(serviceRef);
			CFRelease(accountRef);
			CFRelease(passwordRef);
			return -1;
		}

		OSStatus addStatus = SecItemAdd(addItem, NULL);
		CFRelease(addItem);
		CFRelease(query);
		CFRelease(serviceRef);
		CFRelease(accountRef);
		CFRelease(passwordRef);
		return (int)addStatus;
	}

	CFRelease(query);
	CFRelease(serviceRef);
	CFRelease(accountRef);
	CFRelease(passwordRef);
	return (int)updateStatus;
}
*/
import "C"

import (
	"fmt"
	"unsafe"
)

const keychainServicePrefix = "com.strata.keychain"

type macOSKeychainProvider struct{}

type macOSContainerKeychainProvider struct {
	service string
}

func newPlatformKeychain() KeychainProvider {
	return &macOSKeychainProvider{}
}

func (m *macOSKeychainProvider) Container(namespace string) ContainerKeychainProvider {
	service := keychainServicePrefix
	if len(namespace) > 0 {
		service = fmt.Sprintf("%s.%s", keychainServicePrefix, namespace)
	}

	return &macOSContainerKeychainProvider{
		service: service,
	}
}

func (m *macOSContainerKeychainProvider) Get(key string) string {
	if len(key) == 0 {
		return ""
	}

	servicePtr, serviceLen, freeService := cBytes(m.service)
	defer freeService()

	accountPtr, accountLen, freeAccount := cBytes(key)
	defer freeAccount()

	var outData *C.char
	var outLen C.size_t

	status := C.strata_keychain_get(
		servicePtr,
		serviceLen,
		accountPtr,
		accountLen,
		&outData,
		&outLen,
	)

	if status != 0 {
		return ""
	}

	if outData == nil || outLen == 0 {
		return ""
	}
	defer C.free(unsafe.Pointer(outData))

	bytes := C.GoBytes(unsafe.Pointer(outData), C.int(outLen))
	return string(bytes)
}

func (m *macOSContainerKeychainProvider) Set(key, value string) {
	if len(key) == 0 {
		return
	}

	servicePtr, serviceLen, freeService := cBytes(m.service)
	defer freeService()

	accountPtr, accountLen, freeAccount := cBytes(key)
	defer freeAccount()

	valuePtr, valueLen, freeValue := cBytes(value)
	defer freeValue()

	status := C.strata_keychain_set(
		servicePtr,
		serviceLen,
		accountPtr,
		accountLen,
		valuePtr,
		valueLen,
	)

	if status != 0 {
		return
	}
}

func cBytes(value string) (*C.char, C.size_t, func()) {
	bytes := []byte(value)
	if len(bytes) == 0 {
		return nil, 0, func() {}
	}

	ptr := C.CBytes(bytes)
	return (*C.char)(ptr), C.size_t(len(bytes)), func() {
		C.free(ptr)
	}
}
