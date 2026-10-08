package user_test

import (
	"net/http"
	"testing"

	tests "github.com/MamangRust/microservice-ecommerce-test"
	"github.com/stretchr/testify/suite"
)

type UserGraphqlTestSuite struct {
	tests.BaseTestSuite
	handler http.Handler
}

func (s *UserGraphqlTestSuite) SetupSuite() {
	s.BaseTestSuite.SetupSuite()
	s.SetupUserService()
	s.handler = s.NewGraphQLHandler()
}

func (s *UserGraphqlTestSuite) TestUserGraphqlLifecycle() {
	email := "graphql.user@example.com"

	// 1. Create
	create := s.GQL(s.handler, `
		mutation Create($input: CreateUserInput!) {
			createUser(input: $input) { status message data { id firstname email } }
		}`, map[string]interface{}{"input": map[string]interface{}{
		"firstname":        "GraphQL",
		"lastname":         "User",
		"email":            email,
		"password":         "password123",
		"confirm_password": "password123",
	}})
	created := s.Payload(create, "createUser")
	userID := s.Int(created, "id")
	s.Equal(email, s.Str(created, "email"))
	s.Require().NotZero(userID)

	// 2. FindAll
	all := s.GQL(s.handler, `query { findAllUsers(input: {page: 1, page_size: 10}) { status data { id } } }`, nil)
	s.NotEmpty(s.Items(all, "findAllUsers"))

	// 3. FindById
	found := s.GQL(s.handler, `
		query Find($input: FindByIdUserInput!) { findByIdUser(input: $input) { status data { id email } } }`,
		map[string]interface{}{"input": map[string]interface{}{"id": userID}})
	s.Equal(userID, s.Int(s.Payload(found, "findByIdUser"), "id"))

	// 4. FindByActive
	active := s.GQL(s.handler, `query { findByActiveUsers(input: {page: 1, page_size: 10}) { status } }`, nil)
	s.Equal("success", s.Status(active, "findByActiveUsers"))

	// 5. Update
	update := s.GQL(s.handler, `
		mutation Update($input: UpdateUserInput!) {
			updateUser(input: $input) { status data { id firstname } }
		}`, map[string]interface{}{"input": map[string]interface{}{
		"id":               userID,
		"firstname":        "Updated",
		"lastname":         "User",
		"email":            email,
		"password":         "password123",
		"confirm_password": "password123",
	}})
	s.Equal("Updated", s.Str(s.Payload(update, "updateUser"), "firstname"))

	// 6. Trash
	trashed := s.GQL(s.handler, `
		mutation Trash($input: FindByIdUserInput!) { trashedUser(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"id": userID}})
	s.Equal("success", s.Status(trashed, "trashedUser"))

	// 7. FindByTrashed
	trashedList := s.GQL(s.handler, `query { findByTrashedUsers(input: {page: 1, page_size: 10}) { status } }`, nil)
	s.Equal("success", s.Status(trashedList, "findByTrashedUsers"))

	// 8. Restore
	restored := s.GQL(s.handler, `
		mutation Restore($input: FindByIdUserInput!) { restoreUser(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"id": userID}})
	s.Equal("success", s.Status(restored, "restoreUser"))

	// 9. DeletePermanent
	s.GQL(s.handler, `
		mutation Trash($input: FindByIdUserInput!) { trashedUser(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"id": userID}})
	permanent := s.GQL(s.handler, `
		mutation DeletePermanent($input: FindByIdUserInput!) { deleteUserPermanent(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"id": userID}})
	s.Equal("success", s.Status(permanent, "deleteUserPermanent"))

	// 10. RestoreAll
	restoreAll := s.GQL(s.handler, `mutation { restoreAllUser { status } }`, nil)
	s.Equal("success", s.Status(restoreAll, "restoreAllUser"))

	// 11. DeleteAll
	deleteAll := s.GQL(s.handler, `mutation { deleteAllUserPermanent { status } }`, nil)
	s.Equal("success", s.Status(deleteAll, "deleteAllUserPermanent"))
}

func TestUserGraphqlSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(UserGraphqlTestSuite))
}
