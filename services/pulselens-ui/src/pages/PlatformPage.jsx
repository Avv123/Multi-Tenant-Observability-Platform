import { EmptyState, SectionLoader, useAsyncData } from "../lib/hooks.jsx";
import { queryApi } from "../lib/api";

function StatusBadge({ value }) {
  if (!value && value !== 0) return <span className="badge badge-neutral">—</span>;
  if (Number(value) > 1000) return <span className="badge badge-danger">{value}</span>;
  if (Number(value) > 100)  return <span className="badge badge-warning">{value}</span>;
  return <span className="badge badge-success">{value}</span>;
}

export default function PlatformPage({ state }) {
  const token = state.token;

  const { data: platform, loading: platformLoading, refetch } =
    useAsyncData(() => queryApi.platformOverview(token), [token], { skip: !token });

  const { data: lag, loading: lagLoading } =
    useAsyncData(() => queryApi.kafkaLag(token), [token], { skip: !token });

  const { data: bp, loading: bpLoading } =
    useAsyncData(() => queryApi.platformBackpressure(token), [token], { skip: !token });

  const { data: cleanups, loading: cleanupsLoading } =
    useAsyncData(() => queryApi.cleanupRuns(token), [token], { skip: !token });

  return (
    <div style={{ display:"flex", flexDirection:"column", gap:"1.5rem" }}>
      <div style={{ display:"flex", justifyContent:"space-between", alignItems:"flex-end" }}>
        <div>
          <h1 style={{ fontSize:"1.4rem", fontWeight:700, marginBottom:"0.2rem" }}>Infrastructure</h1>
          <p style={{ color:"var(--text-2)", fontSize:"0.875rem" }}>Runtime health, Kafka lag, and queue backpressure.</p>
        </div>
        <button id="btn-refresh-platform" className="btn btn-secondary btn-sm" onClick={refetch}>↺ Refresh</button>
      </div>

      {/* Runtime nodes */}
      <div className="panel">
        <div className="panel-header">
          <div>
            <div className="panel-title">Runtime Nodes</div>
            <div className="panel-desc">Live heartbeats from all services</div>
          </div>
        </div>
        {platformLoading ? <SectionLoader /> : (
          platform?.runtime?.length ? (
            <div className="table-wrap">
              <table className="data-table" id="table-runtime">
                <thead>
                  <tr><th>Service</th><th>Mode</th><th>Port</th><th>Last Seen</th><th>Status</th></tr>
                </thead>
                <tbody>
                  {platform.runtime.map((n, i) => (
                    <tr key={i}>
                      <td style={{fontFamily:"var(--font-mono)",fontSize:"0.8rem"}}>{n.service_name}</td>
                      <td><span className="badge badge-info">{n.mode || "—"}</span></td>
                      <td style={{fontFamily:"var(--font-mono)",color:"var(--text-2)"}}>{n.port || "—"}</td>
                      <td style={{color:"var(--text-2)",fontSize:"0.8rem"}}>{n.last_seen ? new Date(n.last_seen).toLocaleString() : "—"}</td>
                      <td><span className="status-dot success" /></td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          ) : (
            <EmptyState icon="⚙" title="No runtime nodes" body="Services register heartbeats via Redis. Start the full stack to see nodes." />
          )
        )}
      </div>

      <div className="grid-2">
        {/* Kafka lag */}
        <div className="panel">
          <div className="panel-title" style={{ marginBottom:"0.75rem" }}>Kafka Consumer Lag</div>
          {lagLoading ? <SectionLoader /> : (
            lag?.length ? (
              <div className="table-wrap">
                <table className="data-table" id="table-kafka-lag">
                  <thead><tr><th>Topic</th><th>Partition</th><th>Lag</th></tr></thead>
                  <tbody>
                    {lag.map((row, i) => (
                      <tr key={i}>
                        <td style={{fontFamily:"var(--font-mono)",fontSize:"0.75rem",color:"var(--text-2)"}}>{row.topic}</td>
                        <td>{row.partition}</td>
                        <td><StatusBadge value={row.lag} /></td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            ) : (
              <EmptyState icon="◼" title="No Kafka lag data" />
            )
          )}
        </div>

        {/* Backpressure */}
        <div className="panel">
          <div className="panel-title" style={{ marginBottom:"0.75rem" }}>Queue Backpressure</div>
          {bpLoading ? <SectionLoader /> : (
            bp && Object.keys(bp).length ? (
              <div style={{ display:"flex", flexDirection:"column", gap:"0.5rem" }}>
                {Object.entries(bp).map(([k, v]) => (
                  <div key={k} style={{ display:"flex", justifyContent:"space-between", alignItems:"center", padding:"0.5rem 0.75rem", borderRadius:"var(--r-sm)", background:"var(--surface)" }}>
                    <span style={{fontFamily:"var(--font-mono)",fontSize:"0.8rem",color:"var(--text-2)"}}>{k}</span>
                    <StatusBadge value={v} />
                  </div>
                ))}
              </div>
            ) : (
              <EmptyState icon="≋" title="No backpressure data" />
            )
          )}
        </div>
      </div>

      {/* Cleanup runs */}
      <div className="panel">
        <div className="panel-title" style={{ marginBottom:"0.75rem" }}>Recent Cleanup Runs</div>
        {cleanupsLoading ? <SectionLoader /> : (
          cleanups?.length ? (
            <div className="table-wrap">
              <table className="data-table" id="table-cleanup-runs">
                <thead><tr><th>Time</th><th>Logs Deleted</th><th>Metrics</th><th>Traces</th><th>DLQ</th></tr></thead>
                <tbody>
                  {cleanups.map((run) => (
                    <tr key={run.id}>
                      <td style={{color:"var(--text-2)",fontSize:"0.8rem"}}>{new Date(run.created_at).toLocaleString()}</td>
                      <td>{run.deleted_logs ?? "—"}</td>
                      <td>{run.deleted_metrics ?? "—"}</td>
                      <td>{run.deleted_traces ?? "—"}</td>
                      <td>{run.deleted_dlq ?? "—"}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          ) : (
            <EmptyState icon="♻" title="No cleanup runs" body="The retention cleanup job runs on a schedule and deletes events older than the tenant's retention window." />
          )
        )}
      </div>
    </div>
  );
}
