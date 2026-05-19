package npmctl

type memorySecretStore struct {
	values map[string]string
	err    error
}

func newMemorySecretStore() *memorySecretStore {
	return &memorySecretStore{values: map[string]string{}}
}

func (s *memorySecretStore) Get(secretName string) (string, bool, error) {
	if s.err != nil {
		return "", false, s.err
	}
	value, ok := s.values[secretName]
	return value, ok && value != "", nil
}

func (s *memorySecretStore) Set(secretName, value string) error {
	if s.err != nil {
		return s.err
	}
	s.values[secretName] = value
	return nil
}

func (s *memorySecretStore) Delete(secretName string) error {
	if s.err != nil {
		return s.err
	}
	delete(s.values, secretName)
	return nil
}
