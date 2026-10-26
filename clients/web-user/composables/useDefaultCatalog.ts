// clients/web-user/composables/useDefaultCatalog.ts
import type { AggregatedHomeResponse, TitleDto } from '~/types/netflix';

export function useDefaultCatalog(): {
  getCuratedCatalog: () => AggregatedHomeResponse;
  getCuratedTitleDetail: (titleId: string) => TitleDto;
} {
  const heroPoster = 'https://images.unsplash.com/photo-1574375927938-d5a98e8ffe85?w=1600&auto=format&fit=crop&q=80';
  const rowPoster1 = 'https://images.unsplash.com/photo-1626814026160-2237a95fc5a0?w=600&auto=format&fit=crop&q=80';
  const rowPoster2 = 'https://images.unsplash.com/photo-1536440136628-849c177e76a1?w=600&auto=format&fit=crop&q=80';
  const rowPoster3 = 'https://images.unsplash.com/photo-1518676599649-7c4e3e3b7b4a?w=600&auto=format&fit=crop&q=80';
  const rowPoster4 = 'https://images.unsplash.com/photo-1509281373149-e957c6296406?w=600&auto=format&fit=crop&q=80';
  const rowPoster5 = 'https://images.unsplash.com/photo-1594909122845-11baa439b7bf?w=600&auto=format&fit=crop&q=80';

  const getCuratedCatalog = (): AggregatedHomeResponse => ({
    lolomo: {
      snapshot_id: 'curated-v2026',
      hero_billboard: {
        title_id: 'title-squid-game-2',
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
            { title_id: 'title-glory', personalized_artwork_url: rowPoster2, match_score_pct: 97, rank_number: 2, bookmark_ms: 0, duration_ms: 3300000 },
            { title_id: 'title-parasite', personalized_artwork_url: rowPoster3, match_score_pct: 96, rank_number: 3, bookmark_ms: 2400000, duration_ms: 7920000 },
            { title_id: 'title-phys100', personalized_artwork_url: rowPoster4, match_score_pct: 94, rank_number: 4, bookmark_ms: 0, duration_ms: 3000000 },
            { title_id: 'title-kingdom', personalized_artwork_url: rowPoster5, match_score_pct: 93, rank_number: 5, bookmark_ms: 800000, duration_ms: 3200000 }
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
            { title_id: 'title-parasite', personalized_artwork_url: rowPoster3, match_score_pct: 92, rank_number: 4, bookmark_ms: 0, duration_ms: 7920000 },
            { title_id: 'title-phys100', personalized_artwork_url: rowPoster4, match_score_pct: 91, rank_number: 5, bookmark_ms: 0, duration_ms: 3000000 }
          ]
        },
        {
          row_id: 'row-originals',
          row_index: 2,
          row_type: 'NETFLIX_ORIGINALS',
          header_text: '넷플릭스 오리지널',
          items: [
            { title_id: 'title-sq2', personalized_artwork_url: rowPoster1, match_score_pct: 99, rank_number: 1, bookmark_ms: 0, duration_ms: 3600000 },
            { title_id: 'title-glory', personalized_artwork_url: rowPoster2, match_score_pct: 97, rank_number: 2, bookmark_ms: 0, duration_ms: 3300000 },
            { title_id: 'title-kingdom', personalized_artwork_url: rowPoster5, match_score_pct: 96, rank_number: 3, bookmark_ms: 0, duration_ms: 3200000 },
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
  });

  const getCuratedTitleDetail = (titleId: string): TitleDto => ({
    id: titleId,
    slug: 'squid-game-season-2',
    type: 'SERIES',
    localized_title: '오징어 게임: 시즌 2',
    localized_synopsis: '456억 원의 상금이 걸린 의문의 서바이벌 게임에 다시 뛰어든 기훈. 주최자들의 정체를 파헤치기 위해 목숨을 건 복수가 시작된다.',
    release_year: 2026,
    maturity_rating: '19+',
    maturity_reasons: ['폭력성', '대사', '모방위험'],
    is_netflix_original: true,
    max_resolution: '4K Ultra HD · Dolby Vision · Dolby Atmos',
    supports_hdr10: true,
    supports_dolby_vision: true,
    supports_dolby_atmos: true,
    hero_artwork_url: heroPoster,
    logo_png_url: '',
    boxart_url: rowPoster1,
    match_score_pct: 99,
    in_my_list: true,
    user_rating: 'THUMBS_UP'
  });

  return { getCuratedCatalog, getCuratedTitleDetail };
}
