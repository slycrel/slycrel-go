import { hashPassword, newSalt, createSession, sessionCookieHeader, isSecure } from '../_lib/auth.js';
import { jsonResponse, errorResponse } from '../_lib/json.js';

// POST { bbsName, password } — verifies an existing player, or registers a new
// one on first login. Auto-register matches the BBS feel where "logging in
// with a name nobody's used yet" is how you claim it; password locks the
// claim. Returns { bbsName, registered, hasCharacter }.
export async function onRequestPost({ request, env }) {
  const body = await request.json().catch(() => null);
  if (!body || typeof body.bbsName !== 'string' || typeof body.password !== 'string') {
    return errorResponse(400, 'bbsName and password required');
  }
  const bbsName = body.bbsName.trim();
  if (!bbsName) return errorResponse(400, 'bbsName required');
  if (!body.password) return errorResponse(400, 'password required');

  const row = await env.DB.prepare(
    'SELECT bbs_name, password_hash, password_salt, char_name_lower FROM characters WHERE bbs_name = ?'
  ).bind(bbsName).first();

  let registered = false;
  let hasCharacter = false;
  if (!row) {
    const salt = newSalt();
    const hash = await hashPassword(body.password, salt);
    const now = Math.floor(Date.now() / 1000);
    const blank = JSON.stringify({ bbsName });
    await env.DB.prepare(
      'INSERT INTO characters (bbs_name, password_hash, password_salt, data, updated_at) VALUES (?, ?, ?, ?, ?)'
    ).bind(bbsName, hash, salt, blank, now).run();
    registered = true;
  } else {
    const hash = await hashPassword(body.password, row.password_salt);
    if (hash !== row.password_hash) return errorResponse(401, 'bad password');
    hasCharacter = row.char_name_lower != null;
  }

  const token = await createSession(env, bbsName);
  return jsonResponse({ bbsName, registered, hasCharacter }, {
    headers: { 'Set-Cookie': sessionCookieHeader(token, isSecure(request.url)) },
  });
}
