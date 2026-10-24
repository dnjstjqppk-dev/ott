package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	iamv1 "github.com/netflix-clone/proto/iam/v1"
	"github.com/netflix-clone/services/user-profile-svc/internal/security"
	"github.com/netflix-clone/services/user-profile-svc/internal/service"
	"google.golang.org/grpc"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	dbURL := os.Getenv("DATABASE_URL")
	jwtSecret := os.Getenv("JWT_HMAC_SECRET")
	grpcAddr := os.Getenv("IAM_GRPC_ADDR")
	if dbURL == "" || jwtSecret == "" || grpcAddr == "" {
		log.Fatal("DATABASE_URL, JWT_HMAC_SECRET, and IAM_GRPC_ADDR must be set")
	}

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("unable to connect to postgres: %v", err)
	}
	defer pool.Close()

	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("failed to bind tcp %s: %v", grpcAddr, err)
	}

	tokenMgr := security.NewTokenManager(jwtSecret, 15*time.Minute)
	grpcServer := grpc.NewServer()
	iamv1.RegisterIamServiceServer(grpcServer, service.NewIamGrpcServer(pool, tokenMgr))

	go func() {
		log.Printf("[user-profile-svc] listening on %s", grpcAddr)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("grpc serve error: %v", err)
		}
	}()

	<-ctx.Done()
	grpcServer.GracefulStop()
}

