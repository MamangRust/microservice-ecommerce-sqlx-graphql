package merchant_test

import (
	"context"
	"net/http"
	"testing"

	tests "github.com/MamangRust/microservice-ecommerce-test"
	"github.com/stretchr/testify/suite"
)

type MerchantGraphqlTestSuite struct {
	tests.BaseTestSuite
	handler http.Handler
	userID  int
}

func (s *MerchantGraphqlTestSuite) SetupSuite() {
	s.BaseTestSuite.SetupSuite()
	s.SetupUserService()
	s.SetupMerchantService()
	s.handler = s.NewGraphQLHandler()
	s.userID = s.SeedUser(context.Background())
}

func (s *MerchantGraphqlTestSuite) TestMerchantGraphqlLifecycle() {
	// 1. Create
	create := s.GQL(s.handler, `
		mutation Create($input: CreateMerchantInput!) {
			createMerchant(input: $input) { status message data { id name } }
		}`, map[string]interface{}{"input": map[string]interface{}{
		"user_id":       s.userID,
		"name":          "GQL Merchant",
		"description":   "Testing via GraphQL",
		"address":       "GQL Address",
		"contact_email": "gql.merchant@example.com",
		"contact_phone": "08123456789",
		"status":        "active",
	}})
	created := s.Payload(create, "createMerchant")
	merchantID := s.Int(created, "id")
	s.Equal("GQL Merchant", s.Str(created, "name"))
	s.Require().NotZero(merchantID)

	// 2. FindById
	found := s.GQL(s.handler, `
		query Find($input: FindByIdMerchantInput!) { findMerchantById(input: $input) { status data { id name } } }`,
		map[string]interface{}{"input": map[string]interface{}{"id": merchantID}})
	s.Equal(merchantID, s.Int(s.Payload(found, "findMerchantById"), "id"))

	// 3. FindAll
	all := s.GQL(s.handler, `query { findAllMerchants(input: {page: 1, page_size: 10}) { status data { id } } }`, nil)
	s.NotEmpty(s.Items(all, "findAllMerchants"))

	// 4. FindByActive
	active := s.GQL(s.handler, `query { findActiveMerchants(input: {page: 1, page_size: 10}) { status } }`, nil)
	s.Equal("success", s.Status(active, "findActiveMerchants"))

	// 5. Update
	update := s.GQL(s.handler, `
		mutation Update($input: UpdateMerchantInput!) {
			updateMerchant(input: $input) { status data { id name } }
		}`, map[string]interface{}{"input": map[string]interface{}{
		"merchant_id":   merchantID,
		"user_id":       s.userID,
		"name":          "Updated GQL Merchant",
		"description":   "Updated via GraphQL",
		"address":       "GQL Address",
		"contact_email": "gql.merchant@example.com",
		"contact_phone": "08123456789",
		"status":        "active",
	}})
	s.Equal("Updated GQL Merchant", s.Str(s.Payload(update, "updateMerchant"), "name"))

	// 6. Trash
	trashed := s.GQL(s.handler, `
		mutation Trash($input: FindByIdMerchantInput!) { trashMerchant(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"id": merchantID}})
	s.Equal("success", s.Status(trashed, "trashMerchant"))

	// 7. FindByTrashed
	trashedList := s.GQL(s.handler, `query { findTrashedMerchants(input: {page: 1, page_size: 10}) { status } }`, nil)
	s.Equal("success", s.Status(trashedList, "findTrashedMerchants"))

	// 8. Restore
	restored := s.GQL(s.handler, `
		mutation Restore($input: FindByIdMerchantInput!) { restoreMerchant(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"id": merchantID}})
	s.Equal("success", s.Status(restored, "restoreMerchant"))

	// 9. DeletePermanent
	s.GQL(s.handler, `
		mutation Trash($input: FindByIdMerchantInput!) { trashMerchant(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"id": merchantID}})
	permanent := s.GQL(s.handler, `
		mutation DeletePermanent($input: FindByIdMerchantInput!) { deleteMerchantPermanent(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"id": merchantID}})
	s.Equal("success", s.Status(permanent, "deleteMerchantPermanent"))

	// 10. RestoreAll
	restoreAll := s.GQL(s.handler, `mutation { restoreAllMerchants { status } }`, nil)
	s.Equal("success", s.Status(restoreAll, "restoreAllMerchants"))

	// 11. DeleteAll
	deleteAll := s.GQL(s.handler, `mutation { deleteAllMerchantsPermanent { status } }`, nil)
	s.Equal("success", s.Status(deleteAll, "deleteAllMerchantsPermanent"))
}

func TestMerchantGraphqlSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(MerchantGraphqlTestSuite))
}
