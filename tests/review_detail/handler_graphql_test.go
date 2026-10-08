package review_detail_test

import (
	"context"
	"net/http"
	"testing"

	tests "github.com/MamangRust/microservice-ecommerce-test"
	"github.com/stretchr/testify/suite"
)

type ReviewDetailGraphqlTestSuite struct {
	tests.BaseTestSuite
	handler  http.Handler
	reviewID int
}

func (s *ReviewDetailGraphqlTestSuite) SetupSuite() {
	s.BaseTestSuite.SetupSuite()
	s.SetupUserService()
	s.SetupCategoryService()
	s.SetupMerchantService()
	s.SetupProductService()
	s.SetupReviewService()
	s.SetupReviewDetailService()
	s.handler = s.NewGraphQLHandler()

	ctx := context.Background()
	userID := s.SeedUser(ctx)
	catID := s.SeedCategory(ctx)
	merchID := s.SeedMerchant(ctx, userID)
	prodID := s.SeedProduct(ctx, merchID, catID)
	s.reviewID = s.SeedReview(ctx, userID, prodID)
}

func (s *ReviewDetailGraphqlTestSuite) TestReviewDetailGraphqlLifecycle() {
	// 1. Create
	create := s.GQL(s.handler, `
		mutation Create($input: CreateReviewDetailInput!) {
			createReviewDetail(input: $input) { status message data { id reviewId caption } }
		}`, map[string]interface{}{"input": map[string]interface{}{
		"reviewId": s.reviewID,
		"type":     "photo",
		"url":      "https://example.com/review.jpg",
		"caption":  "GQL Review Detail",
	}})
	created := s.Payload(create, "createReviewDetail")
	detailID := s.Int(created, "id")
	s.Require().NotZero(detailID)

	// 2. FindById
	found := s.GQL(s.handler, `
		query Find($input: FindByIdReviewDetailInput!) { findReviewDetailById(input: $input) { status data { id } } }`,
		map[string]interface{}{"input": map[string]interface{}{"id": detailID}})
	s.Equal(detailID, s.Int(s.Payload(found, "findReviewDetailById"), "id"))

	// 3. FindAll
	all := s.GQL(s.handler, `query { findAllReviewDetails(input: {page: 1, pageSize: 10}) { status data { id } } }`, nil)
	s.Equal("success", s.Status(all, "findAllReviewDetails"))

	// 4. FindByActive
	active := s.GQL(s.handler, `query { findActiveReviewDetails(input: {page: 1, pageSize: 10}) { status } }`, nil)
	s.Equal("success", s.Status(active, "findActiveReviewDetails"))

	// 5. Update
	update := s.GQL(s.handler, `
		mutation Update($input: UpdateReviewDetailInput!) {
			updateReviewDetail(input: $input) { status data { id caption } }
		}`, map[string]interface{}{"input": map[string]interface{}{
		"reviewDetailId": detailID,
		"type":           "video",
		"url":            "https://example.com/review-updated.jpg",
		"caption":        "Updated GQL Review Detail",
	}})
	s.Equal("Updated GQL Review Detail", s.Str(s.Payload(update, "updateReviewDetail"), "caption"))

	// 6. Trash
	trashed := s.GQL(s.handler, `
		mutation Trash($input: FindByIdReviewDetailInput!) { trashedReviewDetail(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"id": detailID}})
	s.Equal("success", s.Status(trashed, "trashedReviewDetail"))

	// 7. FindByTrashed
	trashedList := s.GQL(s.handler, `query { findTrashedReviewDetails(input: {page: 1, pageSize: 10}) { status } }`, nil)
	s.Equal("success", s.Status(trashedList, "findTrashedReviewDetails"))

	// 8. Restore
	restored := s.GQL(s.handler, `
		mutation Restore($input: FindByIdReviewDetailInput!) { restoreReviewDetail(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"id": detailID}})
	s.Equal("success", s.Status(restored, "restoreReviewDetail"))

	// 9. DeletePermanent
	s.GQL(s.handler, `
		mutation Trash($input: FindByIdReviewDetailInput!) { trashedReviewDetail(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"id": detailID}})
	permanent := s.GQL(s.handler, `
		mutation DeletePermanent($input: FindByIdReviewDetailInput!) { deleteReviewDetailPermanent(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"id": detailID}})
	s.Equal("success", s.Status(permanent, "deleteReviewDetailPermanent"))

	// 10. RestoreAll
	restoreAll := s.GQL(s.handler, `mutation { restoreAllReviewDetails { status } }`, nil)
	s.Equal("success", s.Status(restoreAll, "restoreAllReviewDetails"))

	// 11. DeleteAll
	deleteAll := s.GQL(s.handler, `mutation { deleteAllReviewDetailsPermanent { status } }`, nil)
	s.Equal("success", s.Status(deleteAll, "deleteAllReviewDetailsPermanent"))
}

func TestReviewDetailGraphqlSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(ReviewDetailGraphqlTestSuite))
}
