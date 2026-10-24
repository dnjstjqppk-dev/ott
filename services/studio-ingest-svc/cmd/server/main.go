package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	ingestv1 "github.com/netflix-clone/proto/ingest/v1"
	"github.com/netflix-clone/services/studio-ingest-svc/internal/service"
	"github.com/netflix-clone/services/studio-ingest-svc/internal/storage"
	"github.com/segmentio/kafka-go"
	"google.golang.org/grpc"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	dbURL := os.Getenv("DATABASE_URL")
	minioEndpoint := os.Getenv("MINIO_ENDPOINT")
	minioKey := os.Getenv("MINIO_ACCESS_KEY")
	minioSecret := os.Getenv("MINIO_SECRET_KEY")
	kafkaBroker := os.Getenv("KAFKA_BROKER")
	grpcAddr := os.Getenv("INGEST_GRPC_ADDR")
	publicURL := os.Getenv("INGEST_PUBLIC_URL")

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("postgres connect error: %v", err)
	}
	defer pool.Close()

	store, err := storage.NewMinioStore(minioEndpoint, minioKey, minioSecret, false)
	if err != nil {
		log.Fatalf("minio init error: %v", err)
	}

	kw := &kafka.Writer{
		Addr:     kafka.TCP(kafkaBroker),
		Topic:    "ingest.completed",
		Balancer: &kafka.LeastBytes{},
	}
	defer kw.Close()

	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("listen error: %v", err)
	}

	grpcServer := grpc.NewServer(grpc.MaxRecvMsgSize(32 * 1024 * 1024))
	ingestv1.RegisterStudioIngestServiceServer(grpcServer, service.NewStudioIngestGrpcServer(pool, store, kw, publicURL))

	go func() {
		log.Printf("[studio-ingest-svc] listening on %s", grpcAddr)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("grpc serve error: %v", err)
		}
	}()

	<-ctx.Done()
	grpcServer.GracefulStop()
}

