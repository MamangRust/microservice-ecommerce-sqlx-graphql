package slider_test

import (
	"net/http"
	"testing"

	tests "github.com/MamangRust/microservice-ecommerce-test"
	"github.com/stretchr/testify/suite"
)

type SliderGraphqlTestSuite struct {
	tests.BaseTestSuite
	handler http.Handler
}

func (s *SliderGraphqlTestSuite) SetupSuite() {
	s.BaseTestSuite.SetupSuite()
	s.SetupSliderService()
	s.handler = s.NewGraphQLHandler()
}

func (s *SliderGraphqlTestSuite) TestSliderGraphqlLifecycle() {
	img := []byte("dummy image content")

	// 1. Create
	create := s.GQLUpload(s.handler, `
		mutation Create($input: CreateSliderRequest!) {
			createSlider(input: $input) { status message data { id name } }
		}`, map[string]interface{}{"input": map[string]interface{}{"name": "GQL Slider"}},
		tests.GraphQLFile{Path: "input.image", Filename: "slider.jpg", Content: img})
	created := s.Payload(create, "createSlider")
	sliderID := s.Int(created, "id")
	s.Equal("GQL Slider", s.Str(created, "name"))
	s.Require().NotZero(sliderID)

	// 2. FindAll
	all := s.GQL(s.handler, `query { findAllSliders(input: {page: 1, page_size: 10}) { status data { id } } }`, nil)
	s.NotEmpty(s.Items(all, "findAllSliders"))

	// 3. FindByActive
	active := s.GQL(s.handler, `query { findActiveSliders(input: {page: 1, page_size: 10}) { status } }`, nil)
	s.Equal("success", s.Status(active, "findActiveSliders"))

	// 4. Update
	update := s.GQLUpload(s.handler, `
		mutation Update($input: UpdateSliderRequest!) {
			updateSlider(input: $input) { status data { id name } }
		}`, map[string]interface{}{"input": map[string]interface{}{"id": sliderID, "name": "Updated GQL Slider"}},
		tests.GraphQLFile{Path: "input.image", Filename: "slider_updated.jpg", Content: img})
	s.Equal("Updated GQL Slider", s.Str(s.Payload(update, "updateSlider"), "name"))

	// 5. Trash
	trashed := s.GQL(s.handler, `
		mutation Trash($input: FindByIdSliderRequest!) { trashedSlider(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"id": sliderID}})
	s.Equal("success", s.Status(trashed, "trashedSlider"))

	// 6. FindByTrashed
	trashedList := s.GQL(s.handler, `query { findTrashedSliders(input: {page: 1, page_size: 10}) { status } }`, nil)
	s.Equal("success", s.Status(trashedList, "findTrashedSliders"))

	// 7. Restore
	restored := s.GQL(s.handler, `
		mutation Restore($input: FindByIdSliderRequest!) { restoreSlider(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"id": sliderID}})
	s.Equal("success", s.Status(restored, "restoreSlider"))

	// 8. DeletePermanent
	s.GQL(s.handler, `
		mutation Trash($input: FindByIdSliderRequest!) { trashedSlider(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"id": sliderID}})
	permanent := s.GQL(s.handler, `
		mutation DeletePermanent($input: FindByIdSliderRequest!) { deleteSliderPermanent(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"id": sliderID}})
	s.Equal("success", s.Status(permanent, "deleteSliderPermanent"))

	// 9. RestoreAll
	restoreAll := s.GQL(s.handler, `mutation { restoreAllSliders { status } }`, nil)
	s.Equal("success", s.Status(restoreAll, "restoreAllSliders"))

	// 10. DeleteAll
	deleteAll := s.GQL(s.handler, `mutation { deleteAllSlidersPermanent { status } }`, nil)
	s.Equal("success", s.Status(deleteAll, "deleteAllSlidersPermanent"))
}

func TestSliderGraphqlSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(SliderGraphqlTestSuite))
}
