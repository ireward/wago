package openapi

// Set of default Security-Parameters, that can be used
// in the Security-Scheme Object when modeling a path.
//
// If you want to use a custom-ID, just assign it before using it:
// WithBearerAuth.ID = "my-custom-id"
var (
	WithBearerAuth = SecurityParam{
		ID:             "BearerAuth",
		SecurityType:   SecurityType_Http,
		SecurityScheme: SecurityScheme_Bearer,
		BearerFormat:   "JWT",
	}

	WithBasicAuth = SecurityParam{
		ID:             "BasicAuth",
		SecurityType:   SecurityType_Http,
		SecurityScheme: SecurityScheme_Basic,
	}

	WithNoAuth = SecurityParam{
		ID:             "NoAuth",
		SecurityType:   SecurityType_Http,
		SecurityScheme: SecurityScheme_NoAuth,
	}
)
