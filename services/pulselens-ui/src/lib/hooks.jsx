// F4: useAsyncData — per-section data fetching hook with independent loading state.
// F9: returns an empty state component when data is null/empty.
// Each page imports this hook instead of sharing one global loading boolean.

import { useCallback, useEffect, useRef, useState } from "react";

/**
 * @param {Function} fetcher  - async function that returns the data
 * @param {Array}    deps     - dependency array (re-fetches when deps change)
 * @param {Object}   options  - { skip: bool }
 */
export function useAsyncData(fetcher, deps = [], options = {}) {
  const [data, setData]       = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError]     = useState(null);
  const mountedRef = useRef(true);

  useEffect(() => {
    mountedRef.current = true;
    return () => { mountedRef.current = false; };
  }, []);

  const fetch = useCallback(async () => {
    if (options.skip) return;
    setLoading(true);
    setError(null);
    try {
      const result = await fetcher();
      if (mountedRef.current) setData(result);
    } catch (err) {
      if (mountedRef.current) setError(err.message);
    } finally {
      if (mountedRef.current) setLoading(false);
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [...deps, options.skip]);

  useEffect(() => { fetch(); }, [fetch]);

  return { data, loading, error, refetch: fetch };
}

/** Inline loading indicator for a section. */
export function SectionLoader() {
  return (
    <div style={{ display: "flex", flexDirection: "column", gap: "0.5rem", padding: "0.5rem 0" }}>
      <div className="skeleton" style={{ height: "18px", width: "60%" }} />
      <div className="skeleton" style={{ height: "18px", width: "80%" }} />
      <div className="skeleton" style={{ height: "18px", width: "50%" }} />
    </div>
  );
}

/** F9: Empty state with contextual guidance. */
export function EmptyState({ icon = "◌", title, body, action }) {
  return (
    <div className="empty-state">
      <div className="empty-state__icon" aria-hidden="true">{icon}</div>
      <div className="empty-state__title">{title}</div>
      {body  && <div className="empty-state__body">{body}</div>}
      {action}
    </div>
  );
}

/** Inline error message. */
export function ErrorMessage({ message }) {
  if (!message) return null;
  return (
    <div
      role="alert"
      style={{
        padding: "0.75rem",
        background: "var(--color-danger-bg)",
        border: "1px solid var(--color-danger)",
        borderRadius: "var(--radius-md)",
        color: "var(--color-danger)",
        fontSize: "0.83rem",
      }}
    >
      {message}
    </div>
  );
}
