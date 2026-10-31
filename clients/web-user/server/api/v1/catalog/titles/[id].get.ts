import { defineEventHandler } from 'h3';

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

const titlesMap: Record<string, any> = {
  'title-sq2': {
    id: 'title-sq2',
    slug: 'squid-game-season-2',
    type: 'SERIES',
    localized_title: '오징어 게임: 시즌 2',
    localized_synopsis:
      '456억 원의 상금이 걸린 의문의 서바이벌 게임에 다시 뛰어든 기훈. 주최자들의 정체를 파헤치기 위해 목숨을 건 복수가 시작된다.',
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
  },
  'title-glory': {
    id: 'title-glory',
    slug: 'the-glory',
    type: 'SERIES',
    localized_title: '더 글로리',
    localized_synopsis:
      '유년 시절 폭력으로 영혼까지 부서진 한 여자가 온 생을 걸어 치밀하게 준비한 처절한 복수와 그 소용돌이에 빠져드는 이들의 이야기.',
    release_year: 2023,
    maturity_rating: '19+',
    maturity_reasons: ['폭력성', '약물', '모방위험'],
    is_netflix_original: true,
    max_resolution: '4K Ultra HD · HDR',
    supports_hdr10: true,
    supports_dolby_vision: true,
    supports_dolby_atmos: true,
    hero_artwork_url: rowPoster2,
    logo_png_url: '',
    boxart_url: rowPoster2,
    match_score_pct: 98,
    in_my_list: true,
    user_rating: 'THUMBS_UP'
  },
  'title-parasite': {
    id: 'title-parasite',
    slug: 'parasite',
    type: 'MOVIE',
    localized_title: '기생충',
    localized_synopsis:
      '전원백수로 살 길 막막하지만 사이는 좋은 기택 가족. 장남 기우가 고액 과외 면접을 위해 박사장 집에 발을 들이면서 두 가족의 만남이 시작된다.',
    release_year: 2019,
    maturity_rating: '15+',
    maturity_reasons: ['주제', '대사'],
    is_netflix_original: false,
    max_resolution: '4K Ultra HD',
    supports_hdr10: true,
    supports_dolby_vision: false,
    supports_dolby_atmos: true,
    hero_artwork_url: rowPoster3,
    logo_png_url: '',
    boxart_url: rowPoster3,
    match_score_pct: 96,
    in_my_list: false,
    user_rating: 'THUMBS_UP'
  },
  'title-phys100': {
    id: 'title-phys100',
    slug: 'physical-100',
    type: 'SERIES',
    localized_title: '피지컬: 100',
    localized_synopsis:
      '가장 완벽한 피지컬을 자부하는 100인이 벌이는 극강의 컴피티션. 최후의 1인이 되어 영광과 상금을 거머쥘 피지컬은 누구인가.',
    release_year: 2024,
    maturity_rating: '12+',
    maturity_reasons: ['신체접촉'],
    is_netflix_original: true,
    max_resolution: '4K Ultra HD',
    supports_hdr10: true,
    supports_dolby_vision: false,
    supports_dolby_atmos: false,
    hero_artwork_url: rowPoster4,
    logo_png_url: '',
    boxart_url: rowPoster4,
    match_score_pct: 94,
    in_my_list: false,
    user_rating: 'THUMBS_UP'
  },
  'title-kingdom': {
    id: 'title-kingdom',
    slug: 'kingdom-ashin',
    type: 'SERIES',
    localized_title: '킹덤: 아신전',
    localized_synopsis:
      '조선을 뒤흔든 거대한 비극의 시작. 생사초의 비밀과 아신의 이야기가 마침내 드러난다. 핏빛으로 물든 북방의 대서사시.',
    release_year: 2021,
    maturity_rating: '19+',
    maturity_reasons: ['폭력성', '공포'],
    is_netflix_original: true,
    max_resolution: '4K Ultra HD · Dolby Vision',
    supports_hdr10: true,
    supports_dolby_vision: true,
    supports_dolby_atmos: true,
    hero_artwork_url: rowPoster5,
    logo_png_url: '',
    boxart_url: rowPoster5,
    match_score_pct: 95,
    in_my_list: true,
    user_rating: 'THUMBS_UP'
  }
};

export default defineEventHandler((event) => {
  const id = event.context.params?.id ?? 'title-sq2';
  const title = titlesMap[id] ?? titlesMap['title-sq2'];
  return {
    detail: { title },
    license: {
      eligible: true,
      country_allowed: true,
      requires_parental_approval: false,
      avod_ad_frequency_tier: 'STANDARD'
    }
  };
});
