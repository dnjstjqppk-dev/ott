package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	billingv1 "github.com/netflix-clone/proto/billing/v1"
	catalogv1 "github.com/netflix-clone/proto/catalog/v1"
	iamv1 "github.com/netflix-clone/proto/iam/v1"
	recov1 "github.com/netflix-clone/proto/reco/v1"
	"github.com/netflix-clone/services/api-gateway/internal/clients"
	"github.com/netflix-clone/services/api-gateway/internal/middleware"
	"golang.org/x/sync/errgroup"
)

type BffHandler struct {
	svc *clients.ServiceClients
}

func NewBffHandler(svc *clients.ServiceClients) *BffHandler {
	return &BffHandler{svc: svc}
}

func (h *BffHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var req iamv1.AuthenticateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"malformed json"}`, http.StatusBadRequest)
		return
	}
	req.ClientIp = r.RemoteAddr
	res, err := h.svc.Iam.Authenticate(r.Context(), &req)
	if err != nil {
		http.Error(w, `{"error":"authentication failed"}`, http.StatusUnauthorized)
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (h *BffHandler) HandleGetAggregatedHome(w http.ResponseWriter, r *http.Request) {
	sess, ok := middleware.ExtractSession(r.Context())
	if !ok {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}
	lang := r.URL.Query().Get("lang")
	if lang == "" {
		lang = "ko"
	}
	var (
		lolomo *recov1.GenerateLoLoMoResponse
		sub    *billingv1.SubscriptionStateDto
	)
	g, gctx := errgroup.WithContext(r.Context())
	g.Go(func() error {
		var err error
		lolomo, err = h.svc.Reco.GenerateLoLoMo(gctx, &recov1.GenerateLoLoMoRequest{
			ProfileId:    sess.ProfileID,
			CountryCode:  sess.CountryCode,
			LanguageCode: lang,
			MaxMaturity:  sess.MaxMaturity,
		})
		return err
	})
	g.Go(func() error {
		var err error
		sub, err = h.svc.Billing.GetSubscriptionState(gctx, &billingv1.GetSubscriptionStateRequest{
			UserId: sess.UserID,
		})
		return err
	})
	if err := g.Wait(); err != nil {
		http.Error(w, `{"error":"failed to aggregate home feed"}`, http.StatusBadGateway)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"lolomo":       lolomo,
		"subscription": sub,
	})
}

func (h *BffHandler) HandleGetTitleModal(w http.ResponseWriter, r *http.Request) {
	sess, ok := middleware.ExtractSession(r.Context())
	if !ok {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}
	titleID := chi.URLParam(r, "titleId")
	lang := r.URL.Query().Get("lang")
	if lang == "" {
		lang = "ko"
	}
	var (
		detail  *catalogv1.TitleDetailResponse
		license *catalogv1.CheckLicenseEligibilityResponse
	)
	g, gctx := errgroup.WithContext(r.Context())
	g.Go(func() error {
		var err error
		detail, err = h.svc.Catalog.GetTitleDetail(gctx, &catalogv1.GetTitleDetailRequest{
			TitleId:      titleID,
			ProfileId:    sess.ProfileID,
			LanguageCode: lang,
			CountryCode:  sess.CountryCode,
		})
		return err
	})
	g.Go(func() error {
		var err error
		license, err = h.svc.Catalog.CheckLicenseEligibility(gctx, &catalogv1.CheckLicenseEligibilityRequest{
			TitleId:     titleID,
			CountryCode: sess.CountryCode,
			IsAvodPlan:  false,
		})
		return err
	})
	if err := g.Wait(); err != nil {
		http.Error(w, `{"error":"title not found"}`, http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"detail":  detail,
		"license": license,
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

