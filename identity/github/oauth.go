package github

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/go-github/github"
	log "github.com/sirupsen/logrus"

	"encore.app/identity/models"
	"encore.app/identity/models/generated/identity/public/model"
)

var (
	// ErrAuthorizationPending occurs when the authorization request is pending and
	// the user hasn't entered the user code yet. The app is expected to keep polling
	// the POST https://github.com/login/oauth/access_token request without exceeding
	// the interval, which requires a minimum number of seconds between each request.
	ErrAuthorizationPending = errors.New("authorization pending")

	// ErrSlowDown occurs when you receive the slow_down error, 5 extra seconds are added
	// to the minimum interval or timeframe required between your requests using
	// POST https://github.com/login/oauth/access_token. For example, if the starting
	// interval required at least 5 seconds between requests and you get a slow_down
	// error response, you must now wait a minimum of 10 seconds before making a new
	// request for an OAuth access token. The error response includes the new interval
	// that you must use.
	ErrSlowDown = errors.New("too many requests, slow down interval")

	// ErrExpiredToken occurs if the device code expired, then you will see the token_expired
	// error. You must make a new request for a device code.
	ErrExpiredToken = errors.New("token was expired, please request a new one")

	// ErrInvalidGrantType occurs when the grant type is not urn:ietf:params:oauth:grant-type:device_code or
	// is not included as an input parameter when you poll the OAuth token request
	// POST https://github.com/login/oauth/access_token.
	ErrInvalidGrantType = errors.New("grant type is unsupported, must use `urn:ietf:params:oauth:grant-type:device_code`")

	// ErrIncorrectCredentials occurs when the ClientID field is not given or not valid
	// in the request
	ErrIncorrectCredentials = errors.New("incorrect client credentials")

	// ErrIncorrectDeviceCode occurs when the device_code provided is not valid.
	ErrIncorrectDeviceCode = errors.New("incorrect device code received, please request a new one")

	// ErrAccessDenied occurs when a user clicks cancel during the authorization process, you'll receive
	// a access_denied error and the user won't be able to use the verification code again.
	ErrAccessDenied = errors.New("access was denied to the resource")
)

// OAuthClient is a client implementation for triggering the device code
// OAuth flow with GitHub and getting the user info.
type OAuthClient struct {
	clientID string

	deviceCodeURL  string
	accessTokenURL string
	identityURL    string
}

// NewOAuthClientOptions are the available options to add to the OAuth client
// when creating it.
type NewOAuthClientOptions struct {
	ClientID       string
	DeviceCodeURL  string
	AccessTokenURL string
	IdentityURL    string
}

// NewOAuthClient will create a new client based on the options and return
// an instance of the client.
func NewOAuthClient(options NewOAuthClientOptions) OAuthClient {
	return OAuthClient{
		clientID:       options.ClientID,
		deviceCodeURL:  options.DeviceCodeURL,
		accessTokenURL: options.AccessTokenURL,
		identityURL:    options.IdentityURL,
	}
}

// DeviceCodeResponse is the response given to a POST request to the
// github device code endpoint for device OAuth.
type DeviceCodeResponse struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationUri string `json:"verification_uri"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
}

// AccessTokenResponse is the response given to a POST request to
// get an OAuth token. This response is compatible with any OAuth 2.0
// API.
type AccessTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`

	// These fields will be filled when there is an error
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
	ErrorUri         string `json:"error_uri"`
}

