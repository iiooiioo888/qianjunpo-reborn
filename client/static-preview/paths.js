/**
 * Resolve URLs for static assets under optional nginx mount (e.g. /qjp/).
 * Set <base href="/qjp/"> in index.html when the page is served below site root.
 */
function qjpMountBase() {
  const path = window.location.pathname;
  if (path === '/qjp' || path.startsWith('/qjp/')) {
    return new URL('/qjp/', window.location.origin).href;
  }
  return null;
}

export function pageBaseUrl() {
  const baseEl = document.querySelector('base[href]');
  if (baseEl) {
    return new URL(baseEl.getAttribute('href'), window.location.origin).href;
  }
  const qjp = qjpMountBase();
  if (qjp) {
    return qjp;
  }
  const path = window.location.pathname;
  const dir = path.endsWith('/') ? path : path.replace(/\/[^/]*$/, '/');
  return new URL(dir, window.location.origin).href;
}

/** Relative path from this preview folder, or absolute http(s) / site-root path. */
export function resolveAppUrl(relativeOrAbsolute) {
  if (/^https?:\/\//i.test(relativeOrAbsolute)) {
    return relativeOrAbsolute;
  }
  if (relativeOrAbsolute.startsWith('/')) {
    return relativeOrAbsolute;
  }
  return new URL(relativeOrAbsolute, pageBaseUrl()).href;
}
