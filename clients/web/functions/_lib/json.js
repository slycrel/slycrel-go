export function jsonResponse(body, init = {}) {
  return new Response(JSON.stringify(body), {
    ...init,
    headers: {
      'Content-Type': 'application/json',
      ...(init.headers ?? {}),
    },
  });
}

export function errorResponse(status, message) {
  return jsonResponse({ error: message }, { status });
}
