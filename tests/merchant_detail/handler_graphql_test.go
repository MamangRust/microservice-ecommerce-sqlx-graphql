package merchant_detail_test

import (
	"context"
	"net/http"
	"testing"

	tests "github.com/MamangRust/microservice-ecommerce-test"
	"github.com/stretchr/testify/suite"
)

type MerchantDetailGraphqlTestSuite struct {
	tests.BaseTestSuite
	handler    http.Handler
	merchantID int
}

func (s *MerchantDetailGraphqlTestSuite) SetupSuite() {
	s.BaseTestSuite.SetupSuite()
	s.SetupUserService()
	s.SetupMerchantService()
	s.SetupMerchantDetailService()
	s.handler = s.NewGraphQLHandler()

	ctx := context.Background()
	userID := s.SeedUser(ctx)
	s.merchantID = s.SeedMerchant(ctx, userID)
}

func (s *MerchantDetailGraphqlTestSuite) TestMerchantDetailGraphqlLifecycle() {
	img := []byte("dummy image content")

	// 1. Create
	create := s.GQLUpload(s.handler, `
		mutation Create($input: CreateMerchantDetailInput!) {
			createMerchantDetail(input: $input) { status message data { id display_name } }
		}`, map[string]interface{}{"input": map[string]interface{}{
		"merchant_id":       s.merchantID,
		"display_name":      "GQL Detail",
		"short_description": "Testing via GraphQL",
		"website_url":       "https://example.com",
	}},
		tests.GraphQLFile{Path: "input.cover_image", Filename: "cover.jpg", Content: img},
		tests.GraphQLFile{Path: "input.logo", Filename: "logo.jpg", Content: img},
	)
	created := s.Payload(create, "createMerchantDetail")
	detailID := s.Int(created, "id")
	s.Equal("GQL Detail", s.Str(created, "display_name"))
	s.Require().NotZero(detailID)

	// 2. FindById
	found := s.GQL(s.handler, `
		query Find($input: FindByIdMerchantDetailInput!) { findMerchantDetailById(input: $input) { status data { id display_name } } }`,
		map[string]interface{}{"input": map[string]interface{}{"id": detailID}})
	s.Equal(detailID, s.Int(s.Payload(found, "findMerchantDetailById"), "id"))

	// 3. FindAll
	all := s.GQL(s.handler, `query { findAllMerchantDetails(input: {page: 1, page_size: 10}) { status data { id } } }`, nil)
	s.NotEmpty(s.Items(all, "findAllMerchantDetails"))

	// 4. FindByActive
	active := s.GQL(s.handler, `query { findActiveMerchantDetails(input: {page: 1, page_size: 10}) { status } }`, nil)
	s.Equal("success", s.Status(active, "findActiveMerchantDetails"))

	// 5. Update
	update := s.GQLUpload(s.handler, `
		mutation Update($input: UpdateMerchantDetailInput!) {
			updateMerchantDetail(input: $input) { status data { id display_name } }
		}`, map[string]interface{}{"input": map[string]interface{}{
		"merchant_detail_id": detailID,
		"display_name":       "Updated GQL Detail",
		"short_description":  "Updated via GraphQL",
		"website_url":        "https://updated.example.com",
	}},
		tests.GraphQLFile{Path: "input.cover_image", Filename: "cover_updated.jpg", Content: img},
		tests.GraphQLFile{Path: "input.logo", Filename: "logo_updated.jpg", Content: img},
	)
	s.Equal("Updated GQL Detail", s.Str(s.Payload(update, "updateMerchantDetail"), "display_name"))

	// 6. Trash
	trashed := s.GQL(s.handler, `
		mutation Trash($input: FindByIdMerchantDetailInput!) { trashMerchantDetail(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"id": detailID}})
	s.Equal("success", s.Status(trashed, "trashMerchantDetail"))

	// 7. FindByTrashed
	trashedList := s.GQL(s.handler, `query { findTrashedMerchantDetails(input: {page: 1, page_size: 10}) { status } }`, nil)
	s.Equal("success", s.Status(trashedList, "findTrashedMerchantDetails"))

	// 8. Restore
	restored := s.GQL(s.handler, `
		mutation Restore($input: FindByIdMerchantDetailInput!) { restoreMerchantDetail(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"id": detailID}})
	s.Equal("success", s.Status(restored, "restoreMerchantDetail"))

	// 9. DeletePermanent
	s.GQL(s.handler, `
		mutation Trash($input: FindByIdMerchantDetailInput!) { trashMerchantDetail(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"id": detailID}})
	permanent := s.GQL(s.handler, `
		mutation DeletePermanent($input: FindByIdMerchantDetailInput!) { deleteMerchantDetailPermanent(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"id": detailID}})
	s.Equal("success", s.Status(permanent, "deleteMerchantDetailPermanent"))

	// 10. RestoreAll
	restoreAll := s.GQL(s.handler, `mutation { restoreAllMerchantDetails { status } }`, nil)
	s.Equal("success", s.Status(restoreAll, "restoreAllMerchantDetails"))

	// 11. DeleteAll
	deleteAll := s.GQL(s.handler, `mutation { deleteAllMerchantDetailsPermanent { status } }`, nil)
	s.Equal("success", s.Status(deleteAll, "deleteAllMerchantDetailsPermanent"))
}

func TestMerchantDetailGraphqlSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(MerchantDetailGraphqlTestSuite))
}