// RequestDeviceCode will send a request to GitHub for a new device code so the
// user can start authenticating.
func (c OAuthClient) RequestDeviceCode() (DeviceCodeResponse, error) {
	payload := strings.NewReader(fmt.Sprintf("client_id=%s", c.clientID))

	request, err := http.NewRequest("POST", c.deviceCodeURL, payload)
	if err != nil {
		log.WithError(err).Error("Could not create a request to get a device code")
		return DeviceCodeResponse{}, err
	}

	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-type", "application/x-www-form-urlencoded")

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		log.WithError(err).Error("Could not request device code")
		return DeviceCodeResponse{}, err
	}
	defer response.Body.Close()

	if response.StatusCode > 400 {
		log.WithFields(log.Fields{
			"status_code": response.StatusCode,
		}).WithError(err).Error("Could not request device code, request error")
		return DeviceCodeResponse{}, fmt.Errorf("could not request device code, %s error", response.StatusCode)
	}

	resp := DeviceCodeResponse{}
	err = json.NewDecoder(response.Body).Decode(&resp)
	if err != nil {
		log.WithFields(map[string]interface{}{
			"status": response.Status,
		}).WithError(err).Error("Could not extract response body")
		return DeviceCodeResponse{}, err
	}

	return resp, nil
}

// PollDeviceAuth tries to get the access token for a device code on an OAuth provider.
func (c OAuthClient) PollDeviceAuth(deviceCode string) (AccessTokenResponse, error) {
	payload := strings.NewReader(
		fmt.Sprintf(
			"client_id=%s&device_code=%s&grant_type=urn:ietf:params:oauth:grant-type:device_code",
			c.clientID,
			deviceCode,
		),
	)

	request, err := http.NewRequest("POST", c.accessTokenURL, payload)
	if err != nil {
		log.WithError(err).Error("Could not create a request to get an access token")
		return AccessTokenResponse{}, err
	}

	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-type", "application/x-www-form-urlencoded")

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		log.WithError(err).Error("Could not request access token")
		return AccessTokenResponse{}, err
	}
	defer response.Body.Close()

	resp := AccessTokenResponse{}
	err = json.NewDecoder(response.Body).Decode(&resp)
	if err != nil {
		log.WithError(err).Error("Could not decode error payload from a non 200 response")
		return AccessTokenResponse{}, err
	}

	if resp.Error != "" {
		switch resp.Error {
		case "authorization_pending":
			log.Errorf("authorization_pending: %s", resp.ErrorDescription)
			return AccessTokenResponse{}, ErrAuthorizationPending
		case "slow_down":
			log.Errorf("slow_down: %s", resp.ErrorDescription)
			return AccessTokenResponse{}, ErrSlowDown
		case "expired_token":
			log.Errorf("expired_token: %s", resp.ErrorDescription)
			return AccessTokenResponse{}, ErrExpiredToken
		case "unsupported_grant_type":
			log.Errorf("unsupported_grant_type: %s", resp.ErrorDescription)
			return AccessTokenResponse{}, ErrInvalidGrantType
		case "incorrect_client_credentials":
			log.Errorf("incorrect_client_credentials: %s", resp.ErrorDescription)
			return AccessTokenResponse{}, ErrIncorrectCredentials
		case "incorrect_device_code":
			log.Errorf("incorrect_device_code: %s", resp.ErrorDescription)
			return AccessTokenResponse{}, ErrIncorrectDeviceCode
		case "access_denied":
			log.Errorf("access_denied: %s", resp.ErrorDescription)
			return AccessTokenResponse{}, ErrAccessDenied
		default:
			log.Errorf("Unknown error received, %s: %s", resp.Error, resp.ErrorDescription)
			return AccessTokenResponse{}, fmt.Errorf("error when fetching access token, %s", resp.ErrorDescription)
		}
	}

	return resp, nil
}

