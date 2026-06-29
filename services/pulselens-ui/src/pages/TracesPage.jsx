import { useMemo, useState } from "react";
import { EmptyState, SectionLoader, useAsyncData } from "../lib/hooks";
import { queryApi, ingestApi } from "../lib/api";
import { compactTimestamp } from "../lib/widgets";
import ServiceSelector from "../components/ServiceSelector";

// ─── Utilities ───────────────────────────────────────────────────────────────
const SEV_COLORS = {
  "analytics-service": "var(--cyan)",
  "api-gateway": "var(--primary-2)",
};
function serviceColor(name = "") {
  return SEV_COLORS[name] || `hsl(${[...name].reduce((a, c) => a + c.charCodeAt(0), 0) % 360}, 65%, 55%)`;
}
function statusBadge(status) {
  return status === "error"
    ? <span className="badge badge-danger">ERROR</span>
    : <span className="badge badge-success">OK</span>;
}
function msLabel(ms) {
  if (ms == null || ms === "") return "—";
  const n = Number(ms);
  return n >= 1000 ? `${(n / 1000).toFixed(2)}s` : `${n}ms`;
}

// ─── Sample trace builder ────────────────────────────────────────────────────
function sampleTrace() {
  const traceId = `trace-${Date.now()}-${Math.random().toString(36).slice(2, 6)}`;
  const now = Date.now();
  const scenarios = [
    [
      { span_id: "span-root", parent_span_id: "", operation: "POST /api/v1/reports/generate", service_name: "api-gateway", status: "ok", start_ms: 0, dur_ms: 320 },
      { span_id: "span-auth", parent_span_id: "span-root", operation: "jwt.validate", service_name: "api-gateway", status: "ok", start_ms: 5, dur_ms: 8 },
      { span_id: "span-rpc", parent_span_id: "span-root", operation: "grpc.GenerateReport", service_name: "analytics-service", status: "ok", start_ms: 15, dur_ms: 290 },
      { span_id: "span-db", parent_span_id: "span-rpc", operation: "mongo.aggregate(events)", service_name: "analytics-service", status: "ok", start_ms: 20, dur_ms: 200 },
    ],
    [
      { span_id: "span-root", parent_span_id: "", operation: "GET /api/v1/reports/fetch", service_name: "api-gateway", status: "error", start_ms: 0, dur_ms: 5030 },
      { span_id: "span-rpc", parent_span_id: "span-root", operation: "FetchReport", service_name: "analytics-service", status: "error", start_ms: 12, dur_ms: 5010 },
    ],
  ];
  const spans = scenarios[Math.floor(Math.random() * scenarios.length)];
  return spans.map(s => ({
    event_type: "trace",
    payload: {
      trace_id: traceId, span_id: s.span_id, parent_span_id: s.parent_span_id,
      operation: s.operation, status: s.status, service_name: s.service_name,
      duration_ms: s.dur_ms, environment: "production",
      start_time: new Date(now - spans[0].dur_ms + s.start_ms).toISOString(),
      end_time: new Date(now - spans[0].dur_ms + s.start_ms + s.dur_ms).toISOString(),
      http_method: s.operation.startsWith("POST") ? "POST" : s.operation.startsWith("GET") ? "GET" : undefined,
      http_url: s.operation.match(/\/api\/v1\/.+/) ? s.operation.replace(/^[A-Z]+ /, "") : undefined,
    },
  }));
}

