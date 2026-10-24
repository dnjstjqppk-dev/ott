package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"net/netip"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	billingv1 "github.com/netflix-clone/proto/billing/v1"
	"github.com/netflix-clone/services/billing-svc/internal/db/sqlc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type BillingGrpcServer struct {
	billingv1.UnimplementedBillingServiceServer
	queries *db.Queries
}

func NewBillingGrpcServer(pool *pgxpool.Pool) *BillingGrpcServer {
	return &BillingGrpcServer{queries: db.New(pool)}
}

func (s *BillingGrpcServer) GetSubscriptionState(ctx context.Context, req *billingv1.GetSubscriptionStateRequest) (*billingv1.SubscriptionStateDto, error) {
	uid, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}
	sub, err := s.queries.GetUserSubscriptionWithPlan(ctx, uid)
	if err != nil {
		return nil, status.Error(codes.NotFound, "active subscription not found")
	}
	return &billingv1.SubscriptionStateDto{
		SubscriptionId:       sub.SubscriptionID.String(),
		PlanCode:             string(sub.PlanCode),
		MaxConcurrentStreams: int32(sub.MaxConcurrentStreams),
		MaxDownloadDevices:   int32(sub.MaxDownloadDevices),
		MaxResolution:        sub.MaxResolution,
		HasAds:               sub.HasAds,
		CurrentPeriodEnd:     sub.CurrentPeriodEnd.Format(time.RFC3339),
	}, nil
}

func (s *BillingGrpcServer) ChangeSubscriptionPlan(ctx context.Context, req *billingv1.ChangeSubscriptionPlanRequest) (*billingv1.SubscriptionStateDto, error) {
	uid, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}
	subID, err := s.queries.UpsertUserSubscriptionPlan(ctx, db.UpsertUserSubscriptionPlanParams{
		UserID:   uid,
		PlanCode: db.PlanTier(req.GetNewPlanCode()),
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to change plan: %v", err)
	}
	s.recordAuditLog(ctx, uid, "SUBSCRIPTION_PLAN_CHANGED", "SUBSCRIPTION", subID, req.GetNewPlanCode())
	return s.GetSubscriptionState(ctx, &billingv1.GetSubscriptionStateRequest{UserId: req.GetUserId()})
}

func (s *BillingGrpcServer) CalculateMonthlyPartnerSettlements(ctx context.Context, req *billingv1.CalculateSettlementsRequest) (*billingv1.CalculateSettlementsResponse, error) {
	orgID, err := uuid.Parse(req.GetPartnerOrgId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid partner_org_id")
	}
	contracts, err := s.queries.ListPartnerContractsForOrg(ctx, orgID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to fetch partner contracts")
	}
	ledgers := make([]*billingv1.PartnerSettlementItemDto, 0, len(contracts))
	for _, c := range contracts {
		item, err := s.settleSingleContract(ctx, c, req.GetSettlementMonth())
		if err != nil {
			return nil, err
		}
		ledgers = append(ledgers, item)
	}
	return &billingv1.CalculateSettlementsResponse{Ledgers: ledgers}, nil
}

func (s *BillingGrpcServer) settleSingleContract(ctx context.Context, c db.ListPartnerContractsForOrgRow, month string) (*billingv1.PartnerSettlementItemDto, error) {
	watchedMs, err := s.queries.SumWatchedMillisecondsForTitleMonth(ctx, db.SumWatchedMillisecondsForTitleMonthParams{
		TitleID: c.TitleID,
		Column2: month,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to sum watch history ledger")
	}
	hours := math.Round((float64(watchedMs)/3600000.0)*100) / 100
	gross := computeGrossPayoutKrw(c, hours)
	tax := (gross * 33) / 1000
	net := gross - tax

	var qHoursNum, compNum pgtype.Numeric
	_ = qHoursNum.Scan(fmt.Sprintf("%.2f", hours))
	_ = compNum.Scan("84.50")

	row, err := s.queries.UpsertPartnerSettlementLedger(ctx, db.UpsertPartnerSettlementLedgerParams{
		ContractID:           c.ID,
		PartnerOrgID:         c.PartnerOrgID,
		TitleID:              c.TitleID,
		SettlementMonth:      month,
		TotalWatchedMs:       watchedMs,
		QualifiedHoursViewed: qHoursNum,
		CompletionRatePct:    compNum,
		GrossPayoutKrw:       gross,
		WithholdingTaxKrw:    tax,
		NetPayoutKrw:         net,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "settlement upsert failed: %v", err)
	}
	return &billingv1.PartnerSettlementItemDto{
		LedgerId:             row.ID.String(),
		TitleId:              row.TitleID.String(),
		QualifiedHoursViewed: hours,
		GrossPayoutKrw:       row.GrossPayoutKrw,
		NetPayoutKrw:         row.NetPayoutKrw,
		Status:               row.Status,
	}, nil
}

func computeGrossPayoutKrw(c db.ListPartnerContractsForOrgRow, hours float64) int64 {
	variablePayout := int64(hours * float64(c.RatePerQualifiedHourKrw))
	switch c.Model {
	case db.ContractModelORIGINALBUYOUT:
		return c.MinimumGuaranteeKrw
	case db.ContractModelMGPLUSREVENUESHARE:
		share := (variablePayout * int64(c.RevenueShareBps)) / 10000
		return c.MinimumGuaranteeKrw + share
	default:
		return variablePayout
	}
}

func (s *BillingGrpcServer) recordAuditLog(ctx context.Context, actorID uuid.UUID, action, entityType string, entityID uuid.UUID, afterVal string) {
	afterBytes, _ := json.Marshal(map[string]string{"state": afterVal})
	hash := sha256.Sum256(afterBytes)
	_ = s.queries.InsertImmutableAuditLog(ctx, db.InsertImmutableAuditLogParams{
		ActorUserID:      actorID,
		ActorRole:        db.AccountRoleUSER,
		ActorIp:          netip.MustParseAddr("127.0.0.1"),
		ActionCode:       action,
		TargetEntityType: entityType,
		TargetEntityID:   entityID,
		BeforeState:      []byte(`{}`),
		AfterState:       afterBytes,
		RecordSha256:     hex.EncodeToString(hash[:]),
	})
}

