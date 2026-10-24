package e2e

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	ingestv1 "github.com/netflix-clone/proto/ingest/v1"
	mediav1 "github.com/netflix-clone/proto/media/v1"
	playbackv1 "github.com/netflix-clone/proto/playback/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestConcurrencyStreamLimitRaceGate(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	env := BootstrapLiveEnvironment(ctx, t)
	defer env.Pool.Close()

	pbCli := dialGrpc(t, os.Getenv("PLAYBACK_GRPC_ADDR"), playbackv1.NewPlaybackServiceClient)

	const totalConcurrentAttempts = 12
	var grantedCount int32
	var exhaustedCount int32
	var wg sync.WaitGroup

	for i := 0; i < totalConcurrentAttempts; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := pbCli.AcquirePlaybackSession(ctx, &playbackv1.AcquirePlaybackSessionRequest{
				UserId:      env.UserID.String(),
				ProfileId:   env.ProfileID.String(),
				EpisodeId:   env.EpisodeID.String(),
				DeviceId:    env.DeviceID.String(),
				ClientIp:    "127.0.0.1",
				CountryCode: "KR",
			})
			if err == nil {
				atomic.AddInt32(&grantedCount, 1)
			} else if status.Code(err) == codes.ResourceExhausted {
				atomic.AddInt32(&exhaustedCount, 1)
			}
		}()
	}
	wg.Wait()

	if grantedCount != 2 || exhaustedCount != 10 {
		t.Fatalf("concurrency gate failed: expected 2 granted and 10 exhausted, got granted=%d, exhausted=%d", grantedCount, exhaustedCount)
	}
}

func TestEdgeNodeDrainFailoverUnderLoad(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	env := BootstrapLiveEnvironment(ctx, t)
	defer env.Pool.Close()

	mediaCli := dialGrpc(t, os.Getenv("MEDIA_WORKER_GRPC_ADDR"), mediav1.NewMediaPipelineServiceClient)

	initialSel, err := mediaCli.SelectOptimalEdgeNode(ctx, &mediav1.SelectOptimalEdgeNodeRequest{
		ClientIp:    "127.0.0.1",
		CountryCode: "KR",
		EpisodeId:   env.EpisodeID.String(),
	})
	if err != nil || initialSel.GetEdgeNodeId() != env.PrimaryEdgeID.String() {
		t.Fatalf("expected primary edge node %s initially, got %v", env.PrimaryEdgeID, initialSel)
	}

	_, err = mediaCli.SetEdgeNodeDrainMode(ctx, &mediav1.SetEdgeNodeDrainModeRequest{
		EdgeNodeId:     env.PrimaryEdgeID.String(),
		TargetStatus:   "DRAINING",
		OperatorUserId: env.UserID.String(),
	})
	if err != nil {
		t.Fatalf("SetEdgeNodeDrainMode failed: %v", err)
	}

	failoverSel, err := mediaCli.SelectOptimalEdgeNode(ctx, &mediav1.SelectOptimalEdgeNodeRequest{
		ClientIp:    "127.0.0.1",
		CountryCode: "KR",
		EpisodeId:   env.EpisodeID.String(),
	})
	if err != nil || failoverSel.GetEdgeNodeId() != env.BackupEdgeID.String() {
		t.Fatalf("failover routing failed: expected backup node %s, got %v", env.BackupEdgeID, failoverSel)
	}
}

func TestTusInterruptedOffsetRecovery(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	env := BootstrapLiveEnvironment(ctx, t)
	defer env.Pool.Close()

	ingestCli := dialGrpc(t, os.Getenv("INGEST_GRPC_ADDR"), ingestv1.NewStudioIngestServiceClient)

	sess, err := ingestCli.CreateTusSession(ctx, &ingestv1.CreateTusSessionRequest{
		EpisodeId:       env.EpisodeID.String(),
		UserId:          env.UserID.String(),
		PartnerOrgId:    env.PartnerOrgID.String(),
		FileName:        "chaos_recovery_master.mov",
		SourceContainer: "MOV_PRORES",
		TotalByteSize:   int64(len(env.RawMasterBytes)),
		ExpectedSha256:  env.RawMasterSHA,
	})
	if err != nil {
		t.Fatalf("CreateTusSession failed: %v", err)
	}

	chunkSize := int(sess.GetChunkSizeBytes())
	firstSlice := env.RawMasterBytes[:chunkSize]
	firstHash := sha256.Sum256(firstSlice)

	_, err = ingestCli.AppendTusChunk(ctx, &ingestv1.AppendTusChunkRequest{
		IngestSessionId: sess.GetIngestSessionId(),
		ChunkIndex:      0,
		StartByte:       0,
		Payload:         firstSlice,
		ChunkSha256:     hex.EncodeToString(firstHash[:]),
	})
	if err != nil {
		t.Fatalf("chunk 0 upload failed: %v", err)
	}

	_, wrongOffsetErr := ingestCli.AppendTusChunk(ctx, &ingestv1.AppendTusChunkRequest{
		IngestSessionId: sess.GetIngestSessionId(),
		ChunkIndex:      1,
		StartByte:       int64(chunkSize - 512),
		Payload:         env.RawMasterBytes[chunkSize:],
		ChunkSha256:     "invalid_hash",
	})
	if status.Code(wrongOffsetErr) != codes.FailedPrecondition {
		t.Fatalf("expected FailedPrecondition on misaligned byte offset, got %v", wrongOffsetErr)
	}

	secondSlice := env.RawMasterBytes[chunkSize:]
	secondHash := sha256.Sum256(secondSlice)
	_, err = ingestCli.AppendTusChunk(ctx, &ingestv1.AppendTusChunkRequest{
		IngestSessionId: sess.GetIngestSessionId(),
		ChunkIndex:      1,
		StartByte:       int64(chunkSize),
		Payload:         secondSlice,
		ChunkSha256:     hex.EncodeToString(secondHash[:]),
	})
	if err != nil {
		t.Fatalf("resumed chunk 1 upload failed: %v", err)
	}

	comp, err := ingestCli.CompleteTusUpload(ctx, &ingestv1.CompleteTusUploadRequest{
		IngestSessionId: sess.GetIngestSessionId(),
	})
	if err != nil || !comp.GetSha256Verified() {
		t.Fatalf("resumed upload final SHA-256 verification failed: %v", err)
	}
}