// ─── Waterfall component ─────────────────────────────────────────────────────
function TraceWaterfall({ spans, onSelectSpan, selectedSpanId }) {
  if (!spans?.length) return <EmptyState icon="⊕" title="No spans" body="No span data found for this trace." />;

  const minTime = Math.min(...spans.map(s => new Date(s.occurred_at || s.start_time || 0).getTime()));
  const totalDur = spans.reduce((mx, s) => Math.max(mx, s.duration_ms || 0), 0) || 1;

  return (
    <div style={{ fontFamily: "var(--font-mono)", fontSize: "0.78rem" }}>
      {/* Header */}
      <div style={{ display: "grid", gridTemplateColumns: "200px 1fr", gap: "0.5rem", padding: "0.4rem 0.75rem", color: "var(--text-3)", fontWeight: 700, textTransform: "uppercase", fontSize: "0.68rem", letterSpacing: "0.07em", borderBottom: "1px solid var(--border)" }}>
        <span>Service / Operation</span>
        <span>Timeline</span>
      </div>
      {spans.map((span, i) => {
        const startOffset = Math.max(0, (new Date(span.occurred_at || 0).getTime() - minTime));
        const durMs = span.duration_ms || 0;
        const leftPct = Math.min(95, (startOffset / (totalDur * 1.1)) * 100);
        const widthPct = Math.max(0.5, Math.min(100 - leftPct, (durMs / (totalDur * 1.1)) * 100));
        const color = serviceColor(span.service_name);
        const isSelected = selectedSpanId === span.span_id;
        return (
          <div
            key={span.span_id || i}
            onClick={() => onSelectSpan(span)}
            style={{
              display: "grid", gridTemplateColumns: "200px 1fr", gap: "0.5rem",
              padding: "0.35rem 0.75rem", alignItems: "center", cursor: "pointer",
              background: isSelected ? "rgba(99,102,241,0.12)" : i % 2 === 0 ? "rgba(255,255,255,0.02)" : "transparent",
              borderLeft: isSelected ? "2px solid var(--primary-2)" : "2px solid transparent",
              transition: "background 0.15s",
            }}
          >
            <div style={{ overflow: "hidden" }}>
              <div style={{ color, fontWeight: 600, fontSize: "0.73rem", whiteSpace: "nowrap", overflow: "hidden", textOverflow: "ellipsis" }}>
                {span.service_name || "—"}
              </div>
              <div style={{ color: "var(--text-2)", fontSize: "0.71rem", whiteSpace: "nowrap", overflow: "hidden", textOverflow: "ellipsis" }}>
                {span.operation || "—"}
              </div>
            </div>
            <div style={{ position: "relative", height: "20px" }}>
              <div style={{ position: "absolute", left: `${leftPct}%`, width: `${widthPct}%`, height: "100%", background: color, borderRadius: "3px", opacity: 0.85, display: "flex", alignItems: "center", paddingLeft: "4px", minWidth: "2px" }}>
                {widthPct > 8 && <span style={{ color: "#fff", fontSize: "0.64rem", fontWeight: 600, whiteSpace: "nowrap" }}>{msLabel(durMs)}</span>}
              </div>
              {widthPct <= 8 && (
                <span style={{ position: "absolute", left: `calc(${leftPct}% + ${widthPct}% + 4px)`, color: "var(--text-3)", fontSize: "0.68rem", top: "2px" }}>{msLabel(durMs)}</span>
              )}
            </div>
          </div>
        );
      })}
    </div>
  );
}

