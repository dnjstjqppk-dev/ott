import { describe, it, expect } from 'vitest';
import { useDefaultCatalog } from '../../clients/web-user/composables/useDefaultCatalog';

describe('useDefaultCatalog', () => {
  it('returns valid curated catalog containing hero and rows', () => {
    const { getCuratedCatalog } = useDefaultCatalog();
    const catalog = getCuratedCatalog();

    expect(catalog.lolomo).toBeDefined();
    expect(catalog.lolomo.hero_billboard.title_id).toBe('title-sq2');
    expect(catalog.lolomo.rows.length).toBeGreaterThan(0);
  });

  it('retrieves detailed title information correctly', () => {
    const { getCuratedTitleDetail } = useDefaultCatalog();
    const title = getCuratedTitleDetail('title-glory');

    expect(title.id).toBe('title-glory');
    expect(title.localized_title).toBe('더 글로리');
    expect(title.maturity_rating).toBe('19+');
  });

  it('provides a valid fallback poster url', () => {
    const { getFallbackPoster } = useDefaultCatalog();
    const fallback = getFallbackPoster();

    expect(fallback).toMatch(/^https:\/\//);
  });
});
