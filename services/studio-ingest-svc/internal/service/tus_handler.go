package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	ingestv1 "github.com/netflix-clone/proto/ingest/v1"
	"github.com/netflix-clone/services/studio-ingest-svc/internal/db/sqlc"
	"github.com/netflix-clone/services/studio-ingest-svc/internal/storage"
	"github.com/segmentio/kafka-go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const defaultTusChunkBytes int32 = 16 * 1024 * 1024

type StudioIngestGrpcServer struct {
	ingestv1.UnimplementedStudioIngestServiceServer
	pool        *pgxpool.Pool
	queries     *db.Queries
	store       *storage.MinioStore
	kafkaWriter *kafka.Writer
	publicURL   string
}

func NewStudioIngestGrpcServer(pool *pgxpool.Pool, store *storage.MinioStore, kw *kafka.Writer, pubURL string) *StudioIngestGrpcServer {
	return &StudioIngestGrpcServer{
		pool:        pool,
		queries:     db.New(pool),
		store:       store,
		kafkaWriter: kw,
		publicURL:   pubURL,
	}
}

func (s *StudioIngestGrpcServer) CreateTusSession(ctx context.Context, req *ingestv1.CreateTusSessionRequest) (*ingestv1.CreateTusSessionResponse, error) {
	epID, err := uuid.Parse(req.GetEpisodeId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid episode_id")
	}
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}
	orgID, err := uuid.Parse(req.GetPartnerOrgId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid partner_org_id")
	}
	if req.GetTotalByteSize() <= 0 || len(req.GetExpectedSha256()) != 64 {
		return nil, status.Error(codes.InvalidArgument, "invalid byte size or sha256")
	}
	objKey := fmt.Sprintf("raw/%s/%s/%s", orgID.String(), epID.String(), req.GetFileName())
	sess, err := s.queries.CreateIngestSession(ctx, db.CreateIngestSessionParams{
		EpisodeID:       epID,
		UploadedBy:      userID,
		PartnerOrgID:    orgID,
		FileName:        req.GetFileName(),
		SourceContainer: req.GetSourceContainer(),
		TotalByteSize:   req.GetTotalByteSize(),
		ChunkSizeBytes:  defaultTusChunkBytes,
		ExpectedSha256:  req.GetExpectedSha256(),
		MinioBucket:     "nf-raw-ingest",
		MinioObjectKey:  objKey,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create tus session: %v", err)
	}
	return &ingestv1.CreateTusSessionResponse{
		IngestSessionId:   sess.ID.String(),
		ChunkSizeBytes:    sess.ChunkSizeBytes,
		CurrentByteOffset: sess.UploadedByteOffset,
		UploadEndpointUrl: fmt.Sprintf("%s/tus/%s", s.publicURL, sess.ID.String()),
	}, nil
}

func (s *StudioIngestGrpcServer) AppendTusChunk(ctx context.Context, req *ingestv1.AppendTusChunkRequest) (*ingestv1.AppendTusChunkResponse, error) {
	sessID, err := uuid.Parse(req.GetIngestSessionId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid ingest_session_id")
	}
	sess, err := s.queries.GetIngestSessionByID(ctx, sessID)
	if err != nil {
		return nil, status.Error(codes.NotFound, "ingest session not found")
	}
	if sess.UploadedByteOffset != req.GetStartByte() {
		return nil, status.Errorf(codes.FailedPrecondition, "byte offset mismatch: expected %d, got %d", sess.UploadedByteOffset, req.GetStartByte())
	}
	payload := req.GetPayload()
	if len(payload) == 0 {
		return nil, status.Error(codes.InvalidArgument, "empty chunk payload")
	}
	sum := sha256.Sum256(payload)
	computedHex := hex.EncodeToString(sum[:])
	if computedHex != req.GetChunkSha256() {
		return nil, status.Error(codes.DataLoss, "chunk sha256 checksum mismatch")
	}
	etag, err := s.store.PutChunkPart(ctx, sess.MinioBucket, sess.ID.String(), req.GetChunkIndex(), payload)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "minio put part failed: %v", err)
	}
	return s.commitChunkOffset(ctx, sess.ID, req.GetChunkIndex(), req.GetStartByte(), int32(len(payload)), computedHex, etag)
}

