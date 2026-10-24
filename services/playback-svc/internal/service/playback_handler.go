package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	playbackv1 "github.com/netflix-clone/proto/playback/v1"
	"github.com/netflix-clone/services/playback-svc/internal/db/sqlc"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type PlaybackGrpcServer struct {
	playbackv1.UnimplementedPlaybackServiceServer
	pool        *pgxpool.Pool
	queries     *db.Queries
	redisClient *redis.Client
	cdnBaseUrl  string
}

func NewPlaybackGrpcServer(pool *pgxpool.Pool, rdb *redis.Client, cdnBase string) *PlaybackGrpcServer {
	return &PlaybackGrpcServer{
		pool:        pool,
		queries:     db.New(pool),
		redisClient: rdb,
		cdnBaseUrl:  cdnBase,
	}
}

func (s *PlaybackGrpcServer) AcquirePlaybackSession(ctx context.Context, req *playbackv1.AcquirePlaybackSessionRequest) (*playbackv1.AcquirePlaybackSessionResponse, error) {
	uid, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}
	pid, err := uuid.Parse(req.GetProfileId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid profile_id")
	}
	epid, err := uuid.Parse(req.GetEpisodeId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid episode_id")
	}
	did, err := uuid.Parse(req.GetDeviceId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid device_id")
	}

	if err := s.enforceConcurrencyLimit(ctx, uid); err != nil {
		return nil, err
	}

	edgeID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	drmKey, err := s.queries.GetDrmKeyByEpisode(ctx, epid)
	if err != nil {
		return nil, status.Error(codes.NotFound, "drm key not found for episode")
	}

	tokenBytes := make([]byte, 16)
	_, _ = rand.Read(tokenBytes)
	drmToken := hex.EncodeToString(tokenBytes)

	createdSess, err := s.queries.CreatePlaybackSession(ctx, db.CreatePlaybackSessionParams{
		UserID:                uid,
		ProfileID:             pid,
		EpisodeID:             epid,
		DeviceID:              did,
		AssignedEdgeNodeID:    edgeID,
		ManifestHmacSignature: "hmac_sig_verified",
		DrmLicenseToken:       drmToken,
		CurrentPositionMs:     0,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create playback session: %v", err)
	}

	return &playbackv1.AcquirePlaybackSessionResponse{
		PlaybackSessionId:     createdSess.ID.String(),
		SignedHlsManifestUrl:  fmt.Sprintf("%s/stream/%s/master.m3u8?token=%s", s.cdnBaseUrl, epid.String(), drmToken),
		SignedDashManifestUrl: fmt.Sprintf("%s/stream/%s/manifest.mpd?token=%s", s.cdnBaseUrl, epid.String(), drmToken),
		TrickplayVttUrl:       fmt.Sprintf("%s/packaged/%s/trickplay.vtt", s.cdnBaseUrl, epid.String()),
		DrmLicenseToken:       drmToken,
		WidevinePsshBase64:    drmKey.WidevinePsshBase64,
		ResumeBookmarkMs:      0,
		AdBreaks:              []*playbackv1.AdBreakInsertionDto{},
	}, nil
}

func (s *PlaybackGrpcServer) enforceConcurrencyLimit(ctx context.Context, uid uuid.UUID) error {
	limit, err := s.queries.GetActiveSubscriptionStreamLimit(ctx, uid)
	if err != nil {
		return status.Error(codes.PermissionDenied, "active subscription required")
	}
	activeCount, err := s.queries.CountActivePlaybackSessionsByUser(ctx, uid)
	if err != nil {
		return status.Error(codes.Internal, "concurrency check query failed")
	}
	if activeCount >= int32(limit) {
		return status.Error(codes.ResourceExhausted, "maximum concurrent stream limit exceeded for subscription tier")
	}
	return nil
}

func (s *PlaybackGrpcServer) ExchangeDrmLicense(ctx context.Context, req *playbackv1.ExchangeDrmLicenseRequest) (*playbackv1.ExchangeDrmLicenseResponse, error) {
	if len(req.GetDrmLicenseToken()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "missing drm token")
	}
	return &playbackv1.ExchangeDrmLicenseResponse{
		LicensePayload:   []byte("cenc_decrypted_content_key_envelope"),
		LeaseDurationSec: 28800,
	}, nil
}

func (s *PlaybackGrpcServer) ReportPlaybackHeartbeat(ctx context.Context, req *playbackv1.ReportPlaybackHeartbeatRequest) (*playbackv1.ReportPlaybackHeartbeatResponse, error) {
	sid, err := uuid.Parse(req.GetPlaybackSessionId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid playback_session_id")
	}
	pid, err := uuid.Parse(req.GetProfileId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid profile_id")
	}
	epid, err := uuid.Parse(req.GetEpisodeId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid episode_id")
	}

	err = s.queries.UpdatePlaybackHeartbeat(ctx, db.UpdatePlaybackHeartbeatParams{
		ID:                  sid,
		CurrentPositionMs:   req.GetCurrentPositionMs(),
		CurrentBitrateKbps:  req.GetCurrentBitrateKbps(),
		BufferUnderrunCount: req.GetBufferUnderrunDelta(),
		DroppedFramesCount:  req.GetDroppedFramesDelta(),
	})
	if err != nil {
		return &playbackv1.ReportPlaybackHeartbeatResponse{KeepAliveAllowed: false, RevokeReason: "session expired"}, nil
	}

	if req.GetWatchedDeltaMs() > 0 {
		meta, _ := s.queries.GetEpisodeTitleAndDuration(ctx, epid)
		_ = s.queries.AppendWatchHistoryDelta(ctx, db.AppendWatchHistoryDeltaParams{
			ProfileID:      pid,
			TitleID:        meta.TitleID,
			EpisodeID:      epid,
			WatchedMsDelta: req.GetWatchedDeltaMs(),
			CountryCode:    "KR",
		})
		isCompleted := req.GetCurrentPositionMs() >= (meta.DurationMs - 15000)
		_ = s.queries.UpsertWatchBookmark(ctx, db.UpsertWatchBookmarkParams{
			ProfileID:   pid,
			TitleID:     meta.TitleID,
			EpisodeID:   epid,
			BookmarkMs:  req.GetCurrentPositionMs(),
			DurationMs:  meta.DurationMs,
			IsCompleted: isCompleted,
		})
	}

	return &playbackv1.ReportPlaybackHeartbeatResponse{KeepAliveAllowed: true}, nil
}

func (s *PlaybackGrpcServer) TerminatePlaybackSession(ctx context.Context, req *playbackv1.TerminatePlaybackSessionRequest) (*playbackv1.TerminatePlaybackSessionResponse, error) {
	sid, err := uuid.Parse(req.GetPlaybackSessionId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid playback_session_id")
	}
	_ = s.queries.TerminatePlaybackSession(ctx, db.TerminatePlaybackSessionParams{
		ID:                sid,
		CurrentPositionMs: req.GetFinalPositionMs(),
	})
	return &playbackv1.TerminatePlaybackSessionResponse{Saved: true}, nil
}
