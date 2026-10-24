package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	catalogv1 "github.com/netflix-clone/proto/catalog/v1"
	"github.com/netflix-clone/services/catalog-svc/internal/db/sqlc"
	"github.com/netflix-clone/services/catalog-svc/internal/search"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *CatalogGrpcServer) CheckLicenseEligibility(ctx context.Context, req *catalogv1.CheckLicenseEligibilityRequest) (*catalogv1.CheckLicenseEligibilityResponse, error) {
	tid, err := uuid.Parse(req.GetTitleId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid title_id")
	}
	win, err := s.queries.CheckActiveLicenseWindow(ctx, db.CheckActiveLicenseWindowParams{
		TitleID:     tid,
		CountryCode: req.GetCountryCode(),
	})
	if err != nil {
		return &catalogv1.CheckLicenseEligibilityResponse{IsPlayable: false}, nil
	}
	if req.GetIsAvodPlan() && !win.AllowAvodTier {
		return &catalogv1.CheckLicenseEligibilityResponse{IsPlayable: false}, nil
	}
	return &catalogv1.CheckLicenseEligibilityResponse{
		IsPlayable:       true,
		IsDownloadable:   win.AllowDownload && !req.GetIsAvodPlan(),
		LicenseExpiresAt: win.EndsAt.Format(time.RFC3339),
	}, nil
}

func (s *CatalogGrpcServer) UpsertTitleMetadata(ctx context.Context, req *catalogv1.UpsertTitleMetadataRequest) (*catalogv1.TitleDto, error) {
	partnerID, err := uuid.Parse(req.GetPartnerOrgId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid partner_org_id")
	}
	locTitlesJSON, _ := json.Marshal(req.GetLocalizedTitles())
	synopsisJSON, _ := json.Marshal(req.GetSynopsis())
	row, err := s.queries.UpsertTitleRecord(ctx, db.UpsertTitleRecordParams{
		PartnerOrgID:      partnerID,
		Slug:              req.GetSlug(),
		Type:              db.TitleType(req.GetType()),
		OriginalTitle:     req.GetOriginalTitle(),
		LocalizedTitles:   locTitlesJSON,
		Synopsis:          synopsisJSON,
		ReleaseYear:       int16(req.GetReleaseYear()),
		MaturityRating:    db.MaturityLevel(req.GetMaturityRating()),
		MaturityReasons:   req.GetMaturityReasons(),
		IsNetflixOriginal: req.GetIsNetflixOriginal(),
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to upsert title: %v", err)
	}
	_ = s.meili.IndexTitleDocument(ctx, search.TitleSearchDocument{
		ID:                row.ID.String(),
		Slug:              row.Slug,
		Type:              string(row.Type),
		OriginalTitle:     row.OriginalTitle,
		LocalizedTitles:   req.GetLocalizedTitles(),
		Synopsis:          req.GetSynopsis(),
		ReleaseYear:       int32(row.ReleaseYear),
		MaturityRating:    string(row.MaturityRating),
		IsNetflixOriginal: row.IsNetflixOriginal,
	})
	return &catalogv1.TitleDto{
		Id:             row.ID.String(),
		Slug:           row.Slug,
		Type:           string(row.Type),
		LocalizedTitle: row.OriginalTitle,
		ReleaseYear:    int32(row.ReleaseYear),
		MaturityRating: string(row.MaturityRating),
	}, nil
}
