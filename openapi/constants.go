package openapi

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
