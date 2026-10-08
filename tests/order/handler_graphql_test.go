package order_test

import (
	"context"
	"net/http"
	"testing"

	tests "github.com/MamangRust/microservice-ecommerce-test"
	"github.com/stretchr/testify/suite"
)

type OrderGraphqlTestSuite struct {
	tests.BaseTestSuite
	handler    http.Handler
	userID     int
	merchantID int
	productID  int
}

func (s *OrderGraphqlTestSuite) SetupSuite() {
	s.BaseTestSuite.SetupSuite()
	s.SetupUserService()
	s.SetupCategoryService()
	s.SetupMerchantService()
	s.SetupProductService()
	s.SetupOrderItemService()
	s.SetupShippingAddressService()
	s.SetupTransactionService()
	s.SetupOrderService()
	s.handler = s.NewGraphQLHandler()

	ctx := context.Background()
	s.userID = s.SeedUser(ctx)
	catID := s.SeedCategory(ctx)
	s.merchantID = s.SeedMerchant(ctx, s.userID)
	s.productID = s.SeedProduct(ctx, s.merchantID, catID)
}

func (s *OrderGraphqlTestSuite) TestOrderGraphqlLifecycle() {
	// 1. Create
	create := s.GQL(s.handler, `
		mutation Create($input: CreateOrderInput!) {
			createOrder(input: $input) { status message data { id user_id merchant_id total_price } }
		}`, map[string]interface{}{"input": map[string]interface{}{
		"merchant_id": s.merchantID,
		"user_id":     s.userID,
		"total_price": 15000,
		"items": []map[string]interface{}{
			{"product_id": s.productID, "quantity": 1, "price": 15000},
		},
		"shipping": map[string]interface{}{
			"alamat":          "GQL Address",
			"provinsi":        "GQL Province",
			"kota":            "GQL City",
			"courier":         "GQL Courier",
			"shipping_method": "GQL Method",
			"shipping_cost":   1000,
			"negara":          "GQL Country",
		},
	}})
	created := s.Payload(create, "createOrder")
	orderID := s.Int(created, "id")
	s.Equal(s.userID, s.Int(created, "user_id"))
	s.Require().NotZero(orderID)

	// 2. FindById
	found := s.GQL(s.handler, `
		query Find($input: FindByIdOrderInput!) { findOrderById(input: $input) { status data { id } } }`,
		map[string]interface{}{"input": map[string]interface{}{"id": orderID}})
	s.Equal(orderID, s.Int(s.Payload(found, "findOrderById"), "id"))

	// 3. FindAll
	all := s.GQL(s.handler, `query { findAllOrders(input: {page: 1, page_size: 10}) { status data { id } } }`, nil)
	s.NotEmpty(s.Items(all, "findAllOrders"))

	// 4. FindByActive
	active := s.GQL(s.handler, `query { findActiveOrders(input: {page: 1, page_size: 10}) { status } }`, nil)
	s.Equal("success", s.Status(active, "findActiveOrders"))

	// 5. Update — UpdateOrderInput requires the existing item/shipping ids.
	// The service recomputes total_price as sum(product.price * quantity) +
	// shipping_cost, i.e. 10000*2 + 1000 = 21000 (item price comes from the
	// product, not the request).
	items := s.GQL(s.handler, `
		query Items($input: FindByIdOrderItemInput!) { findOrderItemsByOrder(input: $input) { data { id } } }`,
		map[string]interface{}{"input": map[string]interface{}{"id": orderID}})
	itemList := s.Items(items, "findOrderItemsByOrder")
	s.Require().NotEmpty(itemList)
	orderItemID := s.Int(itemList[0].(map[string]interface{}), "id")

	shipping := s.GQL(s.handler, `
		query Ship($input: FindByIdShippingRequest!) { findShippingByOrder(input: $input) { data { id } } }`,
		map[string]interface{}{"input": map[string]interface{}{"id": orderID}})
	shippingID := s.Int(s.Payload(shipping, "findShippingByOrder"), "id")

	update := s.GQL(s.handler, `
		mutation Update($input: UpdateOrderInput!) {
			updateOrder(input: $input) { status data { id total_price } }
		}`, map[string]interface{}{"input": map[string]interface{}{
		"order_id":    orderID,
		"user_id":     s.userID,
		"total_price": 20000,
		"items": []map[string]interface{}{
			{"order_item_id": orderItemID, "product_id": s.productID, "quantity": 2, "price": 20000},
		},
		"shipping": map[string]interface{}{
			"shipping_id":     shippingID,
			"alamat":          "GQL Address Updated",
			"provinsi":        "GQL Province",
			"kota":            "GQL City",
			"courier":         "GQL Courier",
			"shipping_method": "GQL Method",
			"shipping_cost":   1000,
			"negara":          "GQL Country",
		},
	}})
	s.Equal(21000, s.Int(s.Payload(update, "updateOrder"), "total_price"))

	// 6. Trash
	trashed := s.GQL(s.handler, `
		mutation Trash($input: FindByIdOrderInput!) { trashOrder(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"id": orderID}})
	s.Equal("success", s.Status(trashed, "trashOrder"))

	// 7. FindByTrashed
	trashedList := s.GQL(s.handler, `query { findTrashedOrders(input: {page: 1, page_size: 10}) { status } }`, nil)
	s.Equal("success", s.Status(trashedList, "findTrashedOrders"))

	// 8. Restore
	restored := s.GQL(s.handler, `
		mutation Restore($input: FindByIdOrderInput!) { restoreOrder(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"id": orderID}})
	s.Equal("success", s.Status(restored, "restoreOrder"))

	// 9. DeletePermanent
	s.GQL(s.handler, `
		mutation Trash($input: FindByIdOrderInput!) { trashOrder(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"id": orderID}})
	permanent := s.GQL(s.handler, `
		mutation DeletePermanent($input: FindByIdOrderInput!) { deleteOrderPermanent(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"id": orderID}})
	s.Equal("success", s.Status(permanent, "deleteOrderPermanent"))

	// 10. RestoreAll
	restoreAll := s.GQL(s.handler, `mutation { restoreAllOrders { status } }`, nil)
	s.Equal("success", s.Status(restoreAll, "restoreAllOrders"))

	// 11. DeleteAll
	deleteAll := s.GQL(s.handler, `mutation { deleteAllOrdersPermanent { status } }`, nil)
	s.Equal("success", s.Status(deleteAll, "deleteAllOrdersPermanent"))
}

func TestOrderGraphqlSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(OrderGraphqlTestSuite))
}
