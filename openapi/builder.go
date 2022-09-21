package openapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	log "github.com/ireward/wago/logger"
	"github.com/ireward/wago/logger/tag"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3gen"
)

const (
	schemasPath         = "#/components/schemas"
	responsesPath       = "#/components/responses"
	parametersPath      = "#/components/parameters"
	requestBodiesPath   = "#/components/requestBodies"
	headersPath         = "#/components/headers"
	securitySchemesPath = "#/components/securitySchemes"
)

// Spec is a wrapper around the openapi3.T model.
// Spec allows the generation of a byte-slice or a
// JSON string from the model.
type Spec struct {
	Model openapi3.T
}

// ToBytes returns the spec as a byte-slice.
func (s *Spec) ToBytes() ([]byte, error) {
	var d []byte
	if m, err := json.MarshalIndent(s.Model, "", " "); err != nil {
		return nil, err
	} else {
		d = append(d, m...)
	}
	return d, nil
}

// ToJSON returns the spec as a YAML string.
func (s *Spec) ToJSON() (string, error) {
	d, err := s.ToBytes()
	if err != nil {
		return "", err
	}
	return string(d), nil
}

// builder is an object which can be used to build a valid OpenAPI spec
// from the API provided.
type builder struct {
	ctx       context.Context
	conf      *BuilderConfig
	Spec      *Spec
	logger    log.Logger
	generator *openapi3gen.Generator

	// cache used to store the generated schemas for enums
	enumCache  map[string]bool
	modelCache map[string]reflect.Type
}

type BuilderConfig struct {
	API        Impl
	Customizer openapi3gen.SchemaCustomizerFn

	*openapi3.Info
	*openapi3.Servers
	*openapi3.Schemas
}

// NewBuilder creates a new builder.
func NewBuilder(ctx context.Context, v string, conf *BuilderConfig) *builder {
	logger := log.FromCtx(ctx)

	if v == "" {
		v = "3.0.0"
		logger.Info("no OpenAPI version specified. Falling back to default", tag.NewStringTag("version", v))
	}

	b := &builder{
		ctx:    ctx,
		logger: logger,
		Spec: &Spec{
			Model: openapi3.T{
				OpenAPI: v,
			},
		},
		conf:       conf,
		enumCache:  make(map[string]bool),
		modelCache: make(map[string]reflect.Type),
	}

	// confgigure the openapi generator
	var opts []openapi3gen.Option
	opts = append(opts, openapi3gen.UseAllExportedFields())
	if conf.Customizer != nil {
		opts = append(opts, openapi3gen.SchemaCustomizer(chainCustomizer(b.customizer, conf.Customizer)))
	} else {
		opts = append(opts, openapi3gen.SchemaCustomizer(b.customizer))
	}
	b.generator = openapi3gen.NewGenerator(opts...)
	return b
}

// Build builds a valid OpenAPI spec.
func (b *builder) Build() error {

	b.Spec.Model.Info = b.conf.Info
	b.Spec.Model.Servers = *b.conf.Servers

	// initialize the components, that include:
	// - request bodies
	// - parameters
	// - responses
	// - schemas
	// - headers
	// - security schemes
	c := openapi3.NewComponents()
	c.RequestBodies = openapi3.RequestBodies{}
	c.Parameters = openapi3.ParametersMap{}
	c.Responses = openapi3.NewResponses()
	c.Schemas = openapi3.Schemas{}
	c.Headers = openapi3.Headers{}
	c.SecuritySchemes = openapi3.SecuritySchemes{}
	b.Spec.Model.Components = c

	if err := b.buildPaths(); err != nil {
		return err
	}
	b.resolveRefPaths()
	b.markPropsAsRequired()
	return nil
}

// buildPaths builds the paths from the provided API implementation.
// For each path found, it iterates over the operations and creates an OpenAPI operation.
// The operation is then added to the path.
func (b *builder) buildPaths() error {
	paths := openapi3.Paths{}
	for _, p := range b.conf.API.GetPaths() {
		if err := p.isValid(); err != nil {
			return err
		}
		// for each operation in the path, build the operation
		// and create the components
		path := &openapi3.PathItem{}
		for _, o := range p.Operations {
			op, err := b.newOpenApiOperation(o)
			if err != nil {
				return err
			}
			switch o.Method {
			case http.MethodGet:
				path.Get = op
			case http.MethodPost:
				path.Post = op
			case http.MethodPut:
				path.Put = op
			case http.MethodDelete:
				path.Delete = op
			}
		}
		paths[p.Template] = path
	}
	b.Spec.Model.Paths = paths
	return nil
}

