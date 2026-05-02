export interface Env {
  NOTES: R2Bucket;
  WORKER_SECRET: string;
  ALLOWED_ORIGIN: string;
}

export default {
  async fetch(request: Request, env: Env): Promise<Response> {
    // ── 1. Only allow GET ─────────────────────────────────────────────────
    if (request.method !== "GET") {
      return new Response("Method not allowed", { status: 405 });
    }

    // ── 2. Verify the shared secret ───────────────────────────────────────
    // Your SvelteKit frontend sets this header when embedding the viewer.
    // Without it, the Worker returns 403 — blocking hotlinks and scrapers.
    const secret = request.headers.get("X-Notes-Secret");
    if (!secret || secret !== env.WORKER_SECRET) {
      return new Response("Forbidden", { status: 403 });
    }

    // ── 3. Check Referer ──────────────────────────────────────────────────
    // Secondary check — the request must originate from your domain.
    const referer = request.headers.get("Referer") ?? "";
    if (!referer.startsWith(env.ALLOWED_ORIGIN)) {
      return new Response("Forbidden", { status: 403 });
    }

    // ── 4. Extract the R2 object key from the URL path ────────────────────
    // URL pattern: /view/notes/{subjectId}/{noteId}.{ext}
    const url = new URL(request.url);
    const key = url.pathname.replace(/^\/view\//, "");

    if (!key || key.includes("..")) {
      return new Response("Bad request", { status: 400 });
    }

    // ── 5. Fetch from R2 ──────────────────────────────────────────────────
    const object = await env.NOTES.get(key);
    if (!object) {
      return new Response("Not found", { status: 404 });
    }

    // ── 6. Stream the response with protective headers ────────────────────
    const headers = new Headers();

    // Tell the browser to render inline, not download.
    headers.set("Content-Disposition", "inline");

    // Only allow embedding in your own domain — blocks <iframe> on other sites.
    headers.set("X-Frame-Options", "SAMEORIGIN");

    // No caching — each request must go through the Worker's secret check.
    headers.set("Cache-Control", "private, no-store, no-cache");

    // Content type from the stored object metadata.
    headers.set(
      "Content-Type",
      object.httpMetadata?.contentType ?? "application/octet-stream",
    );

    // Content-Security-Policy: restricts what the embedded document can do.
    headers.set(
      "Content-Security-Policy",
      "default-src 'none'; style-src 'unsafe-inline'; sandbox allow-same-origin",
    );

    // No CORS — blocks fetch()/XHR from any other origin trying to read bytes.
    // Intentionally NOT setting Access-Control-Allow-Origin.

    return new Response(object.body, {
      status: 200,
      headers,
    });
  },
};
