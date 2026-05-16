import { useMemo, useState } from "react";
import LineChart from "../components/LineChart";
import ServiceMap from "../components/ServiceMap";
import { EmptyState, SectionLoader, useAsyncData } from "../lib/hooks";
import { queryApi, ingestApi } from "../lib/api";
import { buildTraceFilters } from "../lib/queryFilters";
import { compactTimestamp } from "../lib/widgets";

function sampleTrace() {
  return [{
    event_type: "trace",
    payload: {
      trace_id: `trace-${Date.now()}`,
      span_id: "root-span",
      parent_span_id: "",
      operation: "checkout",
      status: "error",
      service_name: "checkout-service",
      environment: "production",
      start_time: new Date(Date.now() - 120).toISOString(),
      end_time: new Date().toISOString(),
    },
  }];
}

export default function TracesPage({ state, notify }) {
  const token = state.token;

  const [filters, setFilters] = useState({
    trace_id: "",
    service_id: "",
    environment: "",
    lookback_minutes: "120",
  });

  const { data: traces, loading, error, refetch } = useAsyncData(
    () => queryApi.tracesWithFilters(token, buildTraceFilters(filters)),
    [token, JSON.stringify(filters)],
    { skip: !token },
  );

  const { data: latency } = useAsyncData(
    () => queryApi.traceLatencyWithFilters(token, { limit: 20, ...buildTraceFilters(filters) }),
    [token, JSON.stringify(filters)],
    { skip: !token },
  );

  const chart = useMemo(() => (
    [...(latency || [])]
      .sort((a, b) => new Date(a.bucket_start) - new Date(b.bucket_start))
      .slice(-15)
      .map((row) => ({
        label: compactTimestamp(row.bucket_start),
        value: Number(row.average_duration_ms || 0),
      }))
  ), [latency]);

  async function handleIngest() {
    if (!state.apiKey) {
      return notify("API Key missing! Please generate an Ingestion key in Settings first.", "error");
    }
    try {
      await ingestApi.ingest(state.apiKey, sampleTrace());
      notify("Sample trace ingested.", "success");
      setTimeout(refetch, 1000);
    } catch (err) {
      notify(err.message, "error");
    }
  }

  return (
    <div style={{ display:"flex", flexDirection:"column", gap:"1.25rem" }}>
      {/* Header */}
      <div style={{ display:"flex", justifyContent:"space-between", alignItems:"flex-end" }}>
        <div>
          <h1 style={{ fontSize:"1.4rem", fontWeight:700, marginBottom:"0.2rem" }}>Trace Explorer</h1>
          <p style={{ color:"var(--text-2)", fontSize:"0.875rem" }}>Distributed request flows and spans.</p>
        </div>
        <div style={{ display:"flex", gap:"0.65rem" }}>
          <button id="btn-ingest-trace" className="btn btn-primary btn-sm" onClick={handleIngest}>
            + Ingest Sample Trace
          </button>
          <button id="btn-refresh-traces" className="btn btn-secondary btn-sm" onClick={refetch}>↺ Refresh</button>
        </div>
      </div>

      {error && (
        <div style={{ padding:"0.875rem", background:"var(--danger-soft)", border:"1px solid rgba(239,68,68,0.25)", borderRadius:"var(--r-md)", color:"var(--danger)", fontSize:"0.875rem" }}>⚠ {error}</div>
      )}

      {/* Filter panel */}
      <div className="panel" style={{ padding:"1rem" }}>
        <div className="form-grid">
          {[
            { key: "trace_id",         label: "Trace ID",       placeholder: "trace-xxxx" },
            { key: "service_id",       label: "Service ID",     placeholder: "all" },
            { key: "environment",      label: "Environment",    placeholder: "production" },
            { key: "lookback_minutes", label: "Lookback (min)", placeholder: "120", type: "number" },
          ].map(({ key, label, placeholder, type = "text" }) => (
            <div key={key} className="form-group" style={{ marginBottom:0 }}>
              <label className="form-label" htmlFor={`trace-filter-${key}`}>{label}</label>
              <input
                className="form-input"
                id={`trace-filter-${key}`}
                type={type}
                value={filters[key]}
                placeholder={placeholder}
                onChange={(e) => setFilters((f) => ({ ...f, [key]: e.target.value }))}
              />
            </div>
          ))}
        </div>
      </div>

      <div className="grid-2">
        <div className="panel">
          <div className="panel-title" style={{ marginBottom: "1rem" }}>Avg Span Latency (ms)</div>
          {chart.length
            ? <LineChart data={chart} color="var(--warning)" />
            : <EmptyState icon="⊕" title="No latency data yet" />}
        </div>

        <div className="panel" style={{ gridColumn: "1 / -1" }}>
          <div className="panel-title" style={{ marginBottom: "1rem" }}>Service Topology</div>
          {loading ? <SectionLoader /> : <ServiceMap traces={traces} />}
        </div>

        <div className="panel" style={{ gridColumn: "1 / -1" }}>
          <div className="panel-title" style={{ marginBottom: "1rem" }}>Recent Spans</div>
          {loading ? <SectionLoader /> : (
            traces?.length ? (
              <div className="table-wrap">
                <table className="data-table" id="table-traces">
                  <thead>
                    <tr>
                      <th>Time</th>
                      <th>Trace ID</th>
                      <th>Operation</th>
                      <th>Status</th>
                      <th>Duration</th>
                    </tr>
                  </thead>
                  <tbody>
                    {traces.map((row, i) => (
                      <tr key={row.event_id || i}>
                        <td style={{ color:"var(--text-2)", fontSize:"0.8rem" }}>{compactTimestamp(row.occurred_at)}</td>
                        <td style={{ fontFamily:"var(--font-mono)", fontSize:"0.75rem" }}>{row.trace_id?.slice(0, 16) || "—"}</td>
                        <td>{row.operation || row.service_name || "—"}</td>
                        <td>
                          <span className={`badge ${row.status === "error" ? "badge-danger" : "badge-success"}`}>
                            {row.status || "ok"}
                          </span>
                        </td>
                        <td>{row.duration_ms != null ? `${row.duration_ms}ms` : "—"}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            ) : (
              <EmptyState
                icon="⊕"
                title="No trace spans found"
                body="Ingest a trace event to start visualising distributed request flows."
                action={state.apiKey && (
                  <button id="btn-ingest-trace-empty" className="btn btn-primary btn-sm" onClick={handleIngest}>
                    + Ingest Sample Trace
                  </button>
                )}
              />
            )
          )}
        </div>
      </div>
    </div>
  );
}
