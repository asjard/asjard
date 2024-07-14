package security

type testCipher struct{}

func (testCipher) Encrypt(data string, opts *Options) (string, error) {
	return "", nil
}

func (testCipher) Decrypt(data string, opts *Options) (string, error) {
	return "", nil
}

func newTestCipher(_ string) (Cipher, error) {
	return &testCipher{}, nil
}
