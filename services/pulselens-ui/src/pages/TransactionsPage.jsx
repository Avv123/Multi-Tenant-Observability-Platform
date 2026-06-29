import { useState, useMemo } from "react";
import { SectionLoader, EmptyState, useAsyncData } from "../lib/hooks";
import { queryApi } from "../lib/api";
import LineChart from "../components/LineChart";
import ServiceSelector from "../components/ServiceSelector";

// ─── Utilities ───────────────────────────────────────────────────────────────
function msLabel(ms) {
  if (!ms && ms !== 0) return "—";
  const n = Number(ms);
  return n >= 1000 ? `${(n / 1000).toFixed(2)}s` : `${Math.round(n)}ms`;
}

function ErrorBadge({ pct }) {
  const n = Number(pct || 0);
  if (n < 1) return <span className="badge badge-success">{n.toFixed(1)}%</span>;
  if (n < 5) return <span className="badge badge-warning">{n.toFixed(1)}%</span>;
  return <span className="badge badge-danger">{n.toFixed(1)}%</span>;
}

function SortIcon({ col, sortCol, sortDir }) {
  if (sortCol !== col) return <span style={{ color: "var(--text-3)", fontSize: "0.65rem" }}>⇅</span>;
  return <span style={{ color: "var(--primary-2)", fontSize: "0.65rem" }}>{sortDir === "asc" ? "↑" : "↓"}</span>;
}

// ─── Transaction Detail Drawer ────────────────────────────────────────────────
function TxnDetail({ tx, token, onClose }) {
  const { data: traces } = useAsyncData(
    () => queryApi.tracesWithFilters(token, { service_name: tx.service_name, limit: 5 }),
    [tx.operation, tx.service_name], { skip: !token }
  );

  return (
    <>
      {/* Blurred dark dimming backdrop overlay */}
      <div className="drawer-backdrop" onClick={onClose} />

      {/* Solid deep space drawer panel */}
      <div className="drawer-panel" style={{ height: "48vh" }}>
        <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-start", marginBottom: "1rem", borderBottom: "1px solid var(--border)", paddingBottom: "0.75rem", flexShrink: 0 }}>
          <div>
            <div style={{ fontFamily: "var(--font-mono)", fontSize: "1.05rem", fontWeight: 700, color: "var(--text-1)" }}>{tx.operation}</div>
            <div style={{ color: "var(--text-3)", fontSize: "0.8rem", marginTop: "0.2rem" }}>
              Service: <span className="badge badge-info" style={{ fontSize: "0.7rem", textTransform: "none" }}>{tx.service_name}</span>
            </div>
          </div>
          <button className="btn btn-ghost btn-sm" onClick={onClose}>✕ Close</button>
        </div>

        <div style={{ display: "grid", gridTemplateColumns: "repeat(5, 1fr)", gap: "1rem", marginBottom: "1.5rem", flexShrink: 0 }}>
          {[
            { label: "Total Calls", value: tx.total_calls?.toLocaleString() },
            { label: "Avg Latency", value: msLabel(tx.avg_duration_ms) },
            { label: "Max Latency", value: msLabel(tx.max_duration_ms) },
            { label: "Error Rate", value: <ErrorBadge pct={tx.error_pct} /> },
            { label: "Total Errors", value: tx.total_errors?.toLocaleString() },
          ].map(({ label, value }) => (
            <div key={label} style={{ background: "rgba(255,255,255,0.02)", border: "1px solid var(--border)", borderRadius: "var(--r-md)", padding: "0.75rem", textAlign: "center" }}>
              <div style={{ color: "var(--text-3)", fontSize: "0.72rem", marginBottom: "0.3rem", fontWeight: 500 }}>{label}</div>
              <div style={{ fontWeight: 700, fontSize: "1rem", fontFamily: "var(--font-mono)" }}>{value}</div>
            </div>
          ))}
        </div>

        <div style={{ flex: 1, overflowY: "auto" }}>
          <div style={{ fontSize: "0.72rem", color: "var(--text-3)", fontWeight: 700, textTransform: "uppercase", letterSpacing: "0.07em", marginBottom: "0.5rem" }}>
            Recent Traces for This Operation
          </div>
          {traces?.length ? (
            <div style={{ display: "flex", flexDirection: "column", gap: "0.35rem" }}>
              {traces.slice(0, 5).map((t, i) => (
                <div key={i} style={{ display: "flex", justifyContent: "space-between", padding: "0.45rem 0.65rem", background: "rgba(255,255,255,0.03)", border: "1px solid var(--border)", borderRadius: "var(--r-sm)", fontSize: "0.78rem" }}>
                  <span style={{ fontFamily: "var(--font-mono)", color: "var(--primary-2)" }}>{t.trace_id}</span>
                  <span style={{ color: "var(--text-3)", fontSize: "0.75rem" }}>{t.span_count} spans</span>
                </div>
              ))}
            </div>
          ) : (
            <div style={{ color: "var(--text-3)", fontSize: "0.8rem", padding: "1rem 0" }}>No recent traces available.</div>
          )}
        </div>
      </div>
    </>
  );
}

