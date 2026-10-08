package transaction_test

import (
	"context"
	"net/http"
	"testing"

	tests "github.com/MamangRust/microservice-ecommerce-test"
	"github.com/stretchr/testify/suite"
)

type TransactionGraphqlTestSuite struct {
	tests.BaseTestSuite
	handler    http.Handler
	userID     int
	merchantID int
	orderID    int
}

func (s *TransactionGraphqlTestSuite) SetupSuite() {
	s.BaseTestSuite.SetupSuite()
	s.SetupRoleService()
	s.SetupUserService()
	s.SetupCategoryService()
	s.SetupMerchantService()
	s.SetupProductService()
	s.SetupShippingAddressService()
	s.SetupOrderItemService()
	s.SetupOrderService()
	s.SetupTransactionService()
	s.handler = s.NewGraphQLHandler()

	ctx := context.Background()
	s.userID = s.SeedUser(ctx)
	s.merchantID = s.SeedMerchant(ctx, s.userID)
	catID := s.SeedCategory(ctx)
	prodID := s.SeedProduct(ctx, s.merchantID, catID)
	// SeedOrder already creates the order's line item and shipping address.
	s.orderID = s.SeedOrder(ctx, s.userID, s.merchantID, prodID)
}

func (s *TransactionGraphqlTestSuite) TestTransactionGraphqlLifecycle() {
	// 1. Create
	create := s.GQL(s.handler, `
		mutation Create($input: CreateTransactionRequest!) {
			createTransaction(input: $input) { status message data { id payment_method amount } }
		}`, map[string]interface{}{"input": map[string]interface{}{
		"order_id":       s.orderID,
		"merchant_id":    s.merchantID,
		"user_id":        s.userID,
		"payment_method": "credit_card",
		"amount":         15000,
	}})
	created := s.Payload(create, "createTransaction")
	transactionID := s.Int(created, "id")
	s.Equal("credit_card", s.Str(created, "payment_method"))
	s.Require().NotZero(transactionID)

	// 2. FindById
	found := s.GQL(s.handler, `
		query Find($input: FindByIdTransactionRequest!) { findTransactionById(input: $input) { status data { id } } }`,
		map[string]interface{}{"input": map[string]interface{}{"id": transactionID}})
	s.Equal(transactionID, s.Int(s.Payload(found, "findTransactionById"), "id"))

	// 3. FindAll
	all := s.GQL(s.handler, `query { findAllTransaction(input: {page: 1, page_size: 10}) { status data { id } } }`, nil)
	s.NotEmpty(s.Items(all, "findAllTransaction"))

	// 4. FindByActive
	active := s.GQL(s.handler, `query { findByActive(input: {page: 1, page_size: 10}) { status } }`, nil)
	s.Equal("success", s.Status(active, "findByActive"))

	// 5. Update — the service recomputes the stored amount from the order
	// (merchandise 10000 + shipping 1000, plus 11% tax = 12210); the requested
	// amount only has to cover it.
	update := s.GQL(s.handler, `
		mutation Update($input: UpdateTransactionRequest!) {
			updateTransaction(input: $input) { status data { id amount } }
		}`, map[string]interface{}{"input": map[string]interface{}{
		"transaction_id": transactionID,
		"order_id":       s.orderID,
		"merchant_id":    s.merchantID,
		"payment_method": "bank_transfer",
		"amount":         25000,
	}})
	s.Equal(12210, s.Int(s.Payload(update, "updateTransaction"), "amount"))

	// 6. Trash
	trashed := s.GQL(s.handler, `
		mutation Trash($input: FindByIdTransactionRequest!) { trashedTransaction(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"id": transactionID}})
	s.Equal("success", s.Status(trashed, "trashedTransaction"))

	// 7. FindByTrashed
	trashedList := s.GQL(s.handler, `query { findByTrashed(input: {page: 1, page_size: 10}) { status } }`, nil)
	s.Equal("success", s.Status(trashedList, "findByTrashed"))

	// 8. Restore
	restored := s.GQL(s.handler, `
		mutation Restore($input: FindByIdTransactionRequest!) { restoreTransaction(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"id": transactionID}})
	s.Equal("success", s.Status(restored, "restoreTransaction"))

	// 9. DeletePermanent
	s.GQL(s.handler, `
		mutation Trash($input: FindByIdTransactionRequest!) { trashedTransaction(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"id": transactionID}})
	permanent := s.GQL(s.handler, `
		mutation DeletePermanent($input: FindByIdTransactionRequest!) { deleteTransactionPermanent(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"id": transactionID}})
	s.Equal("success", s.Status(permanent, "deleteTransactionPermanent"))

	// 10. RestoreAll
	restoreAll := s.GQL(s.handler, `mutation { restoreAllTransaction { status } }`, nil)
	s.Equal("success", s.Status(restoreAll, "restoreAllTransaction"))

	// 11. DeleteAll
	deleteAll := s.GQL(s.handler, `mutation { deleteAllTransactionPermanent { status } }`, nil)
	s.Equal("success", s.Status(deleteAll, "deleteAllTransactionPermanent"))
}

func TestTransactionGraphqlSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(TransactionGraphqlTestSuite))
}
