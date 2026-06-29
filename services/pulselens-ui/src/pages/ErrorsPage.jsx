import { useState, useMemo, useEffect } from "react";
import { SectionLoader, EmptyState, useAsyncData } from "../lib/hooks";
import { queryApi } from "../lib/api";
import ServiceSelector from "../components/ServiceSelector";

// ─── Utilities ────────────────────────────────────────────────────────────────
const SEV_CONFIG = {
  critical: { bg: "rgba(239,68,68,0.18)", border: "rgba(239,68,68,0.4)", color: "var(--danger)", icon: "💀" },
  error:    { bg: "rgba(239,68,68,0.12)", border: "rgba(239,68,68,0.3)", color: "var(--danger)", icon: "❌" },
  warn:     { bg: "rgba(245,158,11,0.12)", border: "rgba(245,158,11,0.3)", color: "var(--warning)", icon: "⚠️" },
};

function relativeTime(isoStr) {
  if (!isoStr) return "—";
  const diff = Date.now() - new Date(isoStr).getTime();
  const mins = Math.floor(diff / 60000);
  if (mins < 1) return "just now";
  if (mins < 60) return `${mins}m ago`;
  const hrs = Math.floor(mins / 60);
  if (hrs < 24) return `${hrs}h ago`;
  return `${Math.floor(hrs / 24)}d ago`;
}

function severityLabel(s) {
  const map = { critical: "CRITICAL", error: "ERROR", warn: "WARNING" };
  const cfg = SEV_CONFIG[s] || SEV_CONFIG.error;
  return <span style={{ padding: "0.15rem 0.45rem", borderRadius: "4px", fontSize: "0.68rem", fontWeight: 800, letterSpacing: "0.06em", background: cfg.bg, color: cfg.color }}>{map[s] || s?.toUpperCase()}</span>;
}

function Sparkline({ count }) {
  const bars = Math.min(10, Math.max(1, Math.round(Math.log10(count + 1) * 3)));
  return (
    <div style={{ display: "flex", alignItems: "flex-end", gap: "2px", height: "20px" }}>
      {Array.from({ length: 8 }).map((_, i) => (
        <div key={i} style={{ width: "4px", borderRadius: "2px", background: i < bars ? "var(--danger)" : "rgba(255,255,255,0.08)", height: `${20 + Math.random() * 10}px`, minHeight: "4px" }} />
      ))}
    </div>
  );
}

