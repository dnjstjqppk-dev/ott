package e2e

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type LiveTestEnvironment struct {
	Pool           *pgxpool.Pool
	MinioClient    *minio.Client
	PartnerOrgID   uuid.UUID
	UserID         uuid.UUID
	ProfileID      uuid.UUID
	DeviceID       uuid.UUID
	TitleID        uuid.UUID
	SeasonID       uuid.UUID
	EpisodeID      uuid.UUID
	TimedTrackID   uuid.UUID
	PrimaryEdgeID  uuid.UUID
	BackupEdgeID   uuid.UUID
	RawMasterBytes []byte
	RawMasterSHA   string
}

func BootstrapLiveEnvironment(ctx context.Context, t *testing.T) *LiveTestEnvironment {
	t.Helper()
	pool, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		t.Fatalf("failed to connect to live postgres: %v", err)
	}
	mc, err := minio.New(os.Getenv("MINIO_ENDPOINT"), &minio.Options{
		Creds:  credentials.NewStaticV4(os.Getenv("MINIO_ACCESS_KEY"), os.Getenv("MINIO_SECRET_KEY"), ""),
		Secure: false,
	})
	if err != nil {
		t.Fatalf("failed to initialize minio client: %v", err)
	}
	ensureBucketsExist(ctx, t, mc, []string{"nf-raw-ingest", "nf-media-packaged"})
	rawPayload, shaHex := generateDeterministicMasterFile(16*1024*1024 + 524288)

	env := &LiveTestEnvironment{
		Pool:           pool,
		MinioClient:    mc,
		PartnerOrgID:   uuid.New(),
		UserID:         uuid.New(),
		ProfileID:      uuid.New(),
		DeviceID:       uuid.New(),
		TitleID:        uuid.New(),
		SeasonID:       uuid.New(),
		EpisodeID:      uuid.New(),
		TimedTrackID:   uuid.New(),
		PrimaryEdgeID:  uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		BackupEdgeID:   uuid.MustParse("22222222-2222-2222-2222-222222222222"),
		RawMasterBytes: rawPayload,
		RawMasterSHA:   shaHex,
	}
	seedDatabaseEntities(ctx, t, env)
	return env
}

func ensureBucketsExist(ctx context.Context, t *testing.T, mc *minio.Client, buckets []string) {
	t.Helper()
	for _, b := range buckets {
		exists, err := mc.BucketExists(ctx, b)
		if err != nil {
			t.Fatalf("bucket check failed for %s: %v", b, err)
		}
		if !exists {
			if err := mc.MakeBucket(ctx, b, minio.MakeBucketOptions{}); err != nil {
				t.Fatalf("make bucket failed for %s: %v", b, err)
			}
		}
	}
}

func generateDeterministicMasterFile(totalBytes int) ([]byte, string) {
	buf := make([]byte, totalBytes)
	for i := 0; i < totalBytes; i++ {
		buf[i] = byte((i*31 + 17) % 251)
	}
	sum := sha256.Sum256(buf)
	return buf, hex.EncodeToString(sum[:])
}

func seedDatabaseEntities(ctx context.Context, t *testing.T, env *LiveTestEnvironment) {
	t.Helper()
	seedIamAndSubscription(ctx, t, env)
	seedCatalogAndContracts(ctx, t, env)
	seedEdgeNodesAndPackagedStream(ctx, t, env)
}

func seedIamAndSubscription(ctx context.Context, t *testing.T, env *LiveTestEnvironment) {
	t.Helper()
	_, err := env.Pool.Exec(ctx, `
		INSERT INTO subscription_plans (code, display_name, monthly_price_krw, monthly_price_usd_cents, max_concurrent_streams, max_download_devices, max_resolution, supports_spatial_audio, has_ads)
		VALUES ('STANDARD', '{"ko":"스탠다드"}', 13500, 1350, 2, 2, '1080p', FALSE, FALSE)
		ON CONFLICT (code) DO NOTHING;

		INSERT INTO partner_organizations (id, legal_name, organization_type, country_code, tax_registration_no, watermark_template)
		VALUES ($1, 'Studio Dragon E2E Corp', 'STUDIO', 'KR', '120-88-99999', '{{EMAIL}}|{{IP}}');

		INSERT INTO users (id, email, password_hash, role, status, country_code)
		VALUES ($2, 'e2e_architect@netflix.com', '$2a$12$R9h/cIPz0gi.URNNX3kh2OPST9/PgBkqquzi.Ss7KIUgO2t0jWMUW', 'USER', 'ACTIVE', 'KR');

		INSERT INTO profiles (id, user_id, display_name, avatar_key, is_kids, max_maturity)
		VALUES ($3, $2, 'Lead Architect', 'avatar_01', FALSE, 'AGE_19');

		INSERT INTO user_devices (id, user_id, device_fingerprint, device_name, platform, drm_security_level, last_ip)
		VALUES ($4, $2, 'fp-e2e-001', 'Chrome Workstation', 'WEB', 'L1', '127.0.0.1');

		INSERT INTO subscriptions (user_id, plan_code, current_period_start, current_period_end)
		VALUES ($2, 'STANDARD', NOW() - INTERVAL '1 day', NOW() + INTERVAL '29 days');
	`, env.PartnerOrgID, env.UserID, env.ProfileID, env.DeviceID)
	if err != nil {
		t.Fatalf("failed to seed IAM and subscription: %v", err)
	}
}

