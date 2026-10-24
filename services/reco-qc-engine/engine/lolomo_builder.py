import os
import psycopg2
import redis
from uuid import UUID

class LoLoMoEngine:
    def __init__(self):
        self.conn = psycopg2.connect(os.getenv("DATABASE_URL"))
        self.rdb = redis.Redis.from_url(os.getenv("REDIS_URL"))

    def generate_lolomo(self, profile_id: str, country_code: str, language_code: str, max_maturity: str) -> dict:
        cur = self.conn.cursor()
        
        # 1. Hero Billboard
        cur.execute("""
            SELECT t.id, t.original_title, a.storage_key
            FROM titles t
            JOIN artworks a ON a.title_id = t.id AND a.kind = 'BILLBOARD_HERO' AND a.language_code = %s
            WHERE t.state = 'PUBLISHED' AND t.maturity_rating <= %s::maturity_level
            ORDER BY t.popularity_score DESC LIMIT 1
        """, (language_code, max_maturity))
        hero_row = cur.fetchone()
        
        hero_item = {
            "title_id": str(hero_row[0]) if hero_row else "",
            "personalized_artwork_url": f"https://cdn.netflix-clone.internal/{hero_row[2]}" if hero_row else "",
            "match_score_pct": 98,
            "rank_number": 1,
            "bookmark_ms": 0,
            "duration_ms": 5400000
        }

        # 2. Continue Watching Row Items
        cur.execute("""
            SELECT wb.title_id, wb.episode_id, wb.bookmark_ms, wb.duration_ms, t.original_title, a.storage_key
            FROM watch_bookmarks wb
            JOIN titles t ON t.id = wb.title_id
            LEFT JOIN artworks a ON a.title_id = t.id AND a.kind = 'BOXART_HORIZONTAL' AND a.language_code = %s
            WHERE wb.profile_id = %s AND wb.is_completed = FALSE AND wb.is_hidden_from_row = FALSE
            ORDER BY wb.updated_at DESC LIMIT 10
        """, (language_code, profile_id))
        cw_rows = cur.fetchall()
        
        cw_items = []
        for r in cw_rows:
            cw_items.append({
                "title_id": str(r[0]),
                "personalized_artwork_url": f"https://cdn.netflix-clone.internal/{r[5]}" if r[5] else "",
                "match_score_pct": 95,
                "rank_number": len(cw_items) + 1,
                "bookmark_ms": r[2],
                "duration_ms": r[3]
            })

        rows = []
        if cw_items:
            rows.append({
                "row_id": "row_continue_watching",
                "row_index": 0,
                "row_type": "CONTINUE_WATCHING",
                "header_text": "시청 중인 콘텐츠",
                "items": cw_items
            })

        cur.close()
        return {
            "snapshot_id": str(UUID(int=0)),
            "hero_billboard": hero_item,
            "rows": rows
        }

    def update_bandit(self, profile_id: str, artwork_id: str, event_type: str):
        cur = self.conn.cursor()
        cluster_id = 1
        if event_type == "IMPRESSION":
            cur.execute("""
                UPDATE artwork_bandit_stats SET impressions_count = impressions_count + 1,
                beta_param = beta_param + 0.1, updated_at = NOW()
                WHERE artwork_id = %s AND segment_cluster_id = %s
            """, (artwork_id, cluster_id))
        elif event_type == "PLAY_START":
            cur.execute("""
                UPDATE artwork_bandit_stats SET play_starts_count = play_starts_count + 1,
                alpha_param = alpha_param + 1.0, updated_at = NOW()
                WHERE artwork_id = %s AND segment_cluster_id = %s
            """, (artwork_id, cluster_id))
        self.conn.commit()
        cur.close()