func (s *StudioIngestGrpcServer) commitChunkOffset(ctx context.Context, sessID uuid.UUID, idx int32, start int64, length int32, shaHex, etag string) (*ingestv1.AppendTusChunkResponse, error) {
	endByte := start + int64(length) - 1
	err := s.queries.InsertIngestChunk(ctx, db.InsertIngestChunkParams{
		IngestSessionID: sessID,
		ChunkIndex:      idx,
		StartByte:       start,
		EndByte:         endByte,
		ByteLength:      length,
		ChunkSha256:     shaHex,
		Etag:            etag,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "chunk db insert error: %v", err)
	}
	newOffset, err := s.queries.AdvanceIngestByteOffset(ctx, db.AdvanceIngestByteOffsetParams{
		ID:                   sessID,
		UploadedByteOffset:   start + int64(length),
		UploadedByteOffset_2: start,
	})
	if err != nil {
		return nil, status.Error(codes.Aborted, "concurrent offset modification detected")
	}
	return &ingestv1.AppendTusChunkResponse{NewByteOffset: newOffset, Etag: etag}, nil
}

func (s *StudioIngestGrpcServer) CompleteTusUpload(ctx context.Context, req *ingestv1.CompleteTusUploadRequest) (*ingestv1.CompleteTusUploadResponse, error) {
	sessID, err := uuid.Parse(req.GetIngestSessionId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid ingest_session_id")
	}
	sess, err := s.queries.GetIngestSessionByID(ctx, sessID)
	if err != nil {
		return nil, status.Error(codes.NotFound, "session not found")
	}
	if sess.UploadedByteOffset != sess.TotalByteSize {
		return nil, status.Errorf(codes.FailedPrecondition, "upload incomplete: %d / %d bytes", sess.UploadedByteOffset, sess.TotalByteSize)
	}
	chunks, err := s.queries.ListIngestChunksOrdered(ctx, sessID)
	if err != nil || len(chunks) == 0 {
		return nil, status.Error(codes.Internal, "failed to list uploaded chunks")
	}
	fullHash, err := s.store.ComposeAndHashObject(ctx, sess.MinioBucket, sess.ID.String(), sess.MinioObjectKey, len(chunks))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "object composition failed: %v", err)
	}
	if fullHash != sess.ExpectedSha256 {
		return &ingestv1.CompleteTusUploadResponse{Sha256Verified: false, ComputedSha256: fullHash}, nil
	}
	_ = s.queries.MarkIngestCompleted(ctx, db.MarkIngestCompletedParams{
		ID:             sess.ID,
		ComputedSha256: pgtype.Text{String: fullHash, Valid: true},
	})
	qcID, _ := s.queries.CreateInitialQcReport(ctx, db.CreateInitialQcReportParams{
		IngestSessionID: sess.ID,
		EpisodeID:       sess.EpisodeID,
	})
	s.publishIngestCompletedEvent(ctx, sess.ID.String(), sess.EpisodeID.String(), sess.MinioBucket, sess.MinioObjectKey)
	return &ingestv1.CompleteTusUploadResponse{
		Sha256Verified:      true,
		ComputedSha256:      fullHash,
		TriggeredQcReportId: qcID.String(),
	}, nil
}

func (s *StudioIngestGrpcServer) publishIngestCompletedEvent(ctx context.Context, sessionID, episodeID, bucket, key string) {
	payload, _ := json.Marshal(map[string]string{
		"ingest_session_id": sessionID,
		"episode_id":        episodeID,
		"bucket":            bucket,
		"object_key":        key,
	})
	_ = s.kafkaWriter.WriteMessages(ctx, kafka.Message{
		Key:   []byte(episodeID),
		Value: payload,
	})
}

