package tests

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strconv"

	graphtest "github.com/MamangRust/microservice-ecommerce-grpc-apigateway/graphtest"
)

// GraphQLFile describes a file part for a GraphQL multipart (Upload) request.
//
// Path is the dot-notation location of the upload inside the variables object,
// without the leading "variables." segment, e.g. "input.image_category".
type GraphQLFile struct {
	Path     string
	Filename string
	Content  []byte
}

// NewGraphQLHandler builds a GraphQL HTTP handler wired to the suite's gRPC
// connections, cache store and logger. It mirrors the apigateway runtime wiring
// (see service/apigateway/graphtest) so the suites exercise the real resolvers.
func (s *BaseTestSuite) NewGraphQLHandler() http.Handler {
	return graphtest.NewTestHandler(s.Conns, s.GetCacheStore(), s.Log)
}

// GQL executes a JSON GraphQL request and fails the test on transport errors or
// GraphQL errors.
func (s *BaseTestSuite) GQL(h http.Handler, query string, vars map[string]interface{}) *graphtest.GraphQLResponse {
	s.T().Helper()
	resp, err := graphtest.ExecuteGraphQL(h, query, vars, "")
	s.Require().NoError(err)
	s.Require().Emptyf(resp.Errors, "unexpected graphql errors: %+v", resp.Errors)
	return resp
}

// GQLWithUser executes a JSON GraphQL request against a handler that injects
// userID into the request context, emulating what AuthMiddleware does for an
// authenticated caller. Use it for resolvers that read the caller identity from
// the context (e.g. deleteAllCarts) instead of from the input.
func (s *BaseTestSuite) GQLWithUser(h http.Handler, userID int, query string, vars map[string]interface{}) *graphtest.GraphQLResponse {
	s.T().Helper()
	return s.GQL(graphtest.WithUser(h, userID), query, vars)
}

// GQLUpload executes a multipart GraphQL request carrying Upload scalars.
func (s *BaseTestSuite) GQLUpload(h http.Handler, query string, vars map[string]interface{}, files ...GraphQLFile) *graphtest.GraphQLResponse {
	s.T().Helper()
	body, contentType, err := buildMultipartGraphQLRequest(query, vars, files)
	s.Require().NoError(err)

	req := httptest.NewRequest(http.MethodPost, "/query", bytes.NewReader(body))
	req.Header.Set("Content-Type", contentType)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	var resp graphtest.GraphQLResponse
	s.Require().NoErrorf(json.Unmarshal(rec.Body.Bytes(), &resp), "response body: %s", rec.Body.String())
	s.Require().Emptyf(resp.Errors, "unexpected graphql errors: %+v (body: %s)", resp.Errors, rec.Body.String())
	return &resp
}

// buildMultipartGraphQLRequest encodes a request per the GraphQL multipart
// request spec (operations, map, then one file part per upload).
func buildMultipartGraphQLRequest(query string, vars map[string]interface{}, files []GraphQLFile) ([]byte, string, error) {
	if vars == nil {
		vars = map[string]interface{}{}
	}
	operations := map[string]interface{}{"query": query, "variables": vars}
	opsJSON, err := json.Marshal(operations)
	if err != nil {
		return nil, "", err
	}

	mapping := make(map[string][]string, len(files))
	for i, f := range files {
		mapping[strconv.Itoa(i)] = []string{"variables." + f.Path}
	}
	mapJSON, err := json.Marshal(mapping)
	if err != nil {
		return nil, "", err
	}

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	if err := w.WriteField("operations", string(opsJSON)); err != nil {
		return nil, "", err
	}
	if err := w.WriteField("map", string(mapJSON)); err != nil {
		return nil, "", err
	}
	for i, f := range files {
		part, err := w.CreateFormFile(strconv.Itoa(i), f.Filename)
		if err != nil {
			return nil, "", err
		}
		if _, err := part.Write(f.Content); err != nil {
			return nil, "", err
		}
	}
	if err := w.Close(); err != nil {
		return nil, "", err
	}
	return buf.Bytes(), w.FormDataContentType(), nil
}

// Status returns resp.Data[field].status.
func (s *BaseTestSuite) Status(resp *graphtest.GraphQLResponse, field string) string {
	s.T().Helper()
	return s.Str(s.Obj(resp.Data, field), "status")
}

// Obj extracts data[field] as a JSON object.
func (s *BaseTestSuite) Obj(data map[string]interface{}, field string) map[string]interface{} {
	s.T().Helper()
	v, ok := data[field].(map[string]interface{})
	s.Require().Truef(ok, "expected %q to be an object, got %T", field, data[field])
	return v
}

// Payload returns resp.Data[field].data as an object (the ApiResponse shape).
func (s *BaseTestSuite) Payload(resp *graphtest.GraphQLResponse, field string) map[string]interface{} {
	s.T().Helper()
	return s.Obj(s.Obj(resp.Data, field), "data")
}

// Items returns resp.Data[field].data as a list.
func (s *BaseTestSuite) Items(resp *graphtest.GraphQLResponse, field string) []interface{} {
	s.T().Helper()
	obj := s.Obj(resp.Data, field)
	list, ok := obj["data"].([]interface{})
	s.Require().Truef(ok, "expected %s.data to be a list, got %T", field, obj["data"])
	return list
}

// Int reads a numeric field from a JSON object.
func (s *BaseTestSuite) Int(m map[string]interface{}, key string) int {
	s.T().Helper()
	v, ok := m[key].(float64)
	s.Require().Truef(ok, "expected %q to be a number, got %T", key, m[key])
	return int(v)
}

// Str reads a string field from a JSON object.
func (s *BaseTestSuite) Str(m map[string]interface{}, key string) string {
	s.T().Helper()
	v, ok := m[key].(string)
	s.Require().Truef(ok, "expected %q to be a string, got %T", key, m[key])
	return v
}

// Bool reads a boolean field from a JSON object.
func (s *BaseTestSuite) Bool(m map[string]interface{}, key string) bool {
	s.T().Helper()
	v, ok := m[key].(bool)
	s.Require().Truef(ok, "expected %q to be a bool, got %T", key, m[key])
	return v
}
