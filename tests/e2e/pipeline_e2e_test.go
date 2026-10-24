package e2e

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	billingv1 "github.com/netflix-clone/proto/billing/v1"
	ingestv1 "github.com/netflix-clone/proto/ingest/v1"
	mediav1 "github.com/netflix-clone/proto/media/v1"
	playbackv1 "github.com/netflix-clone/proto/playback/v1"
	recov1 "github.com/netflix-clone/proto/reco/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestFullNetflixLifecycleE2E(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	env := BootstrapLiveEnvironment(ctx, t)
	defer env.Pool.Close()

	ingestCli := dialGrpc(t, os.Getenv("INGEST_GRPC_ADDR"), ingestv1.NewStudioIngestServiceClient)
	mediaCli := dialGrpc(t, os.Getenv("MEDIA_WORKER_GRPC_ADDR"), mediav1.NewMediaPipelineServiceClient)
	playbackCli := dialGrpc(t, os.Getenv("PLAYBACK_GRPC_ADDR"), playbackv1.NewPlaybackServiceClient)
	recoCli := dialGrpc(t, os.Getenv("RECO_GRPC_ADDR"), recov1.NewRecommendationServiceClient)
	billingCli := dialGrpc(t, os.Getenv("BILLING_GRPC_ADDR"), billingv1.NewBillingServiceClient)

	ingestSessionID, qcReportID := step1TusChunkedIngest(ctx, t, ingestCli, env)
	step2QcAndTimedTextValidation(ctx, t, ingestCli, env, qcReportID)
	step3MediaEncodingDispatch(ctx, t, mediaCli, env, ingestSessionID)
	playbackSessionID := step4AcquireSessionAndVerifyHttp206(ctx, t, playbackCli, mediaCli, env)
	step5HeartbeatAndLoLoMoSync(ctx, t, playbackCli, recoCli, env, playbackSessionID)
	step6VerifyPartnerRoyaltySettlement(ctx, t, billingCli, env)
}

