import { readSessionCookie, lookupSession } from '../_lib/auth.js';
import { errorResponse } from '../_lib/json.js';

// /api/characters is open so the pre-login title screen can show the
// Guild Rankings without forcing a login first (matches BBS feel where
// the lobby screen showed the leaderboard).
const OPEN_PATHS = new Set(['/api/login', '/api/characters']);

export async function onRequest(context) {
  const url = new URL(context.request.url);
  if (OPEN_PATHS.has(url.pathname)) return context.next();
  const token = readSessionCookie(context.request);
  const bbsName = await lookupSession(context.env, token);
  if (!bbsName) return errorResponse(401, 'not authenticated');
  context.data.bbsName = bbsName;
  context.data.sessionToken = token;
  return context.next();
}
