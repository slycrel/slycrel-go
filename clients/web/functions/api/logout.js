import { destroySession, clearCookieHeader, isSecure } from '../_lib/auth.js';
import { jsonResponse } from '../_lib/json.js';

export async function onRequestPost({ request, env, data }) {
  await destroySession(env, data.sessionToken);
  return jsonResponse({ ok: true }, {
    headers: { 'Set-Cookie': clearCookieHeader(isSecure(request.url)) },
  });
}
