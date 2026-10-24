package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	billingv1 "github.com/netflix-clone/proto/billing/v1"
	"github.com/netflix-clone/services/billing-svc/internal/service"
	"google.golang.org/grpc"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	dbURL := os.Getenv("DATABASE_URL")
	grpcAddr := os.Getenv("BILLING_GRPC_ADDR")

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("postgres connection error: %v", err)
	}
	defer pool.Close()

	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("failed to listen on %s: %v", grpcAddr, err)
	}

	grpcServer := grpc.NewServer()
	billingv1.RegisterBillingServiceServer(grpcServer, service.NewBillingGrpcServer(pool))

	go func() {
		log.Printf("[billing-svc] listening on %s", grpcAddr)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("grpc serve error: %v", err)
		}
	}()

	<-ctx.Done()
	grpcServer.GracefulStop()
}