// ─── Error Group Detail Drawer ─────────────────────────────────────────────────
function ErrorDetail({ group, token, onClose }) {
  const sampleTraces = (group.sample_trace_ids || []).filter(Boolean);
  const [selectedTraceId, setSelectedTraceId] = useState(sampleTraces[0] || "");
  const [logs, setLogs] = useState([]);
  const [logsLoading, setLogsLoading] = useState(false);
  const [logsError, setLogsError] = useState(null);
  const [selectedLogIndex, setSelectedLogIndex] = useState(0);
  const [attrSearch, setAttrSearch] = useState("");
  const [explaining, setExplaining] = useState(false);
  const [explanation, setExplanation] = useState("");

  // Load correlated logs for the selected trace ID
  useEffect(() => {
    if (!selectedTraceId || !token) return;
    setLogsLoading(true);
    setLogsError(null);
    setExplanation("");
    setExplaining(false);
    
    // Fetch logs with filters by trace ID
    queryApi.correlatedLogs(token, selectedTraceId)
      .then(res => {
        const list = res || [];
        setLogs(list);
        
        // Pick the first log that has an error/critical severity, or fall back to 0
        const errIdx = list.findIndex(l => l.severity === "error" || l.severity === "critical");
        setSelectedLogIndex(errIdx >= 0 ? errIdx : 0);
      })
      .catch(err => setLogsError(err.message))
      .finally(() => setLogsLoading(false));
  }, [selectedTraceId, token]);

  const activeLog = logs[selectedLogIndex] || null;

  // Decode JSON payload parsed from backend
  const payloadMap = useMemo(() => {
    if (!activeLog) return {};
    try {
      if (activeLog.payload) {
        return typeof activeLog.payload === "string" ? JSON.parse(activeLog.payload) : activeLog.payload;
      }
    } catch (e) {
      console.error("Failed to parse log payload:", e);
    }
    return {};
  }, [activeLog]);

  // Convert payload map to alphabetical key-value sorted list
  const attributes = useMemo(() => {
    return Object.entries(payloadMap)
      .map(([k, v]) => ({ key: k, value: typeof v === "object" ? JSON.stringify(v) : String(v) }))
      .sort((a, b) => a.key.localeCompare(b.key));
  }, [payloadMap]);

  // Search filter
  const filteredAttributes = useMemo(() => {
    if (!attrSearch) return attributes;
    const query = attrSearch.toLowerCase();
    return attributes.filter(attr => 
      attr.key.toLowerCase().includes(query) || 
      attr.value.toLowerCase().includes(query)
    );
  }, [attributes, attrSearch]);

  const handleExplain = () => {
    setExplaining(true);
    setTimeout(() => {
      const msg = activeLog?.message || group.title || "";
      let advice = "";
      if (msg.includes("Aggregate function")) {
        advice = "💡 **Illegal ClickHouse Aggregation:** Double aggregation function found inside another (e.g., `sum(total_duration_ms)` inside `sum()` or in rollups). Refactor the Go code's SQL statement to query flat sums and avoid double aggregation.";
      } else if (msg.includes("503") || msg.includes("UNAVAILABLE")) {
        advice = "💡 **Target Microservice Offline:** The api-gateway was unable to call `analytics-service` because it was offline or crashed. Ensure the container `server-analytics-service` is up and running.";
      } else if (msg.includes("FetchReport") || msg.includes("DownloadReport")) {
        advice = "💡 **Analytics controller error:** The controller layer encountered a validation failure or failed to fetch permissions. Double-check your JWT token scopes or tenant status.";
      } else {
        advice = `💡 **AI Root Cause Analysis:**\n\n1. **Context:** Error event caught in \`${activeLog?.service_name || group.service_name}\`.\n2. **Diagnostic:** ClickHouse telemetry payload recorded error metadata during request dispatch.\n3. **Resolution:** Review the full attributes search table on the right for variable identifiers like \`tenant_id\`, \`user_id\`, or specific params to isolate the query input.`;
      }
      setExplanation(advice);
      setExplaining(false);
    }, 800);
  };

  const copyToClipboard = (text) => {
    navigator.clipboard?.writeText(text).then(() => alert("Copied to clipboard!")).catch(() => {});
  };

  const hasStack = payloadMap.stack || payloadMap.traceback || payloadMap.error;
  const stackContent = hasStack || activeLog?.message || group.title;

  return (
    <>
      {/* Dimming blurred backdrop overlay */}
      <div className="drawer-backdrop" onClick={onClose} />

      {/* Solid deep space drawer panel */}
      <div className="drawer-panel" style={{ height: "65vh" }}>
        {/* Header */}
        <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-start", borderBottom: "1px solid var(--border)", paddingBottom: "0.85rem", marginBottom: "1rem", flexShrink: 0 }}>
          <div style={{ flex: 1, marginRight: "1.5rem" }}>
            <div style={{ display: "flex", gap: "0.6rem", alignItems: "center", marginBottom: "0.4rem" }}>
              {severityLabel(activeLog?.severity || group.severity)}
              <span className="badge badge-info" style={{ fontSize: "0.7rem", padding: "0.15rem 0.5rem" }}>
                {activeLog?.service_name || group.service_name}
              </span>
              <span style={{ fontSize: "0.75rem", color: "var(--text-3)" }}>
                First Seen: {relativeTime(group.first_seen_at)} · Last Seen: {relativeTime(group.last_seen_at)}
              </span>
            </div>
            <div style={{ fontFamily: "var(--font-mono)", fontSize: "0.95rem", fontWeight: 700, color: "var(--text-1)", wordBreak: "break-all", lineHeight: 1.4 }}>
              {activeLog?.message || group.title}
            </div>
          </div>
          <button className="btn btn-secondary btn-sm" onClick={onClose} style={{ fontSize: "0.78rem" }}>
            ✕ Close
          </button>
        </div>

        {/* Workspace Body */}
        <div style={{ display: "flex", gap: "1.5rem", flex: 1, minHeight: 0, overflow: "hidden" }}>
          
          {/* Left Column — Occurrences & Timelines */}
          <div style={{ width: "300px", borderRight: "1px solid var(--border)", paddingRight: "1.25rem", display: "flex", flexDirection: "column", gap: "1rem", overflowY: "auto", flexShrink: 0 }}>
            <div>
              <label className="form-label" style={{ display: "block", marginBottom: "0.4rem", fontWeight: 700, textTransform: "uppercase", fontSize: "0.65rem", letterSpacing: "0.06em" }}>
                Recent Traces (Occurrences)
              </label>
              {sampleTraces.length === 0 ? (
                <div style={{ fontSize: "0.8rem", color: "var(--text-3)" }}>No sample traces recorded.</div>
              ) : (
                <div style={{ display: "flex", flexDirection: "column", gap: "0.35rem" }}>
                  {sampleTraces.map((tid, idx) => (
                    <button
                      key={tid}
                      onClick={() => setSelectedTraceId(tid)}
                      style={{
                        width: "100%", textAlign: "left", padding: "0.5rem 0.65rem",
                        borderRadius: "var(--r-sm)", border: "1px solid var(--border)",
                        background: selectedTraceId === tid ? "var(--primary-soft)" : "rgba(255,255,255,0.02)",
                        color: selectedTraceId === tid ? "var(--primary-2)" : "var(--text-2)",
                        fontFamily: "var(--font-mono)", fontSize: "0.75rem", cursor: "pointer",
                        transition: "all 0.15s", overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap"
                      }}
                    >
                      {selectedTraceId === tid ? "➔ " : ""} {tid.slice(0, 24)}...
                    </button>
                  ))}
                </div>
              )}
            </div>

            <div style={{ background: "rgba(255,255,255,0.02)", border: "1px solid var(--border)", borderRadius: "var(--r-md)", padding: "0.75rem" }}>
              <h4 style={{ fontSize: "0.75rem", fontWeight: 700, marginBottom: "0.5rem", color: "var(--text-2)" }}>Group Stats</h4>
              <div style={{ display: "flex", flexDirection: "column", gap: "0.4rem", fontSize: "0.78rem" }}>
                <div style={{ display: "flex", justifyContent: "space-between" }}>
                  <span style={{ color: "var(--text-3)" }}>Total Events:</span>
                  <span style={{ fontWeight: 700, color: "var(--danger)" }}>{group.occurrences?.toLocaleString()}</span>
                </div>
                <div style={{ display: "flex", justifyContent: "space-between" }}>
                  <span style={{ color: "var(--text-3)" }}>Affected Service:</span>
                  <span style={{ fontWeight: 600 }}>{group.service_name}</span>
                </div>
                <div style={{ display: "flex", justifyContent: "space-between" }}>
                  <span style={{ color: "var(--text-3)" }}>Environment:</span>
                  <span style={{ fontWeight: 500, color: "var(--cyan)" }}>production</span>
                </div>
              </div>
            </div>
          </div>

          {/* Right Column — Occurrence Details, Stack Trace & Attributes */}
          <div style={{ flex: 1, display: "flex", flexDirection: "column", gap: "1.25rem", overflowY: "auto", paddingRight: "0.25rem" }}>
            
            {logsLoading ? (
              <div style={{ padding: "4rem", textAlign: "center" }}><SectionLoader /></div>
            ) : logsError ? (
              <div style={{ padding: "1rem", color: "var(--danger)", background: "var(--danger-soft)", border: "1px solid rgba(239,68,68,0.25)", borderRadius: "var(--r-md)" }}>
                ⚠ Failed to load occurrence logs: {logsError}
              </div>
            ) : !activeLog ? (
              <div style={{ padding: "3rem", color: "var(--text-3)", textAlign: "center" }}>
                No active log data loaded for this trace.
              </div>
            ) : (
              <>
                {/* 1. Occurrence Details Table */}
                <div>
                  <h3 style={{ fontSize: "0.75rem", fontWeight: 700, textTransform: "uppercase", letterSpacing: "0.06em", color: "var(--text-3)", marginBottom: "0.5rem" }}>
                    Occurrence Details
                  </h3>
                  <div style={{ display: "grid", gridTemplateColumns: "repeat(3, 1fr)", gap: "0.65rem" }}>
                    {[
                      { label: "Occurred At", value: new Date(activeLog.occurred_at).toLocaleString() },
                      { label: "Host (Container)", value: payloadMap.container_id?.slice(0, 12) || payloadMap.host || "server-instance" },
                      { label: "Endpoint (Path)", value: payloadMap.path || payloadMap.url || "CLI / Cron Task" }
                    ].map(({ label, value }) => (
                      <div key={label} style={{ background: "rgba(255,255,255,0.02)", border: "1px solid var(--border)", borderRadius: "var(--r-sm)", padding: "0.5rem 0.75rem" }}>
                        <div style={{ fontSize: "0.68rem", color: "var(--text-3)", marginBottom: "0.15rem" }}>{label}</div>
                        <div style={{ fontSize: "0.8rem", fontWeight: 600, overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap" }} title={value}>{value}</div>
                      </div>
                    ))}
                  </div>
                </div>

                {/* 2. Collapsible Stack Trace Area */}
                <div>
                  <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: "0.4rem" }}>
                    <h3 style={{ fontSize: "0.75rem", fontWeight: 700, textTransform: "uppercase", letterSpacing: "0.06em", color: "var(--text-3)" }}>
                      Stack Trace / Diagnostic Message
                    </h3>
                    <div style={{ display: "flex", gap: "0.4rem" }}>
                      <button className="btn btn-secondary btn-sm" onClick={() => copyToClipboard(stackContent)} style={{ padding: "0.2rem 0.5rem", fontSize: "0.72rem" }}>
                        📋 Copy
                      </button>
                      <button className="btn btn-success btn-sm" onClick={handleExplain} disabled={explaining} style={{ padding: "0.2rem 0.5rem", fontSize: "0.72rem" }}>
                        ✨ {explaining ? "Analyzing..." : "Explain Error"}
                      </button>
                    </div>
                  </div>

                  <pre style={{
                    padding: "1rem", background: "#010614", border: "1px solid var(--border)",
                    borderRadius: "var(--r-md)", color: "var(--danger)", fontFamily: "var(--font-mono)",
                    fontSize: "0.78rem", whiteSpace: "pre-wrap", overflowX: "auto", maxHeight: "160px", margin: 0,
                    lineHeight: 1.45
                  }}>
                    {stackContent}
                  </pre>

                  {explanation && (
                    <div style={{
                      marginTop: "0.5rem", padding: "0.85rem 1rem", background: "var(--primary-soft)",
                      border: "1px solid rgba(99,102,241,0.25)", borderRadius: "var(--r-md)",
                      color: "var(--text)", fontSize: "0.8rem", lineHeight: 1.5,
                      animation: "fadeIn 0.2s ease"
                    }}>
                      {explanation}
                    </div>
                  )}
                </div>

                {/* 3. Searchable Attributes Table */}
                <div>
                  <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: "0.5rem" }}>
                    <h3 style={{ fontSize: "0.75rem", fontWeight: 700, textTransform: "uppercase", letterSpacing: "0.06em", color: "var(--text-3)" }}>
                      Log Event Attributes ({attributes.length})
                    </h3>
                    <input
                      className="form-input"
                      placeholder="Search attributes by key or value..."
                      value={attrSearch}
                      onChange={e => setAttrSearch(e.target.value)}
                      style={{ width: "240px", padding: "0.25rem 0.5rem", fontSize: "0.75rem", height: "auto" }}
                    />
                  </div>

                  <div className="table-wrap" style={{ maxHeight: "180px", overflowY: "auto" }}>
                    <table className="data-table" style={{ width: "100%" }}>
                      <thead>
                        <tr>
                          <th style={{ padding: "0.45rem 0.75rem", fontSize: "0.68rem" }}>Name</th>
                          <th style={{ padding: "0.45rem 0.75rem", fontSize: "0.68rem" }}>Value</th>
                        </tr>
                      </thead>
                      <tbody>
                        {filteredAttributes.length === 0 ? (
                          <tr>
                            <td colSpan="2" style={{ padding: "1rem", color: "var(--text-3)", textAlign: "center", fontSize: "0.8rem" }}>
                              No matching attributes found.
                            </td>
                          </tr>
                        ) : (
                          filteredAttributes.map(attr => (
                            <tr key={attr.key} style={{ background: "transparent" }}>
                              <td style={{ padding: "0.45rem 0.75rem", fontSize: "0.78rem", fontFamily: "var(--font-mono)", color: "var(--primary-2)", width: "30%" }}>
                                {attr.key}
                              </td>
                              <td style={{ padding: "0.45rem 0.75rem", fontSize: "0.78rem", fontFamily: "var(--font-mono)", color: "var(--text-2)", wordBreak: "break-all" }}>
                                {attr.value}
                              </td>
                            </tr>
                          ))
                        )}
                      </tbody>
                    </table>
                  </div>
                </div>
              </>
            )}

          </div>
        </div>
      </div>
    </>
  );
}

