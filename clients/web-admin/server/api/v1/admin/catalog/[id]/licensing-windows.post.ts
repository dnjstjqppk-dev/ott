import { defineEventHandler, readBody } from 'h3';

interface CreateLicensingWindowBody {
  title_id?: string;
  country_code?: string;
  starts_at?: string;
  ends_at?: string;
  allow_download?: boolean;
  allow_avod_tier?: boolean;
}

export default defineEventHandler(async (event) => {
  const body = await readBody<CreateLicensingWindowBody>(event);
  const titleId = event.context.params?.id || body?.title_id || '00000000-0000-0000-0000-000000000001';

  return {
    id: `win-${Date.now()}`,
    title_id: titleId,
    country_code: body?.country_code || 'KR',
    starts_at: body?.starts_at || new Date().toISOString(),
    ends_at: body?.ends_at || new Date(Date.now() + 365 * 86400 * 1000).toISOString(),
    allow_download: Boolean(body?.allow_download),
    allow_avod_tier: Boolean(body?.allow_avod_tier)
  };
});