// GetUserInfo calls the /user endpoint on the GitHub API to get some basic user information
// user the given access token.
func (c OAuthClient) GetUserInfo(ctx context.Context, accessToken string) (github.User, error) {
	request, err := http.NewRequestWithContext(ctx, "GET", c.identityURL, nil)
	if err != nil {
		log.WithError(err).Error("Could not create a request to get the current user")
		return github.User{}, err
	}

	request.Header = map[string][]string{
		"Accept":        {"application/vnd.github.v3+json"},
		"Authorization": {fmt.Sprintf("Bearer %s", accessToken)},
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil || response.StatusCode != 200 {
		log.WithError(err).Error("Could not make a request for user info to Github.")
		return github.User{}, err
	}

	defer response.Body.Close()

	resp := github.User{}
	err = json.NewDecoder(response.Body).Decode(&resp)
	if err != nil {
		log.WithError(err).Error("Could not decode JSON body.")
		return github.User{}, err
	}

	return resp, err
}

// HandleDeviceCodePolling handles the flow of polling the OAuth provider for an OAuth
// access token.
func (c OAuthClient) HandleDeviceCodePolling(ctx context.Context, user *model.Users,
	deviceCode DeviceCodeResponse, secretKey string) {

	currentTime := time.Now()

	// As long as the code is not expired
	for currentTime.Before(currentTime.Add(time.Duration(deviceCode.ExpiresIn) * time.Second)) {
		pollingResp, err := c.PollDeviceAuth(deviceCode.DeviceCode)
		if err != nil {
			// If we got an error, process it as intended
			switch {
			case errors.Is(err, ErrAuthorizationPending):
				// If the request is still pending, wait then continue
				time.Sleep(time.Duration(deviceCode.Interval) * time.Second)
				continue
			case errors.Is(err, ErrSlowDown):
				// If the request is still pending and we went too fast, wait then continue
				time.Sleep(time.Duration(deviceCode.Interval) * 2 * time.Second)
				continue
			case errors.Is(err, ErrAccessDenied) ||
				errors.Is(err, ErrExpiredToken):
				// If the request was denied, set the user to denied then cancel
				user.Status = model.UserStatus_Denied

				err = models.SaveUser(ctx, user)
				if err != nil {
					log.WithError(err).Error("Could not save user in the database.")
					dropUser(ctx, user)
				}
				return
			case errors.Is(err, ErrIncorrectDeviceCode) ||
				errors.Is(err, ErrIncorrectCredentials) ||
				errors.Is(err, ErrInvalidGrantType):
				// If the request was invalid, drop the user and cance
				log.WithError(err).Error("Error when trying to pool for user access token, cancelling")
				dropUser(ctx, user)
				return
			case err != nil:
				log.WithError(err).Error("Unknown error when trying to pool for user access")
				dropUser(ctx, user)
				return
			}
		}

		userInfo, err := c.GetUserInfo(ctx, pollingResp.AccessToken)
		if err != nil {
			log.WithError(err).Error("Could not make a request for user info to Github.")
			dropUser(ctx, user)
			return
		}

		token, err := EncryptAccessToken(pollingResp.AccessToken, secretKey)
		if err != nil {
			log.WithError(err).Error("Could not encrypt access token for db storage.")
			dropUser(ctx, user)
			return
		}

		existingUser, _ := models.GetUserByUniqueID(ctx, *userInfo.NodeID)
		// Make sure to not create duplicates if this is a reauth
		if existingUser != nil {
			// Transfer the API keys from the temp user to the existing user so we don't lose the newly created key
			err := models.TransferApiKeys(ctx, user.ID, existingUser.ID)
			if err != nil {
				// If we got an error, we likely could not transfer the keys. This user is now in a bad state.
				log.WithError(err).WithFields(map[string]interface{}{
					"corrupted_user_id": existingUser.ID,
				}).Error("Could not transfer api keys to the exiting account for this user, key will be deleted")
			}

			dropUser(ctx, user)
			user = existingUser
		}

		user.Token = &token
		user.Username = userInfo.Login
		user.UniqueID = userInfo.NodeID
		user.Status = model.UserStatus_Accepted

		err = models.SaveUser(ctx, user)
		if err != nil {
			log.WithError(err).Error("Could not save user in the database, user record may be corrupted.")
		}
		return
	}
}

func dropUser(ctx context.Context, user *model.Users) {
	err := models.DeleteUser(ctx, user)
	if err != nil {
		log.WithError(err).Error("Could not drop user")
	}
}
