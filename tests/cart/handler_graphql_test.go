package cart_test

import (
	"context"
	"net/http"
	"testing"

	tests "github.com/MamangRust/microservice-ecommerce-test"
	"github.com/stretchr/testify/suite"
)

type CartGraphqlTestSuite struct {
	tests.BaseTestSuite
	handler   http.Handler
	userID    int
	productID int
}

func (s *CartGraphqlTestSuite) SetupSuite() {
	s.BaseTestSuite.SetupSuite()
	s.SetupUserService()
	s.SetupCategoryService()
	s.SetupMerchantService()
	s.SetupProductService()
	s.SetupCartService()
	s.handler = s.NewGraphQLHandler()

	ctx := context.Background()
	s.userID = s.SeedUser(ctx)
	catID := s.SeedCategory(ctx)
	merchID := s.SeedMerchant(ctx, s.userID)
	s.productID = s.SeedProduct(ctx, merchID, catID)
}

func (s *CartGraphqlTestSuite) TestCartGraphqlLifecycle() {
	// 1. Create
	create := s.GQL(s.handler, `
		mutation Create($input: CreateCartInput!) {
			createCart(input: $input) { status message data { id user_id product_id quantity } }
		}`, map[string]interface{}{"input": map[string]interface{}{
		"user_id":    s.userID,
		"product_id": s.productID,
		"quantity":   2,
	}})
	created := s.Payload(create, "createCart")
	cartID := s.Int(created, "id")
	s.Equal(s.userID, s.Int(created, "user_id"))
	s.Require().NotZero(cartID)

	// 2. FindAll
	all := s.GQL(s.handler, `
		query All($input: FindAllCartInput!) { findAllCarts(input: $input) { status data { id } } }`,
		map[string]interface{}{"input": map[string]interface{}{"user_id": s.userID, "page": 1, "page_size": 10}})
	s.NotEmpty(s.Items(all, "findAllCarts"))

	// 3. Delete
	deleted := s.GQL(s.handler, `
		mutation Delete($input: DeleteCartInput!) { deleteCart(input: $input) { status message } }`,
		map[string]interface{}{"input": map[string]interface{}{"cart_id": cartID, "user_id": s.userID}})
	s.Equal("success", s.Status(deleted, "deleteCart"))

	// 4. DeleteAll — the resolver scopes deletion to the authenticated caller,
	// so the request must carry the user in the context.
	deleteAll := s.GQLWithUser(s.handler, s.userID, `
		mutation DeleteAll($input: DeleteCartsInput!) { deleteAllCarts(input: $input) { status message } }`,
		map[string]interface{}{"input": map[string]interface{}{"cart_ids": []int{cartID}}})
	s.Equal("success", s.Status(deleteAll, "deleteAllCarts"))
}

func TestCartGraphqlSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(CartGraphqlTestSuite))
}
