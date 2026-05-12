import { jsonResponse } from '../../_lib/json.js';

export async function onRequestGet({ env, params }) {
  const bbsName = decodeURIComponent(params.bbsName);
  const row = await env.DB.prepare(
    'SELECT data, char_name_lower FROM characters WHERE bbs_name = ?'
  ).bind(bbsName).first();
  if (!row || row.char_name_lower == null) return jsonResponse(null);
  return jsonResponse(JSON.parse(row.data));
}
