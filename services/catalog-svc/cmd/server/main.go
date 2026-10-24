package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	catalogv1 "github.com/netflix-clone/proto/catalog/v1"
	"github.com/netflix-clone/services/catalog-svc/internal/search"
	"github.com/netflix-clone/services/catalog-svc/internal/service"
	"google.golang.org/grpc"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	dbURL := os.Getenv("DATABASE_URL")
	meiliURL := os.Getenv("MEILI_URL")
	meiliKey := os.Getenv("MEILI_MASTER_KEY")
	cdnBase := os.Getenv("CDN_ASSET_BASE_URL")
	grpcAddr := os.Getenv("CATALOG_GRPC_ADDR")

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("postgres connect failed: %v", err)
	}
	defer pool.Close()

	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("listen error: %v", err)
	}

	meiliClient := search.NewMeiliClient(meiliURL, meiliKey)
	grpcServer := grpc.NewServer()
	catalogv1.RegisterCatalogServiceServer(grpcServer, service.NewCatalogGrpcServer(pool, meiliClient, cdnBase))

	go func() {
		log.Printf("[catalog-svc] listening on %s", grpcAddr)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("serve failed: %v", err)
		}
	}()

	<-ctx.Done()
	grpcServer.GracefulStop()
}

