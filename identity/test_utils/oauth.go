package test_utils

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"

	gogithub "github.com/google/go-github/github"

	"encore.app/identity/github"
)

var (
	DefaultDeviceCodeResponse = github.DeviceCodeResponse{
		DeviceCode:      "1234",
		UserCode:        "1234",
		VerificationUri: "https://example.com",
		ExpiresIn:       2,
		Interval:        1,
	}
	DefaultAccessTokenResponse = github.AccessTokenResponse{
		AccessToken: "1234",
		TokenType:   "bearer",
		Scope:       "test",
	}
	DefaultIdentityResponse = gogithub.User{
		Login:  StringPointer("test"),
		NodeID: StringPointer("1234"),
	}
)

type GithubTestOAuthServer struct {
	deviceCodeStatus   int
	deviceCodeResponse github.DeviceCodeResponse

	accessTokenStatus   int
	accessTokenResponse github.AccessTokenResponse

	identityStatus   int
	identityResponse gogithub.User
}

func CreateTestGithubOAuthServer() (*httptest.Server, *GithubTestOAuthServer) {
	handler := &GithubTestOAuthServer{
		deviceCodeStatus:    http.StatusOK,
		deviceCodeResponse:  DefaultDeviceCodeResponse,
		accessTokenStatus:   http.StatusOK,
		accessTokenResponse: DefaultAccessTokenResponse,
		identityStatus:      http.StatusOK,
		identityResponse:    DefaultIdentityResponse,
	}
	server := httptest.NewServer(handler)

	return server, handler
}

func (s *GithubTestOAuthServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/login/device/code":
		w.WriteHeader(s.deviceCodeStatus)
		_ = json.NewEncoder(w).Encode(s.deviceCodeResponse)
	case "/login/oauth/access_token":
		w.WriteHeader(s.accessTokenStatus)
		_ = json.NewEncoder(w).Encode(s.accessTokenResponse)
	case "/user":
		w.WriteHeader(s.identityStatus)
		_ = json.NewEncoder(w).Encode(s.identityResponse)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func (s *GithubTestOAuthServer) DeviceCodeResponse(statusCode int, response github.DeviceCodeResponse) {
	s.deviceCodeStatus = statusCode
	s.deviceCodeResponse = response
}

func (s *GithubTestOAuthServer) AccessTokenResponse(statusCode int, response github.AccessTokenResponse) {
	s.accessTokenStatus = statusCode
	s.accessTokenResponse = response
}

func (s *GithubTestOAuthServer) IdentityResponse(statusCode int, response gogithub.User) {
	s.identityStatus = statusCode
	s.identityResponse = response
}
