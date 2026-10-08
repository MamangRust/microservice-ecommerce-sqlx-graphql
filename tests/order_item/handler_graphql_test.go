package order_item_test

import (
	"context"
	"net/http"
	"testing"

	tests "github.com/MamangRust/microservice-ecommerce-test"
	"github.com/stretchr/testify/suite"
)

type OrderItemGraphqlTestSuite struct {
	tests.BaseTestSuite
	handler http.Handler
	orderID int
}

func (s *OrderItemGraphqlTestSuite) SetupSuite() {
	s.BaseTestSuite.SetupSuite()
	s.SetupUserService()
	s.SetupCategoryService()
	s.SetupMerchantService()
	s.SetupProductService()
	s.SetupShippingAddressService()
	s.SetupOrderItemService()
	s.SetupOrderService()
	s.handler = s.NewGraphQLHandler()

	ctx := context.Background()
	userID := s.SeedUser(ctx)
	catID := s.SeedCategory(ctx)
	merchID := s.SeedMerchant(ctx, userID)
	prodID := s.SeedProduct(ctx, merchID, catID)
	s.orderID = s.SeedOrder(ctx, userID, merchID, prodID)
}

func (s *OrderItemGraphqlTestSuite) TestOrderItemGraphqlLifecycle() {
	// 1. FindAll
	all := s.GQL(s.handler, `query { findAllOrderItems(input: {page: 1, page_size: 10}) { status data { id } } }`, nil)
	s.Equal("success", s.Status(all, "findAllOrderItems"))

	// 2. FindByOrder
	byOrder := s.GQL(s.handler, `
		query ByOrder($input: FindByIdOrderItemInput!) { findOrderItemsByOrder(input: $input) { status data { id order_id } } }`,
		map[string]interface{}{"input": map[string]interface{}{"id": s.orderID}})
	s.NotEmpty(s.Items(byOrder, "findOrderItemsByOrder"))

	// 3. FindByActive
	active := s.GQL(s.handler, `query { findActiveOrderItems(input: {page: 1, page_size: 10}) { status } }`, nil)
	s.Equal("success", s.Status(active, "findActiveOrderItems"))

	// 4. FindByTrashed
	trashed := s.GQL(s.handler, `query { findTrashedOrderItems(input: {page: 1, page_size: 10}) { status } }`, nil)
	s.Equal("success", s.Status(trashed, "findTrashedOrderItems"))
}

func TestOrderItemGraphqlSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(OrderItemGraphqlTestSuite))
}
