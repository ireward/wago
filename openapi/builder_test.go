package openapi

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"path"
	"reflect"
	"testing"

	"github.com/ireward/wago/logger"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3gen"
	"github.com/stretchr/testify/assert"
)

func TestBuildFromModel(t *testing.T) {
	ctx := context.Background()
	b := NewBuilder(ctx, "3.0.3", &BuilderConfig{
		API:     &TestAPI{},
		Info:    &info,
		Servers: &servers,
	})
	err := b.Build()
	assert.NoError(t, err)
	assert.NotNil(t, b.Spec)
	assert.NotNil(t, b.Spec.Model)
	assert.NotNil(t, b.Spec.Model.Components)

	bytes, err := b.Spec.ToBytes()
	assert.NoError(t, err)
	assert.NotEmpty(t, string(bytes))

	path := path.Join(t.TempDir(), "test-build-api.json")
	err = os.WriteFile(path, bytes, 0644)
	assert.NoError(t, err)

	loader := openapi3.NewLoader()
	m, err := loader.LoadFromFile(path)
	assert.NoError(t, err)
	assert.NotNil(t, m)

	err = m.Validate(ctx)
	assert.NoError(t, err)
}

func TestSchemaRefs(t *testing.T) {
	ctx := context.Background()
	b := builder{
		ctx:    ctx,
		logger: logger.FromCtx(ctx),
		Spec: &Spec{
			Model: openapi3.T{
				Components: openapi3.Components{
					Schemas: openapi3.Schemas{},
				},
			},
		},
		enumCache:  make(map[string]bool),
		modelCache: make(map[string]reflect.Type),
	}

	b.generator = openapi3gen.NewGenerator(openapi3gen.UseAllExportedFields(), openapi3gen.SchemaCustomizer(b.customizer))
	t.Run("test-enum-ref", func(t *testing.T) {
		ref, err := b.buildSchema(reflect.TypeOf(TestResponse{}))
		assert.NoError(t, err)
		assert.NotNil(t, ref)

		b.resolveRefPaths()
		assert.Equal(t, fmt.Sprintf("%s/%s", schemasPath, "TestEnum"), ref.Value.Properties["test_enum"].Ref)
	})

}

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
)

var _ Enum = (*TestEnum)(nil)

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
	ReqProp1 string `json:"req_prop1"`
	ReqProp2 string `json:"req_prop2"`
}

type EmbeddedModel struct {
	EmbeddedProp1 int    `json:"embedded_prop1"`
	EmbeddedProp2 string `json:"embedded_prop2"`
}

type TestResponse struct {
	EmbeddedModel
	RespProp1 string   `json:"resp_prop1"`
	RespProp2 float64  `json:"resp_prop2"`
	TestEnum  TestEnum `json:"test_enum"`
}

type TestAPI struct{}

func (t *TestAPI) GetPaths() []*Path {
	tag1 := []string{"tag1"}
	tag2 := []string{"tag2"}
	tags := [][]string{tag1, tag2}

	return []*Path{
		{
			Template: "/api/v1/test",
			Operations: []*Operation{
				NewOperation(http.MethodPost, nil, tags[rand.Intn(len(tags))], "TestPost",
					NewRequestBody(SchemeType_Object, TestRequest{}),
					[]*Parameter{
						{
							ID:          "Param1",
							Name:        "param1",
							SchemeType:  SchemeType_String,
							In:          ParameterLocation_InHeader,
							Description: "param1 description",
						},
					},
					t.withJWTAuth(),
					&OperationMeta{
						Summary:     "Performs a test post",
						Description: "This is a test post operation.\n",
					},
					NewResponse(http.StatusOK, SchemeType_Object, TestResponse{}, nil,
						&ResponseMeta{
							Name:        "TestPostResponse",
							Description: "Returns the object after POSTing it to the server.",
						}),
				),
			},
		},
	}
}

func (t *TestAPI) withJWTAuth() *SecurityParam {
	return &SecurityParam{
		ID:             "BearerAuth",
		SecurityType:   SecurityType(SecurityType_Http),
		SecurityScheme: SecurityScheme_Bearer,
		BearerFormat:   "JWT",
	}
}
