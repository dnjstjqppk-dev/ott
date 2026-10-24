package service

import (
	"context"
	"net/netip"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	iamv1 "github.com/netflix-clone/proto/iam/v1"
	"github.com/netflix-clone/services/user-profile-svc/internal/db/sqlc"
	"github.com/netflix-clone/services/user-profile-svc/internal/security"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type IamGrpcServer struct {
	iamv1.UnimplementedIamServiceServer
	pool    *pgxpool.Pool
	queries *db.Queries
	tokens  *security.TokenManager
}

func NewIamGrpcServer(pool *pgxpool.Pool, tokens *security.TokenManager) *IamGrpcServer {
	return &IamGrpcServer{
		pool:    pool,
		queries: db.New(pool),
		tokens:  tokens,
	}
}

func (s *IamGrpcServer) Authenticate(ctx context.Context, req *iamv1.AuthenticateRequest) (*iamv1.AuthenticateResponse, error) {
	user, err := s.queries.GetUserByEmail(ctx, req.GetEmail())
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid email or password")
	}
	if user.LockedUntil.Valid && user.LockedUntil.Time.After(time.Now().UTC()) {
		return nil, status.Error(codes.PermissionDenied, "account temporarily locked due to failed attempts")
	}
	if !security.CompareBcrypt(user.PasswordHash, req.GetPassword()) {
		_ = s.queries.IncrementLoginFailure(ctx, user.ID)
		return nil, status.Error(codes.Unauthenticated, "invalid email or password")
	}
	_ = s.queries.ResetLoginFailure(ctx, user.ID)
	return s.issueSessionForUser(ctx, user, req)
}

func (s *IamGrpcServer) issueSessionForUser(ctx context.Context, user db.GetUserByEmailRow, req *iamv1.AuthenticateRequest) (*iamv1.AuthenticateResponse, error) {
	ipAddr, err := netip.ParseAddr(req.GetClientIp())
	if err != nil {
		ipAddr = netip.MustParseAddr("127.0.0.1")
	}
	dev, err := s.queries.UpsertUserDevice(ctx, db.UpsertUserDeviceParams{
		UserID:            user.ID,
		DeviceFingerprint: req.GetDeviceFingerprint(),
		DeviceName:        req.GetDeviceName(),
		Platform:          db.DevicePlatform(req.GetPlatform()),
		DrmSecurityLevel:  req.GetDrmSecurityLevel(),
		LastIp:            ipAddr,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "device upsert failed: %v", err)
	}
	rawRefresh, refreshHash, err := security.GenerateRefreshTokenPair()
	if err != nil {
		return nil, status.Error(codes.Internal, "entropy error")
	}
	_, err = s.queries.InsertRefreshToken(ctx, db.InsertRefreshTokenParams{
		UserID:      user.ID,
		DeviceID:    dev.ID,
		TokenSha256: refreshHash,
		ExpiresAt:   time.Now().UTC().Add(30 * 24 * time.Hour),
	})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to persist refresh token")
	}
	access, ttl, err := s.tokens.IssueAccessToken(user.ID.String(), "", dev.ID.String(), string(user.Role), "AGE_19", user.CountryCode)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to sign jwt")
	}
	return &iamv1.AuthenticateResponse{
		AccessToken:   access,
		RefreshToken:  rawRefresh,
		ExpiresInSec:  ttl,
		UserId:        user.ID.String(),
		DeviceId:      dev.ID.String(),
		Role:          string(user.Role),
		AccountStatus: string(user.Status),
	}, nil
}

func (s *IamGrpcServer) RefreshSession(ctx context.Context, req *iamv1.RefreshSessionRequest) (*iamv1.AuthenticateResponse, error) {
	incomingHash := security.HashSHA256(req.GetRefreshToken())
	stored, err := s.queries.GetActiveRefreshTokenByHash(ctx, incomingHash)
	if err != nil || stored.IsRevoked || stored.DeviceFingerprint != req.GetDeviceFingerprint() {
		return nil, status.Error(codes.Unauthenticated, "refresh token invalid or device mismatch")
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "tx start failed")
	}
	defer func() { _ = tx.Rollback(ctx) }()
	qtx := s.queries.WithTx(tx)

	rawNew, hashNew, err := security.GenerateRefreshTokenPair()
	if err != nil {
		return nil, status.Error(codes.Internal, "token generation failed")
	}
	newRow, err := qtx.InsertRefreshToken(ctx, db.InsertRefreshTokenParams{
		UserID:      stored.UserID,
		DeviceID:    stored.DeviceID,
		TokenSha256: hashNew,
		ExpiresAt:   time.Now().UTC().Add(30 * 24 * time.Hour),
	})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to insert rotated token")
	}
	if err := qtx.RevokeRefreshToken(ctx, db.RevokeRefreshTokenParams{ID: stored.ID, ReplacedByTokenID: newRow.ID}); err != nil {
		return nil, status.Error(codes.Internal, "failed to revoke old token")
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, status.Error(codes.Internal, "tx commit failed")
	}
	access, ttl, err := s.tokens.IssueAccessToken(stored.UserID.String(), "", stored.DeviceID.String(), string(stored.Role), "AGE_19", stored.CountryCode)
	if err != nil {
		return nil, status.Error(codes.Internal, "jwt sign failed")
	}
	return &iamv1.AuthenticateResponse{
		AccessToken:   access,
		RefreshToken:  rawNew,
		ExpiresInSec:  ttl,
		UserId:        stored.UserID.String(),
		DeviceId:      stored.DeviceID.String(),
		Role:          string(stored.Role),
		AccountStatus: string(stored.Status),
	}, nil
}

func (s *IamGrpcServer) ValidateAccessToken(ctx context.Context, req *iamv1.ValidateAccessTokenRequest) (*iamv1.ValidateAccessTokenResponse, error) {
	claims, err := s.tokens.VerifyAccessToken(req.GetAccessToken())
	if err != nil {
		return &iamv1.ValidateAccessTokenResponse{IsValid: false}, nil
	}
	return &iamv1.ValidateAccessTokenResponse{
		IsValid:     true,
		UserId:      claims.UserID,
		ProfileId:   claims.ProfileID,
		DeviceId:    claims.DeviceID,
		Role:        claims.Role,
		MaxMaturity: claims.MaxMaturity,
		CountryCode: claims.CountryCode,
	}, nil
}

func (s *IamGrpcServer) RevokeDeviceSession(ctx context.Context, req *iamv1.RevokeDeviceSessionRequest) (*iamv1.RevokeDeviceSessionResponse, error) {
	uid, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}
	did, err := uuid.Parse(req.GetTargetDeviceId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid target_device_id")
	}
	_ = s.queries.RevokeDeviceAndTokens(ctx, db.RevokeDeviceAndTokensParams{UserID: uid, ID: did})
	_ = s.queries.RevokeAllTokensForDevice(ctx, db.RevokeAllTokensForDeviceParams{UserID: uid, DeviceID: did})
	return &iamv1.RevokeDeviceSessionResponse{Revoked: true}, nil
}

