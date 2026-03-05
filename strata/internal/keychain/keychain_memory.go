package keychain

import "sync"

type memoryKeychainProvider struct {
	mu     sync.RWMutex
	values map[string]map[string]string
}

type memoryContainerKeychainProvider struct {
	provider  *memoryKeychainProvider
	namespace string
}

func (m *memoryKeychainProvider) Container(namespace string) ContainerKeychainProvider {
	return &memoryContainerKeychainProvider{
		provider:  m,
		namespace: namespace,
	}
}

func (m *memoryContainerKeychainProvider) Get(key string) string {
	m.provider.mu.RLock()
	defer m.provider.mu.RUnlock()

	ns, ok := m.provider.values[m.namespace]
	if !ok {
		return ""
	}

	return ns[key]
}

func (m *memoryContainerKeychainProvider) Set(key, value string) {
	m.provider.mu.Lock()
	defer m.provider.mu.Unlock()

	ns, ok := m.provider.values[m.namespace]
	if !ok {
		ns = map[string]string{}
		m.provider.values[m.namespace] = ns
	}

	ns[key] = value
}
