import { useState, useMemo } from "react";
import { queryApi, ingestApi } from "../lib/api";
import { useAsyncData, SectionLoader, EmptyState } from "../lib/hooks.jsx";
import ServiceSelector from "../components/ServiceSelector";

function Sparkline({ points, color = "var(--cyan)" }) {
  if (!points || points.length < 2) return null;
  const vals = points.map(p => Number(p.value) || 0);
  const min = Math.min(...vals);
  const max = Math.max(...vals, min + 1);
  const W = 200; const H = 50;
  const pts = points.map((_, i) => {
    const x = (i / (points.length - 1)) * W;
    const y = H - ((vals[i] - min) / (max - min)) * H;
    return `${x},${y}`;
  }).join(" ");
  return (
    <svg viewBox={`0 0 ${W} ${H}`} style={{ width:"100%", height:"50px", overflow:"visible" }} preserveAspectRatio="none">
      <defs>
        <linearGradient id="sg" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" stopColor={color} stopOpacity="0.3" />
          <stop offset="100%" stopColor={color} stopOpacity="0" />
        </linearGradient>
      </defs>
      <polygon points={`0,${H} ${pts} ${W},${H}`} fill="url(#sg)" />
      <polyline points={pts} fill="none" stroke={color} strokeWidth="1.5" />
    </svg>
  );
}

function MetricCard({ name, points }) {
  const vals = points.map(p => Number(p.value) || 0);
  const last = vals[vals.length - 1] ?? 0;
  const prev = vals[vals.length - 2] ?? last;
  const delta = last - prev;
  const unit = points[points.length - 1]?.unit || points[0]?.unit || "";
  return (
    <div className="panel panel--glow" style={{ gap:"0.5rem" }}>
      <div style={{ display:"flex", justifyContent:"space-between", alignItems:"flex-start" }}>
        <div style={{ fontSize:"0.75rem", color:"var(--text-3)", fontWeight:700, textTransform:"uppercase", letterSpacing:"0.07em" }}>{name}</div>
        <span className={`badge ${delta >= 0 ? "badge-success" : "badge-danger"}`}>{delta >= 0 ? "▲" : "▼"} Live</span>
      </div>
      <div style={{ fontSize:"2rem", fontWeight:700, letterSpacing:"-0.04em", lineHeight:1.1 }}>
        {last.toLocaleString()} <span style={{ fontSize:"0.9rem", color:"var(--text-3)", fontWeight:500 }}>{unit}</span>
      </div>
      <Sparkline points={points} color="var(--cyan)" />
      <div style={{ fontSize:"0.75rem", color:"var(--text-3)" }}>
        Latest @ {points[points.length-1]?.occurred_at ? new Date(points[points.length-1].occurred_at).toLocaleTimeString("en",{hour12:false}) : "—"}
      </div>
    </div>
  );
}

