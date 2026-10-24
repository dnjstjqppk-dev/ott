package service

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	ingestv1 "github.com/netflix-clone/proto/ingest/v1"
	"github.com/netflix-clone/services/studio-ingest-svc/internal/db/sqlc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *StudioIngestGrpcServer) GetQcReport(ctx context.Context, req *ingestv1.GetQcReportRequest) (*ingestv1.QcReportDto, error) {
	epID, err := uuid.Parse(req.GetEpisodeId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid episode_id")
	}
	rep, err := s.queries.GetLatestQcReportByEpisode(ctx, epID)
	if err != nil {
		return nil, status.Error(codes.NotFound, "qc report not found")
	}
	return s.hydrateQcReportDto(ctx, rep)
}

func (s *StudioIngestGrpcServer) SubmitManualQcDecision(ctx context.Context, req *ingestv1.SubmitManualQcDecisionRequest) (*ingestv1.QcReportDto, error) {
	repID, err := uuid.Parse(req.GetQcReportId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid qc_report_id")
	}
	revID, err := uuid.Parse(req.GetReviewerUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid reviewer_user_id")
	}
	rep, err := s.queries.UpdateQcReportDecision(ctx, db.UpdateQcReportDecisionParams{
		ID:          repID,
		Verdict:     db.QcVerdict(req.GetVerdict()),
		ReviewedBy:  revID,
		ReviewNotes: pgtype.Text{String: req.GetReviewNotes(), Valid: true},
	})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to update qc decision")
	}
	return s.hydrateQcReportDto(ctx, db.GetLatestQcReportByEpisodeRow(rep))
}

func (s *StudioIngestGrpcServer) hydrateQcReportDto(ctx context.Context, rep db.GetLatestQcReportByEpisodeRow) (*ingestv1.QcReportDto, error) {
	issues, err := s.queries.ListQcIssuesByReport(ctx, rep.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list qc issues")
	}
	issueDtos := make([]*ingestv1.QcIssueDto, 0, len(issues))
	for _, iss := range issues {
		issueDtos = append(issueDtos, &ingestv1.QcIssueDto{
			Id:              iss.ID.String(),
			IssueType:       iss.IssueType,
			Severity:        string(iss.Severity),
			StartMs:         iss.StartMs,
			EndMs:           iss.EndMs,
			StartFrame:      iss.StartFrame,
			EndFrame:        iss.EndFrame,
			MeasuredValue:   iss.MeasuredValue,
			ThresholdValue:  iss.ThresholdValue,
			ReviewerComment: iss.ReviewerComment,
			IsResolved:      iss.IsResolved,
		})
	}
	lufs, _ := rep.IntegratedLufs.Float64Value()
	peak, _ := rep.TruePeakDbtp.Float64Value()
	return &ingestv1.QcReportDto{
		Id:              rep.ID.String(),
		IngestSessionId: rep.IngestSessionID.String(),
		EpisodeId:       rep.EpisodeID.String(),
		Verdict:         string(rep.Verdict),
		DetectedWidth:   rep.DetectedWidth,
		DetectedHeight:  rep.DetectedHeight,
		DetectedFpsNum:  rep.DetectedFpsNum,
		DetectedFpsDen:  rep.DetectedFpsDen,
		ColorSpace:      rep.ColorSpace,
		HdrFormat:       rep.HdrFormat,
		AudioChannels:   rep.AudioChannels,
		IntegratedLufs:  lufs.Float64,
		TruePeakDbtp:    peak.Float64,
		Issues:          issueDtos,
	}, nil
}

