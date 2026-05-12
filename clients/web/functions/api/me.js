import { jsonResponse } from '../_lib/json.js';

// GET → { bbsName, character | null }. character is null if registered but
// not yet character-created.
export async function onRequestGet({ env, data }) {
  const row = await env.DB.prepare(
    'SELECT data, char_name_lower FROM characters WHERE bbs_name = ?'
  ).bind(data.bbsName).first();
  const character = row && row.char_name_lower != null ? JSON.parse(row.data) : null;
  return jsonResponse({ bbsName: data.bbsName, character });
}
