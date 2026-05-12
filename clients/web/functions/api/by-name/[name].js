import { jsonResponse } from '../../_lib/json.js';

export async function onRequestGet({ env, params }) {
  // Pages Functions hand us the raw, URL-encoded segment in params.
  const name = decodeURIComponent(params.name);
  const row = await env.DB.prepare(
    'SELECT data FROM characters WHERE char_name_lower = ?'
  ).bind(name.toLowerCase()).first();
  return jsonResponse(row ? JSON.parse(row.data) : null);
}
