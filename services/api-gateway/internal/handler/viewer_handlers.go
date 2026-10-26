package handler

import (
 "encoding/json"
 "net/http"

 "github.com/go-chi/chi/v5"
 catalogv1 "github.com/netflix-clone/proto/catalog/v1"
 iamv1 "github.com/netflix-clone/proto/iam/v1"
 playbackv1 "github.com/netflix-clone/proto/playback/v1"
 "github.com/netflix-clone/services/api-gateway/internal/middleware"
 "google.golang.org/grpc/codes"
 "google.golang.org/grpc/status"
)

func rpcResponse(w http.ResponseWriter, payload any, err error) {
 if err == nil { writeJSON(w, http.StatusOK, payload); return }
 code := http.StatusBadGateway
 switch status.Code(err) {
 case codes.InvalidArgument: code = http.StatusBadRequest
 case codes.NotFound: code = http.StatusNotFound
 case codes.PermissionDenied: code = http.StatusForbidden
 case codes.Unauthenticated: code = http.StatusUnauthorized
 case codes.ResourceExhausted: code = http.StatusTooManyRequests
 }
 writeJSON(w, code, map[string]string{"error": http.StatusText(code)})
}

func (h *BffHandler) HandleProfiles(w http.ResponseWriter, r *http.Request) {
 sess, _ := middleware.ExtractSession(r.Context())
 res, err := h.svc.Iam.ListProfiles(r.Context(), &iamv1.ListProfilesRequest{UserId: sess.UserID})
 rpcResponse(w, res, err)
}

func (h *BffHandler) HandleVerifyPin(w http.ResponseWriter, r *http.Request) {
 sess, _ := middleware.ExtractSession(r.Context())
 profiles, err := h.svc.Iam.ListProfiles(r.Context(), &iamv1.ListProfilesRequest{UserId: sess.UserID})
 if err != nil { rpcResponse(w, nil, err); return }
 id := chi.URLParam(r, "profileId")
 owned := false
 for _, p := range profiles.GetProfiles() { if p.GetId() == id { owned = true; break } }
 if !owned { http.Error(w, "profile not found", http.StatusNotFound); return }
 var req iamv1.VerifyProfilePinRequest
 if json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&req) != nil { http.Error(w, "invalid request", 400); return }
 req.ProfileId = id
 res, err := h.svc.Iam.VerifyProfilePin(r.Context(), &req)
 rpcResponse(w, res, err)
}

func (h *BffHandler) HandleEpisodes(w http.ResponseWriter, r *http.Request) {
 sess, _ := middleware.ExtractSession(r.Context())
 res, err := h.svc.Catalog.ListSeasonEpisodes(r.Context(), &catalogv1.ListSeasonEpisodesRequest{
  SeasonId: chi.URLParam(r, "seasonId"), ProfileId: sess.ProfileID, LanguageCode: "ko",
 })
 rpcResponse(w, res, err)
}

func (h *BffHandler) HandleAcquirePlayback(w http.ResponseWriter, r *http.Request) {
 sess, _ := middleware.ExtractSession(r.Context())
 var req playbackv1.AcquirePlaybackSessionRequest
 if json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&req) != nil { http.Error(w, "invalid request", 400); return }
 if sess.ProfileID == "" { http.Error(w, "select a profile", 403); return }
 req.UserId, req.ProfileId, req.DeviceId = sess.UserID, sess.ProfileID, sess.DeviceID
 req.ClientIp, req.CountryCode = r.RemoteAddr, sess.CountryCode
 res, err := h.svc.Playback.AcquirePlaybackSession(r.Context(), &req)
 rpcResponse(w, res, err)
}
