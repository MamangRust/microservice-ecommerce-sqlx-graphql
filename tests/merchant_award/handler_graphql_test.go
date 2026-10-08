package merchant_award_test

import (
	"context"
	"net/http"
	"testing"

	tests "github.com/MamangRust/microservice-ecommerce-test"
	"github.com/stretchr/testify/suite"
)

type MerchantAwardGraphqlTestSuite struct {
	tests.BaseTestSuite
	handler    http.Handler
	merchantID int
}

func (s *MerchantAwardGraphqlTestSuite) SetupSuite() {
	s.BaseTestSuite.SetupSuite()
	s.SetupUserService()
	s.SetupMerchantService()
	s.SetupMerchantAwardService()
	s.handler = s.NewGraphQLHandler()

	ctx := context.Background()
	userID := s.SeedUser(ctx)
	s.merchantID = s.SeedMerchant(ctx, userID)
}

func (s *MerchantAwardGraphqlTestSuite) TestMerchantAwardGraphqlLifecycle() {
	// 1. Create
	create := s.GQL(s.handler, `
		mutation Create($input: CreateMerchantAwardInput!) {
			createMerchantAward(input: $input) { status message data { id title } }
		}`, map[string]interface{}{"input": map[string]interface{}{
		"merchant_id":     s.merchantID,
		"title":           "GQL Award",
		"description":     "Testing via GraphQL",
		"issued_by":       "GQL Issuer",
		"issue_date":      "2024-01-01",
		"expiry_date":     "2025-01-01",
		"certificate_url": "https://example.com/cert.pdf",
	}})
	created := s.Payload(create, "createMerchantAward")
	awardID := s.Int(created, "id")
	s.Equal("GQL Award", s.Str(created, "title"))
	s.Require().NotZero(awardID)

	// 2. FindById
	found := s.GQL(s.handler, `
		query Find($input: FindByIdMerchantAwardInput!) { findMerchantAwardById(input: $input) { status data { id title } } }`,
		map[string]interface{}{"input": map[string]interface{}{"id": awardID}})
	s.Equal(awardID, s.Int(s.Payload(found, "findMerchantAwardById"), "id"))

	// 3. FindAll
	all := s.GQL(s.handler, `query { findAllMerchantAwards(input: {page: 1, page_size: 10}) { status data { id } } }`, nil)
	s.NotEmpty(s.Items(all, "findAllMerchantAwards"))

	// 4. FindByActive
	active := s.GQL(s.handler, `query { findActiveMerchantAwards(input: {page: 1, page_size: 10}) { status } }`, nil)
	s.Equal("success", s.Status(active, "findActiveMerchantAwards"))

	// 5. Update
	update := s.GQL(s.handler, `
		mutation Update($input: UpdateMerchantAwardInput!) {
			updateMerchantAward(input: $input) { status data { id title } }
		}`, map[string]interface{}{"input": map[string]interface{}{
		"merchant_certification_id": awardID,
		"title":                     "Updated GQL Award",
		"description":               "Updated via GraphQL",
		"issued_by":                 "GQL Issuer",
		"issue_date":                "2024-01-01",
		"expiry_date":               "2025-01-01",
		"certificate_url":           "https://example.com/cert.pdf",
	}})
	s.Equal("Updated GQL Award", s.Str(s.Payload(update, "updateMerchantAward"), "title"))

	// 6. Trash
	trashed := s.GQL(s.handler, `
		mutation Trash($input: FindByIdMerchantAwardInput!) { trashMerchantAward(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"id": awardID}})
	s.Equal("success", s.Status(trashed, "trashMerchantAward"))

	// 7. FindByTrashed
	trashedList := s.GQL(s.handler, `query { findTrashedMerchantAwards(input: {page: 1, page_size: 10}) { status } }`, nil)
	s.Equal("success", s.Status(trashedList, "findTrashedMerchantAwards"))

	// 8. Restore
	restored := s.GQL(s.handler, `
		mutation Restore($input: FindByIdMerchantAwardInput!) { restoreMerchantAward(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"id": awardID}})
	s.Equal("success", s.Status(restored, "restoreMerchantAward"))

	// 9. DeletePermanent
	s.GQL(s.handler, `
		mutation Trash($input: FindByIdMerchantAwardInput!) { trashMerchantAward(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"id": awardID}})
	permanent := s.GQL(s.handler, `
		mutation DeletePermanent($input: FindByIdMerchantAwardInput!) { deleteMerchantAwardPermanent(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"id": awardID}})
	s.Equal("success", s.Status(permanent, "deleteMerchantAwardPermanent"))

	// 10. RestoreAll
	restoreAll := s.GQL(s.handler, `mutation { restoreAllMerchantAwards { status } }`, nil)
	s.Equal("success", s.Status(restoreAll, "restoreAllMerchantAwards"))

	// 11. DeleteAll
	deleteAll := s.GQL(s.handler, `mutation { deleteAllMerchantAwardsPermanent { status } }`, nil)
	s.Equal("success", s.Status(deleteAll, "deleteAllMerchantAwardsPermanent"))
}

func TestMerchantAwardGraphqlSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(MerchantAwardGraphqlTestSuite))
}
