const CACHE_NAME = "absensi-static-v1";

const STATIC_ASSETS = [
  "/",
  "/pwa/manifest.json",

  // icons (PASTIKAN ADA)
  "/pwa/icons/maonkscode-logo250.png",
  "/pwa/icons/maonkscode-logo.png",

  // external libs
  "https://cdn.tailwindcss.com",
  "https://unpkg.com/htmx.org@1.9.10"
];

// ===== INSTALL =====
self.addEventListener("install", (event) => {
  console.log("🟢 SW installing");

  event.waitUntil(
    caches.open(CACHE_NAME)
      .then(cache => {
        return cache.addAll(STATIC_ASSETS);
      })
      .catch(err => {
        console.error("❌ Cache failed:", err);
      })
  );

  self.skipWaiting();
});

// ===== ACTIVATE =====
self.addEventListener("activate", (event) => {
  console.log("🔵 SW activated");

  event.waitUntil(
    caches.keys().then(keys =>
      Promise.all(
        keys
          .filter(k => k !== CACHE_NAME)
          .map(k => caches.delete(k))
      )
    )
  );

  self.clients.claim();
});

// ===== FETCH =====
self.addEventListener("fetch", (event) => {
  const req = event.request;
  const url = new URL(req.url);

  // ❌ JANGAN sentuh websocket / api / htmx
  if (
    url.protocol === "ws:" ||
    url.protocol === "wss:" ||
    url.pathname.startsWith("/api") ||
    url.pathname.startsWith("/websocket")
  ) {
    return;
  }

  // ✅ cache-first hanya untuk GET
  if (req.method !== "GET") return;

  event.respondWith(
    caches.match(req).then(cached => {
      return cached || fetch(req);
    })
  );
});
