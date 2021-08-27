package identity

import (
	"context"
	"testing"
	"time"

	"encore.dev/beta/errs"

	"encore.app/identity/models/generated/identity/public/table"
	"encore.dev/storage/sqldb"
	"github.com/go-jet/jet/v2/postgres"
	"github.com/stretchr/testify/require"

	gogithub "github.com/google/go-github/github"
	"github.com/stretchr/testify/assert"

	"encore.app/identity/github"
	"encore.app/identity/models/generated/identity/public/model"
	"encore.app/identity/test_utils"
)

func TestSignIn(t *testing.T) {
	type deviceCodeResponse struct {
		statusCode int
		response   github.DeviceCodeResponse
	}

	type accessTokenResponse struct {
		statusCode int
		response   github.AccessTokenResponse
	}

	type identityResponse struct {
		statusCode int
		response   gogithub.User
	}

	type expected struct {
		response *SignInResponse
		user     *model.Users
		err      error
	}

	tcs := []struct {
		scenario            string
		deviceCodeResponse  *deviceCodeResponse
		accessTokenResponse *accessTokenResponse
		identityResponse    *identityResponse
		expected            expected
	}{
		{
			scenario: "Will create a temporary account and get the access token data from the server",
			expected: expected{
				response: &SignInResponse{},
				user: &model.Users{
					Username: test_utils.DefaultIdentityResponse.Login,
					Token:    test_utils.StringPointer(test_utils.DefaultAccessTokenResponse.AccessToken),
					UniqueID: test_utils.DefaultIdentityResponse.NodeID,
					Status:   model.UserStatus_Accepted,
				},
			},
		},
		{
			scenario: "Fails when the device code fetching fails",
			deviceCodeResponse: &deviceCodeResponse{
				statusCode: 500,
				response:   github.DeviceCodeResponse{},
			},
			expected: expected{
				err: &errs.Error{
					Code: errs.Internal,
					Message: "Could not fetch a device code for your sign-in request, " +
						"please try again later or contact the administrator.",
				},
			},
		},
		{
			scenario: "Keep the user as pending if we wait too long",
			accessTokenResponse: &accessTokenResponse{
				statusCode: 200,
				response: github.AccessTokenResponse{
					Error: "authorization_pending",
				},
			},
			expected: expected{
				response: &SignInResponse{},
				user: &model.Users{
					Status: model.UserStatus_Pending,
				},
			},
		},
		{
			scenario: "Keep the user as denied if user was denied access",
			accessTokenResponse: &accessTokenResponse{
				statusCode: 200,
				response: github.AccessTokenResponse{
					Error: "access_denied",
				},
			},
			expected: expected{
				response: &SignInResponse{},
				user: &model.Users{
					Status: model.UserStatus_Denied,
				},
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.scenario, func(t *testing.T) {
			ctx := context.Background()

			defer test_utils.Cleanup(ctx)

			server, handler := test_utils.CreateTestGithubOAuthServer()
			defer server.Close()

			githubOAuthDeviceCodeURL = server.URL + "/login/device/code"
			githubOAuthAccessTokenURL = server.URL + "/login/oauth/access_token"
			githubOAuthIdentityURL = server.URL + "/user"

			if tc.deviceCodeResponse != nil {
				handler.DeviceCodeResponse(tc.deviceCodeResponse.statusCode, tc.deviceCodeResponse.response)
			}
			if tc.accessTokenResponse != nil {
				handler.AccessTokenResponse(tc.accessTokenResponse.statusCode, tc.accessTokenResponse.response)
			}
			if tc.identityResponse != nil {
				handler.IdentityResponse(tc.identityResponse.statusCode, tc.identityResponse.response)
			}

			response, err := SignIn(ctx)
			// Sleep for a few seconds to make sure the goroutine can finish
			time.Sleep(time.Duration(test_utils.DefaultDeviceCodeResponse.ExpiresIn) * time.Second)

			if err != nil {
				assert.Equal(t, tc.expected.err, err)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotEqual(t, "", response.Message)
				assert.NotEqual(t, "", response.ApiKey)
			}

			if tc.expected.user != nil {
				query, args := postgres.SELECT(
					table.Users.ID,
					table.Users.Username,
					table.Users.Token,
					table.Users.UniqueID,
					table.Users.Status,
					table.Users.UpdatedAt,
					table.Users.CreatedAt,
				).FROM(table.Users).LIMIT(1).Sql()

				user := model.Users{}
				err := sqldb.
					QueryRow(ctx, query, args...).
					Scan(
						&user.ID,
						&user.Username,
						&user.Token,
						&user.UniqueID,
						&user.Status,
						&user.UpdatedAt,
						&user.CreatedAt,
					)
				require.NoError(t, err)

				assert.Equal(t, tc.expected.user.Status, user.Status)
				// Only check this if the token request was successful
				if user.Token != nil {
					token, err := github.DecryptAccessToken(*user.Token, secrets.SecretKey)
					require.NoError(t, err)

					assert.Equal(t, *tc.expected.user.Username, *user.Username)
					assert.Equal(t, *tc.expected.user.Token, token)
					assert.Equal(t, *tc.expected.user.UniqueID, *user.UniqueID)
				}
			}
		})
	}
}
