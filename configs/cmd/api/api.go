package api

import (
	"log/slog"
	"net/http"
)

type API struct {
	addr   string
	logger *slog.Logger
}

func NewApi(addr string, logger *slog.Logger) *API {
	return &API{
		addr:   addr,
		logger: logger,
	}
}

func (a *API) Run() error {
	router := http.NewServeMux()

	_ = router
	server := &http.Server{
		Addr: a.addr,
	}

	return server.ListenAndServe()
}
