import { cookies } from 'next/headers';

/**
 * Reads the current session from the Go API.
 *
 * The session cookie is httpOnly, so the browser cannot read it and neither can
 * client components — which is the point. Only server components call this, and
 * they forward the cookie by hand because a server-side fetch does not inherit
 * the incoming request's cookie jar.
 *
 * Go remains the only thing that can validate a session token. Next.js never
 * parses or trusts it; it asks.
 */

export interface Session {
  authenticated: boolean;
  email?: string;
  tenantId?: string;
}

const API_ORIGIN = process.env.GO_API_ORIGIN ?? 'http://127.0.0.1:8080';

export async function getSession(): Promise<Session> {
  const jar = await cookies();
  const token = jar.get('cg_session');

  // No cookie means signed out. Skipping the round trip here matters: this runs
  // on every render of every Guard page, and most visitors are anonymous.
  if (!token) return { authenticated: false };

  try {
    const res = await fetch(`${API_ORIGIN}/api/auth/session`, {
      headers: { cookie: `cg_session=${token.value}` },
      // Session state changes the moment someone logs out. Caching it would
      // keep showing an account menu to a signed-out visitor.
      cache: 'no-store',
    });

    if (!res.ok) return { authenticated: false };
    return (await res.json()) as Session;
  } catch (err) {
    // If the API is unreachable, render the signed-out view rather than an
    // error page. A visitor reading the docs does not care that the session
    // service is down, and the marketing site should survive it.
    console.error('[session] could not reach the API', err);
    return { authenticated: false };
  }
}
