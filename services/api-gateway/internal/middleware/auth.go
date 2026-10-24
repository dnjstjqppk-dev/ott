package middleware

import (
	"context"
	"net/http"
	"strings"

	iamv1 "github.com/netflix-clone/proto/iam/v1"
)

type ctxKey string

const AuthSessionKey ctxKey = "nf_auth_session"

type SessionContext struct {
	UserID      string
	ProfileID   string
	DeviceID    string
	Role        string
	MaxMaturity string
	CountryCode string
}

func RequireAuthentication(iamClient iamv1.IamServiceClient) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, `{"error":"missing bearer token"}`, http.StatusUnauthorized)
				return
			}
			rawToken := strings.TrimPrefix(authHeader, "Bearer ")
			res, err := iamClient.ValidateAccessToken(r.Context(), &iamv1.ValidateAccessTokenRequest{
				AccessToken: rawToken,
			})
			if err != nil || !res.GetIsValid() {
				http.Error(w, `{"error":"invalid or expired token"}`, http.StatusUnauthorized)
				return
			}
			profileID := r.Header.Get("X-Netflix-Profile-Id")
			if profileID == "" {
				profileID = res.GetProfileId()
			}
			sess := SessionContext{
				UserID:      res.GetUserId(),
				ProfileID:   profileID,
				DeviceID:    res.GetDeviceId(),
				Role:        res.GetRole(),
				MaxMaturity: res.GetMaxMaturity(),
				CountryCode: res.GetCountryCode(),
			}
			ctx := context.WithValue(r.Context(), AuthSessionKey, sess)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func ExtractSession(ctx context.Context) (SessionContext, bool) {
	sess, ok := ctx.Value(AuthSessionKey).(SessionContext)
	return sess, ok
}

