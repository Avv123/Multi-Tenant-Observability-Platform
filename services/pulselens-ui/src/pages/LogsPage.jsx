import { useState } from "react";
import { queryApi, ingestApi } from "../lib/api";
import { useAsyncData, SectionLoader, EmptyState } from "../lib/hooks.jsx";

const SEV_BADGE = { error:"badge-danger", warn:"badge-warning", info:"badge-info", debug:"badge-neutral" };

export default function LogsPage({ state, notify }) {
  const token = state.token;
  const [filters, setFilters] = useState({ service_id:"", severity:"", search:"", lookback_minutes:"120" });
  const { data: logs, loading, error, refetch } = useAsyncData(
    () => queryApi.logsWithFilters(token, { ...filters, lookback_minutes: parseInt(filters.lookback_minutes)||120 }),
    [token, JSON.stringify(filters)], { skip: !token }
  );

  const f = (k, v) => setFilters(p => ({ ...p, [k]: v }));

  async function sendTestLog() {
    try {
      await ingestApi.ingest(state.apiKey, [{
        event_type: "log",
        severity: ["error","warn","info","debug"][Math.floor(Math.random()*4)],
        payload: { message: `Test event at ${new Date().toISOString()}`, service: "test-client", trace_id: `tr-${Math.random().toString(36).slice(2,10)}` }
      }]);
      notify("Test log ingested", "success");
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
          {state.apiKey && (
            <button className="btn btn-primary btn-sm" onClick={sendTestLog}>
              + Test Log
            </button>
          )}
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
            <input className="form-input" placeholder="service-id…" value={filters.service_id} onChange={e=>f("service_id",e.target.value)} />
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
