package role_test

import (
	"net/http"
	"testing"

	tests "github.com/MamangRust/microservice-ecommerce-test"
	"github.com/stretchr/testify/suite"
)

type RoleGraphqlTestSuite struct {
	tests.BaseTestSuite
	handler http.Handler
}

func (s *RoleGraphqlTestSuite) SetupSuite() {
	s.BaseTestSuite.SetupSuite()
	s.SetupRoleService()
	s.handler = s.NewGraphQLHandler()
}

func (s *RoleGraphqlTestSuite) TestRoleGraphqlLifecycle() {
	// 1. Create
	create := s.GQL(s.handler, `
		mutation Create($input: CreateRoleInput!) {
			createRole(input: $input) { status message data { id name } }
		}`, map[string]interface{}{"input": map[string]interface{}{"name": "GQL_ROLE"}})
	created := s.Payload(create, "createRole")
	roleID := s.Int(created, "id")
	s.Equal("GQL_ROLE", s.Str(created, "name"))
	s.Require().NotZero(roleID)

	// 2. FindAll
	all := s.GQL(s.handler, `query { findAllRole(input: {page: 1, page_size: 10}) { status data { id } } }`, nil)
	s.NotEmpty(s.Items(all, "findAllRole"))

	// 3. FindById
	found := s.GQL(s.handler, `
		query Find($input: FindByIdRoleInput!) { findByIdRole(input: $input) { status data { id name } } }`,
		map[string]interface{}{"input": map[string]interface{}{"role_id": roleID}})
	s.Equal(roleID, s.Int(s.Payload(found, "findByIdRole"), "id"))

	// 4. FindByActive
	active := s.GQL(s.handler, `query { findByActiveRole(input: {page: 1, page_size: 10}) { status } }`, nil)
	s.Equal("success", s.Status(active, "findByActiveRole"))

	// 5. FindByTrashed
	trashedList := s.GQL(s.handler, `query { findByTrashedRole(input: {page: 1, page_size: 10}) { status } }`, nil)
	s.Equal("success", s.Status(trashedList, "findByTrashedRole"))

	// 6. FindByUserId
	byUser := s.GQL(s.handler, `
		query ByUser($input: FindByIdUserRoleInput!) { findByUserIdRole(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"user_id": 1}})
	s.Equal("success", s.Status(byUser, "findByUserIdRole"))

	// 7. Update
	update := s.GQL(s.handler, `
		mutation Update($input: UpdateRoleInput!) {
			updateRole(input: $input) { status data { id name } }
		}`, map[string]interface{}{"input": map[string]interface{}{"id": roleID, "name": "GQL_ROLE_UPDATED"}})
	s.Equal("GQL_ROLE_UPDATED", s.Str(s.Payload(update, "updateRole"), "name"))

	// 8. Trash
	trashed := s.GQL(s.handler, `
		mutation Trash($input: FindByIdRoleInput!) { trashedRole(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"role_id": roleID}})
	s.Equal("success", s.Status(trashed, "trashedRole"))

	// 9. Restore
	restored := s.GQL(s.handler, `
		mutation Restore($input: FindByIdRoleInput!) { restoreRole(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"role_id": roleID}})
	s.Equal("success", s.Status(restored, "restoreRole"))

	// 10. DeletePermanent
	s.GQL(s.handler, `
		mutation Trash($input: FindByIdRoleInput!) { trashedRole(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"role_id": roleID}})
	permanent := s.GQL(s.handler, `
		mutation DeletePermanent($input: FindByIdRoleInput!) { deleteRolePermanent(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"role_id": roleID}})
	s.Equal("success", s.Status(permanent, "deleteRolePermanent"))

	// 11. RestoreAll
	restoreAll := s.GQL(s.handler, `mutation { restoreAllRole { status } }`, nil)
	s.Equal("success", s.Status(restoreAll, "restoreAllRole"))

	// 12. DeleteAll
	deleteAll := s.GQL(s.handler, `mutation { deleteAllRolePermanent { status } }`, nil)
	s.Equal("success", s.Status(deleteAll, "deleteAllRolePermanent"))
}

func TestRoleGraphqlSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(RoleGraphqlTestSuite))
}
