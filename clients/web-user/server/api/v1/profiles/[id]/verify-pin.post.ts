import { defineEventHandler, readBody } from 'h3';

interface VerifyPinBody {
  raw_pin_code?: string;
}

export default defineEventHandler(async (event) => {
  const profileId = event.context.params?.id ?? '';
  const body = await readBody<VerifyPinBody>(event);
  const pin = (body?.raw_pin_code ?? '').trim();

  // Accept standard demo PIN 1234
  const isPinValid = pin === '1234';

  if (!isPinValid) {
    return {
      verified: false,
      profile_scoped_jwt: ''
    };
  }

  const tokenHeader = btoa(JSON.stringify({ alg: 'HS256', typ: 'JWT' }));
  const tokenPayload = btoa(
    JSON.stringify({
      sub: 'usr-netflix-viewer-2026',
      prof: profileId,
      role: 'SUBSCRIBER',
      maturity: 'AGE_19',
      exp: Math.floor(Date.now() / 1000) + 86400
    })
  );
  const signature = btoa('verified_netflix_pin_token_2026');
  const scopedJwt = `${tokenHeader}.${tokenPayload}.${signature}`;

  return {
    verified: true,
    profile_scoped_jwt: scopedJwt
  };
});
