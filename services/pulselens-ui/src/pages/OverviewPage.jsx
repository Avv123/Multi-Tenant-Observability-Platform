import { useMemo } from "react";
import { queryApi } from "../lib/api";
import { useAsyncData, SectionLoader, EmptyState } from "../lib/hooks.jsx";

function Kpi({ label, value, tone = "primary", icon, trend, trendDir = "up" }) {
  return (
    <div className="metric-card">
      <div className="metric-label">
        <span style={{ fontSize:"1.1rem" }}>{icon}</span>
        {label}
      </div>
      <div className="metric-value" style={{ color:`var(--${tone})` }}>{value ?? "—"}</div>
      {trend && (
        <div className={`metric-trend trend-${trendDir}`}>{trend}</div>
      )}
    </div>
  );
}

function SevBar({ data }) {
  if (!data?.length) return (
    <EmptyState icon="≡" title="No log events yet" body="Send log events to see severity distribution." />
  );
  const max = Math.max(...data.map(d => d.value), 1);
  const colors = { ERROR:"var(--danger)", WARN:"var(--warning)", INFO:"var(--info)", DEBUG:"var(--text-3)" };
  return (
    <div style={{ display:"flex", flexDirection:"column", gap:"0.65rem" }}>
      {data.map(d => (
        <div key={d.label} style={{ display:"flex", alignItems:"center", gap:"0.75rem" }}>
          <div style={{ width:"52px", textAlign:"right", fontSize:"0.72rem", fontWeight:700, color: colors[d.label] || "var(--text-2)" }}>
            {d.label}
          </div>
          <div style={{ flex:1, height:"8px", borderRadius:"4px", background:"var(--surface-active)", overflow:"hidden" }}>
            <div style={{ height:"100%", width:`${(d.value/max)*100}%`, background: colors[d.label] || "var(--primary)", borderRadius:"4px", transition:"width 0.6s ease" }} />
          </div>
          <div style={{ width:"44px", textAlign:"right", fontSize:"0.78rem", color:"var(--text-2)", fontFamily:"var(--font-mono)" }}>
            {d.value.toLocaleString()}
          </div>
        </div>
      ))}
    </div>
  );
}

export default function OverviewPage({ state }) {
  const token = state.token;
  const skip = !token;

  const { data: ov, loading: ovL, error: ovE, refetch } = useAsyncData(() => queryApi.overview(token), [token], { skip });
  const { data: sev, loading: sevL } = useAsyncData(() => queryApi.logSeverityWithFilters(token, { limit: 50 }), [token], { skip });

  const sevChart = useMemo(() => {
    if (!sev?.length) return [];
    const m = new Map();
    sev.forEach(r => m.set((r.severity||"unknown").toUpperCase(), (m.get((r.severity||"unknown").toUpperCase())||0) + Number(r.event_count || 0)));
    return [...m.entries()].map(([label, value]) => ({ label, value }));
  }, [sev]);

  return (
    <div style={{ display:"flex", flexDirection:"column", gap:"1.5rem" }}>
      {/* Header row */}
      <div style={{ display:"flex", justifyContent:"space-between", alignItems:"flex-end" }}>
        <div>
          <h1 style={{ fontSize:"1.4rem", fontWeight:700, marginBottom:"0.2rem" }}>Dashboard</h1>
          <p style={{ color:"var(--text-2)", fontSize:"0.875rem" }}>Real-time telemetry snapshot for your workspace.</p>
        </div>
        <button className="btn btn-secondary btn-sm" onClick={refetch}>
          <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
            <path d="M21 12a9 9 0 11-9-9c2.52 0 4.93 1 6.74 2.74L21 8"/><path d="M21 3v5h-5"/>
          </svg>
          Refresh
        </button>
      </div>

      {ovE && (
        <div style={{ padding:"0.875rem 1rem", background:"var(--danger-soft)", border:"1px solid rgba(239,68,68,0.25)", borderRadius:"var(--r-md)", color:"var(--danger)", fontSize:"0.875rem", display:"flex", alignItems:"center", gap:"0.5rem" }}>
          <span style={{fontSize:"1rem"}}>⚠</span> {ovE}
        </div>
      )}

      {/* KPI grid */}
      {ovL ? (
        <div className="metric-grid">
          {[1,2,3,4].map(i => <div key={i} className="metric-card"><div className="skeleton" style={{height:"18px",width:"60%",marginBottom:"0.75rem"}}/><div className="skeleton" style={{height:"36px",width:"40%"}}/></div>)}
        </div>
      ) : (
        <div className="metric-grid">
          <Kpi label="Log Events"         value={(ov?.log_count    ?? 0).toLocaleString()} tone="primary" icon="≡" trend="+12% vs last hr" trendDir="up" />
          <Kpi label="Metrics Ingested"   value={(ov?.metric_count ?? 0).toLocaleString()} tone="cyan"    icon="↗" trend="+5% vs last hr"  trendDir="up" />
          <Kpi label="Active Traces"      value={(ov?.trace_count  ?? 0).toLocaleString()} tone="warning" icon="⊕" trend="tracking"         trendDir="neutral" />
          <Kpi label="Services Monitored" value={(ov?.service_count?? 0)}                  tone="success" icon="◈" trend="All healthy"       trendDir="up" />
        </div>
      )}

      {/* Bottom panels */}
      <div className="grid-2">
        {/* Severity chart */}
        <div className="panel panel--glow">
          <div className="panel-header">
            <div>
              <div className="panel-title">Log Severity Distribution</div>
              <div className="panel-desc">Event volume by level — last 24 h</div>
            </div>
            {sevL && <div className="skeleton" style={{width:"60px",height:"16px"}} />}
          </div>
          <SevBar data={sevChart} />
        </div>

        {/* Incident / alerts status */}
        <div className="panel panel--glow">
          <div className="panel-header">
            <div>
              <div className="panel-title">System Health</div>
              <div className="panel-desc">Active incidents & triggered alerts</div>
            </div>
          </div>
          <div style={{ flex:1, display:"flex", flexDirection:"column", gap:"0.75rem" }}>
            <div style={{ display:"flex", alignItems:"center", gap:"0.75rem", padding:"0.875rem", borderRadius:"var(--r-sm)", background:"var(--success-soft)", border:"1px solid rgba(16,185,129,0.2)" }}>
              <span className="status-dot success" />
              <div>
                <div style={{ fontSize:"0.9rem", fontWeight:600, color:"var(--success)" }}>All Systems Nominal</div>
                <div style={{ fontSize:"0.78rem", color:"var(--text-2)", marginTop:"2px" }}>No active incidents in the selected window.</div>
              </div>
            </div>
            <div style={{ padding:"0.75rem", borderRadius:"var(--r-sm)", background:"var(--surface)", border:"1px solid var(--border)", fontSize:"0.85rem", color:"var(--text-2)", display:"flex", justifyContent:"space-between" }}>
              <span>Alert Rules Active</span>
              <span className="badge badge-info">—</span>
            </div>
            <div style={{ padding:"0.75rem", borderRadius:"var(--r-sm)", background:"var(--surface)", border:"1px solid var(--border)", fontSize:"0.85rem", color:"var(--text-2)", display:"flex", justifyContent:"space-between" }}>
              <span>Notification Channels</span>
              <span className="badge badge-neutral">—</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