func dialGrpc[T any](t *testing.T, addr string, ctor func(grpc.ClientConnInterface) T) T {
	t.Helper()
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("grpc dial failed for %s: %v", addr, err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	return ctor(conn)
}

func step1TusChunkedIngest(ctx context.Context, t *testing.T, cli ingestv1.StudioIngestServiceClient, env *LiveTestEnvironment) (string, string) {
	t.Helper()
	sess, err := cli.CreateTusSession(ctx, &ingestv1.CreateTusSessionRequest{
		EpisodeId:       env.EpisodeID.String(),
		UserId:          env.UserID.String(),
		PartnerOrgId:    env.PartnerOrgID.String(),
		FileName:        "kingdom_s01e01_master.mov",
		SourceContainer: "MOV_PRORES",
		TotalByteSize:   int64(len(env.RawMasterBytes)),
		ExpectedSha256:  env.RawMasterSHA,
	})
	if err != nil {
		t.Fatalf("CreateTusSession failed: %v", err)
	}
	chunkSize := int(sess.GetChunkSizeBytes())
	for idx, offset := 0, 0; offset < len(env.RawMasterBytes); idx++ {
		end := min(offset+chunkSize, len(env.RawMasterBytes))
		slice := env.RawMasterBytes[offset:end]
		sum := sha256.Sum256(slice)
		res, err := cli.AppendTusChunk(ctx, &ingestv1.AppendTusChunkRequest{
			IngestSessionId: sess.GetIngestSessionId(),
			ChunkIndex:      int32(idx),
			StartByte:       int64(offset),
			Payload:         slice,
			ChunkSha256:     hex.EncodeToString(sum[:]),
		})
		if err != nil || res.GetNewByteOffset() != int64(end) {
			t.Fatalf("AppendTusChunk #%d failed: %v", idx, err)
		}
		offset = end
	}
	comp, err := cli.CompleteTusUpload(ctx, &ingestv1.CompleteTusUploadRequest{IngestSessionId: sess.GetIngestSessionId()})
	if err != nil || !comp.GetSha256Verified() || comp.GetComputedSha256() != env.RawMasterSHA {
		t.Fatalf("CompleteTusUpload SHA-256 verification failed: %v", err)
	}
	return sess.GetIngestSessionId(), comp.GetTriggeredQcReportId()
}

func step2QcAndTimedTextValidation(ctx context.Context, t *testing.T, cli ingestv1.StudioIngestServiceClient, env *LiveTestEnvironment, qcReportID string) {
	t.Helper()
	qcRep, err := cli.SubmitManualQcDecision(ctx, &ingestv1.SubmitManualQcDecisionRequest{
		QcReportId:     qcReportID,
		ReviewerUserId: env.UserID.String(),
		Verdict:        "PASSED",
		ReviewNotes:    "E2E frame-accurate QC approved",
	})
	if err != nil || qcRep.GetVerdict() != "PASSED" {
		t.Fatalf("SubmitManualQcDecision failed: %v", err)
	}
	ttRes, err := cli.SaveTimedTextCues(ctx, &ingestv1.SaveTimedTextCuesRequest{
		TrackId:      env.TimedTrackID.String(),
		AuthorUserId: env.UserID.String(),
		Cues: []*ingestv1.TimedTextCueDto{
			{CueSequence: 1, StartMs: 1200, EndMs: 3500, TextLines: []string{"왕이 죽었소", "역병이 시작되었소"}},
		},
	})
	if err != nil || ttRes.GetCpsViolationCount() != 0 || ttRes.GetShotChangeViolationCount() != 0 {
		t.Fatalf("SaveTimedTextCues validation failed: %v", err)
	}
}

func step3MediaEncodingDispatch(ctx context.Context, t *testing.T, cli mediav1.MediaPipelineServiceClient, env *LiveTestEnvironment, ingestSessionID string) {
	t.Helper()
	job, err := cli.DispatchEncodingJob(ctx, &mediav1.DispatchEncodingJobRequest{
		EpisodeId:       env.EpisodeID.String(),
		IngestSessionId: ingestSessionID,
		MinTargetVmaf:   92.0,
		TargetCodecs:    []string{"H264_AVC", "H265_HEVC", "AV1"},
	})
	if err != nil || job.GetJobId() == "" {
		t.Fatalf("DispatchEncodingJob failed: %v", err)
	}
}

func step4AcquireSessionAndVerifyHttp206(ctx context.Context, t *testing.T, pbCli playbackv1.PlaybackServiceClient, mediaCli mediav1.MediaPipelineServiceClient, env *LiveTestEnvironment) string {
	t.Helper()
	ticket, err := pbCli.AcquirePlaybackSession(ctx, &playbackv1.AcquirePlaybackSessionRequest{
		UserId:      env.UserID.String(),
		ProfileId:   env.ProfileID.String(),
		EpisodeId:   env.EpisodeID.String(),
		DeviceId:    env.DeviceID.String(),
		ClientIp:    "127.0.0.1",
		CountryCode: "KR",
	})
	if err != nil {
		t.Fatalf("AcquirePlaybackSession failed: %v", err)
	}
	edgeSel, err := mediaCli.SelectOptimalEdgeNode(ctx, &mediav1.SelectOptimalEdgeNodeRequest{
		ClientIp:    "127.0.0.1",
		CountryCode: "KR",
		EpisodeId:   env.EpisodeID.String(),
	})
	if err != nil {
		t.Fatalf("SelectOptimalEdgeNode failed: %v", err)
	}
	verifyEdgeHttp206ByteRange(t, edgeSel, env)
	return ticket.GetPlaybackSessionId()
}

func verifyEdgeHttp206ByteRange(t *testing.T, edgeSel *mediav1.SelectOptimalEdgeNodeResponse, env *LiveTestEnvironment) {
	t.Helper()
	segPath := fmt.Sprintf("packaged/%s/H265_HEVC_1080/segment_0.m4s", env.EpisodeID.String())
	streamURL := fmt.Sprintf("%s/stream/%s?token=%s&exp=%d&ep=%s&ip=127.0.0.1",
		edgeSel.GetEdgeBaseUrl(), segPath, edgeSel.GetSignedHmacToken(), edgeSel.GetTokenExpiresAtUnix(), env.EpisodeID.String())

	req, err := http.NewRequest(http.MethodGet, streamURL, nil)
	if err != nil {
		t.Fatalf("new request failed: %v", err)
	}
	req.Header.Set("Range", "bytes=1024-5119")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("edge http request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusPartialContent {
		t.Fatalf("expected HTTP 206 Partial Content, got %d", resp.StatusCode)
	}
	bodyBytes, _ := io.ReadAll(resp.Body)
	expectedSlice := env.RawMasterBytes[1024 : 5119+1]
	if !bytes.Equal(bodyBytes, expectedSlice) {
		t.Fatalf("1-byte precision mismatch on HTTP 206 range response: got %d bytes", len(bodyBytes))
	}
}

func step5HeartbeatAndLoLoMoSync(ctx context.Context, t *testing.T, pbCli playbackv1.PlaybackServiceClient, recoCli recov1.RecommendationServiceClient, env *LiveTestEnvironment, sessionID string) {
	t.Helper()
	hb, err := pbCli.ReportPlaybackHeartbeat(ctx, &playbackv1.ReportPlaybackHeartbeatRequest{
		PlaybackSessionId:   sessionID,
		ProfileId:           env.ProfileID.String(),
		EpisodeId:           env.EpisodeID.String(),
		CurrentPositionMs:   1800000,
		WatchedDeltaMs:      1800000,
		CurrentBitrateKbps:  5800,
		BufferUnderrunDelta: 0,
		DroppedFramesDelta:  0,
		IsPaused:            false,
	})
	if err != nil || !hb.GetKeepAliveAllowed() {
		t.Fatalf("ReportPlaybackHeartbeat failed: %v", err)
	}
	lolomo, err := recoCli.GenerateLoLoMo(ctx, &recov1.GenerateLoLoMoRequest{
		ProfileId:    env.ProfileID.String(),
		CountryCode:  "KR",
		LanguageCode: "ko",
		MaxMaturity:  "AGE_19",
	})
	if err != nil || len(lolomo.GetRows()) == 0 || lolomo.GetRows()[0].GetItems()[0].GetBookmarkMs() != 1800000 {
		t.Fatalf("LoLoMo Continue Watching bookmark mismatch: %v", err)
	}
}

func step6VerifyPartnerRoyaltySettlement(ctx context.Context, t *testing.T, bilCli billingv1.BillingServiceClient, env *LiveTestEnvironment) {
	t.Helper()
	monthStr := time.Now().UTC().Format("2006-01")
	res, err := bilCli.CalculateMonthlyPartnerSettlements(ctx, &billingv1.CalculateSettlementsRequest{
		SettlementMonth: monthStr,
		PartnerOrgId:    env.PartnerOrgID.String(),
	})
	if err != nil || len(res.GetLedgers()) == 0 {
		t.Fatalf("CalculateMonthlyPartnerSettlements failed: %v", err)
	}
	ledger := res.GetLedgers()[0]
	if ledger.GetQualifiedHoursViewed() != 0.5 || ledger.GetGrossPayoutKrw() != 10001250 {
		t.Fatalf("partner settlement calculation mismatch: hours=%v, gross=%d", ledger.GetQualifiedHoursViewed(), ledger.GetGrossPayoutKrw())
	}
}

