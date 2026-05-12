import { jsonResponse } from '../_lib/json.js';

// GET → all *created* characters (skips registered-but-uncreated rows so
// the Guild Lists don't show empty placeholders).
export async function onRequestGet({ env }) {
  const { results } = await env.DB.prepare(
    'SELECT data FROM characters WHERE char_name_lower IS NOT NULL ORDER BY bbs_name'
  ).all();
  return jsonResponse((results ?? []).map(r => JSON.parse(r.data)));
}
