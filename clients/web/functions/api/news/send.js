import { jsonResponse, errorResponse } from '../../_lib/json.js';

// POST { to, body } — append one news entry for another player. We don't
// require the recipient row to exist; that's a soft-fail (the message just
// never gets read), matching local.js writeNews semantics.
export async function onRequestPost({ request, env }) {
  const body = await request.json().catch(() => null);
  if (!body || typeof body.to !== 'string' || typeof body.body !== 'string') {
    return errorResponse(400, 'to and body required');
  }
  const now = Math.floor(Date.now() / 1000);
  await env.DB.prepare(
    'INSERT INTO news (recipient, body, created_at) VALUES (?, ?, ?)'
  ).bind(body.to, body.body, now).run();
  return jsonResponse({ ok: true });
}
