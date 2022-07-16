package identity

import (
	"context"
	"fmt"

	"encore.app/permissions"
	"encore.dev/beta/errs"
	log "github.com/sirupsen/logrus"

	"encore.app/identity/github"
	"encore.app/identity/models"
)

var (
	githubOAuthDeviceCodeURL  = "https://github.com/login/device/code"
	githubOAuthAccessTokenURL = "https://github.com/login/oauth/access_token"
	githubOAuthIdentityURL    = "https://api.github.com/user"
)

// SignInResponse is the response from the sign-in endpoint
type SignInResponse struct {
	// A message to inform the user of the result of the operation
	Message string

	// The API key to be user for all authenticated requests going forward
	ApiKey string
}

// SignIn runs the console signIn process using the device code process from GitHub.
//encore:api public
func SignIn(ctx context.Context) (*SignInResponse, error) {
	client := github.NewOAuthClient(github.NewOAuthClientOptions{
		ClientID:       secrets.GithubClientID,
		DeviceCodeURL:  githubOAuthDeviceCodeURL,
		AccessTokenURL: githubOAuthAccessTokenURL,
		IdentityURL:    githubOAuthIdentityURL,
	})

	deviceCode, err := client.RequestDeviceCode()
	if err != nil {
		log.WithError(err).Error("Could not request a device code")
		return nil, &errs.Error{
			Code: errs.Internal,
			Message: "Could not fetch a device code for your sign-in request, " +
				"please try again later or contact the administrator.",
		}
	}

	user := models.NewPendingUser()
	err = models.SaveUser(ctx, user)
	if err != nil {
		log.WithError(err).Error("Could not save the temporary user record")
		return nil, &errs.Error{
			Code:    errs.Internal,
			Message: "Could not create temporary user account",
		}
	}

	apiKey, key, err := createKeyForUser(ctx, user)
	if err != nil {
		log.WithError(err).Error("Could not create an API key for the temporary user")
		return nil, err
	}

	_, err = permissions.AddPermissionSet(ctx, &permissions.AddPermissionSetParams{
		KeyID:  key.ID,
		UserID: user.ID,
		Role:   "admin",
	})
	if err != nil {
		log.WithError(err).Error("Could not create permission set for api key")
		return nil, err
	}

	// FIXME: This won't work long term since the oauth process will
	// die if the server dies. Should probably make this into a queue of some
	// kind.
	go client.HandleDeviceCodePolling(
		context.Background(),
		user,
		deviceCode,
		secrets.SecretKey,
	)

	return &SignInResponse{
		Message: fmt.Sprintf(
			"Sign-in process started, please open the following URL in your "+
				"browser to authenticate with GitHub. %s and copy this device code "+
				"when prompted %s. All future requests should use the API key from this "+
				"response for authentication or a newly created key. Save this key somewhere, "+
				"it will not be available again and you will not be able to recreate an admin key.",
			deviceCode.VerificationUri,
			deviceCode.UserCode,
		),
		ApiKey: apiKey,
	}, nil
}

// OAuthRedirect is the endpoint to be called when redirecting users from an
// OAuth request using a web flow..
//encore:api public
func OAuthRedirect(ctx context.Context) (*SignInResponse, error) {
	return &SignInResponse{
		Message: "Not implemented, please use the API flow instead",
	}, nil
}