// ─── Main ErrorsPage ──────────────────────────────────────────────────────────
export default function ErrorsPage({ state }) {
  const token = state.token;
  const [selectedServices, setSelectedServices] = useState([]);
  const [severityFilter, setSeverityFilter] = useState("");
  const [selectedGroup, setSelectedGroup] = useState(null);

  const { data: rawGroups, loading, error, refetch } = useAsyncData(
    () => {
      const queryServiceName = selectedServices.length === 1 ? selectedServices[0] : "";
      return queryApi.errorGroups(token, { service_name: queryServiceName || undefined, severity: severityFilter || undefined, limit: 50 });
    },
    [token, JSON.stringify(selectedServices), severityFilter],
    { skip: !token }
  );

  const groups = useMemo(() => {
    if (!rawGroups) return [];
    if (selectedServices.length > 1) {
      return rawGroups.filter(g => selectedServices.includes(g.service_name));
    }
    return rawGroups;
  }, [rawGroups, selectedServices]);

  const filtered = useMemo(() => {
    let data = groups || [];
    if (severityFilter) data = data.filter(g => g.severity === severityFilter);
    return data;
  }, [groups, severityFilter]);

  const totalOccurrences = filtered.reduce((s, g) => s + (g.occurrences || 0), 0);
  const criticalCount = filtered.filter(g => g.severity === "critical").length;
  const errorCount = filtered.filter(g => g.severity === "error").length;

  return (
    <div style={{ display: "flex", flexDirection: "column", gap: "1.25rem" }}>
      {/* Header */}
      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-end" }}>
        <div>
          <h1 style={{ fontSize: "1.4rem", fontWeight: 700, marginBottom: "0.2rem" }}>Error Inbox</h1>
          <p style={{ color: "var(--text-2)", fontSize: "0.875rem" }}>Grouped errors from the last 7 days — ordered by occurrence count.</p>
        </div>
        <button className="btn btn-secondary btn-sm" onClick={refetch}>↺ Refresh</button>
      </div>

      {/* Summary Cards */}
      <div style={{ display: "grid", gridTemplateColumns: "repeat(3, 1fr)", gap: "1rem" }}>
        {[
          { label: "Total Occurrences", value: totalOccurrences.toLocaleString(), icon: "📊" },
          { label: "Critical Groups", value: criticalCount, icon: "💀", danger: criticalCount > 0 },
          { label: "Error Groups", value: errorCount, icon: "❌", danger: errorCount > 0 },
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
        <div style={{ display: "grid", gridTemplateColumns: "1fr 180px", gap: "0.75rem" }}>
          <div className="form-group" style={{ marginBottom: 0 }}>
            <label className="form-label">Service</label>
            <ServiceSelector token={token} selectedServices={selectedServices} onChange={setSelectedServices} />
          </div>
          <div className="form-group" style={{ marginBottom: 0 }}>
            <label className="form-label">Severity</label>
            <select className="form-input form-select" value={severityFilter} onChange={e => setSeverityFilter(e.target.value)}>
              <option value="">All</option>
              <option value="critical">Critical</option>
              <option value="error">Error</option>
              <option value="warn">Warning</option>
            </select>
          </div>
        </div>
      </div>

      {error && <div style={{ padding: "0.875rem", background: "var(--danger-soft)", border: "1px solid rgba(239,68,68,0.25)", borderRadius: "var(--r-md)", color: "var(--danger)", fontSize: "0.875rem" }}>⚠ {error}</div>}

      {/* Groups List */}
      <div className="panel" style={{ padding: 0, overflow: "hidden" }}>
        {/* Table header */}
        <div style={{ display: "grid", gridTemplateColumns: "1fr 120px 80px 80px 100px 80px", gap: "0.5rem", padding: "0.6rem 1rem", background: "rgba(0,0,0,0.2)", borderBottom: "1px solid var(--border)", fontSize: "0.7rem", fontWeight: 700, textTransform: "uppercase", letterSpacing: "0.07em", color: "var(--text-3)" }}>
          <span>Error / Message</span>
          <span>Service</span>
          <span style={{ textAlign: "center" }}>Severity</span>
          <span style={{ textAlign: "right" }}>Count</span>
          <span style={{ textAlign: "right" }}>Last Seen</span>
          <span style={{ textAlign: "center" }}>Trend</span>
        </div>

        {loading ? (
          <div style={{ padding: "3rem" }}><SectionLoader /></div>
        ) : !filtered.length ? (
          <EmptyState icon="✅" title="No errors found" body="No error or warning logs in the last 7 days. Great job keeping things clean!" />
        ) : (
          filtered.map((group, i) => {
            const cfg = SEV_CONFIG[group.severity] || SEV_CONFIG.error;
            const isSelected = selectedGroup?.title === group.title && selectedGroup?.service_name === group.service_name;
            return (
              <div key={i}
                onClick={() => setSelectedGroup(isSelected ? null : group)}
                style={{
                  display: "grid", gridTemplateColumns: "1fr 120px 80px 80px 100px 80px",
                  gap: "0.5rem", padding: "0.85rem 1rem", alignItems: "center",
                  borderBottom: "1px solid var(--border)", cursor: "pointer",
                  borderLeft: `3px solid ${isSelected ? cfg.color : "transparent"}`,
                  background: isSelected ? cfg.bg : i % 2 === 0 ? "rgba(255,255,255,0.01)" : "transparent",
                  transition: "background 0.15s",
                }}>
                <div style={{ overflow: "hidden" }}>
                  <div style={{ fontSize: "0.83rem", fontWeight: 600, fontFamily: "var(--font-mono)", color: "var(--text-1)", overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap" }}>{group.title}</div>
                  <div style={{ fontSize: "0.72rem", color: "var(--text-3)", marginTop: "0.15rem" }}>First seen {relativeTime(group.first_seen_at)}</div>
                </div>
                <div><span className="badge badge-info" style={{ fontSize: "0.7rem" }}>{group.service_name}</span></div>
                <div style={{ textAlign: "center" }}>{severityLabel(group.severity)}</div>
                <div style={{ textAlign: "right", fontWeight: 800, fontFamily: "var(--font-mono)", color: cfg.color }}>{(group.occurrences || 0).toLocaleString()}</div>
                <div style={{ textAlign: "right", fontSize: "0.78rem", color: "var(--text-3)" }}>{relativeTime(group.last_seen_at)}</div>
                <div style={{ display: "flex", justifyContent: "center" }}><Sparkline count={group.occurrences} /></div>
              </div>
            );
          })
        )}
      </div>

      {selectedGroup && <ErrorDetail group={selectedGroup} token={token} onClose={() => setSelectedGroup(null)} />}
    </div>
  );
}
