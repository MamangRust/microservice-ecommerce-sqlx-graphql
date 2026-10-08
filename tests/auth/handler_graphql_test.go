package auth_test

import (
	"net/http"
	"testing"

	graphtest "github.com/MamangRust/microservice-ecommerce-grpc-apigateway/graphtest"
	tests "github.com/MamangRust/microservice-ecommerce-test"
	"github.com/stretchr/testify/suite"
)

type AuthGraphqlTestSuite struct {
	tests.BaseTestSuite
	handler  http.Handler
	email    string
	password string
}

func (s *AuthGraphqlTestSuite) SetupSuite() {
	s.BaseTestSuite.SetupSuite()
	s.SetupAuthService()
	s.handler = s.NewGraphQLHandler()
	s.email = "auth.graphql.test@example.com"
	s.password = "password123"
}

func (s *AuthGraphqlTestSuite) TestAuthGraphqlLifecycle() {
	// 1. Register
	register := s.GQL(s.handler, `
		mutation Register($input: RegisterInput!) {
			registerUser(input: $input) { status message data { id email } }
		}`, map[string]interface{}{"input": map[string]interface{}{
		"firstname":        "Auth",
		"lastname":         "GraphQL",
		"email":            s.email,
		"password":         s.password,
		"confirm_password": s.password,
	}})
	registered := s.Payload(register, "registerUser")
	userID := s.Int(registered, "id")
	s.Equal(s.email, s.Str(registered, "email"))
	s.Require().NotZero(userID)

	// 2. Login
	login := s.GQL(s.handler, `
		mutation Login($input: LoginInput!) {
			loginUser(input: $input) { status data { access_token refresh_token } }
		}`, map[string]interface{}{"input": map[string]interface{}{"email": s.email, "password": s.password}})
	tokens := s.Payload(login, "loginUser")
	accessToken := s.Str(tokens, "access_token")
	refreshToken := s.Str(tokens, "refresh_token")
	s.Require().NotEmpty(accessToken)
	s.Require().NotEmpty(refreshToken)

	// 3. GetMe — the resolver reads the user id from the request context, which
	// AuthMiddleware normally populates; WithUser emulates that.
	authed := graphtest.WithUser(s.handler, userID)
	me := s.GQL(authed, `
		query Me($input: GetMeInput!) { getMe(input: $input) { status data { id email } } }`,
		map[string]interface{}{"input": map[string]interface{}{"access_token": accessToken}})
	s.Equal(s.email, s.Str(s.Payload(me, "getMe"), "email"))

	// 4. RefreshToken
	refresh := s.GQL(s.handler, `
		mutation Refresh($input: RefreshTokenInput!) {
			refreshToken(input: $input) { status data { access_token } }
		}`, map[string]interface{}{"input": map[string]interface{}{"refresh_token": refreshToken}})
	s.Equal("success", s.Status(refresh, "refreshToken"))
}

func TestAuthGraphqlSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(AuthGraphqlTestSuite))
}
