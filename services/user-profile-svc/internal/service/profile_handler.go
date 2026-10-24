package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	iamv1 "github.com/netflix-clone/proto/iam/v1"
	"github.com/netflix-clone/services/user-profile-svc/internal/db/sqlc"
	"github.com/netflix-clone/services/user-profile-svc/internal/security"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *IamGrpcServer) ListProfiles(ctx context.Context, req *iamv1.ListProfilesRequest) (*iamv1.ListProfilesResponse, error) {
	uid, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}
	rows, err := s.queries.ListProfilesByUser(ctx, uid)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to query profiles")
	}
	result := make([]*iamv1.ProfileDto, 0, len(rows))
	for _, r := range rows {
		result = append(result, mapProfileToDto(r))
	}
	return &iamv1.ListProfilesResponse{Profiles: result}, nil
}

func (s *IamGrpcServer) UpsertProfile(ctx context.Context, req *iamv1.UpsertProfileRequest) (*iamv1.ProfileDto, error) {
	uid, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}
	var pinHash pgtype.Text
	if len(req.GetRawPinCode()) == 4 {
		hashed, err := security.HashSecretBcrypt(req.GetRawPinCode())
		if err != nil {
			return nil, status.Error(codes.Internal, "failed to hash profile pin")
		}
		pinHash = pgtype.Text{String: hashed, Valid: true}
	}
	if req.GetProfileId() == "" {
		return s.createNewProfile(ctx, uid, req, pinHash)
	}
	return s.updateExistingProfile(ctx, uid, req, pinHash)
}

func (s *IamGrpcServer) createNewProfile(ctx context.Context, uid uuid.UUID, req *iamv1.UpsertProfileRequest, pinHash pgtype.Text) (*iamv1.ProfileDto, error) {
	count, err := s.queries.CountProfilesByUser(ctx, uid)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to count profiles")
	}
	if count >= 5 {
		return nil, status.Error(codes.ResourceExhausted, "maximum of 5 profiles allowed per account")
	}
	created, err := s.queries.CreateProfile(ctx, db.CreateProfileParams{
		UserID:                uid,
		DisplayName:           req.GetDisplayName(),
		AvatarKey:             req.GetAvatarKey(),
		IsKids:                req.GetIsKids(),
		MaxMaturity:           db.MaturityLevel(req.GetMaxMaturity()),
		PinBcryptHash:         pinHash,
		PreferredAudioLang:    req.GetPreferredAudioLang(),
		PreferredSubtitleLang: req.GetPreferredSubtitleLang(),
		AutoplayNextEpisode:   req.GetAutoplayNextEpisode(),
		AutoplayPreviews:      req.GetAutoplayPreviews(),
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "create profile error: %v", err)
	}
	return mapProfileToDto(db.ListProfilesByUserRow(created)), nil
}

func (s *IamGrpcServer) updateExistingProfile(ctx context.Context, uid uuid.UUID, req *iamv1.UpsertProfileRequest, pinHash pgtype.Text) (*iamv1.ProfileDto, error) {
	pid, err := uuid.Parse(req.GetProfileId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid profile_id")
	}
	updated, err := s.queries.UpdateProfile(ctx, db.UpdateProfileParams{
		ID:                    pid,
		UserID:                uid,
		DisplayName:           req.GetDisplayName(),
		AvatarKey:             req.GetAvatarKey(),
		IsKids:                req.GetIsKids(),
		MaxMaturity:           db.MaturityLevel(req.GetMaxMaturity()),
		PinBcryptHash:         pinHash,
		PreferredAudioLang:    req.GetPreferredAudioLang(),
		PreferredSubtitleLang: req.GetPreferredSubtitleLang(),
		AutoplayNextEpisode:   req.GetAutoplayNextEpisode(),
		AutoplayPreviews:      req.GetAutoplayPreviews(),
	})
	if err != nil {
		return nil, status.Error(codes.NotFound, "profile not found")
	}
	return mapProfileToDto(db.ListProfilesByUserRow(updated)), nil
}

func (s *IamGrpcServer) VerifyProfilePin(ctx context.Context, req *iamv1.VerifyProfilePinRequest) (*iamv1.VerifyProfilePinResponse, error) {
	pid, err := uuid.Parse(req.GetProfileId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid profile_id")
	}
	prof, err := s.queries.GetProfileByID(ctx, pid)
	if err != nil {
		return nil, status.Error(codes.NotFound, "profile not found")
	}
	if prof.PinBcryptHash.Valid && !security.CompareBcrypt(prof.PinBcryptHash.String, req.GetRawPinCode()) {
		return &iamv1.VerifyProfilePinResponse{Verified: false}, nil
	}
	scopedJwt, _, err := s.tokens.IssueAccessToken(prof.UserID.String(), prof.ID.String(), "", "USER", string(prof.MaxMaturity), "KR")
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to sign profile token")
	}
	return &iamv1.VerifyProfilePinResponse{
		Verified:         true,
		ProfileScopedJwt: scopedJwt,
	}, nil
}

func mapProfileToDto(r db.ListProfilesByUserRow) *iamv1.ProfileDto {
	fontScale, _ := r.SubtitleFontScale.Float64Value()
	bgOpacity, _ := r.SubtitleBgOpacity.Float64Value()
	return &iamv1.ProfileDto{
		Id:                    r.ID.String(),
		UserId:                r.UserID.String(),
		DisplayName:           r.DisplayName,
		AvatarKey:             r.AvatarKey,
		IsKids:                r.IsKids,
		MaxMaturity:           string(r.MaxMaturity),
		HasPin:                r.PinBcryptHash.Valid && r.PinBcryptHash.String != "",
		PreferredAudioLang:    r.PreferredAudioLang,
		PreferredSubtitleLang: r.PreferredSubtitleLang,
		AutoplayNextEpisode:   r.AutoplayNextEpisode,
		AutoplayPreviews:      r.AutoplayPreviews,
		SubtitleFontScale:     fontScale.Float64,
		SubtitleBgOpacity:     bgOpacity.Float64,
		SubtitleEdgeStyle:     r.SubtitleEdgeStyle,
	}
}

