import { defineEventHandler } from 'h3';

export default defineEventHandler(() => {
  const heroPoster =
    'https://images.unsplash.com/photo-1578632767115-351597cf2477?w=1600&auto=format&fit=crop&q=80';
  const rowPoster1 =
    'https://images.unsplash.com/photo-1626814026160-2237a95fc5a0?w=600&auto=format&fit=crop&q=80';
  const rowPoster2 =
    'https://images.unsplash.com/photo-1536440136628-849c177e76a1?w=600&auto=format&fit=crop&q=80';
  const rowPoster3 =
    'https://images.unsplash.com/photo-1574267432553-4b4628081c31?w=600&auto=format&fit=crop&q=80';
  const rowPoster4 =
    'https://images.unsplash.com/photo-1509281373149-e957c6296406?w=600&auto=format&fit=crop&q=80';
  const rowPoster5 =
    'https://images.unsplash.com/photo-1594909122845-11baa439b7bf?w=600&auto=format&fit=crop&q=80';

  return {
    lolomo: {
      snapshot_id: 'live-cf-v2026',
      hero_billboard: {
        title_id: 'title-sq2',
        personalized_artwork_url: heroPoster,
        match_score_pct: 99,
        rank_number: 1,
        bookmark_ms: 0,
        duration_ms: 3600000
      },
      rows: [
        {
          row_id: 'row-top-10',
          row_index: 0,
          row_type: 'TOP_10_COUNTRY',
          header_text: '오늘 대한민국 TOP 10 시리즈',
          items: [
            { title_id: 'title-sq2', personalized_artwork_url: rowPoster1, match_score_pct: 99, rank_number: 1, bookmark_ms: 1200000, duration_ms: 3600000 },
            { title_id: 'title-glory', personalized_artwork_url: rowPoster2, match_score_pct: 98, rank_number: 2, bookmark_ms: 0, duration_ms: 3300000 },
            { title_id: 'title-parasite', personalized_artwork_url: rowPoster3, match_score_pct: 96, rank_number: 3, bookmark_ms: 2400000, duration_ms: 7920000 },
            { title_id: 'title-phys100', personalized_artwork_url: rowPoster4, match_score_pct: 94, rank_number: 4, bookmark_ms: 0, duration_ms: 3000000 },
            { title_id: 'title-kingdom', personalized_artwork_url: rowPoster5, match_score_pct: 95, rank_number: 5, bookmark_ms: 800000, duration_ms: 3200000 }
          ]
        },
        {
          row_id: 'row-trending',
          row_index: 1,
          row_type: 'TRENDING_NOW',
          header_text: '지금 뜨는 콘텐츠',
          items: [
            { title_id: 'title-glory', personalized_artwork_url: rowPoster2, match_score_pct: 98, rank_number: 1, bookmark_ms: 0, duration_ms: 3300000 },
            { title_id: 'title-kingdom', personalized_artwork_url: rowPoster5, match_score_pct: 95, rank_number: 2, bookmark_ms: 0, duration_ms: 3200000 },
            { title_id: 'title-sq2', personalized_artwork_url: rowPoster1, match_score_pct: 99, rank_number: 3, bookmark_ms: 0, duration_ms: 3600000 },
            { title_id: 'title-parasite', personalized_artwork_url: rowPoster3, match_score_pct: 96, rank_number: 4, bookmark_ms: 0, duration_ms: 7920000 },
            { title_id: 'title-phys100', personalized_artwork_url: rowPoster4, match_score_pct: 94, rank_number: 5, bookmark_ms: 0, duration_ms: 3000000 }
          ]
        },
        {
          row_id: 'row-originals',
          row_index: 2,
          row_type: 'NETFLIX_ORIGINALS',
          header_text: '넷플릭스 오리지널',
          items: [
            { title_id: 'title-sq2', personalized_artwork_url: rowPoster1, match_score_pct: 99, rank_number: 1, bookmark_ms: 0, duration_ms: 3600000 },
            { title_id: 'title-glory', personalized_artwork_url: rowPoster2, match_score_pct: 98, rank_number: 2, bookmark_ms: 0, duration_ms: 3300000 },
            { title_id: 'title-kingdom', personalized_artwork_url: rowPoster5, match_score_pct: 95, rank_number: 3, bookmark_ms: 0, duration_ms: 3200000 },
            { title_id: 'title-phys100', personalized_artwork_url: rowPoster4, match_score_pct: 94, rank_number: 4, bookmark_ms: 0, duration_ms: 3000000 }
          ]
        }
      ]
    },
    subscription: {
      subscription_id: 'sub-premium-active',
      plan_code: 'PREMIUM_UHD',
      max_concurrent_streams: 4,
      max_resolution: '4K Ultra HD',
      has_ads: false
    }
  };
});
