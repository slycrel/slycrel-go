// PBKDF2 + session-cookie helpers. PBKDF2 over Web Crypto runs natively on
// Workers; 100k iterations is a reasonable hash cost without the bcrypt /
// argon2 dependency drag.

const SESSION_COOKIE = 'slycrel_session';
const SESSION_TTL_SECONDS = 60 * 60 * 24 * 30;
const PBKDF2_ITER = 100000;
const PBKDF2_KEYLEN = 32;

const encoder = new TextEncoder();

function toHex(bytes) {
  return Array.from(new Uint8Array(bytes), b => b.toString(16).padStart(2, '0')).join('');
}

function fromHex(hex) {
  const out = new Uint8Array(hex.length / 2);
  for (let i = 0; i < hex.length; i += 2) out[i / 2] = parseInt(hex.slice(i, i + 2), 16);
  return out;
}

export async function hashPassword(password, saltHex) {
  const key = await crypto.subtle.importKey(
    'raw', encoder.encode(password), 'PBKDF2', false, ['deriveBits']
  );
  const bits = await crypto.subtle.deriveBits(
    { name: 'PBKDF2', salt: fromHex(saltHex), iterations: PBKDF2_ITER, hash: 'SHA-256' },
    key, PBKDF2_KEYLEN * 8
  );
  return toHex(bits);
}

export function newSalt() {
  return toHex(crypto.getRandomValues(new Uint8Array(16)));
}

export function newSessionToken() {
  return toHex(crypto.getRandomValues(new Uint8Array(32)));
}

export function readSessionCookie(req) {
  const cookie = req.headers.get('Cookie') ?? '';
  for (const part of cookie.split(';')) {
    const idx = part.indexOf('=');
    if (idx < 0) continue;
    const k = part.slice(0, idx).trim();
    if (k === SESSION_COOKIE) return part.slice(idx + 1).trim();
  }
  return null;
}

export function sessionCookieHeader(token, secure) {
  const attrs = [
    `${SESSION_COOKIE}=${token}`,
    'Path=/',
    `Max-Age=${SESSION_TTL_SECONDS}`,
    'HttpOnly',
    'SameSite=Lax',
  ];
  if (secure) attrs.push('Secure');
  return attrs.join('; ');
}

export function clearCookieHeader(secure) {
  const attrs = [`${SESSION_COOKIE}=`, 'Path=/', 'Max-Age=0', 'HttpOnly', 'SameSite=Lax'];
  if (secure) attrs.push('Secure');
  return attrs.join('; ');
}

export async function lookupSession(env, token) {
  if (!token) return null;
  const now = Math.floor(Date.now() / 1000);
  const row = await env.DB.prepare(
    'SELECT bbs_name FROM sessions WHERE token = ? AND expires_at > ?'
  ).bind(token, now).first();
  return row ? row.bbs_name : null;
}

export async function createSession(env, bbsName) {
  const token = newSessionToken();
  const now = Math.floor(Date.now() / 1000);
  await env.DB.prepare(
    'INSERT INTO sessions (token, bbs_name, created_at, expires_at) VALUES (?, ?, ?, ?)'
  ).bind(token, bbsName, now, now + SESSION_TTL_SECONDS).run();
  return token;
}

export async function destroySession(env, token) {
  if (!token) return;
  await env.DB.prepare('DELETE FROM sessions WHERE token = ?').bind(token).run();
}

export function isSecure(url) {
  return new URL(url).protocol === 'https:';
}