func (s *StudioIngestGrpcServer) SaveTimedTextCues(ctx context.Context, req *ingestv1.SaveTimedTextCuesRequest) (*ingestv1.SaveTimedTextCuesResponse, error) {
	trackID, err := uuid.Parse(req.GetTrackId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid track_id")
	}
	authorID, err := uuid.Parse(req.GetAuthorUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid author_user_id")
	}
	if err := s.queries.DeleteTimedTextCuesByTrack(ctx, trackID); err != nil {
		return nil, status.Error(codes.Internal, "failed to clear old cues")
	}
	cpsViolations, shotViolations, vttBuf := s.persistAndValidateCues(ctx, trackID, req.GetCues())
	vttKey := fmt.Sprintf("subtitles/%s/track.vtt", trackID.String())
	if err := s.store.PutWebVTT(ctx, "nf-media-packaged", vttKey, vttBuf.Bytes()); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to upload webvtt: %v", err)
	}
	verdict := db.QcVerdictPASSED
	if cpsViolations > 0 || shotViolations > 0 {
		verdict = db.QcVerdictWARNING
	}
	_ = s.queries.UpdateTimedTextTrackStorage(ctx, db.UpdateTimedTextTrackStorageParams{
		ID:               trackID,
		AuthorUserID:     authorID,
		WebvttStorageKey: pgtype.Text{String: vttKey, Valid: true},
		Verdict:          verdict,
	})
	return &ingestv1.SaveTimedTextCuesResponse{
		TotalCues:                int32(len(req.GetCues())),
		CpsViolationCount:        cpsViolations,
		ShotChangeViolationCount: shotViolations,
		GeneratedWebvttUrl:       fmt.Sprintf("%s/%s", s.publicURL, vttKey),
	}, nil
}

func (s *StudioIngestGrpcServer) persistAndValidateCues(ctx context.Context, trackID uuid.UUID, cues []*ingestv1.TimedTextCueDto) (int32, int32, *bytes.Buffer) {
	var cpsViolations, shotViolations int32
	vtt := bytes.NewBufferString("WEBVTT\n\n")
	for _, c := range cues {
		durSec := math.Max(float64(c.GetEndMs()-c.GetStartMs())/1000.0, 0.001)
		totalChars, maxLine := countCueRunes(c.GetTextLines())
		cps := math.Round((float64(totalChars)/durSec)*100) / 100
		violatesCPS := cps > 12.0 || c.GetEndMs()-c.GetStartMs() < 833
		violatesShot := (c.GetStartMs()%1000) < 83 && c.GetStartMs() > 0
		if violatesCPS {
			cpsViolations++
		}
		if violatesShot {
			shotViolations++
		}
		var cpsNum pgtype.Numeric
		_ = cpsNum.Scan(fmt.Sprintf("%.2f", cps))
		_ = s.queries.InsertTimedTextCue(ctx, db.InsertTimedTextCueParams{
			TrackID:            trackID,
			CueSequence:        c.GetCueSequence(),
			StartMs:            c.GetStartMs(),
			EndMs:              c.GetEndMs(),
			TextLines:          c.GetTextLines(),
			ComputedCps:        cpsNum,
			MaxLineLength:      int16(maxLine),
			ViolatesShotChange: violatesShot,
			ViolatesCpsLimit:   violatesCPS,
		})
		writeVttCueBlock(vtt, c)
	}
	return cpsViolations, shotViolations, vtt
}

func countCueRunes(lines []string) (int, int) {
	total, maxLine := 0, 0
	for _, line := range lines {
		rc := utf8.RuneCountInString(line)
		total += rc
		if rc > maxLine {
			maxLine = rc
		}
	}
	return total, maxLine
}

func writeVttCueBlock(buf *bytes.Buffer, c *ingestv1.TimedTextCueDto) {
	fmt.Fprintf(buf, "%d\n%s --> %s\n", c.GetCueSequence(), formatVttTime(c.GetStartMs()), formatVttTime(c.GetEndMs()))
	for _, line := range c.GetTextLines() {
		buf.WriteString(line + "\n")
	}
	buf.WriteString("\n")
}

func formatVttTime(ms int64) string {
	h := ms / 3600000
	m := (ms % 3600000) / 60000
	s := (ms % 60000) / 1000
	rem := ms % 1000
	return fmt.Sprintf("%02d:%02d:%02d.%03d", h, m, s, rem)
}

