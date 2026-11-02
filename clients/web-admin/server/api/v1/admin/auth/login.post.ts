import { defineEventHandler, readBody, createError } from 'h3';

interface AdminLoginBody {
  email?: string;
  password?: string;
}

export default defineEventHandler(async (event) => {
  const body = await readBody<AdminLoginBody>(event);
  const email = (body?.email ?? '').trim();
  const password = (body?.password ?? '').trim();

  // Validate admin credentials
  const isValid =
    (email === 'admin@netflix.com' || email === 'admin') &&
    password === 'netflix2026!strict';

  if (!isValid) {
    throw createError({
      statusCode: 401,
      statusMessage: 'Invalid admin credentials. Access denied.'
    });
  }

  const tokenHeader = btoa(JSON.stringify({ alg: 'HS256', typ: 'JWT' }));
  const tokenPayload = btoa(
    JSON.stringify({
      sub: 'adm-netflix-sre-001',
      email: 'admin@netflix.com',
      role: 'SUPER_ADMIN',
      permissions: ['CATALOG_WRITE', 'MEDIA_OPERATOR', 'EDGE_DRAIN', 'AUDIT_READ'],
      exp: Math.floor(Date.now() / 1000) + 86400 * 7
    })
  );
  const signature = btoa('verified_netflix_admin_signature_2026');
  const token = `${tokenHeader}.${tokenPayload}.${signature}`;

  return {
    success: true,
    token,
    admin: {
      id: 'adm-netflix-sre-001',
      email: 'admin@netflix.com',
      role: 'SUPER_ADMIN',
      name: '시스템 최고 관리자 (SRE Control Tower)'
    }
  };
});