export default function MetricsPage({ state, notify }) {
  const token = state.token;
  const [selectedServices, setSelectedServices] = useState([]);
  const [filters, setFilters] = useState({ metric_name:"", lookback_minutes:"120" });
  const { data: rawMetrics, loading, error, refetch } = useAsyncData(
    () => {
      const queryServiceName = selectedServices.length === 1 ? selectedServices[0] : "";
      return queryApi.metricsWithFilters(token, {
        ...filters,
        service_name: queryServiceName,
        lookback_minutes: parseInt(filters.lookback_minutes) || 120
      });
    },
    [token, JSON.stringify(filters), JSON.stringify(selectedServices)], { skip: !token }
  );

  const metrics = useMemo(() => {
    if (!rawMetrics) return [];
    if (selectedServices.length > 1) {
      return rawMetrics.filter(m => selectedServices.includes(m.service_name));
    }
    return rawMetrics;
  }, [rawMetrics, selectedServices]);

  const f = (k, v) => setFilters(p => ({ ...p, [k]: v }));

  async function sendTestMetric() {
    if (!state.apiKey) {
      return notify("API Key missing! Please generate an Ingestion key in Settings first.", "error");
    }

    // Realistic metrics from api-gateway and analytics-service
    const metricBurst = [
      { metric_name: "http_requests_total",    value: Math.floor(Math.random()*500)+50,    unit: "req/s",   service_name: "api-gateway"       },
      { metric_name: "http_error_rate",         value: parseFloat((Math.random()*5).toFixed(2)),            unit: "%",     service_name: "api-gateway"       },
      { metric_name: "http_latency_p99_ms",     value: Math.floor(Math.random()*800)+100,   unit: "ms",      service_name: "api-gateway"       },
      { metric_name: "http_latency_p50_ms",     value: Math.floor(Math.random()*200)+20,    unit: "ms",      service_name: "api-gateway"       },
      { metric_name: "report_generation_ms",    value: Math.floor(Math.random()*3000)+200,  unit: "ms",      service_name: "analytics-service" },
      { metric_name: "mongodb_query_duration",  value: Math.floor(Math.random()*500)+10,    unit: "ms",      service_name: "analytics-service" },
      { metric_name: "mongodb_connections",     value: Math.floor(Math.random()*80)+10,     unit: "conn",    service_name: "analytics-service" },
      { metric_name: "cache_hit_ratio",         value: parseFloat((Math.random()*40+60).toFixed(1)),        unit: "%",     service_name: "analytics-service" },
      { metric_name: "queue_depth",             value: Math.floor(Math.random()*200),        unit: "msgs",    service_name: "api-gateway"       },
      { metric_name: "active_users",            value: Math.floor(Math.random()*300)+20,    unit: "users",   service_name: "api-gateway"       },
    ];

    // Pick 4 random metrics to send this burst
    const picks = metricBurst.sort(() => Math.random()-0.5).slice(0, 4);

    const events = picks.map(m => ({
      event_type: "metric",
      payload: { ...m, environment: "production" }
    }));

    try {
      await ingestApi.ingest(state.apiKey, events);
      notify(`${events.length} metrics ingested: ${picks.map(p=>p.metric_name).join(", ")}`, "success");
      setTimeout(refetch, 1200);
    } catch(e) { notify(e.message, "error"); }
  }

  const grouped = useMemo(() => {
    if (!metrics) return {};
    const g = {};
    [...metrics].sort((a,b)=>new Date(a.occurred_at)-new Date(b.occurred_at)).forEach(m => {
      const name = m.name || m.metric_name || m.payload?.name || "unknown";
      if (!g[name]) g[name] = [];
      g[name].push({ value: m.value ?? m.payload?.value ?? 0, unit: m.unit || m.payload?.unit || "", occurred_at: m.occurred_at });
    });
    return g;
  }, [metrics]);

  const metricNames = Object.keys(grouped);

  return (
    <div style={{ display:"flex", flexDirection:"column", gap:"1.25rem" }}>
      <div style={{ display:"flex", justifyContent:"space-between", alignItems:"flex-end" }}>
        <div>
          <h1 style={{ fontSize:"1.4rem", fontWeight:700, marginBottom:"0.2rem" }}>Metrics Explorer</h1>
          <p style={{ color:"var(--text-2)", fontSize:"0.875rem" }}>Time-series data and performance trends.</p>
        </div>
        <div style={{ display:"flex", gap:"0.65rem" }}>
          {state.apiKey && <button className="btn btn-primary btn-sm" onClick={sendTestMetric}>+ Test Metric</button>}
          <button className="btn btn-secondary btn-sm" onClick={refetch}>↺ Refresh</button>
        </div>
      </div>

      {/* Filters */}
      <div className="panel" style={{ padding:"1rem" }}>
        <div className="form-grid">
          <div className="form-group" style={{ marginBottom:0 }}>
            <label className="form-label">Metric Name</label>
            <input className="form-input" placeholder="e.g. http_requests_total" value={filters.metric_name} onChange={e=>f("metric_name",e.target.value)} />
          </div>
          <div className="form-group" style={{ marginBottom:0 }}>
            <label className="form-label">Service</label>
            <ServiceSelector token={token} selectedServices={selectedServices} onChange={setSelectedServices} />
          </div>
          <div className="form-group" style={{ marginBottom:0 }}>
            <label className="form-label">Lookback (mins)</label>
            <input className="form-input" type="number" min="1" value={filters.lookback_minutes} onChange={e=>f("lookback_minutes",e.target.value)} />
          </div>
        </div>
      </div>

      {error && (
        <div style={{ padding:"0.875rem", background:"var(--danger-soft)", border:"1px solid rgba(239,68,68,0.25)", borderRadius:"var(--r-md)", color:"var(--danger)", fontSize:"0.875rem" }}>⚠ {error}</div>
      )}

      {loading ? (
        <div className="grid-2">
          {[1,2,3,4].map(i => <div key={i} className="panel" style={{height:"160px"}}><SectionLoader /></div>)}
        </div>
      ) : metricNames.length > 0 ? (
        <div className="grid-2">
          {metricNames.map(name => <MetricCard key={name} name={name} points={grouped[name]} />)}
        </div>
      ) : (
        <div className="panel" style={{ padding:"4rem 2rem" }}>
          <EmptyState icon="↗" title="No metrics found" body="Ingest metric events using the API key, or click 'Test Metric' to send a sample." />
        </div>
      )}
    </div>
  );
}