// ─── Span detail panel ───────────────────────────────────────────────────────
function SpanDetail({ span, token, onClose }) {
  const payload = useMemo(() => {
    try { return typeof span.payload === "string" ? JSON.parse(span.payload) : (span.payload || {}); }
    catch { return {}; }
  }, [span.payload]);

  const { data: logs } = useAsyncData(
    () => span.trace_id ? queryApi.correlatedLogs(token, span.trace_id) : Promise.resolve([]),
    [span.trace_id], { skip: !token || !span.trace_id }
  );

  const attrs = [
    ["Service", span.service_name],
    ["Status", span.status],
    ["Duration", msLabel(span.duration_ms)],
    ["Started", span.occurred_at ? new Date(span.occurred_at).toLocaleTimeString("en", { hour12: false }) : "—"],
    ["Trace ID", span.trace_id?.slice(0, 24)],
    ["Span ID", span.span_id?.slice(0, 16)],
  ];

  const httpAttrs = [
    ["Method", payload.http_method],
    ["URL", payload.http_url],
    ["Status Code", payload.http_status_code],
  ].filter(([, v]) => v);

  const headers = payload.http_request_headers || {};
  const customAttrs = Object.entries(payload).filter(([k]) =>
    !["span_id","parent_span_id","operation","status","start_time","end_time",
      "service_name","http_method","http_url","http_status_code","http_request_headers",
      "error_message","error_type"].includes(k) && payload[k] != null
  );

  return (
    <div style={{ borderTop: "1px solid var(--border)", background: "rgba(0,0,0,0.3)", padding: "1.25rem" }}>
      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: "1rem" }}>
        <div>
          <div style={{ fontWeight: 700, fontSize: "1rem" }}>{span.operation}</div>
          <div style={{ color: "var(--text-3)", fontSize: "0.8rem" }}>Span Detail</div>
        </div>
        <button className="btn btn-ghost btn-sm" onClick={onClose} style={{ fontSize: "1.2rem", lineHeight: 1 }}>✕</button>
      </div>

      <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: "1.25rem" }}>
        {/* Left — core attributes */}
        <div>
          <div style={{ fontSize: "0.72rem", color: "var(--text-3)", fontWeight: 700, textTransform: "uppercase", letterSpacing: "0.07em", marginBottom: "0.5rem" }}>Span Attributes</div>
          <div style={{ display: "grid", gridTemplateColumns: "auto 1fr", gap: "0.25rem 1rem" }}>
            {attrs.map(([k, v]) => v && (
              <>
                <span key={k + "k"} style={{ color: "var(--text-3)", fontSize: "0.8rem" }}>{k}</span>
                <span key={k + "v"} style={{ color: "var(--text-1)", fontSize: "0.8rem", fontFamily: "var(--font-mono)", wordBreak: "break-all" }}>
                  {k === "Status" ? statusBadge(v) : v}
                </span>
              </>
            ))}
          </div>

          {httpAttrs.length > 0 && (
            <>
              <div style={{ fontSize: "0.72rem", color: "var(--text-3)", fontWeight: 700, textTransform: "uppercase", letterSpacing: "0.07em", margin: "1rem 0 0.5rem" }}>HTTP Attributes</div>
              <div style={{ display: "grid", gridTemplateColumns: "auto 1fr", gap: "0.25rem 1rem" }}>
                {httpAttrs.map(([k, v]) => (
                  <>
                    <span key={k + "k"} style={{ color: "var(--text-3)", fontSize: "0.8rem" }}>{k}</span>
                    <span key={k + "v"} style={{ color: "var(--cyan)", fontSize: "0.8rem", fontFamily: "var(--font-mono)" }}>{String(v)}</span>
                  </>
                ))}
              </div>
            </>
          )}

          {Object.keys(headers).length > 0 && (
            <>
              <div style={{ fontSize: "0.72rem", color: "var(--text-3)", fontWeight: 700, textTransform: "uppercase", letterSpacing: "0.07em", margin: "1rem 0 0.5rem" }}>Request Headers</div>
              <div style={{ display: "grid", gridTemplateColumns: "auto 1fr", gap: "0.25rem 1rem" }}>
                {Object.entries(headers).map(([k, v]) => (
                  <>
                    <span key={k + "k"} style={{ color: "var(--text-3)", fontSize: "0.8rem" }}>{k}</span>
                    <span key={k + "v"} style={{ color: "var(--text-1)", fontSize: "0.8rem", fontFamily: "var(--font-mono)", wordBreak: "break-all" }}>{v}</span>
                  </>
                ))}
              </div>
            </>
          )}

          {payload.error_message && (
            <>
              <div style={{ fontSize: "0.72rem", color: "var(--danger)", fontWeight: 700, textTransform: "uppercase", letterSpacing: "0.07em", margin: "1rem 0 0.5rem" }}>Error</div>
              <div style={{ background: "var(--danger-soft)", border: "1px solid rgba(239,68,68,0.25)", borderRadius: "var(--r-md)", padding: "0.6rem 0.75rem", fontSize: "0.8rem", color: "var(--danger)", fontFamily: "var(--font-mono)" }}>
                {payload.error_message}
              </div>
            </>
          )}
        </div>

        {/* Right — custom attrs + correlated logs */}
        <div>
          {customAttrs.length > 0 && (
            <>
              <div style={{ fontSize: "0.72rem", color: "var(--text-3)", fontWeight: 700, textTransform: "uppercase", letterSpacing: "0.07em", marginBottom: "0.5rem" }}>Custom Attributes</div>
              <div style={{ display: "grid", gridTemplateColumns: "auto 1fr", gap: "0.25rem 1rem", marginBottom: "1rem" }}>
                {customAttrs.map(([k, v]) => (
                  <>
                    <span key={k + "k"} style={{ color: "var(--text-3)", fontSize: "0.8rem" }}>{k}</span>
                    <span key={k + "v"} style={{ color: "var(--text-1)", fontSize: "0.8rem", fontFamily: "var(--font-mono)", wordBreak: "break-all" }}>{String(v)}</span>
                  </>
                ))}
              </div>
            </>
          )}

          <div style={{ fontSize: "0.72rem", color: "var(--text-3)", fontWeight: 700, textTransform: "uppercase", letterSpacing: "0.07em", marginBottom: "0.5rem" }}>
            Correlated Logs {logs?.length ? `(${logs.length})` : ""}
          </div>
          {logs?.length ? (
            <div style={{ maxHeight: "200px", overflowY: "auto", borderRadius: "var(--r-md)", border: "1px solid var(--border)" }}>
              {logs.map((log, i) => (
                <div key={i} style={{ display: "grid", gridTemplateColumns: "60px 60px 1fr", gap: "0.4rem", padding: "0.3rem 0.5rem", borderBottom: "1px solid var(--border)", fontSize: "0.75rem" }}>
                  <span style={{ color: "var(--text-3)" }}>{log.occurred_at ? new Date(log.occurred_at).toLocaleTimeString("en", { hour12: false }) : "—"}</span>
                  <span className={`badge ${log.severity === "error" ? "badge-danger" : log.severity === "warn" ? "badge-warning" : "badge-info"}`}>{log.severity}</span>
                  <span style={{ color: "var(--text-2)", overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap" }}>{log.message}</span>
                </div>
              ))}
            </div>
          ) : (
            <div style={{ color: "var(--text-3)", fontSize: "0.8rem" }}>No correlated logs found for this trace.</div>
          )}
        </div>
      </div>
    </div>
  );
}

