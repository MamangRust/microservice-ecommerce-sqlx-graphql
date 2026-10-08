package product_test

import (
	"context"
	"net/http"
	"testing"

	tests "github.com/MamangRust/microservice-ecommerce-test"
	"github.com/stretchr/testify/suite"
)

type ProductGraphqlTestSuite struct {
	tests.BaseTestSuite
	handler    http.Handler
	merchantID int
	categoryID int
}

func (s *ProductGraphqlTestSuite) SetupSuite() {
	s.BaseTestSuite.SetupSuite()
	s.SetupUserService()
	s.SetupCategoryService()
	s.SetupMerchantService()
	s.SetupProductService()
	s.handler = s.NewGraphQLHandler()

	ctx := context.Background()
	userID := s.SeedUser(ctx)
	s.categoryID = s.SeedCategory(ctx)
	s.merchantID = s.SeedMerchant(ctx, userID)
}

func (s *ProductGraphqlTestSuite) TestProductGraphqlLifecycle() {
	img := []byte("dummy image content")

	// 1. Create
	create := s.GQLUpload(s.handler, `
		mutation Create($input: CreateProductInput!) {
			createProduct(input: $input) { status message data { id name } }
		}`, map[string]interface{}{"input": map[string]interface{}{
		"merchantId":   s.merchantID,
		"categoryId":   s.categoryID,
		"name":         "GQL Product",
		"description":  "Testing via GraphQL",
		"price":        15000,
		"countInStock": 50,
		"brand":        "GQL Brand",
		"weight":       1000,
		"rating":       5,
		"slugProduct":  "gql-product",
	}}, tests.GraphQLFile{Path: "input.imageProduct", Filename: "product.jpg", Content: img})
	created := s.Payload(create, "createProduct")
	productID := s.Int(created, "id")
	s.Equal("GQL Product", s.Str(created, "name"))
	s.Require().NotZero(productID)

	// 2. FindById
	found := s.GQL(s.handler, `
		query Find($input: FindByIdProductInput!) { findProductById(input: $input) { status data { id name } } }`,
		map[string]interface{}{"input": map[string]interface{}{"id": productID}})
	s.Equal(productID, s.Int(s.Payload(found, "findProductById"), "id"))

	// 3. FindAll
	all := s.GQL(s.handler, `query { findAllProducts(input: {page: 1, pageSize: 10}) { status data { id } } }`, nil)
	s.NotEmpty(s.Items(all, "findAllProducts"))

	// 4. FindByMerchant
	byMerchant := s.GQL(s.handler, `
		query ByMerchant($input: FindAllProductMerchantInput!) {
			findProductsByMerchant(input: $input) { status data { id } }
		}`, map[string]interface{}{"input": map[string]interface{}{"merchantId": s.merchantID, "page": 1, "pageSize": 10}})
	s.NotEmpty(s.Items(byMerchant, "findProductsByMerchant"))

	// 5. FindByActive
	active := s.GQL(s.handler, `query { findActiveProducts(input: {page: 1, pageSize: 10}) { status } }`, nil)
	s.Equal("success", s.Status(active, "findActiveProducts"))

	// 6. Update
	update := s.GQLUpload(s.handler, `
		mutation Update($input: UpdateProductInput!) {
			updateProduct(input: $input) { status data { id name } }
		}`, map[string]interface{}{"input": map[string]interface{}{
		"productId":    productID,
		"merchantId":   s.merchantID,
		"categoryId":   s.categoryID,
		"name":         "Updated GQL Product",
		"description":  "Updated via GraphQL",
		"price":        20000,
		"countInStock": 40,
		"brand":        "GQL Brand",
		"weight":       1200,
		"rating":       4,
		"slugProduct":  "gql-product-updated",
	}}, tests.GraphQLFile{Path: "input.imageProduct", Filename: "product_updated.jpg", Content: img})
	s.Equal("Updated GQL Product", s.Str(s.Payload(update, "updateProduct"), "name"))

	// 7. Trash
	trashed := s.GQL(s.handler, `
		mutation Trash($input: FindByIdProductInput!) { trashedProduct(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"id": productID}})
	s.Equal("success", s.Status(trashed, "trashedProduct"))

	// 8. FindByTrashed
	trashedList := s.GQL(s.handler, `query { findTrashedProducts(input: {page: 1, pageSize: 10}) { status } }`, nil)
	s.Equal("success", s.Status(trashedList, "findTrashedProducts"))

	// 9. Restore
	restored := s.GQL(s.handler, `
		mutation Restore($input: FindByIdProductInput!) { restoreProduct(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"id": productID}})
	s.Equal("success", s.Status(restored, "restoreProduct"))

	// 10. DeletePermanent
	s.GQL(s.handler, `
		mutation Trash($input: FindByIdProductInput!) { trashedProduct(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"id": productID}})
	permanent := s.GQL(s.handler, `
		mutation DeletePermanent($input: FindByIdProductInput!) { deleteProductPermanent(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"id": productID}})
	s.Equal("success", s.Status(permanent, "deleteProductPermanent"))

	// 11. RestoreAll
	restoreAll := s.GQL(s.handler, `mutation { restoreAllProducts { status } }`, nil)
	s.Equal("success", s.Status(restoreAll, "restoreAllProducts"))

	// 12. DeleteAll
	deleteAll := s.GQL(s.handler, `mutation { deleteAllProductsPermanent { status } }`, nil)
	s.Equal("success", s.Status(deleteAll, "deleteAllProductsPermanent"))
}

func TestProductGraphqlSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(ProductGraphqlTestSuite))
}
