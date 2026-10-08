package shipping_address_test

import (
	"context"
	"net/http"
	"testing"

	tests "github.com/MamangRust/microservice-ecommerce-test"
	"github.com/stretchr/testify/suite"
)

type ShippingAddressGraphqlTestSuite struct {
	tests.BaseTestSuite
	handler   http.Handler
	orderID   int
	shippingID int
}

func (s *ShippingAddressGraphqlTestSuite) SetupSuite() {
	s.BaseTestSuite.SetupSuite()
	s.SetupUserService()
	s.SetupCategoryService()
	s.SetupMerchantService()
	s.SetupProductService()
	s.SetupOrderItemService()
	s.SetupShippingAddressService()
	s.SetupOrderService()
	s.handler = s.NewGraphQLHandler()

	ctx := context.Background()
	userID := s.SeedUser(ctx)
	catID := s.SeedCategory(ctx)
	merchID := s.SeedMerchant(ctx, userID)
	prodID := s.SeedProduct(ctx, merchID, catID)
	s.orderID = s.SeedOrder(ctx, userID, merchID, prodID)
}

func (s *ShippingAddressGraphqlTestSuite) TestShippingAddressGraphqlLifecycle() {
	// 1. FindByOrder
	byOrder := s.GQL(s.handler, `
		query ByOrder($input: FindByIdShippingRequest!) { findShippingByOrder(input: $input) { status data { id order_id alamat } } }`,
		map[string]interface{}{"input": map[string]interface{}{"id": s.orderID}})
	shipping := s.Payload(byOrder, "findShippingByOrder")
	s.shippingID = s.Int(shipping, "id")
	s.Require().NotZero(s.shippingID)

	// 2. FindById
	found := s.GQL(s.handler, `
		query Find($input: FindByIdShippingRequest!) { findShippingById(input: $input) { status data { id } } }`,
		map[string]interface{}{"input": map[string]interface{}{"id": s.shippingID}})
	s.Equal(s.shippingID, s.Int(s.Payload(found, "findShippingById"), "id"))

	// 3. FindAll
	all := s.GQL(s.handler, `query { findAllShipping(input: {page: 1, page_size: 10}) { status data { id } } }`, nil)
	s.NotEmpty(s.Items(all, "findAllShipping"))

	// 4. FindByActive
	active := s.GQL(s.handler, `query { findActiveShipping(input: {page: 1, page_size: 10}) { status } }`, nil)
	s.Equal("success", s.Status(active, "findActiveShipping"))

	// 5. Trash
	trashed := s.GQL(s.handler, `
		mutation Trash($input: FindByIdShippingRequest!) { trashedShipping(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"id": s.shippingID}})
	s.Equal("success", s.Status(trashed, "trashedShipping"))

	// 6. FindByTrashed
	trashedList := s.GQL(s.handler, `query { findTrashedShipping(input: {page: 1, page_size: 10}) { status } }`, nil)
	s.Equal("success", s.Status(trashedList, "findTrashedShipping"))

	// 7. Restore
	restored := s.GQL(s.handler, `
		mutation Restore($input: FindByIdShippingRequest!) { restoreShipping(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"id": s.shippingID}})
	s.Equal("success", s.Status(restored, "restoreShipping"))

	// 8. DeletePermanent
	s.GQL(s.handler, `
		mutation Trash($input: FindByIdShippingRequest!) { trashedShipping(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"id": s.shippingID}})
	permanent := s.GQL(s.handler, `
		mutation DeletePermanent($input: FindByIdShippingRequest!) { deleteShippingPermanent(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"id": s.shippingID}})
	s.Equal("success", s.Status(permanent, "deleteShippingPermanent"))

	// 9. RestoreAll
	restoreAll := s.GQL(s.handler, `mutation { restoreAllShipping { status } }`, nil)
	s.Equal("success", s.Status(restoreAll, "restoreAllShipping"))

	// 10. DeleteAll
	deleteAll := s.GQL(s.handler, `mutation { deleteAllShippingPermanent { status } }`, nil)
	s.Equal("success", s.Status(deleteAll, "deleteAllShippingPermanent"))
}

func TestShippingAddressGraphqlSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(ShippingAddressGraphqlTestSuite))
}
