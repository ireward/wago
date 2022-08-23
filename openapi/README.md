## OpenAPI Spec Generator

Manual management of OpenAPI specifications is cumbersome and error-prone, leading to inconsistencies between the actual implementation and
the documentation. Currently (as of my knowledge) there is no out of the box solution for generating OpenAPI specifications in Go based on code.
This package is intended to remedy this, by providing the possibility to generate `json` based OpenAPI specifications using code. In order to
generate OpenAPI specifications, this package uses the amazing `kin-openapi` project and custom type definitions, which can be found in `types.go`.