// newOpenApiOperation creates a new OpenAPI operation from a user provided
// operation, by creating a request body if necessary, a response, as well
// as the parameters defined and the security requiremens.
func (b *builder) newOpenApiOperation(in *Operation) (*openapi3.Operation, error) {
	reqBody, err := b.buildReqBody(in.RequestBody)
	if err != nil {
		return nil, err
	}
	resp, err := b.buildResp(in.Responses)
	if err != nil {
		return nil, err
	}
	params := b.buildParams(in.Parameters)
	security := b.buildSecReq(in.SecurityParam)

	o := &openapi3.Operation{
		Tags:        in.Tags,
		OperationID: in.OperationID,
		Parameters:  params,
		RequestBody: reqBody,
		Responses:   resp,
		Security:    security,
	}

	if in.Meta != nil {
		o.Summary = in.Meta.Summary
		o.Description = in.Meta.Description
	}
	return o, nil
}

// markPropsAsRequired marks the properties of a schema as required if they
// contain the json tag contains `omitempty`.
func (b *builder) markPropsAsRequired() {
	for k := range b.Spec.Model.Components.Schemas {
		requiredProps := make([]string, 0)
		if t := b.modelCache[k]; t != nil {
			requiredProps = append(requiredProps, b.extractRequiredProps(t)...)
		}
		b.Spec.Model.Components.Schemas[k].Value.Required = requiredProps
	}
}

func (b *builder) extractRequiredProps(t reflect.Type) []string {
	requiredProps := make([]string, 0)
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.Anonymous {
			// if the field is anonymous, we need to extract the required properties
			// from the embedded struct which should also be in the model cache
			if e, ok := b.modelCache[f.Type.Name()]; ok {
				requiredProps = append(requiredProps, b.extractRequiredProps(e)...)
			}
		} else {
			// we make a property as required if the json tag does not contain
			// the "omitempty" option
			if isRequiredProp(f.Tag.Get("json")) {
				split := strings.Split(f.Tag.Get("json"), ",")
				if len(split) > 0 {
					requiredProps = append(requiredProps, split[0])
				}
			}
		}
	}
	return requiredProps
}

func isRequiredProp(tag string) bool {
	return !strings.Contains(tag, "omitempty")
}

func (b *builder) resolveRefPaths() {
	for _, ref := range b.Spec.Model.Components.Schemas {
		for _, propRef := range ref.Value.Properties {
			if strings.Contains(propRef.Ref, schemasPath) {
				continue
			}
			// resolver logic:
			// - if the ref is found in the enumCache: make a ref to that schema
			// - if the propRef.Type is a primitiveType or empty, set the propRef.Ref empty
			//	  so that ref.Value is used, when writing the specs.
			if b.enumCache[propRef.Ref] {
				propRef.Ref = fmt.Sprintf("%s/%s", schemasPath, propRef.Ref)
			} else if isPrimitiveTypeOrEmpty(propRef.Value.Type) {
				propRef.Ref = ""
			} else if propRef.Value.Type == string(SchemeType_Array) {
				propRef.Value.Items.Ref = formatSchemaRefPath(propRef.Value.Items, propRef.Value.Items.Ref)
			} else if propRef.Value.AdditionalProperties != nil {
				if len(propRef.Value.AdditionalProperties.Ref) != 0 {
					propRef.Value.AdditionalProperties.Ref = formatSchemaRefPath(propRef.Value.AdditionalProperties, propRef.Value.AdditionalProperties.Ref)
				} else if propRef.Value.AdditionalProperties.Value.Type == string(SchemeType_Array) {
					propRef.Value.AdditionalProperties.Value.Items.Ref = formatSchemaRefPath(propRef.Value.AdditionalProperties, propRef.Value.AdditionalProperties.Value.Items.Ref)
				} else if len(propRef.Ref) != 0 {
					propRef.Ref = formatSchemaRefPath(propRef, propRef.Ref)
				}
			} else {
				propRef.Ref = formatSchemaRefPath(propRef, propRef.Ref)
			}
		}
	}
}

func (b *builder) buildParams(params []*Parameter) openapi3.Parameters {
	parameters := openapi3.Parameters{}
	for _, param := range params {
		if ex, ok := b.Spec.Model.Components.Parameters[param.ID]; ok {
			parameters = append(parameters, &openapi3.ParameterRef{
				Ref:   formatParameterRefPath(param),
				Value: ex.Value,
			})
		} else {
			parameterRef := &openapi3.ParameterRef{
				Value: &openapi3.Parameter{
					Name:        param.Name,
					Description: param.Description,
					Schema: openapi3.NewSchemaRef("", &openapi3.Schema{
						Type: string(param.SchemeType),
					}),
					In:       string(param.In),
					Required: param.Required,
				},
			}
			b.Spec.Model.Components.Parameters[param.ID] = parameterRef
			parameters = append(parameters, &openapi3.ParameterRef{
				Ref:   formatParameterRefPath(param),
				Value: parameterRef.Value,
			})
		}
	}
	return parameters
}

