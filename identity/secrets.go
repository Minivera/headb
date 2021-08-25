package identity

var secrets struct {
	// Client ID for the GitHub OAuth App.
	GithubClientID string

	// Secret key for encryption purposes, should be the AES key, either 16, 24, or 32 bytes
	// to select AES-128, AES-192, or AES-256.
	SecretKey string

	// Secret key for encrypting paseto web tokens, should be 32 bit long.
	SecretPasetoKey string
}
