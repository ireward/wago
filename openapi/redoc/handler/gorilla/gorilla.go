package gorilla

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/ireward/wago/openapi"
	"github.com/ireward/wago/openapi/redoc"
)

type Handler struct {
}

func (h *Handler) Register(ctx context.Context, router *mux.Router, redoc *redoc.Type, conf *openapi.BuilderConfig) error {
	builder := openapi.NewBuilder(ctx, "", conf)
	if err := builder.Build(); err != nil {
		return err
	}
	spec, err := builder.Spec.ToBytes()
	if err != nil {
		return err
	}
	body, err := redoc.Body()
	if err != nil {
		return err
	}

	// Register the spec-handler
	router.HandleFunc(redoc.SpecPath, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write(spec)
	}).Methods("GET")

	// Register the redoc-handler
	router.HandleFunc(redoc.DocPath, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "text/html")
		w.Write(body)
	}).Methods("GET")

	return nil
}