// ─── Main TransactionsPage ────────────────────────────────────────────────────
const COLUMNS = [
  { key: "operation",      label: "Transaction",    align: "left"  },
  { key: "service_name",   label: "Service",        align: "left"  },
  { key: "total_calls",    label: "Calls",          align: "right" },
  { key: "avg_duration_ms",label: "Avg",            align: "right" },
  { key: "max_duration_ms",label: "Max",            align: "right" },
  { key: "error_pct",      label: "Error %",        align: "right" },
];

export default function TransactionsPage({ state, notify }) {
  const token = state.token;
  const [selectedServices, setSelectedServices] = useState([]);
  const [search, setSearch] = useState("");
  const [sortCol, setSortCol] = useState("total_calls");
  const [sortDir, setSortDir] = useState("desc");
  const [selectedTxn, setSelectedTxn] = useState(null);

  const { data: rawRows, loading, error, refetch } = useAsyncData(
    () => {
      const queryServiceName = selectedServices.length === 1 ? selectedServices[0] : "";
      return queryApi.transactions(token, { service_name: queryServiceName || undefined, limit: 100 });
    },
    [token, JSON.stringify(selectedServices)],
    { skip: !token }
  );

  const rows = useMemo(() => {
    if (!rawRows) return [];
    if (selectedServices.length > 1) {
      return rawRows.filter(r => selectedServices.includes(r.service_name));
    }
    return rawRows;
  }, [rawRows, selectedServices]);

  function toggleSort(col) {
    if (sortCol === col) setSortDir(d => d === "asc" ? "desc" : "asc");
    else { setSortCol(col); setSortDir("desc"); }
  }

  const filtered = useMemo(() => {
    let data = rows || [];
    if (search) data = data.filter(r => r.operation?.toLowerCase().includes(search.toLowerCase()));
    return [...data].sort((a, b) => {
      const av = a[sortCol] ?? 0;
      const bv = b[sortCol] ?? 0;
      return sortDir === "asc" ? (av > bv ? 1 : -1) : (av < bv ? 1 : -1);
    });
  }, [rows, search, sortCol, sortDir]);

  const totalCalls = filtered.reduce((s, r) => s + (r.total_calls || 0), 0);
  const totalErrors = filtered.reduce((s, r) => s + (r.total_errors || 0), 0);
  const avgLatency = filtered.length ? filtered.reduce((s, r) => s + (r.avg_duration_ms || 0), 0) / filtered.length : 0;

  return (
    <div style={{ display: "flex", flexDirection: "column", gap: "1.25rem" }}>
      {/* Header */}
      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-end" }}>
        <div>
          <h1 style={{ fontSize: "1.4rem", fontWeight: 700, marginBottom: "0.2rem" }}>Transactions</h1>
          <p style={{ color: "var(--text-2)", fontSize: "0.875rem" }}>Per-endpoint latency, error rate, and call volume — sorted by total cost.</p>
        </div>
        <button className="btn btn-secondary btn-sm" onClick={refetch}>↺ Refresh</button>
      </div>

      {/* Summary Cards */}
      <div style={{ display: "grid", gridTemplateColumns: "repeat(3, 1fr)", gap: "1rem" }}>
        {[
          { label: "Total Calls", value: totalCalls.toLocaleString(), icon: "⚡" },
          { label: "Total Errors", value: totalErrors.toLocaleString(), icon: "⚠", danger: totalErrors > 0 },
          { label: "Avg Latency", value: msLabel(avgLatency), icon: "⏱" },
        ].map(({ label, value, icon, danger }) => (
          <div key={label} className="stat-card">
            <div className="stat-icon">{icon}</div>
            <div>
              <div style={{ fontSize: "0.78rem", color: "var(--text-3)", marginBottom: "0.25rem" }}>{label}</div>
              <div style={{ fontSize: "1.6rem", fontWeight: 800, color: danger ? "var(--danger)" : "var(--text-1)" }}>{value}</div>
            </div>
          </div>
        ))}
      </div>

      {/* Filters */}
      <div className="panel" style={{ padding: "0.875rem" }}>
        <div style={{ display: "grid", gridTemplateColumns: "1fr 240px", gap: "0.75rem" }}>
          <div className="form-group" style={{ marginBottom: 0 }}>
            <label className="form-label">Search Transactions</label>
            <input className="form-input" placeholder="e.g. /api/v1/reports or FetchReport" value={search} onChange={e => setSearch(e.target.value)} />
          </div>
          <div className="form-group" style={{ marginBottom: 0 }}>
            <label className="form-label">Service</label>
            <ServiceSelector token={token} selectedServices={selectedServices} onChange={setSelectedServices} />
          </div>
        </div>
      </div>

      {/* Error */}
      {error && <div style={{ padding: "0.875rem", background: "var(--danger-soft)", border: "1px solid rgba(239,68,68,0.25)", borderRadius: "var(--r-md)", color: "var(--danger)", fontSize: "0.875rem" }}>⚠ {error}</div>}

      {/* Table */}
      <div className="panel" style={{ padding: 0, overflow: "hidden" }}>
        {loading ? (
          <div style={{ padding: "3rem" }}><SectionLoader /></div>
        ) : !filtered.length ? (
          <EmptyState icon="⚡" title="No transactions yet" body="Transactions appear here once traces are ingested. Try adding a sample trace from the Traces page." />
        ) : (
          <div style={{ overflowX: "auto" }}>
            <table style={{ width: "100%", borderCollapse: "collapse" }}>
              <thead>
                <tr style={{ borderBottom: "1px solid var(--border)" }}>
                  {COLUMNS.map(col => (
                    <th key={col.key}
                      onClick={() => toggleSort(col.key)}
                      style={{
                        padding: "0.65rem 1rem", textAlign: col.align,
                        fontSize: "0.72rem", fontWeight: 700, textTransform: "uppercase", letterSpacing: "0.07em",
                        color: sortCol === col.key ? "var(--primary-2)" : "var(--text-3)",
                        cursor: "pointer", userSelect: "none", whiteSpace: "nowrap",
                        background: "rgba(0,0,0,0.15)",
                      }}>
                      {col.label} <SortIcon col={col.key} sortCol={sortCol} sortDir={sortDir} />
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {filtered.map((row, i) => {
                  const isSelected = selectedTxn?.operation === row.operation && selectedTxn?.service_name === row.service_name;
                  return (
                    <tr key={i}
                      onClick={() => setSelectedTxn(isSelected ? null : row)}
                      style={{
                        borderBottom: "1px solid var(--border)",
                        cursor: "pointer",
                        background: isSelected ? "rgba(99,102,241,0.1)" : i % 2 === 0 ? "rgba(255,255,255,0.01)" : "transparent",
                        transition: "background 0.15s",
                      }}>
                      <td style={{ padding: "0.75rem 1rem" }}>
                        <div style={{ fontFamily: "var(--font-mono)", fontSize: "0.82rem", fontWeight: 600, color: "var(--primary-2)" }}>{row.operation}</div>
                      </td>
                      <td style={{ padding: "0.75rem 1rem" }}>
                        <span className="badge badge-info" style={{ fontSize: "0.73rem" }}>{row.service_name}</span>
                      </td>
                      <td style={{ padding: "0.75rem 1rem", textAlign: "right", fontFamily: "var(--font-mono)", fontSize: "0.85rem" }}>
                        {(row.total_calls || 0).toLocaleString()}
                      </td>
                      <td style={{ padding: "0.75rem 1rem", textAlign: "right", fontFamily: "var(--font-mono)", fontSize: "0.85rem", color: Number(row.avg_duration_ms) > 1000 ? "var(--warning)" : "var(--text-1)" }}>
                        {msLabel(row.avg_duration_ms)}
                      </td>
                      <td style={{ padding: "0.75rem 1rem", textAlign: "right", fontFamily: "var(--font-mono)", fontSize: "0.85rem", color: Number(row.max_duration_ms) > 5000 ? "var(--danger)" : "var(--text-1)" }}>
                        {msLabel(row.max_duration_ms)}
                      </td>
                      <td style={{ padding: "0.75rem 1rem", textAlign: "right" }}>
                        <ErrorBadge pct={row.error_pct} />
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* Detail Drawer */}
      {selectedTxn && <TxnDetail tx={selectedTxn} token={token} onClose={() => setSelectedTxn(null)} />}
    </div>
  );
}
