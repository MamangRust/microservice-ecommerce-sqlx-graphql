package banner_test

import (
	"net/http"
	"testing"

	tests "github.com/MamangRust/microservice-ecommerce-test"
	"github.com/stretchr/testify/suite"
)

type BannerGraphqlTestSuite struct {
	tests.BaseTestSuite
	handler http.Handler
}

func (s *BannerGraphqlTestSuite) SetupSuite() {
	s.BaseTestSuite.SetupSuite()
	s.SetupBannerService()
	s.handler = s.NewGraphQLHandler()
}

func (s *BannerGraphqlTestSuite) TestBannerGraphqlLifecycle() {
	// 1. Create
	create := s.GQL(s.handler, `
		mutation Create($input: CreateBannerInput!) {
			createBanner(input: $input) { status message data { banner_id name is_active } }
		}`, map[string]interface{}{"input": map[string]interface{}{
		"name":       "GQL Banner",
		"start_date": "2024-01-01",
		"end_date":   "2024-12-31",
		"start_time": "00:00:00",
		"end_time":   "23:59:59",
		"is_active":  true,
	}})
	created := s.Payload(create, "createBanner")
	bannerID := s.Int(created, "banner_id")
	s.Equal("GQL Banner", s.Str(created, "name"))
	s.Require().NotZero(bannerID)

	// 2. FindById
	found := s.GQL(s.handler, `
		query Find($input: FindByIdBannerInput!) {
			findBannerById(input: $input) { status data { banner_id name } }
		}`, map[string]interface{}{"input": map[string]interface{}{"id": bannerID}})
	s.Equal(bannerID, s.Int(s.Payload(found, "findBannerById"), "banner_id"))

	// 3. FindAll
	all := s.GQL(s.handler, `query { findAllBanners(input: {page: 1, page_size: 10}) { status data { banner_id } } }`, nil)
	s.NotEmpty(s.Items(all, "findAllBanners"))

	// 4. FindByActive
	active := s.GQL(s.handler, `query { findActiveBanners(input: {page: 1, page_size: 10}) { status data { banner_id } } }`, nil)
	s.Equal("success", s.Status(active, "findActiveBanners"))

	// 5. Update
	update := s.GQL(s.handler, `
		mutation Update($input: UpdateBannerInput!) {
			updateBanner(input: $input) { status data { banner_id name is_active } }
		}`, map[string]interface{}{"input": map[string]interface{}{
		"banner_id":  bannerID,
		"name":       "Updated GQL Banner",
		"start_date": "2024-01-01",
		"end_date":   "2024-12-31",
		"start_time": "00:00:00",
		"end_time":   "23:59:59",
		"is_active":  false,
	}})
	s.Equal("Updated GQL Banner", s.Str(s.Payload(update, "updateBanner"), "name"))

	// 6. Trash
	trashed := s.GQL(s.handler, `
		mutation Trash($input: FindByIdBannerInput!) { trashBanner(input: $input) { status data { banner_id } } }`,
		map[string]interface{}{"input": map[string]interface{}{"id": bannerID}})
	s.Equal("success", s.Status(trashed, "trashBanner"))

	// 7. FindByTrashed
	trashedList := s.GQL(s.handler, `query { findTrashedBanners(input: {page: 1, page_size: 10}) { status data { banner_id } } }`, nil)
	s.Equal("success", s.Status(trashedList, "findTrashedBanners"))

	// 8. Restore
	restored := s.GQL(s.handler, `
		mutation Restore($input: FindByIdBannerInput!) { restoreBanner(input: $input) { status data { banner_id } } }`,
		map[string]interface{}{"input": map[string]interface{}{"id": bannerID}})
	s.Equal("success", s.Status(restored, "restoreBanner"))

	// 9. DeletePermanent
	s.GQL(s.handler, `
		mutation Trash($input: FindByIdBannerInput!) { trashBanner(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"id": bannerID}})
	permanent := s.GQL(s.handler, `
		mutation DeletePermanent($input: FindByIdBannerInput!) { deleteBannerPermanent(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"id": bannerID}})
	s.Equal("success", s.Status(permanent, "deleteBannerPermanent"))

	// 10. RestoreAll
	restoreAll := s.GQL(s.handler, `mutation { restoreAllBanners { status message } }`, nil)
	s.Equal("success", s.Status(restoreAll, "restoreAllBanners"))

	// 11. DeleteAll
	deleteAll := s.GQL(s.handler, `mutation { deleteAllBannersPermanent { status message } }`, nil)
	s.Equal("success", s.Status(deleteAll, "deleteAllBannersPermanent"))
}

func TestBannerGraphqlSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(BannerGraphqlTestSuite))
}
