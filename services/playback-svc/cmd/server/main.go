package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	playbackv1 "github.com/netflix-clone/proto/playback/v1"
	"github.com/netflix-clone/services/playback-svc/internal/service"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	dbURL := os.Getenv("DATABASE_URL")
	redisURL := os.Getenv("REDIS_URL")
	cdnBase := os.Getenv("CDN_ASSET_BASE_URL")
	grpcAddr := os.Getenv("PLAYBACK_GRPC_ADDR")

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("postgres connect error: %v", err)
	}
	defer pool.Close()

	opt, _ := redis.ParseURL(redisURL)
	rdb := redis.NewClient(opt)
	defer rdb.Close()

	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("listen error: %v", err)
	}

	grpcServer := grpc.NewServer()
	playbackv1.RegisterPlaybackServiceServer(grpcServer, service.NewPlaybackGrpcServer(pool, rdb, cdnBase))

	go func() {
		log.Printf("[playback-svc] listening on %s", grpcAddr)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("grpc serve error: %v", err)
		}
	}()

	<-ctx.Done()
	grpcServer.GracefulStop()
}