func seedCatalogAndContracts(ctx context.Context, t *testing.T, env *LiveTestEnvironment) {
	t.Helper()
	_, err := env.Pool.Exec(ctx, `
		INSERT INTO titles (id, partner_org_id, slug, type, original_title, localized_titles, synopsis, release_year, maturity_rating, is_netflix_original, state, popularity_score)
		VALUES ($1, $2, 'kingdom-e2e-special', 'SERIES', '킹덤 E2E', '{"ko":"킹덤 E2E"}', '{"ko":"무결성 검증 시리즈"}', 2026, 'AGE_19', TRUE, 'PUBLISHED', 99.9);

		INSERT INTO seasons (id, title_id, season_number, localized_names)
		VALUES ($3, $1, 1, '{"ko":"시즌 1"}');

		INSERT INTO episodes (id, title_id, season_id, episode_number, localized_titles, duration_ms, still_image_key, state)
		VALUES ($4, $1, $3, 1, '{"ko":"1화"}', 3600000, 'stills/ep1.jpg', 'PUBLISHED');

		INSERT INTO drm_keys (episode_id, key_id_hex, content_key_encrypted, iv_hex, cenc_scheme, widevine_pssh_base64, fairplay_skd_uri)
		VALUES ($4, 'fedcba9876543210fedcba9876543210', E'\\x01020304', '00112233445566778899aabbccddeeff', 'cenc', 'AAAANHBzc2gAAAAA7e+LqXnWSs6jyCfc1R0h7QAAABQIARIQ', 'skd://fairplay.netflix.internal/key1');

		INSERT INTO timed_text_tracks (id, episode_id, language_code, track_type, assigned_partner_org_id)
		VALUES ($5, $4, 'ko', 'SUBTITLE', $2);

		INSERT INTO partner_contracts (partner_org_id, title_id, model, minimum_guarantee_krw, rate_per_qualified_hour_krw, revenue_share_bps, effective_from, effective_to)
		VALUES ($2, $1, 'MG_PLUS_REVENUE_SHARE', 10000000, 5000, 5000, CURRENT_DATE - 10, CURRENT_DATE + 300);
	`, env.TitleID, env.PartnerOrgID, env.SeasonID, env.EpisodeID, env.TimedTrackID)
	if err != nil {
		t.Fatalf("failed to seed catalog and contracts: %v", err)
	}
}

func seedEdgeNodesAndPackagedStream(ctx context.Context, t *testing.T, env *LiveTestEnvironment) {
	t.Helper()
	edgeBase := os.Getenv("EDGE_PUBLIC_BASE_URL")
	_, err := env.Pool.Exec(ctx, `
		INSERT INTO open_connect_edge_nodes (id, node_code, isp_asn, region_code, country_code, base_url, max_bandwidth_gbps, current_bandwidth_gbps, cache_hit_ratio, status)
		VALUES
			($1, 'OCA-ICN-001', 4766, 'AP-NORTHEAST-2', 'KR', $3, 100.0, 12.5, 0.9850, 'HEALTHY'),
			($2, 'OCA-ICN-002', 9318, 'AP-NORTHEAST-2', 'KR', $3, 100.0, 25.0, 0.9620, 'HEALTHY')
		ON CONFLICT (id) DO UPDATE SET status = 'HEALTHY', base_url = EXCLUDED.base_url;
	`, env.PrimaryEdgeID, env.BackupEdgeID, edgeBase)
	if err != nil {
		t.Fatalf("failed to seed edge nodes: %v", err)
	}
	cmafBytes := env.RawMasterBytes[:65536]
	objKey := "packaged/" + env.EpisodeID.String() + "/H265_HEVC_1080/segment_0.m4s"
	_, err = env.MinioClient.PutObject(ctx, "nf-media-packaged", objKey, bytes.NewReader(cmafBytes), int64(len(cmafBytes)), minio.PutObjectOptions{})
	if err != nil {
		t.Fatalf("failed to seed cmaf segment: %v", err)
	}
}

