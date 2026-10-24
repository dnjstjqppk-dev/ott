package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	catalogv1 "github.com/netflix-clone/proto/catalog/v1"
	"github.com/netflix-clone/services/catalog-svc/internal/db/sqlc"
	"github.com/netflix-clone/services/catalog-svc/internal/search"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CatalogGrpcServer struct {
	catalogv1.UnimplementedCatalogServiceServer
	queries *db.Queries
	meili   *search.MeiliClient
	cdnBase string
}

func NewCatalogGrpcServer(pool *pgxpool.Pool, meili *search.MeiliClient, cdnBase string) *CatalogGrpcServer {
	return &CatalogGrpcServer{
		queries: db.New(pool),
		meili:   meili,
		cdnBase: cdnBase,
	}
}

func (s *CatalogGrpcServer) GetTitleDetail(ctx context.Context, req *catalogv1.GetTitleDetailRequest) (*catalogv1.TitleDetailResponse, error) {
	tid, err := uuid.Parse(req.GetTitleId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid title_id")
	}
	pid, _ := uuid.Parse(req.GetProfileId())
	tRow, err := s.queries.GetTitleByID(ctx, db.GetTitleByIDParams{
		Lang:      req.GetLanguageCode(),
		ProfileID: pid,
		TitleID:   tid,
	})
	if err != nil {
		return nil, status.Error(codes.NotFound, "title not found")
	}
	resp := &catalogv1.TitleDetailResponse{
		Title: &catalogv1.TitleDto{
			Id:                  tRow.ID.String(),
			Slug:                tRow.Slug,
			Type:                string(tRow.Type),
			LocalizedTitle:      tRow.DisplayTitle,
			LocalizedSynopsis:   tRow.DisplaySynopsis,
			ReleaseYear:         int32(tRow.ReleaseYear),
			MaturityRating:      string(tRow.MaturityRating),
			MaturityReasons:     tRow.MaturityReasons,
			IsNetflixOriginal:   tRow.IsNetflixOriginal,
			MaxResolution:       tRow.MaxResolution,
			SupportsHdr10:       tRow.SupportsHdr10,
			SupportsDolbyVision: tRow.SupportsDolbyVision,
			SupportsDolbyAtmos:  tRow.SupportsDolbyAtmos,
			MatchScorePct:       96,
			InMyList:            tRow.InMyList,
			UserRating:          tRow.UserRating,
		},
	}
	if err := s.hydrateTitleSubResources(ctx, tid, pid, req.GetLanguageCode(), resp); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to load subresources: %v", err)
	}
	return resp, nil
}

func (s *CatalogGrpcServer) hydrateTitleSubResources(ctx context.Context, tid, pid uuid.UUID, lang string, resp *catalogv1.TitleDetailResponse) error {
	g, gctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		arts, err := s.queries.GetActiveArtworksForTitle(gctx, db.GetActiveArtworksForTitleParams{TitleID: tid, LanguageCode: lang})
		if err != nil {
			return err
		}
		for _, a := range arts {
			switch a.Kind {
			case db.ArtworkTypeBILLBOARDHERO:
				resp.Title.HeroArtworkUrl = s.cdnBase + "/" + a.StorageKey
			case db.ArtworkTypeTITLELOGOPNG:
				resp.Title.LogoPngUrl = s.cdnBase + "/" + a.StorageKey
			case db.ArtworkTypeBOXARTHORIZONTAL:
				resp.Title.BoxartUrl = s.cdnBase + "/" + a.StorageKey
			}
		}
		return nil
	})
	g.Go(func() error {
		seasons, err := s.queries.ListSeasonsByTitle(gctx, db.ListSeasonsByTitleParams{TitleID: tid, Column2: lang})
		if err != nil {
			return err
		}
		for _, sn := range seasons {
			resp.Seasons = append(resp.Seasons, &catalogv1.SeasonDto{
				Id:            sn.ID.String(),
				SeasonNumber:  int32(sn.SeasonNumber),
				LocalizedName: sn.DisplayName,
				EpisodeCount:  sn.EpisodeCount,
			})
		}
		return nil
	})
	g.Go(func() error {
		credits, err := s.queries.ListCreditsByTitle(gctx, tid)
		if err != nil {
			return err
		}
		for _, c := range credits {
			resp.Credits = append(resp.Credits, &catalogv1.CreditPersonDto{
				PersonId:      c.PersonID.String(),
				FullName:      c.FullName,
				CreditRole:    c.CreditRole,
				CharacterName: c.CharacterName,
			})
		}
		return nil
	})
	return g.Wait()
}

func (s *CatalogGrpcServer) ListSeasonEpisodes(ctx context.Context, req *catalogv1.ListSeasonEpisodesRequest) (*catalogv1.ListSeasonEpisodesResponse, error) {
	sid, err := uuid.Parse(req.GetSeasonId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid season_id")
	}
	pid, _ := uuid.Parse(req.GetProfileId())
	eps, err := s.queries.ListEpisodesBySeason(ctx, db.ListEpisodesBySeasonParams{
		Lang:      req.GetLanguageCode(),
		ProfileID: pid,
		SeasonID:  sid,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list episodes")
	}
	out := make([]*catalogv1.EpisodeDto, 0, len(eps))
	for _, ep := range eps {
		markers, _ := s.queries.ListTimecodeMarkersByEpisode(ctx, ep.ID)
		mDtos := make([]*catalogv1.TimecodeMarkerDto, 0, len(markers))
		for _, m := range markers {
			mDtos = append(mDtos, &catalogv1.TimecodeMarkerDto{
				MarkerKind: string(m.MarkerKind),
				StartMs:    m.StartMs,
				EndMs:      m.EndMs,
				StartFrame: m.StartFrame,
				EndFrame:   m.EndFrame,
			})
		}
		out = append(out, &catalogv1.EpisodeDto{
			Id:                ep.ID.String(),
			TitleId:           ep.TitleID.String(),
			SeasonId:          ep.SeasonID.String(),
			EpisodeNumber:     int32(ep.EpisodeNumber),
			LocalizedTitle:    ep.DisplayTitle,
			LocalizedSynopsis: ep.DisplaySynopsis,
			DurationMs:        ep.DurationMs,
			StillImageUrl:     s.cdnBase + "/" + ep.StillImageKey,
			UserBookmarkMs:    ep.UserBookmarkMs,
			Markers:           mDtos,
		})
	}
	return &catalogv1.ListSeasonEpisodesResponse{Episodes: out}, nil
}

func (s *CatalogGrpcServer) SearchTitles(ctx context.Context, req *catalogv1.SearchTitlesRequest) (*catalogv1.SearchTitlesResponse, error) {
	hits, total, err := s.meili.Search(ctx, req.GetQuery(), req.GetLimit(), req.GetOffset())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "search engine failure: %v", err)
	}
	results := make([]*catalogv1.TitleDto, 0, len(hits))
	for _, h := range hits {
		titleText := h.LocalizedTitles[req.GetLanguageCode()]
		if titleText == "" {
			titleText = h.OriginalTitle
		}
		results = append(results, &catalogv1.TitleDto{
			Id:                h.ID,
			Slug:              h.Slug,
			Type:              h.Type,
			LocalizedTitle:    titleText,
			LocalizedSynopsis: h.Synopsis[req.GetLanguageCode()],
			ReleaseYear:       h.ReleaseYear,
			MaturityRating:    h.MaturityRating,
			IsNetflixOriginal: h.IsNetflixOriginal,
		})
	}
	return &catalogv1.SearchTitlesResponse{Results: results, TotalHits: total}, nil
}
