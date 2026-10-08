package review_test

import (
	"context"
	"net/http"
	"testing"

	tests "github.com/MamangRust/microservice-ecommerce-test"
	"github.com/stretchr/testify/suite"
)

type ReviewGraphqlTestSuite struct {
	tests.BaseTestSuite
	handler    http.Handler
	userID     int
	productID  int
	merchantID int
}

func (s *ReviewGraphqlTestSuite) SetupSuite() {
	s.BaseTestSuite.SetupSuite()
	s.SetupUserService()
	s.SetupCategoryService()
	s.SetupMerchantService()
	s.SetupProductService()
	s.SetupReviewService()
	s.handler = s.NewGraphQLHandler()

	ctx := context.Background()
	s.userID = s.SeedUser(ctx)
	catID := s.SeedCategory(ctx)
	s.merchantID = s.SeedMerchant(ctx, s.userID)
	s.productID = s.SeedProduct(ctx, s.merchantID, catID)
}

func (s *ReviewGraphqlTestSuite) TestReviewGraphqlLifecycle() {
	// 1. Create
	create := s.GQL(s.handler, `
		mutation Create($input: CreateReviewRequest!) {
			createReview(input: $input) { status message data { id comment rating } }
		}`, map[string]interface{}{"input": map[string]interface{}{
		"user_id":    s.userID,
		"product_id": s.productID,
		"name":       "GQL Reviewer",
		"comment":    "Great product",
		"rating":     5,
	}})
	created := s.Payload(create, "createReview")
	reviewID := s.Int(created, "id")
	s.Equal("Great product", s.Str(created, "comment"))
	s.Require().NotZero(reviewID)

	// 2. FindAll
	all := s.GQL(s.handler, `query { findAllReviews(input: {page: 1, page_size: 10}) { status data { id } } }`, nil)
	s.NotEmpty(s.Items(all, "findAllReviews"))

	// 3. FindByProduct
	byProduct := s.GQL(s.handler, `
		query ByProduct($input: FindAllReviewProductRequest!) { findReviewsByProduct(input: $input) { status data { id } } }`,
		map[string]interface{}{"input": map[string]interface{}{"product_id": s.productID, "page": 1, "page_size": 10}})
	s.Equal("success", s.Status(byProduct, "findReviewsByProduct"))

	// 4. FindByMerchant
	byMerchant := s.GQL(s.handler, `
		query ByMerchant($input: FindAllReviewMerchantRequest!) { findReviewsByMerchant(input: $input) { status data { id } } }`,
		map[string]interface{}{"input": map[string]interface{}{"merchant_id": s.merchantID, "page": 1, "page_size": 10}})
	s.Equal("success", s.Status(byMerchant, "findReviewsByMerchant"))

	// 5. FindByActive
	active := s.GQL(s.handler, `query { findActiveReviews(input: {page: 1, page_size: 10}) { status } }`, nil)
	s.Equal("success", s.Status(active, "findActiveReviews"))

	// 6. Update
	update := s.GQL(s.handler, `
		mutation Update($input: UpdateReviewRequest!) {
			updateReview(input: $input) { status data { id comment rating } }
		}`, map[string]interface{}{"input": map[string]interface{}{
		"review_id": reviewID,
		"name":      "GQL Reviewer",
		"comment":   "Updated comment",
		"rating":    4,
	}})
	s.Equal("Updated comment", s.Str(s.Payload(update, "updateReview"), "comment"))

	// 7. Trash
	trashed := s.GQL(s.handler, `
		mutation Trash($input: FindByIdReviewRequest!) { trashedReview(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"id": reviewID}})
	s.Equal("success", s.Status(trashed, "trashedReview"))

	// 8. FindByTrashed
	trashedList := s.GQL(s.handler, `query { findTrashedReviews(input: {page: 1, page_size: 10}) { status } }`, nil)
	s.Equal("success", s.Status(trashedList, "findTrashedReviews"))

	// 9. Restore
	restored := s.GQL(s.handler, `
		mutation Restore($input: FindByIdReviewRequest!) { restoreReview(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"id": reviewID}})
	s.Equal("success", s.Status(restored, "restoreReview"))

	// 10. DeletePermanent
	s.GQL(s.handler, `
		mutation Trash($input: FindByIdReviewRequest!) { trashedReview(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"id": reviewID}})
	permanent := s.GQL(s.handler, `
		mutation DeletePermanent($input: FindByIdReviewRequest!) { deleteReviewPermanent(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"id": reviewID}})
	s.Equal("success", s.Status(permanent, "deleteReviewPermanent"))

	// 11. RestoreAll
	restoreAll := s.GQL(s.handler, `mutation { restoreAllReviews { status } }`, nil)
	s.Equal("success", s.Status(restoreAll, "restoreAllReviews"))

	// 12. DeleteAll
	deleteAll := s.GQL(s.handler, `mutation { deleteAllReviewsPermanent { status } }`, nil)
	s.Equal("success", s.Status(deleteAll, "deleteAllReviewsPermanent"))
}

func TestReviewGraphqlSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(ReviewGraphqlTestSuite))
}
