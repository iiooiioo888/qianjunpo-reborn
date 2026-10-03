/**
 * Preserve boot query across /qjp → /qjp/ redirects that strip ?live=1 / ?mockVictory=.
 * Run synchronously before module app.js (see index.html).
 */
const STORAGE_KEY = 'qjp_boot_search_v1';

function isQjpPath(pathname) {
  return pathname === '/qjp' || pathname === '/qjp/' || pathname.startsWith('/qjp/');
}

/**
 * After a bad nginx 301 (/qjp?live=1 → /qjp/), Referer often still carries the original query.
 */
export function captureBootQuery() {
  let search = window.location.search;
  if (search && search.length > 1) {
    try {
      sessionStorage.setItem(STORAGE_KEY, search);
    } catch {
      /* private mode */
    }
    return search;
  }
  if (!search && document.referrer) {
    try {
      const ref = new URL(document.referrer);
      if (isQjpPath(ref.pathname) && ref.search) {
        search = ref.search;
        window.history.replaceState(
          null,
          '',
          window.location.pathname + search + window.location.hash,
        );
        try {
          sessionStorage.setItem(STORAGE_KEY, search);
        } catch {
          /* ignore */
        }
        return search;
      }
    } catch {
      /* ignore */
    }
  }
  return search;
}
