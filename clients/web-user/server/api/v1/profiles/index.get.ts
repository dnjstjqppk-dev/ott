import { defineEventHandler } from 'h3';

export default defineEventHandler(() => {
  return {
    profiles: [
      {
        id: 'prof-gihun',
        user_id: 'usr-netflix-viewer-2026',
        display_name: '성기훈',
        avatar_key: 'avatar-red-456',
        is_kids: false,
        max_maturity: 'AGE_19',
        has_pin: false,
        preferred_audio_lang: 'ko',
        preferred_subtitle_lang: 'ko',
        autoplay_next_episode: true,
        autoplay_previews: true,
        subtitle_font_scale: 1.0,
        subtitle_bg_opacity: 0.75,
        subtitle_edge_style: 'DROP_SHADOW'
      },
      {
        id: 'prof-saebyeok',
        user_id: 'usr-netflix-viewer-2026',
        display_name: '강새벽 (PIN 잠금)',
        avatar_key: 'avatar-green-067',
        is_kids: false,
        max_maturity: 'AGE_19',
        has_pin: true,
        preferred_audio_lang: 'ko',
        preferred_subtitle_lang: 'ko',
        autoplay_next_episode: true,
        autoplay_previews: false,
        subtitle_font_scale: 1.0,
        subtitle_bg_opacity: 0.75,
        subtitle_edge_style: 'DROP_SHADOW'
      },
      {
        id: 'prof-junho',
        user_id: 'usr-netflix-viewer-2026',
        display_name: '황준호',
        avatar_key: 'avatar-blue-cop',
        is_kids: false,
        max_maturity: 'AGE_15',
        has_pin: false,
        preferred_audio_lang: 'ko',
        preferred_subtitle_lang: 'ko',
        autoplay_next_episode: true,
        autoplay_previews: true,
        subtitle_font_scale: 1.0,
        subtitle_bg_opacity: 0.75,
        subtitle_edge_style: 'RAISED'
      },
      {
        id: 'prof-kids',
        user_id: 'usr-netflix-viewer-2026',
        display_name: '키즈 전용',
        avatar_key: 'avatar-yellow-robot',
        is_kids: true,
        max_maturity: 'AGE_ALL',
        has_pin: false,
        preferred_audio_lang: 'ko',
        preferred_subtitle_lang: 'ko',
        autoplay_next_episode: true,
        autoplay_previews: true,
        subtitle_font_scale: 1.25,
        subtitle_bg_opacity: 0.9,
        subtitle_edge_style: 'UNIFORM'
      },
      {
        id: 'prof-family',
        user_id: 'usr-netflix-viewer-2026',
        display_name: '가족 공용',
        avatar_key: 'avatar-purple-vip',
        is_kids: false,
        max_maturity: 'AGE_19',
        has_pin: false,
        preferred_audio_lang: 'ko',
        preferred_subtitle_lang: 'ko',
        autoplay_next_episode: true,
        autoplay_previews: true,
        subtitle_font_scale: 1.0,
        subtitle_bg_opacity: 0.75,
        subtitle_edge_style: 'DROP_SHADOW'
      }
    ]
  };
});