// ─── Main TracesPage ─────────────────────────────────────────────────────────
export default function TracesPage({ state, notify }) {
  const token = state.token;
  const [selectedServices, setSelectedServices] = useState([]);
  const [filters, setFilters] = useState({ trace_id: "", environment: "", lookback_minutes: "120" });
  const [selectedTrace, setSelectedTrace] = useState(null);
  const [selectedSpan, setSelectedSpan] = useState(null);

  const { data: rawTraces, loading, error, refetch } = useAsyncData(
    () => {
      const queryServiceName = selectedServices.length === 1 ? selectedServices[0] : "";
      return queryApi.tracesWithFilters(token, {
        ...filters,
        service_name: queryServiceName,
        lookback_minutes: parseInt(filters.lookback_minutes) || 120,
      });
    },
    [token, JSON.stringify(filters), JSON.stringify(selectedServices)],
    { skip: !token }
  );

  const traces = useMemo(() => {
    if (!rawTraces) return [];
    if (selectedServices.length > 1) {
      return rawTraces.filter(t => selectedServices.includes(t.service_name));
    }
    return rawTraces;
  }, [rawTraces, selectedServices]);

  const { data: traceSpans, loading: spansLoading } = useAsyncData(
    () => selectedTrace ? queryApi.traceDetail(token, selectedTrace.trace_id) : Promise.resolve(null),
    [token, selectedTrace?.trace_id],
    { skip: !token || !selectedTrace }
  );

  async function handleIngest() {
    if (!state.apiKey) return notify("API Key missing!", "error");
    try {
      const spans = sampleTrace();
      await ingestApi.ingest(state.apiKey, spans);
      notify(`Sample trace ingested — ${spans.length} spans`, "success");
      setTimeout(refetch, 1000);
    } catch (err) { notify(err.message, "error"); }
  }

  return (
    <div style={{ display: "flex", flexDirection: "column", gap: "1.25rem", height: "100%" }}>
      {/* Header */}
      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-end" }}>
        <div>
          <h1 style={{ fontSize: "1.4rem", fontWeight: 700, marginBottom: "0.2rem" }}>Trace Explorer</h1>
          <p style={{ color: "var(--text-2)", fontSize: "0.875rem" }}>Distributed request flows — click a trace to see the waterfall.</p>
        </div>
        <div style={{ display: "flex", gap: "0.65rem" }}>
          <button className="btn btn-primary btn-sm" onClick={handleIngest}>+ Sample Trace</button>
          <button className="btn btn-secondary btn-sm" onClick={refetch}>↺ Refresh</button>
        </div>
      </div>

      {error && <div style={{ padding: "0.875rem", background: "var(--danger-soft)", border: "1px solid rgba(239,68,68,0.25)", borderRadius: "var(--r-md)", color: "var(--danger)", fontSize: "0.875rem" }}>⚠ {error}</div>}

      {/* Filters */}
      <div className="panel" style={{ padding: "1rem" }}>
        <div className="form-grid">
          <div className="form-group" style={{ marginBottom: 0 }}>
            <label className="form-label">Trace ID</label>
            <input className="form-input" value={filters.trace_id} placeholder="trace-xxxx"
              onChange={e => setFilters(f => ({ ...f, trace_id: e.target.value }))} />
          </div>
          <div className="form-group" style={{ marginBottom: 0 }}>
            <label className="form-label">Service</label>
            <ServiceSelector token={token} selectedServices={selectedServices} onChange={setSelectedServices} />
          </div>
          <div className="form-group" style={{ marginBottom: 0 }}>
            <label className="form-label">Environment</label>
            <input className="form-input" value={filters.environment} placeholder="production"
              onChange={e => setFilters(f => ({ ...f, environment: e.target.value }))} />
          </div>
          <div className="form-group" style={{ marginBottom: 0 }}>
            <label className="form-label">Lookback (min)</label>
            <input className="form-input" type="number" value={filters.lookback_minutes} placeholder="120"
              onChange={e => setFilters(f => ({ ...f, lookback_minutes: e.target.value }))} />
          </div>
        </div>
      </div>

      {/* Two-panel layout */}
      <div style={{ display: "grid", gridTemplateColumns: selectedTrace ? "380px 1fr" : "1fr", gap: "1rem", flex: 1, minHeight: 0 }}>
        {/* Trace List */}
        <div className="panel" style={{ padding: 0, overflow: "hidden", display: "flex", flexDirection: "column" }}>
          <div style={{ padding: "0.75rem 1rem", borderBottom: "1px solid var(--border)", display: "flex", justifyContent: "space-between", alignItems: "center", flexShrink: 0 }}>
            <span style={{ fontWeight: 700, fontSize: "0.875rem" }}>Traces</span>
            <span style={{ fontSize: "0.75rem", color: "var(--text-3)" }}>{traces?.length ?? 0} found</span>
          </div>
          <div style={{ overflowY: "auto", flex: 1 }}>
            {loading ? <div style={{ padding: "2rem" }}><SectionLoader /></div> : traces?.length ? (
              traces.map((trace, i) => {
                const isSelected = selectedTrace?.trace_id === trace.trace_id;
                const hasError = trace.status === "error";
                return (
                  <div key={trace.trace_id || i} onClick={() => { setSelectedTrace(trace); setSelectedSpan(null); }}
                    style={{ padding: "0.75rem 1rem", borderBottom: "1px solid var(--border)", cursor: "pointer", background: isSelected ? "rgba(99,102,241,0.1)" : "transparent", borderLeft: isSelected ? "3px solid var(--primary-2)" : "3px solid transparent", transition: "background 0.15s" }}>
                    <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: "0.3rem" }}>
                      <span style={{ fontFamily: "var(--font-mono)", fontSize: "0.73rem", color: "var(--text-2)" }}>{trace.trace_id?.slice(0, 20) || "—"}</span>
                      {statusBadge(hasError ? "error" : "ok")}
                    </div>
                    <div style={{ display: "flex", justifyContent: "space-between", fontSize: "0.75rem" }}>
                      <span style={{ color: serviceColor(trace.service_name), fontWeight: 600 }}>{trace.service_name || "—"}</span>
                      <span style={{ color: "var(--text-3)" }}>{trace.span_count} spans</span>
                    </div>
                    <div style={{ fontSize: "0.72rem", color: "var(--text-3)", marginTop: "0.2rem" }}>
                      {trace.last_seen_at ? compactTimestamp(trace.last_seen_at) : "—"}
                    </div>
                  </div>
                );
              })
            ) : <div style={{ padding: "3rem 1rem", textAlign: "center", color: "var(--text-3)", fontSize: "0.875rem" }}>No traces found. Adjust filters or ingest a sample.</div>}
          </div>
        </div>

        {/* Waterfall Detail */}
        {selectedTrace && (
          <div className="panel" style={{ padding: 0, overflow: "hidden", display: "flex", flexDirection: "column" }}>
            {/* Trace header */}
            <div style={{ padding: "0.75rem 1rem", borderBottom: "1px solid var(--border)", display: "flex", justifyContent: "space-between", alignItems: "center", flexShrink: 0 }}>
              <div>
                <div style={{ fontWeight: 700, fontSize: "0.875rem" }}>
                  Trace: <span style={{ fontFamily: "var(--font-mono)", color: "var(--text-2)" }}>{selectedTrace.trace_id?.slice(0, 24)}</span>
                </div>
                <div style={{ fontSize: "0.75rem", color: "var(--text-3)", marginTop: "0.15rem" }}>
                  {selectedTrace.span_count} spans · {selectedTrace.service_name}
                </div>
              </div>
              <button className="btn btn-ghost btn-sm" onClick={() => setSelectedTrace(null)}>✕</button>
            </div>

            {/* Waterfall */}
            <div style={{ flex: 1, overflowY: "auto", minHeight: 0 }}>
              {spansLoading ? <div style={{ padding: "2rem" }}><SectionLoader /></div> : (
                <TraceWaterfall
                  spans={traceSpans || []}
                  selectedSpanId={selectedSpan?.span_id}
                  onSelectSpan={setSelectedSpan}
                />
              )}
            </div>

            {/* Span Detail */}
            {selectedSpan && (
              <div style={{ maxHeight: "55%", overflowY: "auto", flexShrink: 0 }}>
                <SpanDetail span={selectedSpan} token={token} onClose={() => setSelectedSpan(null)} />
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
}
