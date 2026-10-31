import { defineEventHandler, readBody, createError } from 'h3';

interface LoginRequestBody {
  email?: string;
  password?: string;
  device_fingerprint?: string;
  device_name?: string;
  platform?: string;
  drm_security_level?: string;
}

export default defineEventHandler(async (event) => {
  const body = await readBody<LoginRequestBody>(event);

  if (!body?.email || !body?.password) {
    throw createError({
      statusCode: 400,
      statusMessage: 'Email and password are required'
    });
  }

  const userId = 'usr-netflix-viewer-2026';
  const deviceId = `dev-${(body.device_fingerprint || 'browser').slice(0, 16)}`;
  const fakeTokenHeader = btoa(JSON.stringify({ alg: 'HS256', typ: 'JWT' }));
  const fakeTokenPayload = btoa(
    JSON.stringify({
      sub: userId,
      dev: deviceId,
      role: 'SUBSCRIBER',
      plan: 'PREMIUM_UHD',
      exp: Math.floor(Date.now() / 1000) + 86400
    })
  );
  const signature = btoa('verified_netflix_signature_2026');
  const accessToken = `${fakeTokenHeader}.${fakeTokenPayload}.${signature}`;
  const refreshToken = `rft_${userId}_${Date.now()}`;

  return {
    access_token: accessToken,
    refresh_token: refreshToken,
    expires_in_sec: 86400,
    user_id: userId,
    device_id: deviceId,
    role: 'SUBSCRIBER',
    account_status: 'ACTIVE'
  };
});
