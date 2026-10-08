package merchant_business_test

import (
	"context"
	"net/http"
	"testing"

	tests "github.com/MamangRust/microservice-ecommerce-test"
	"github.com/stretchr/testify/suite"
)

type MerchantBusinessGraphqlTestSuite struct {
	tests.BaseTestSuite
	handler    http.Handler
	merchantID int
}

func (s *MerchantBusinessGraphqlTestSuite) SetupSuite() {
	s.BaseTestSuite.SetupSuite()
	s.SetupUserService()
	s.SetupMerchantService()
	s.SetupMerchantBusinessService()
	s.handler = s.NewGraphQLHandler()

	ctx := context.Background()
	userID := s.SeedUser(ctx)
	s.merchantID = s.SeedMerchant(ctx, userID)
}

func (s *MerchantBusinessGraphqlTestSuite) TestMerchantBusinessGraphqlLifecycle() {
	// 1. Create
	create := s.GQL(s.handler, `
		mutation Create($input: CreateMerchantBusinessInput!) {
			createMerchantBusiness(input: $input) { status message data { id business_type } }
		}`, map[string]interface{}{"input": map[string]interface{}{
		"merchant_id":         s.merchantID,
		"business_type":       "Retail",
		"tax_id":              "TAX-123",
		"established_year":    2010,
		"number_of_employees": 25,
		"website_url":         "https://example.com",
	}})
	created := s.Payload(create, "createMerchantBusiness")
	businessID := s.Int(created, "id")
	s.Equal("Retail", s.Str(created, "business_type"))
	s.Require().NotZero(businessID)

	// 2. FindById
	found := s.GQL(s.handler, `
		query Find($input: FindByIdMerchantBusinessInput!) { findMerchantBusinessById(input: $input) { status data { id } } }`,
		map[string]interface{}{"input": map[string]interface{}{"id": businessID}})
	s.Equal(businessID, s.Int(s.Payload(found, "findMerchantBusinessById"), "id"))

	// 3. FindAll
	all := s.GQL(s.handler, `query { findAllMerchantBusinesses(input: {page: 1, page_size: 10}) { status data { id } } }`, nil)
	s.NotEmpty(s.Items(all, "findAllMerchantBusinesses"))

	// 4. FindByActive
	active := s.GQL(s.handler, `query { findActiveMerchantBusinesses(input: {page: 1, page_size: 10}) { status } }`, nil)
	s.Equal("success", s.Status(active, "findActiveMerchantBusinesses"))

	// 5. Update
	update := s.GQL(s.handler, `
		mutation Update($input: UpdateMerchantBusinessInput!) {
			updateMerchantBusiness(input: $input) { status data { id business_type } }
		}`, map[string]interface{}{"input": map[string]interface{}{
		"merchant_business_info_id": businessID,
		"business_type":             "Wholesale",
		"tax_id":                    "TAX-456",
		"established_year":          2011,
		"number_of_employees":       30,
		"website_url":               "https://updated.example.com",
	}})
	s.Equal("Wholesale", s.Str(s.Payload(update, "updateMerchantBusiness"), "business_type"))

	// 6. Trash
	trashed := s.GQL(s.handler, `
		mutation Trash($input: FindByIdMerchantBusinessInput!) { trashMerchantBusiness(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"id": businessID}})
	s.Equal("success", s.Status(trashed, "trashMerchantBusiness"))

	// 7. FindByTrashed
	trashedList := s.GQL(s.handler, `query { findTrashedMerchantBusinesses(input: {page: 1, page_size: 10}) { status } }`, nil)
	s.Equal("success", s.Status(trashedList, "findTrashedMerchantBusinesses"))

	// 8. Restore
	restored := s.GQL(s.handler, `
		mutation Restore($input: FindByIdMerchantBusinessInput!) { restoreMerchantBusiness(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"id": businessID}})
	s.Equal("success", s.Status(restored, "restoreMerchantBusiness"))

	// 9. DeletePermanent
	s.GQL(s.handler, `
		mutation Trash($input: FindByIdMerchantBusinessInput!) { trashMerchantBusiness(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"id": businessID}})
	permanent := s.GQL(s.handler, `
		mutation DeletePermanent($input: FindByIdMerchantBusinessInput!) { deleteMerchantBusinessPermanent(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"id": businessID}})
	s.Equal("success", s.Status(permanent, "deleteMerchantBusinessPermanent"))

	// 10. RestoreAll
	restoreAll := s.GQL(s.handler, `mutation { restoreAllMerchantBusinesses { status } }`, nil)
	s.Equal("success", s.Status(restoreAll, "restoreAllMerchantBusinesses"))

	// 11. DeleteAll
	deleteAll := s.GQL(s.handler, `mutation { deleteAllMerchantBusinessesPermanent { status } }`, nil)
	s.Equal("success", s.Status(deleteAll, "deleteAllMerchantBusinessesPermanent"))
}

func TestMerchantBusinessGraphqlSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(MerchantBusinessGraphqlTestSuite))
}