func (b *builder) buildSecReq(param *SecurityParam) *openapi3.SecurityRequirements {
	if param == nil {
		return nil
	}

	var has bool
	var ref *openapi3.SecuritySchemeRef
	reqs := &openapi3.SecurityRequirements{}
	if ref, has = b.Spec.Model.Components.SecuritySchemes[param.Name]; !has {

		switch param.SecurityType {
		case SecurityType_ApiKey:
			ref = &openapi3.SecuritySchemeRef{
				Ref: "",
				Value: &openapi3.SecurityScheme{
					Type: string(SecurityType_ApiKey),
					Name: param.Name,
					In:   string(param.In),
				},
			}
		case SecurityType_Http:
			ref = &openapi3.SecuritySchemeRef{
				Ref: "",
				Value: &openapi3.SecurityScheme{
					Type:         string(SecurityType_Http),
					Scheme:       string(param.SecurityScheme),
					BearerFormat: param.BearerFormat,
				},
			}
		default:
			return nil
		}
		reqs.With(openapi3.SecurityRequirement{
			param.ID: {},
		})
		b.Spec.Model.Components.SecuritySchemes[param.ID] = ref
	} else {
		reqs.With(openapi3.SecurityRequirement{
			ref.Value.Name: {},
		})
	}
	return reqs

}

func (b *builder) buildResp(ins []*Response) (openapi3.Responses, error) {
	resp := openapi3.Responses{}

	for _, in := range ins {
		if in.Meta == nil {
			continue
		}
		ref, has := b.Spec.Model.Components.Responses[in.Meta.Name]
		if !has {
			ref = &openapi3.ResponseRef{
				Ref: "",
				Value: &openapi3.Response{
					Description: &in.Meta.Description,
					Headers:     b.buildHeaders(in.Headers),
				},
			}
			if in.Content != nil {
				schemaRef, err := b.buildSchema(in.Content.Model)
				if err != nil {
					return nil, err
				}

				switch in.Content.SchemaType {
				case SchemeType_Object:
					ref.Value.Content = openapi3.NewContentWithSchemaRef(&openapi3.SchemaRef{
						Ref:   formatSchemaRefPath(schemaRef, in.Content.Model.Name()),
						Value: schemaRef.Value,
					}, []string{in.Content.Type.String()})

				case SchemeType_Array:
					ref.Value.Content = openapi3.NewContentWithSchemaRef(openapi3.NewSchemaRef("", &openapi3.Schema{
						Type: string(SchemeType_Array),
						Items: &openapi3.SchemaRef{
							Ref:   formatSchemaRefPath(schemaRef, in.Content.Model.Name()),
							Value: schemaRef.Value,
						},
					}), []string{in.Content.Type.String()})

				default:
					ref.Value.Content = openapi3.NewContentWithSchemaRef(&openapi3.SchemaRef{
						Ref:   formatSchemaRefPath(schemaRef, in.Content.Model.Name()),
						Value: schemaRef.Value,
					}, []string{in.Content.Type.String()})
				}
			}

		}
		if _, has = b.Spec.Model.Components.Responses[in.Meta.Name]; !has {
			b.Spec.Model.Components.Responses[in.Meta.Name] = ref
		}

		resp[strconv.Itoa(in.Code)] = &openapi3.ResponseRef{
			Ref:   fmt.Sprintf("%s/%s", responsesPath, in.Meta.Name),
			Value: ref.Value,
		}
	}
	return resp, nil
}

func (b *builder) buildHeaders(in []*ResponseHeader) openapi3.Headers {
	refs := openapi3.Headers{}

	for _, h := range in {
		headerRef, has := b.Spec.Model.Components.Headers[h.Name]
		if !has {
			headerRef = &openapi3.HeaderRef{
				Ref: "",
				Value: &openapi3.Header{
					Parameter: openapi3.Parameter{
						Description: h.Description,
						Schema: openapi3.NewSchemaRef("", &openapi3.Schema{
							Type: string(h.SchemaType),
						}),
					},
				},
			}
		}

		refs[h.Name] = headerRef
	}
	return refs
}

