package merchant_policy_test

import (
	"context"
	"net/http"
	"testing"

	tests "github.com/MamangRust/microservice-ecommerce-test"
	"github.com/stretchr/testify/suite"
)

type MerchantPolicyGraphqlTestSuite struct {
	tests.BaseTestSuite
	handler    http.Handler
	merchantID int
}

func (s *MerchantPolicyGraphqlTestSuite) SetupSuite() {
	s.BaseTestSuite.SetupSuite()
	s.SetupUserService()
	s.SetupMerchantService()
	s.SetupMerchantPolicyService()
	s.handler = s.NewGraphQLHandler()

	ctx := context.Background()
	userID := s.SeedUser(ctx)
	s.merchantID = s.SeedMerchant(ctx, userID)
}

func (s *MerchantPolicyGraphqlTestSuite) TestMerchantPolicyGraphqlLifecycle() {
	// 1. Create
	create := s.GQL(s.handler, `
		mutation Create($input: CreateMerchantPoliciesInput!) {
			createMerchantPolicy(input: $input) { status message data { id title } }
		}`, map[string]interface{}{"input": map[string]interface{}{
		"merchant_id": s.merchantID,
		"policy_type": "Return",
		"title":       "GQL Policy",
		"description": "Testing via GraphQL",
	}})
	created := s.Payload(create, "createMerchantPolicy")
	policyID := s.Int(created, "id")
	s.Equal("GQL Policy", s.Str(created, "title"))
	s.Require().NotZero(policyID)

	// 2. FindById
	found := s.GQL(s.handler, `
		query Find($input: FindByIdMerchantPoliciesInput!) { findMerchantPolicyById(input: $input) { status data { id title } } }`,
		map[string]interface{}{"input": map[string]interface{}{"id": policyID}})
	s.Equal(policyID, s.Int(s.Payload(found, "findMerchantPolicyById"), "id"))

	// 3. FindAll
	all := s.GQL(s.handler, `query { findAllMerchantPolicies(input: {page: 1, page_size: 10}) { status data { id } } }`, nil)
	s.NotEmpty(s.Items(all, "findAllMerchantPolicies"))

	// 4. FindByActive
	active := s.GQL(s.handler, `query { findActiveMerchantPolicies(input: {page: 1, page_size: 10}) { status } }`, nil)
	s.Equal("success", s.Status(active, "findActiveMerchantPolicies"))

	// 5. Update
	update := s.GQL(s.handler, `
		mutation Update($input: UpdateMerchantPoliciesInput!) {
			updateMerchantPolicy(input: $input) { status data { id title } }
		}`, map[string]interface{}{"input": map[string]interface{}{
		"merchant_policy_id": policyID,
		"policy_type":        "Privacy",
		"title":              "Updated GQL Policy",
		"description":        "Updated via GraphQL",
	}})
	s.Equal("Updated GQL Policy", s.Str(s.Payload(update, "updateMerchantPolicy"), "title"))

	// 6. Trash
	trashed := s.GQL(s.handler, `
		mutation Trash($input: FindByIdMerchantPoliciesInput!) { trashMerchantPolicy(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"id": policyID}})
	s.Equal("success", s.Status(trashed, "trashMerchantPolicy"))

	// 7. FindByTrashed
	trashedList := s.GQL(s.handler, `query { findTrashedMerchantPolicies(input: {page: 1, page_size: 10}) { status } }`, nil)
	s.Equal("success", s.Status(trashedList, "findTrashedMerchantPolicies"))

	// 8. Restore
	restored := s.GQL(s.handler, `
		mutation Restore($input: FindByIdMerchantPoliciesInput!) { restoreMerchantPolicy(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"id": policyID}})
	s.Equal("success", s.Status(restored, "restoreMerchantPolicy"))

	// 9. DeletePermanent
	s.GQL(s.handler, `
		mutation Trash($input: FindByIdMerchantPoliciesInput!) { trashMerchantPolicy(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"id": policyID}})
	permanent := s.GQL(s.handler, `
		mutation DeletePermanent($input: FindByIdMerchantPoliciesInput!) { deleteMerchantPolicyPermanent(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"id": policyID}})
	s.Equal("success", s.Status(permanent, "deleteMerchantPolicyPermanent"))

	// 10. RestoreAll
	restoreAll := s.GQL(s.handler, `mutation { restoreAllMerchantPolicies { status } }`, nil)
	s.Equal("success", s.Status(restoreAll, "restoreAllMerchantPolicies"))

	// 11. DeleteAll
	deleteAll := s.GQL(s.handler, `mutation { deleteAllMerchantPoliciesPermanent { status } }`, nil)
	s.Equal("success", s.Status(deleteAll, "deleteAllMerchantPoliciesPermanent"))
}

func TestMerchantPolicyGraphqlSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(MerchantPolicyGraphqlTestSuite))
}
