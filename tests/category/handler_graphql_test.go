package category_test

import (
	"net/http"
	"testing"

	tests "github.com/MamangRust/microservice-ecommerce-test"
	"github.com/stretchr/testify/suite"
)

type CategoryGraphqlTestSuite struct {
	tests.BaseTestSuite
	handler http.Handler
}

func (s *CategoryGraphqlTestSuite) SetupSuite() {
	s.BaseTestSuite.SetupSuite()
	s.SetupCategoryService()
	s.handler = s.NewGraphQLHandler()
}

func (s *CategoryGraphqlTestSuite) TestCategoryGraphqlLifecycle() {
	img := []byte("dummy image content")

	// 1. Create
	create := s.GQLUpload(s.handler, `
		mutation Create($input: CreateCategoryInput!) {
			createCategory(input: $input) { status message data { id name } }
		}`, map[string]interface{}{"input": map[string]interface{}{
		"name":          "GQL Category",
		"description":   "Testing via GraphQL",
		"slug_category": "gql-category",
	}}, tests.GraphQLFile{Path: "input.image_category", Filename: "category.jpg", Content: img})
	created := s.Payload(create, "createCategory")
	categoryID := s.Int(created, "id")
	s.Equal("GQL Category", s.Str(created, "name"))
	s.Require().NotZero(categoryID)

	// 2. FindById
	found := s.GQL(s.handler, `
		query Find($input: FindByIdCategoryInput!) { findCategoryById(input: $input) { status data { id name } } }`,
		map[string]interface{}{"input": map[string]interface{}{"id": categoryID}})
	s.Equal(categoryID, s.Int(s.Payload(found, "findCategoryById"), "id"))

	// 3. FindAll
	all := s.GQL(s.handler, `query { findAllCategories(input: {page: 1, page_size: 10}) { status data { id } } }`, nil)
	s.NotEmpty(s.Items(all, "findAllCategories"))

	// 4. FindByActive
	active := s.GQL(s.handler, `query { findActiveCategories(input: {page: 1, page_size: 10}) { status } }`, nil)
	s.Equal("success", s.Status(active, "findActiveCategories"))

	// 5. Update
	update := s.GQLUpload(s.handler, `
		mutation Update($input: UpdateCategoryInput!) {
			updateCategory(input: $input) { status data { id name } }
		}`, map[string]interface{}{"input": map[string]interface{}{
		"category_id":   categoryID,
		"name":          "Updated GQL Category",
		"description":   "Updated via GraphQL",
		"slug_category": "gql-category-updated",
	}}, tests.GraphQLFile{Path: "input.image_category", Filename: "category_updated.jpg", Content: img})
	s.Equal("Updated GQL Category", s.Str(s.Payload(update, "updateCategory"), "name"))

	// 6. Trash
	trashed := s.GQL(s.handler, `
		mutation Trash($input: FindByIdCategoryInput!) { trashCategory(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"id": categoryID}})
	s.Equal("success", s.Status(trashed, "trashCategory"))

	// 7. FindByTrashed
	trashedList := s.GQL(s.handler, `query { findTrashedCategories(input: {page: 1, page_size: 10}) { status } }`, nil)
	s.Equal("success", s.Status(trashedList, "findTrashedCategories"))

	// 8. Restore
	restored := s.GQL(s.handler, `
		mutation Restore($input: FindByIdCategoryInput!) { restoreCategory(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"id": categoryID}})
	s.Equal("success", s.Status(restored, "restoreCategory"))

	// 9. DeletePermanent
	s.GQL(s.handler, `
		mutation Trash($input: FindByIdCategoryInput!) { trashCategory(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"id": categoryID}})
	permanent := s.GQL(s.handler, `
		mutation DeletePermanent($input: FindByIdCategoryInput!) { deleteCategoryPermanent(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"id": categoryID}})
	s.Equal("success", s.Status(permanent, "deleteCategoryPermanent"))

	// 10. RestoreAll
	restoreAll := s.GQL(s.handler, `mutation { restoreAllCategories { status } }`, nil)
	s.Equal("success", s.Status(restoreAll, "restoreAllCategories"))

	// 11. DeleteAll
	deleteAll := s.GQL(s.handler, `mutation { deleteAllCategoriesPermanent { status } }`, nil)
	s.Equal("success", s.Status(deleteAll, "deleteAllCategoriesPermanent"))
}

func TestCategoryGraphqlSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(CategoryGraphqlTestSuite))
}
