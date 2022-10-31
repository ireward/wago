package main

import (
	"context"
	"log"
	"math/rand"
	"net/http"
	"os"

	"github.com/ireward/wago/openapi"
	models "github.com/ireward/wago/openapi/example/pkg"

	"github.com/getkin/kin-openapi/openapi3"
)

// provide some defaults which are used across all the API
var (
	info = openapi3.Info{
		Version:     "1.0.0",
		Title:       "Test API",
		Description: "This is a test API.\n",
		Contact: &openapi3.Contact{
			Email: "john@doe.com",
		},
	}
	servers = openapi3.Servers{
		{
			Description: "Test server",
			URL:         "{schema}://{address}",
			Variables: map[string]*openapi3.ServerVariable{
				"address": {Default: "acme.com"},
				"schema":  {Default: "https", Enum: []string{"http", "https"}},
			},
		},
	}
	xAppIDParam = openapi.NewInHeaderParam("AppIDParam", "X-App-ID", "The application ID used to assign a request to.")
)

// Region API-Models  ///////////////////////////////////////////////////////////

// models are usually defined in another file or are imported into the generator code
var _ openapi.Enum = (*TestEnum)(nil)

type TestEnum int

const (
	TestEnum_ValueOne TestEnum = iota
	TestEnum_ValueTwo
	TestEnum_ValueThree
)

func (e TestEnum) OpenApiValues() []interface{} {
	return []interface{}{
		"ValueOne",
		"ValueTwo",
		"ValueThree",
	}
}

type TestRequest struct {
	ReqProp1 string `json:"req_prop1" descr:"ReqProp1 must be set since the backend useses it to infer some important business logic."`
	ReqProp2 string `json:"req_prop2,omitempty"`
}

type EmbeddedModel struct {
	models.DeepModel
	EmbeddedProp1 int        `json:"embedded_prop1" descr:"this is a description of embedded property 1. This is a very important property."`
	EmbeddedProp2 string     `json:"embedded_prop2"`
	TestEnumSlice []TestEnum `json:"test_enum_slice"`
}

type TestResponse struct {
	RespProp1 string        `json:"resp_prop1"`
	RespProp2 float64       `json:"resp_prop2"`
	RespProp3 EmbeddedModel `json:"resp_prop3"`
	TestEnum  TestEnum      `json:"test_enum"`
}

type TestAPI struct {
}

// End Region  ///////////////////////////////////////////////////////

func (t *TestAPI) GetPaths() []*openapi.Path {
	tag1 := []string{"tag1"}
	tag2 := []string{"tag2"}
	tags := [][]string{tag1, tag2}

	return []*openapi.Path{
		{
			Template: "/api/v1/test",
			Operations: []*openapi.Operation{
				openapi.NewOperation(
					// method of the operation
					http.MethodPost,
					// handler of the operation (currently just a dummy to indicate to the implementer, which handler is used)
					nil,
					// tags where the operation is grouped in
					tags[rand.Intn(len(tags))],
					// operationID
					"TestPost",
					// requestBody
					openapi.NewObjectRequestBody(TestRequest{}),
					// required parameters: either in path, cookie or header
					[]*openapi.Parameter{xAppIDParam},
					// security requirements: eiter basic_auth, api_key or bearer
					&openapi.WithBearerAuth,
					// meta
					&openapi.OperationMeta{
						Summary:     "TestPost",
						Description: "This is a test operation.",
					},
					// response returned by the API
					openapi.NewObjectResponse(
						// status code
						http.StatusOK,
						// response body
						nil,
						// exposed headers (optional)
						nil,
						// description of the response
						&openapi.ResponseMeta{
							Name:        "TestPostResponse",
							Description: "Returns the object after POSTing it to the server.",
						}),
				),
			},
		},
		{
			Template: "/api/v1/test/{id}",
			Operations: []*openapi.Operation{
				openapi.NewOperation(
					// method of the operation
					http.MethodGet,
					// handler of the operation (currently just a dummy to indicate to the implementer, which handler is used)
					nil,
					// tags where the operation is grouped in
					tags[rand.Intn(len(tags))],
					// operationID
					"TestGetById",
					// requestBody
					nil,
					// required parameters: either in path, cookie or header
					[]*openapi.Parameter{xAppIDParam, openapi.NewInPathParam("TestID", "id", "ID of the object to get.")},
					// security requirements: eiter basic_auth, api_key or bearer
					&openapi.WithBearerAuth,
					// meta
					&openapi.OperationMeta{
						Summary:     "TestGetById",
						Description: "This is a test operation.",
					},
					// response returned by the API
					openapi.NewObjectResponse(
						http.StatusOK,
						TestResponse{},
						nil,
						&openapi.ResponseMeta{
							Name:        "TestGetByIdResponse",
							Description: "Returns the object with the given ID.",
						}),
				),
			},
		},
	}
}

func main() {
	ctx := context.Background()
	b := openapi.NewBuilder(ctx, "3.0.3", &openapi.BuilderConfig{
		API:     &TestAPI{},
		Info:    &info,
		Servers: &servers,
	})
	err := b.Build()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	bytes, err := b.Spec.ToBytes()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	if err := os.WriteFile("./generated_example.json", bytes, 0644); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	os.Exit(0)
}
