package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/netflix-clone/services/api-gateway/internal/clients"
	"github.com/netflix-clone/services/api-gateway/internal/handler"
	"github.com/netflix-clone/services/api-gateway/internal/middleware"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	svcClients, err := clients.NewServiceClients(clients.EndpointsConfig{
		IamAddr:      os.Getenv("IAM_GRPC_ADDR"),
		CatalogAddr:  os.Getenv("CATALOG_GRPC_ADDR"),
		BillingAddr:  os.Getenv("BILLING_GRPC_ADDR"),
		PlaybackAddr: os.Getenv("PLAYBACK_GRPC_ADDR"),
		RecoAddr:     os.Getenv("RECO_GRPC_ADDR"),
	})
	if err != nil {
		log.Fatalf("failed to initialize grpc clients: %v", err)
	}
	defer svcClients.CloseAll()

	bff := handler.NewBffHandler(svcClients)
	r := chi.NewRouter()
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Recoverer)

	r.Post("/api/v1/auth/login", bff.HandleLogin)
	r.Group(func(pr chi.Router) {
		pr.Use(middleware.RequireAuthentication(svcClients.Iam))
		pr.Get("/api/v1/browse/lolomo", bff.HandleGetAggregatedHome)
		pr.Get("/api/v1/catalog/titles/{titleId}", bff.HandleGetTitleModal)
	})

	srv := &http.Server{
		Addr:         os.Getenv("GATEWAY_HTTP_ADDR"),
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	go func() {
		log.Printf("[api-gateway] listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http server error: %v", err)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
}

