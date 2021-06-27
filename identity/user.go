package identity

import (
	"context"

	"encore.dev/beta/errs"

	"encore.app/identity/models"
)

// SignInParams is the parameters required to create a user account
type SignInParams struct {
	// The username for the new user
	Username string
}

// SignInResponse is the response from the sign-in endpoint
type SignInResponse struct {
	// A message to inform the user of the result of the operation
	Message string
}

// SignIn creates a user account for the given username.
//encore:api public
func SignIn(ctx context.Context, params *SignInParams) (*SignInResponse, error) {
	user := models.NewUser(params.Username)
	err := user.Save(ctx)
	if err != nil {
		return nil, &errs.Error{
			Code:    errs.Internal,
			Message: "Could not sign-in user, user creation failed",
		}
	}

	return &SignInResponse{
		Message: "Signed-in successfully, use the identity.GenerateApiKey endpoint to generate an API key",
	}, nil
}
