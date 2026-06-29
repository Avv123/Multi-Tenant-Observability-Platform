import { useState, useMemo } from "react";
import { queryApi, ingestApi } from "../lib/api";
import { useAsyncData, SectionLoader, EmptyState } from "../lib/hooks.jsx";
import ServiceSelector from "../components/ServiceSelector";

const SEV_BADGE = { error:"badge-danger", warn:"badge-warning", info:"badge-info", debug:"badge-neutral" };

export default function LogsPage({ state, notify }) {
  const token = state.token;
  const [selectedServices, setSelectedServices] = useState([]);
  const [filters, setFilters] = useState({ severity:"", search:"", lookback_minutes:"120" });
  
  const { data: rawLogs, loading, error, refetch } = useAsyncData(
    () => {
      const queryServiceName = selectedServices.length === 1 ? selectedServices[0] : "";
      return queryApi.logsWithFilters(token, {
        ...filters,
        service_name: queryServiceName,
        lookback_minutes: parseInt(filters.lookback_minutes) || 120
      });
    },
    [token, JSON.stringify(filters), JSON.stringify(selectedServices)], { skip: !token }
  );

  const logs = useMemo(() => {
    if (!rawLogs) return [];
    if (selectedServices.length > 1) {
      return rawLogs.filter(row => selectedServices.includes(row.service_name));
    }
    return rawLogs;
  }, [rawLogs, selectedServices]);

  const f = (k, v) => setFilters(p => ({ ...p, [k]: v }));

  async function sendTestLog() {
    if (!state.apiKey) {
      return notify("API Key missing! Please generate an Ingestion key in Settings first.", "error");
    }

    // Realistic log scenarios across multiple services
    const scenarios = [
      {
        severity: "error",
        service_name: "api-gateway",
        message: "Upstream timeout: analytics-service did not respond within 5000ms",
        http_method: "POST", http_path: "/api/v1/reports/generate",
        http_status: 504, latency_ms: 5023, user_id: `usr_${Math.random().toString(36).slice(2,8)}`,
        request_id: `req_${Math.random().toString(36).slice(2,10)}`
      },
      {
        severity: "error",
        service_name: "analytics-service",
        message: "MongoDB query failed: connection pool exhausted",
        db: "analytics", collection: "report_cache",
        error_code: "ECONNRESET", retry_attempt: 3,
        request_id: `req_${Math.random().toString(36).slice(2,10)}`
      },
      {
        severity: "warn",
        service_name: "api-gateway",
        message: "Rate limit approaching: tenant has used 87% of daily quota",
        tenant_id: "tenant_demo", quota_used: 87400, quota_limit: 100000,
        request_id: `req_${Math.random().toString(36).slice(2,10)}`
      },
      {
        severity: "warn",
        service_name: "analytics-service",
        message: "Slow query detected: report aggregation took 3200ms",
        query_type: "aggregation", duration_ms: 3200, collection: "events",
        index_used: false, docs_scanned: 450000
      },
      {
        severity: "info",
        service_name: "api-gateway",
        message: "Request completed successfully",
        http_method: "GET", http_path: "/api/v1/analytics/dashboard",
        http_status: 200, latency_ms: Math.floor(Math.random()*200)+50,
        user_id: `usr_${Math.random().toString(36).slice(2,8)}`,
        request_id: `req_${Math.random().toString(36).slice(2,10)}`
      },
      {
        severity: "info",
        service_name: "analytics-service",
        message: "Report generation completed",
        report_type: ["daily_sales","inventory_summary","order_fulfillment"][Math.floor(Math.random()*3)],
        tenant_id: "tenant_demo", rows_processed: Math.floor(Math.random()*5000)+100,
        duration_ms: Math.floor(Math.random()*1000)+200
      },
      {
        severity: "debug",
        service_name: "analytics-service",
        message: "Cache miss — fetching from MongoDB",
        cache_key: `report:${Math.random().toString(36).slice(2,8)}`,
        ttl_remaining: 0, strategy: "LRU"
      },
      {
        severity: "debug",
        service_name: "api-gateway",
        message: "JWT validation passed, forwarding to upstream",
        upstream: "analytics-service", auth_method: "Bearer",
        latency_ms: Math.floor(Math.random()*15)+2
      }
    ];

    // Send 3 random realistic logs at once
    const picks = [];
    const used = new Set();
    while (picks.length < 3) {
      const idx = Math.floor(Math.random() * scenarios.length);
      if (!used.has(idx)) { used.add(idx); picks.push(scenarios[idx]); }
    }

    const events = picks.map(s => ({
      event_type: "log",
      severity: s.severity,
      payload: { ...s, environment: "production", trace_id: `tr-${Math.random().toString(36).slice(2,10)}` }
    }));

    try {
      await ingestApi.ingest(state.apiKey, events);
      notify(`${events.length} realistic logs ingested (${picks.map(p=>p.severity).join(", ")})`, "success");
      setTimeout(refetch, 1200);
    } catch(e) { notify(e.message, "error"); }
  }

  return (
    <div style={{ display:"flex", flexDirection:"column", gap:"1.25rem", height:"100%" }}>
      {/* Header */}
      <div style={{ display:"flex", justifyContent:"space-between", alignItems:"flex-end" }}>
        <div>
          <h1 style={{ fontSize:"1.4rem", fontWeight:700, marginBottom:"0.2rem" }}>Logs Explorer</h1>
          <p style={{ color:"var(--text-2)", fontSize:"0.875rem" }}>Live stream of ingested log events.</p>
        </div>
        <div style={{ display:"flex", gap:"0.65rem" }}>
          <button className="btn btn-primary btn-sm" onClick={sendTestLog}>
            + Test Log
          </button>
          <button className="btn btn-secondary btn-sm" onClick={refetch}>↺ Refresh</button>
        </div>
      </div>

      {/* Filters */}
      <div className="panel" style={{ padding:"1rem" }}>
        <div className="form-grid">
          <div className="form-group" style={{ marginBottom:0 }}>
            <label className="form-label">Search</label>
            <input className="form-input" placeholder="keyword / message…" value={filters.search} onChange={e=>f("search",e.target.value)} />
          </div>
          <div className="form-group" style={{ marginBottom:0 }}>
            <label className="form-label">Severity</label>
            <select className="form-input" value={filters.severity} onChange={e=>f("severity",e.target.value)}>
              <option value="">All Levels</option>
              {["error","warn","info","debug"].map(s=><option key={s} value={s}>{s.toUpperCase()}</option>)}
            </select>
          </div>
          <div className="form-group" style={{ marginBottom:0 }}>
            <label className="form-label">Lookback (mins)</label>
            <input className="form-input" type="number" min="1" value={filters.lookback_minutes} onChange={e=>f("lookback_minutes",e.target.value)} />
          </div>
          <div className="form-group" style={{ marginBottom:0 }}>
            <label className="form-label">Service</label>
            <ServiceSelector token={token} selectedServices={selectedServices} onChange={setSelectedServices} />
          </div>
        </div>
      </div>

      {error && (
        <div style={{ padding:"0.875rem", background:"var(--danger-soft)", border:"1px solid rgba(239,68,68,0.25)", borderRadius:"var(--r-md)", color:"var(--danger)", fontSize:"0.875rem" }}>⚠ {error}</div>
      )}

      {/* Log Stream Terminal */}
      <div className="panel panel--glow" style={{ flex:1, padding:0, overflow:"hidden", display:"flex", flexDirection:"column", minHeight:"400px" }}>
        {/* Stream header bar */}
        <div style={{ display:"flex", justifyContent:"space-between", alignItems:"center", padding:"0.65rem 1rem", borderBottom:"1px solid var(--border)", background:"rgba(0,0,0,0.25)", flexShrink:0 }}>
          <div style={{ display:"flex", alignItems:"center", gap:"0.5rem" }}>
            <span className="status-dot success pulse" />
            <span style={{ fontSize:"0.75rem", color:"var(--text-3)", fontWeight:700, letterSpacing:"0.08em", textTransform:"uppercase" }}>Live Stream</span>
          </div>
          <span style={{ fontSize:"0.75rem", color:"var(--text-3)", fontFamily:"var(--font-mono)" }}>
            {loading ? "querying…" : `${logs?.length ?? 0} events`}
          </span>
        </div>

        {/* Log lines */}
        {loading ? (
          <div style={{ padding:"2rem" }}><SectionLoader /></div>
        ) : logs?.length ? (
          <div className="log-stream" style={{ flex:1, overflowY:"auto", padding:"0.5rem 0.25rem", borderRadius:0, border:"none" }}>
            {[...logs].reverse().map((row, i) => (
              <div key={row.event_id || row.id || i} className={`log-line ${row.severity || ""}`}>
                <span className="log-time">
                  {row.occurred_at ? new Date(row.occurred_at).toLocaleTimeString("en",{hour12:false}) : "—"}
                </span>
                <span className={`badge ${SEV_BADGE[row.severity] || "badge-neutral"}`} style={{justifySelf:"start"}}>
                  {row.severity || "log"}
                </span>
                <span className="log-service">{row.service_name || row.service_id || "—"}</span>
                <span className="log-msg">{row.message || (row.payload ? JSON.stringify(row.payload) : "—")}</span>
              </div>
            ))}
          </div>
        ) : (
          <div style={{ flex:1, display:"flex", alignItems:"center", justifyContent:"center" }}>
            <EmptyState icon="≡" title="No logs found" body="Adjust your filters or ingest a test event to see the stream." />
          </div>
        )}
      </div>
    </div>
  );
}