func (b *builder) buildReqBody(r *RequestBody) (*openapi3.RequestBodyRef, error) {
	if r == nil {
		return nil, nil
	}
	// Check if the model is already created in the spec
	// if so, return the ref to the model
	// else create the model and return the ref to the model
	if has, ok := b.Spec.Model.Components.RequestBodies[r.Model.Name()]; ok {
		return has, nil
	}

	rb := openapi3.NewRequestBody()
	sr, err := b.buildSchema(r.Model)
	if err != nil {
		return nil, err
	}

	rb.Required = true
	rb.Content = openapi3.NewContentWithSchemaRef(&openapi3.SchemaRef{
		Ref:   formatSchemaRefPath(sr, r.Model.Name()),
		Value: sr.Value,
	}, []string{ContentType_ApplicationJson.String()})

	b.Spec.Model.Components.RequestBodies[r.Model.Name()] = &openapi3.RequestBodyRef{
		Ref:   "",
		Value: rb,
	}
	return b.Spec.Model.Components.RequestBodies[r.Model.Name()], nil
}

func (b *builder) addToModelCache(model reflect.Type) {
	if _, ok := b.modelCache[model.Name()]; !ok {
		b.modelCache[model.Name()] = model
		for i := 0; i < model.NumField(); i++ {
			field := model.Field(i)
			if field.Anonymous || field.Type.Kind() == reflect.Struct || field.Type.Kind() == reflect.Ptr {
				b.addToModelCache(field.Type)
			}

			if field.Type.Kind() == reflect.Slice {
				t := field.Type.Elem()
				if !isPrimitiveTypeOrEmpty(strings.ReplaceAll(t.String(), "*", "")) {
					b.addToModelCache(t)
				}
			}
		}
	}
}

func (b *builder) buildSchema(model reflect.Type) (*openapi3.SchemaRef, error) {
	var err error
	var ref *openapi3.SchemaRef

	ref, ok := b.Spec.Model.Components.Schemas[model.Name()]
	if !ok {
		// GenerateSchemaRef generates SchemaRefs recursively
		// Assuming this, all embedded models are created in Components.Schemas
		// However, they are not in the modelCache. This leads to problems afterwards,
		// when trying to resolve required properties.
		b.addToModelCache(model)
		ref, err = b.generator.GenerateSchemaRef(model)
		if err != nil {
			return nil, err
		}
	}
	// Add all generated schemas
	for sr := range b.generator.SchemaRefs {
		// omit if this is a primitive type
		if isPrimitiveTypeOrEmpty(sr.Ref) {
			continue
		}
		if len(sr.Ref) > 0 && !strings.Contains(sr.Ref, schemasPath) {
			if _, ok := b.Spec.Model.Components.Schemas[sr.Ref]; !ok {
				b.Spec.Model.Components.Schemas[sr.Ref] = openapi3.NewSchemaRef("", sr.Value)
			}
		}
	}
	return ref, nil
}

func isPrimitiveTypeOrEmpty(t string) bool {
	return t == "" ||
		t == "string" ||
		t == "boolean" ||
		t == "bool" ||
		t == "float" ||
		t == "float32" ||
		t == "float64" ||
		t == "integer" ||
		t == "int" ||
		t == "number" ||
		t == "int32" ||
		t == "int64"

}

func formatSchemaRefPath(ref *openapi3.SchemaRef, name string) string {
	if isPrimitiveTypeOrEmpty(ref.Value.Type) {
		return ""
	}
	if strings.Contains(ref.Ref, schemasPath) {
		return ref.Ref
	}
	return fmt.Sprintf("%s/%s", schemasPath, name)
}

func formatParameterRefPath(param *Parameter) string {
	return fmt.Sprintf("%s/%s", parametersPath, param.ID)
}

// Enum is an interface which must be implemented by types
// that represent an enum in an OpenApi-Document.
type Enum interface {
	OpenApiValues() []interface{}
}

func (b *builder) customizer(name string, t reflect.Type, tag reflect.StructTag, schema *openapi3.Schema) error {
	// Enumeration Customizer
	if t.Implements(reflect.TypeOf((*Enum)(nil)).Elem()) {
		schema.Type = "string"

		m, _ := t.MethodByName("OpenApiValues")
		in := make([]reflect.Value, m.Type.NumIn())
		for i := 0; i < m.Type.NumIn(); i++ {
			in[i] = reflect.Zero(m.Type.In(i))
		}
		res := m.Func.Call(in)
		schema.Enum = res[0].Interface().([]interface{})

		if has := b.enumCache[t.Name()]; !has {
			b.enumCache[t.Name()] = true
		}
	}

	// Description Customizer
	if descr := tag.Get("descr"); descr != "" {
		schema.Description = descr
	}

	return nil
}

func chainCustomizer(customizers ...openapi3gen.SchemaCustomizerFn) openapi3gen.SchemaCustomizerFn {
	return func(name string, t reflect.Type, tag reflect.StructTag, schema *openapi3.Schema) error {
		for _, c := range customizers {
			if err := c(name, t, tag, schema); err != nil {
				return err
			}
		}
		return nil
	}
}
