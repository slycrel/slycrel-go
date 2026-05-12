import { jsonResponse, errorResponse } from '../_lib/json.js';

// PUT body=character — save any registered character. We trust authed
// callers to mutate other players' records (arena win/loss writes to the
// opponent). This matches the original BBS where everyone shared a single
// authoritative game-state file.
export async function onRequestPut({ request, env }) {
  const character = await request.json().catch(() => null);
  if (!character || typeof character.bbsName !== 'string') {
    return errorResponse(400, 'character.bbsName required');
  }
  const nameLower = typeof character.name === 'string' && character.name
    ? character.name.toLowerCase()
    : null;
  const blob = JSON.stringify(character);
  const now = Math.floor(Date.now() / 1000);
  const res = await env.DB.prepare(
    'UPDATE characters SET data = ?, char_name_lower = ?, updated_at = ? WHERE bbs_name = ?'
  ).bind(blob, nameLower, now, character.bbsName).run();
  if (!res.success || (res.meta && res.meta.changes === 0)) {
    return errorResponse(404, 'character not registered');
  }
  return jsonResponse({ ok: true });
}
