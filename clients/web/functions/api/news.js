import { jsonResponse } from '../_lib/json.js';

// GET → atomically fetch + clear all news for the current player. Returns
// { text } with newline-joined bodies. Read-and-clear is fused because the
// only caller (enter_slycrel) always wants both.
export async function onRequestGet({ env, data }) {
  const { results } = await env.DB.prepare(
    'SELECT id, body FROM news WHERE recipient = ? ORDER BY created_at, id'
  ).bind(data.bbsName).all();
  const rows = results ?? [];
  const text = rows.map(r => r.body).join('\n');
  if (rows.length) {
    await env.DB.prepare('DELETE FROM news WHERE recipient = ?').bind(data.bbsName).run();
  }
  return jsonResponse({ text });
}
