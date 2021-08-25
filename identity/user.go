package identity

import (
	"context"
	"fmt"

	"encore.dev/beta/errs"
	log "github.com/sirupsen/logrus"

	"encore.app/identity/github"
	"encore.app/identity/helpers"
	"encore.app/identity/models"
)

var secrets struct {
	// Client ID for the GitHub OAuth App
	GithubClientID string

	// Secret key for encryption purposes
	SecretKey string
}

// SignInResponse is the response from the sign-in endpoint
type SignInResponse struct {
	// A message to inform the user of the result of the operation
	Message string

	// The API key to be user for all authenticated requests going forward
	ApiKey string
}

// SignIn creates a user account for the given username.
//encore:api public
func SignIn(ctx context.Context) (*SignInResponse, error) {
	deviceCode, err := github.RequestDeviceCode(secrets.GithubClientID)
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

	apiKey, keyRecord, err := createKeyForUser(ctx, user)
	if err != nil {
		log.WithError(err).Error("Could not create an API key for the temporary user")
		return nil, err
	}

	// FIXME: This won't work long term since the oauth process will
	// die if the server dies. Should probably make this into a queue of some
	// kind.
	go github.HandleDeviceCodePolling(
		context.Background(),
		user,
		deviceCode,
		secrets.GithubClientID,
		secrets.SecretKey,
	)

	return &SignInResponse{
		Message: fmt.Sprintf(
			"Sign-in process started, please open the following URL in your "+
				"browser to authenticate with GitHub. %s and copy this device code "+
				"when prompted %s. All future requests should use the API key from this "+
				"response for authentication. Save this key somewhere, it will not be "+
				"available again.",
			deviceCode.VerificationUri,
			deviceCode.UserCode,
		),
		ApiKey: helpers.MergeWithKeyID(apiKey, keyRecord.ID),
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
