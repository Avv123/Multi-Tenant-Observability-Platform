// IncidentsPage — enterprise-grade incident management with stepper timeline,
// card layout, and modal-style detail view.

import { useMemo, useState } from "react";
import { EmptyState, SectionLoader, useAsyncData } from "../lib/hooks";
import { alertingApi } from "../lib/api";

// ─── Utilities ───────────────────────────────────────────────────────────────
function relTime(iso) {
  if (!iso) return "—";
  const diff = Date.now() - new Date(iso).getTime();
  const mins = Math.floor(diff / 60000);
  if (mins < 1) return "just now";
  if (mins < 60) return `${mins}m ago`;
  const hrs = Math.floor(mins / 60);
  if (hrs < 24) return `${hrs}h ago`;
  return `${Math.floor(hrs / 24)}d ago`;
}
function fmtDate(iso) {
  if (!iso) return "—";
  return new Date(iso).toLocaleString("en", { month: "short", day: "numeric", hour: "2-digit", minute: "2-digit", hour12: false });
}

const SEV = {
  critical: { bg: "rgba(239,68,68,0.15)", border: "rgba(239,68,68,0.45)", text: "var(--danger)", dot: "#ef4444" },
  error:    { bg: "rgba(239,68,68,0.10)", border: "rgba(239,68,68,0.3)",  text: "var(--danger)", dot: "#f87171" },
  warning:  { bg: "rgba(245,158,11,0.10)", border: "rgba(245,158,11,0.35)", text: "var(--warning)", dot: "#f59e0b" },
};
const STATUS_CFG = {
  open:         { label: "OPEN",         color: "var(--danger)",  cls: "badge-danger"  },
  acknowledged: { label: "IN PROGRESS",  color: "var(--warning)", cls: "badge-warning" },
  resolved:     { label: "RESOLVED",     color: "var(--success)", cls: "badge-success" },
};

function SevBadge({ sev }) {
  const c = SEV[sev] || SEV.warning;
  return (
    <span style={{ padding: "0.2rem 0.5rem", borderRadius: "4px", fontSize: "0.68rem", fontWeight: 800, letterSpacing: "0.07em", background: c.bg, color: c.text, border: `1px solid ${c.border}` }}>
      {sev?.toUpperCase() || "WARNING"}
    </span>
  );
}
function StatusBadge({ status }) {
  const c = STATUS_CFG[status] || STATUS_CFG.open;
  return <span className={`badge ${c.cls}`} style={{ fontSize: "0.7rem", letterSpacing: "0.04em" }}>{c.label}</span>;
}

// ─── Incident Card ────────────────────────────────────────────────────────────
function IncidentCard({ inc, isSelected, onClick, onAck, onResolve }) {
  const sev = SEV[inc.severity] || SEV.warning;
  return (
    <div
      onClick={onClick}
      style={{
        borderRadius: "var(--r-md)", border: `1px solid ${isSelected ? "var(--primary-2)" : sev.border}`,
        background: isSelected ? "rgba(99,102,241,0.1)" : sev.bg,
        padding: "1rem 1.1rem", cursor: "pointer", transition: "all 0.18s",
        boxShadow: isSelected ? "0 0 0 2px rgba(99,102,241,0.35)" : "none",
        borderLeft: `4px solid ${sev.dot}`,
      }}
    >
      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-start", marginBottom: "0.5rem" }}>
        <div style={{ flex: 1, marginRight: "0.75rem" }}>
          <div style={{ fontWeight: 700, fontSize: "0.9rem", lineHeight: 1.35, marginBottom: "0.25rem" }}>
            {inc.title || inc.summary || "Untitled Incident"}
          </div>
          {inc.summary && inc.title && inc.summary !== inc.title && (
            <div style={{ fontSize: "0.78rem", color: "var(--text-2)", lineHeight: 1.4 }}>{inc.summary}</div>
          )}
        </div>
        <div style={{ display: "flex", flexDirection: "column", alignItems: "flex-end", gap: "0.3rem", flexShrink: 0 }}>
          <SevBadge sev={inc.severity} />
          <StatusBadge status={inc.status} />
        </div>
      </div>

      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center" }}>
        <div style={{ display: "flex", gap: "1rem", fontSize: "0.75rem", color: "var(--text-3)" }}>
          <span>⏱ {relTime(inc.triggered_at)}</span>
          {inc.assigned_to && <span>👤 {inc.assigned_to}</span>}
          {inc.observed_value != null && (
            <span style={{ color: sev.text }}>
              {Number(inc.observed_value).toFixed(2)} <span style={{ color: "var(--text-3)" }}>/ {inc.threshold}</span>
            </span>
          )}
        </div>

        <div style={{ display: "flex", gap: "0.4rem" }} onClick={e => e.stopPropagation()}>
          {inc.status === "open" && (
            <button className="btn btn-ghost btn-sm" style={{ fontSize: "0.72rem", padding: "0.2rem 0.6rem" }} onClick={() => onAck(inc.id)}>Acknowledge</button>
          )}
          {inc.status !== "resolved" && (
            <button className="btn btn-success btn-sm" style={{ fontSize: "0.72rem", padding: "0.2rem 0.6rem" }} onClick={() => onResolve(inc.id)}>✓ Resolve</button>
          )}
        </div>
      </div>
    </div>
  );
}

// ─── Stepper Timeline ─────────────────────────────────────────────────────────
const TIMELINE_ICONS = {
  triggered:    { icon: "🚨", color: "var(--danger)"  },
  acknowledged: { icon: "👁",  color: "var(--warning)" },
  resolved:     { icon: "✅",  color: "var(--success)" },
  commented:    { icon: "💬",  color: "var(--cyan)"    },
  assigned:     { icon: "👤",  color: "var(--primary-2)" },
  escalated:    { icon: "📣",  color: "var(--warning)" },
};

function StepperTimeline({ events, loading }) {
  if (loading) return <SectionLoader />;
  if (!events?.length) return <div style={{ color: "var(--text-3)", fontSize: "0.82rem" }}>No timeline events recorded.</div>;

  return (
    <div style={{ position: "relative", paddingLeft: "1.5rem" }}>
      {/* Vertical line */}
      <div style={{ position: "absolute", left: "7px", top: "10px", bottom: "10px", width: "2px", background: "var(--border)", borderRadius: "2px" }} />

      {events.map((ev, i) => {
        const cfg = TIMELINE_ICONS[ev.event_type] || { icon: "•", color: "var(--text-3)" };
        const isLast = i === events.length - 1;
        return (
          <div key={ev.id || i} style={{ display: "flex", gap: "0.75rem", marginBottom: isLast ? 0 : "1.1rem", position: "relative" }}>
            {/* Dot */}
            <div style={{
              position: "absolute", left: "-1.5rem", top: "2px",
              width: "16px", height: "16px", borderRadius: "50%",
              background: "var(--bg)", border: `2px solid ${cfg.color}`,
              display: "flex", alignItems: "center", justifyContent: "center",
              fontSize: "0.55rem", zIndex: 1,
            }}>{cfg.icon}</div>

            <div style={{ flex: 1 }}>
              <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-start" }}>
                <span style={{ fontSize: "0.82rem", fontWeight: 600, color: cfg.color, textTransform: "capitalize" }}>
                  {ev.event_type?.replace(/_/g, " ") || "event"}
                </span>
                <span style={{ fontSize: "0.72rem", color: "var(--text-3)" }}>{fmtDate(ev.created_at)}</span>
              </div>
              {ev.summary && (
                <div style={{ fontSize: "0.78rem", color: "var(--text-2)", marginTop: "0.2rem", lineHeight: 1.4 }}>{ev.summary}</div>
              )}
            </div>
          </div>
        );
      })}
    </div>
  );
}

// ─── Detail Panel ─────────────────────────────────────────────────────────────
function DetailPanel({ incidentId, token, onClose, onAck, onResolve, onAssign, onComment }) {
  const [commentBody, setCommentBody] = useState("");
  const [assignTo, setAssignTo] = useState("");
  const [tab, setTab] = useState("timeline"); // "timeline" | "comments"

  const { data: detail, loading: detailLoading, refetch: refetchDetail } = useAsyncData(
    () => alertingApi.getIncident(token, incidentId),
    [token, incidentId], { skip: !token || !incidentId }
  );
  const { data: timeline, loading: timelineLoading } = useAsyncData(
    () => alertingApi.incidentTimeline(token, incidentId),
    [token, incidentId], { skip: !token || !incidentId }
  );
  const { data: comments, loading: commentsLoading, refetch: refetchComments } = useAsyncData(
    () => alertingApi.listIncidentComments(token, incidentId),
    [token, incidentId], { skip: !token || !incidentId }
  );

  async function submitComment(e) {
    e.preventDefault();
    if (!commentBody.trim()) return;
    await onComment(incidentId, commentBody);
    setCommentBody("");
    refetchComments();
  }
  async function submitAssign(e) {
    e.preventDefault();
    if (!assignTo.trim()) return;
    await onAssign(incidentId, assignTo);
    setAssignTo("");
    refetchDetail();
  }

  if (detailLoading) return <div className="panel" style={{ padding: "2rem" }}><SectionLoader /></div>;
  if (!detail) return null;

  const sev = SEV[detail.severity] || SEV.warning;

  return (
    <div style={{ padding: 0, overflow: "hidden", display: "flex", flexDirection: "column", height: "100%", background: "#0b1120", border: "1px solid var(--border)", borderRadius: "var(--r-md)" }}>
      {/* Header */}
      <div style={{ padding: "1rem 1.25rem", borderBottom: "1px solid var(--border)", background: sev.bg, borderLeft: `4px solid ${sev.dot}`, flexShrink: 0 }}>
        <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-start" }}>
          <div style={{ flex: 1, marginRight: "1rem" }}>
            <div style={{ fontWeight: 700, fontSize: "0.95rem", marginBottom: "0.35rem" }}>{detail.title || detail.summary}</div>
            <div style={{ display: "flex", gap: "0.5rem", flexWrap: "wrap" }}>
              <SevBadge sev={detail.severity} />
              <StatusBadge status={detail.status} />
              {detail.assigned_to && <span className="badge badge-info">👤 {detail.assigned_to}</span>}
            </div>
          </div>
          <button className="btn btn-ghost btn-sm" onClick={onClose} style={{ fontSize: "1.1rem" }}>✕</button>
        </div>
      </div>

      {/* Stats row */}
      <div style={{ display: "grid", gridTemplateColumns: "repeat(3, 1fr)", gap: 0, borderBottom: "1px solid var(--border)", flexShrink: 0 }}>
        {[
          { label: "Triggered", value: fmtDate(detail.triggered_at) },
          { label: "Value / Threshold", value: detail.observed_value != null ? `${Number(detail.observed_value).toFixed(2)} / ${detail.threshold}` : "—" },
          { label: "Rule", value: detail.alert_rule_id ? `#${detail.alert_rule_id.slice(-6)}` : "—" },
        ].map(({ label, value }, i) => (
          <div key={label} style={{ padding: "0.65rem 1rem", borderRight: i < 2 ? "1px solid var(--border)" : "none", textAlign: "center" }}>
            <div style={{ fontSize: "0.68rem", color: "var(--text-3)", fontWeight: 700, textTransform: "uppercase", letterSpacing: "0.07em" }}>{label}</div>
            <div style={{ fontSize: "0.82rem", fontWeight: 600, marginTop: "0.2rem", fontFamily: "var(--font-mono)" }}>{value}</div>
          </div>
        ))}
      </div>

      {/* Action bar */}
      <div style={{ display: "flex", gap: "0.5rem", padding: "0.65rem 1rem", borderBottom: "1px solid var(--border)", flexShrink: 0 }}>
        {detail.status === "open" && (
          <button className="btn btn-secondary btn-sm" onClick={() => onAck(detail.id)}>👁 Acknowledge</button>
        )}
        {detail.status !== "resolved" && (
          <button className="btn btn-success btn-sm" onClick={() => onResolve(detail.id)}>✓ Mark Resolved</button>
        )}

        {/* Inline assign */}
        <form onSubmit={submitAssign} style={{ display: "flex", gap: "0.35rem", marginLeft: "auto" }}>
          <input className="form-input" style={{ padding: "0.3rem 0.6rem", fontSize: "0.78rem", height: "auto", width: "160px" }}
            placeholder="Assign to…" value={assignTo} onChange={e => setAssignTo(e.target.value)} />
          <button type="submit" className="btn btn-ghost btn-sm">Assign</button>
        </form>
      </div>

      {/* Tabs */}
      <div style={{ display: "flex", borderBottom: "1px solid var(--border)", flexShrink: 0 }}>
        {[["timeline", "Timeline"], ["comments", `Comments${comments?.length ? ` (${comments.length})` : ""}`]].map(([t, label]) => (
          <button key={t} onClick={() => setTab(t)}
            style={{ flex: 1, padding: "0.6rem", fontSize: "0.78rem", fontWeight: tab === t ? 700 : 400, color: tab === t ? "var(--primary-2)" : "var(--text-3)", borderBottom: tab === t ? "2px solid var(--primary-2)" : "2px solid transparent", background: "none", cursor: "pointer" }}>
            {label}
          </button>
        ))}
      </div>

      {/* Tab Content */}
      <div style={{ flex: 1, overflowY: "auto", padding: "1rem 1.25rem" }}>
        {tab === "timeline" && (
          <StepperTimeline events={timeline} loading={timelineLoading} />
        )}

        {tab === "comments" && (
          <div style={{ display: "flex", flexDirection: "column", gap: "0.75rem" }}>
            {commentsLoading ? <SectionLoader /> : comments?.length ? (
              comments.map((c, i) => (
                <div key={c.id || i} style={{ background: "rgba(255,255,255,0.04)", border: "1px solid var(--border)", borderRadius: "var(--r-md)", padding: "0.7rem 0.875rem" }}>
                  <div style={{ fontSize: "0.7rem", color: "var(--text-3)", marginBottom: "0.3rem" }}>{fmtDate(c.created_at)}</div>
                  <div style={{ fontSize: "0.83rem", color: "var(--text-1)", lineHeight: 1.5 }}>{c.body}</div>
                </div>
              ))
            ) : <div style={{ color: "var(--text-3)", fontSize: "0.83rem" }}>No comments yet.</div>}

            <form onSubmit={submitComment} style={{ display: "flex", flexDirection: "column", gap: "0.5rem", marginTop: "0.5rem" }}>
              <label className="form-label">Add comment</label>
              <textarea className="form-input" rows={3} value={commentBody}
                onChange={e => setCommentBody(e.target.value)}
                placeholder="What did you investigate? What did you find?" />
              <button type="submit" className="btn btn-primary btn-sm" style={{ alignSelf: "flex-end" }}>Submit</button>
            </form>
          </div>
        )}
      </div>
    </div>
  );
}

// ─── Main IncidentsPage ───────────────────────────────────────────────────────
export default function IncidentsPage({ state, notify }) {
  const token = state.token;
  const [filters, setFilters] = useState({ status: "", severity: "" });
  const [selectedId, setSelectedId] = useState(null);
  const [groupBy, setGroupBy] = useState(false);
  const [expandedGroups, setExpandedGroups] = useState({});

  const { data: incidents, loading, error, refetch } = useAsyncData(
    () => alertingApi.listIncidents(token, filters),
    [token, JSON.stringify(filters)],
    { skip: !token }
  );

  async function handleAck(id) {
    try { await alertingApi.acknowledgeIncident(token, id); notify("Acknowledged.", "success"); refetch(); }
    catch (err) { notify(err.message, "error"); }
  }
  async function handleResolve(id) {
    try { await alertingApi.resolveIncident(token, id); notify("Resolved.", "success"); refetch(); if (selectedId === id) setSelectedId(null); }
    catch (err) { notify(err.message, "error"); }
  }
  async function handleAckAll(incidentIds) {
    try {
      await Promise.all(incidentIds.map(id => alertingApi.acknowledgeIncident(token, id)));
      notify("All incidents in group acknowledged.", "success");
      refetch();
    } catch (err) { notify(err.message, "error"); }
  }
  async function handleResolveAll(incidentIds) {
    try {
      await Promise.all(incidentIds.map(id => alertingApi.resolveIncident(token, id)));
      notify("All incidents in group resolved.", "success");
      refetch();
      setSelectedId(null);
    } catch (err) { notify(err.message, "error"); }
  }
  async function handleAssign(id, to) {
    try { await alertingApi.assignIncident(token, id, { assigned_to: to }); notify("Assigned.", "success"); }
    catch (err) { notify(err.message, "error"); }
  }
  async function handleComment(id, body) {
    try { await alertingApi.addIncidentComment(token, id, { body }); notify("Comment added.", "success"); }
    catch (err) { notify(err.message, "error"); }
  }

  const grouped = useMemo(() => {
    if (!incidents || !groupBy) return null;
    const map = {};
    incidents.forEach(inc => {
      const key = inc.title || inc.summary || "Unknown";
      if (!map[key]) map[key] = { title: key, count: 0, severity: "warning", status: "resolved", incidents: [], latest: 0 };
      const g = map[key];
      g.count++;
      g.incidents.push(inc);
      const t = new Date(inc.triggered_at).getTime();
      if (t > g.latest) g.latest = t;
      if (inc.status === "open") g.status = "open";
      else if (inc.status === "acknowledged" && g.status !== "open") g.status = "acknowledged";
      if (inc.severity === "critical") g.severity = "critical";
      else if (inc.severity === "error" && g.severity !== "critical") g.severity = "error";
    });
    return Object.values(map).sort((a, b) => b.latest - a.latest);
  }, [incidents, groupBy]);

  const openCount = incidents?.filter(i => i.status === "open").length || 0;
  const ackCount = incidents?.filter(i => i.status === "acknowledged").length || 0;
  const criticalCount = incidents?.filter(i => i.severity === "critical").length || 0;

  return (
    <div style={{ display: "flex", flexDirection: "column", gap: "1.25rem", height: "100%" }}>
      {/* Header */}
      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-end" }}>
        <div>
          <h1 style={{ fontSize: "1.4rem", fontWeight: 700, marginBottom: "0.2rem" }}>Active Incidents</h1>
          <p style={{ color: "var(--text-2)", fontSize: "0.875rem" }}>Acknowledge, assign and resolve triggered alert incidents.</p>
        </div>
        <button className="btn btn-secondary btn-sm" onClick={refetch}>↺ Refresh</button>
      </div>

      {/* Summary strip */}
      <div style={{ display: "grid", gridTemplateColumns: "repeat(3, 1fr)", gap: "0.75rem" }}>
        {[
          { label: "Open", value: openCount, danger: openCount > 0, icon: "🚨" },
          { label: "In Progress", value: ackCount, warn: ackCount > 0, icon: "👁" },
          { label: "Critical", value: criticalCount, danger: criticalCount > 0, icon: "💀" },
        ].map(({ label, value, danger, warn, icon }) => (
          <div key={label} className="stat-card" style={{ padding: "0.75rem 1rem" }}>
            <div className="stat-icon" style={{ fontSize: "1.2rem" }}>{icon}</div>
            <div>
              <div style={{ fontSize: "0.75rem", color: "var(--text-3)" }}>{label}</div>
              <div style={{ fontSize: "1.5rem", fontWeight: 800, color: danger ? "var(--danger)" : warn ? "var(--warning)" : "var(--text-1)" }}>{value}</div>
            </div>
          </div>
        ))}
      </div>

      {/* Filters + group toggle */}
      <div className="panel" style={{ padding: "0.875rem" }}>
        <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr auto", gap: "0.75rem", alignItems: "flex-end" }}>
          <div className="form-group" style={{ marginBottom: 0 }}>
            <label className="form-label">Status</label>
            <select className="form-input form-select" value={filters.status} onChange={e => setFilters(f => ({ ...f, status: e.target.value }))}>
              <option value="">All</option>
              {["open", "acknowledged", "resolved"].map(s => <option key={s} value={s}>{s.charAt(0).toUpperCase() + s.slice(1)}</option>)}
            </select>
          </div>
          <div className="form-group" style={{ marginBottom: 0 }}>
            <label className="form-label">Severity</label>
            <select className="form-input form-select" value={filters.severity} onChange={e => setFilters(f => ({ ...f, severity: e.target.value }))}>
              <option value="">All</option>
              {["warning", "error", "critical"].map(s => <option key={s} value={s}>{s.charAt(0).toUpperCase() + s.slice(1)}</option>)}
            </select>
          </div>
          <button className={`btn btn-sm ${groupBy ? "btn-primary" : "btn-secondary"}`} onClick={() => setGroupBy(g => !g)}>
            {groupBy ? "📁 Grouped" : "📋 List"}
          </button>
        </div>
      </div>

      {error && <div style={{ padding: "0.875rem", background: "var(--danger-soft)", border: "1px solid rgba(239,68,68,0.25)", borderRadius: "var(--r-md)", color: "var(--danger)", fontSize: "0.875rem" }}>⚠ {error}</div>}

      {/* Main content — split layout when detail open */}
      <div style={{ display: "grid", gridTemplateColumns: selectedId ? "1fr 420px" : "1fr", gap: "1rem", flex: 1, minHeight: 0 }}>

        {/* Incident list */}
        <div style={{ overflowY: "auto", display: "flex", flexDirection: "column", gap: "0.65rem" }}>
          {loading ? (
            <div className="panel" style={{ padding: "3rem" }}><SectionLoader /></div>
          ) : !incidents?.length ? (
            <EmptyState icon="✅" title="No incidents" body="No incidents match the current filters. Create an alert rule to start monitoring." />
          ) : groupBy && grouped ? (
            grouped.map(g => {
              const isExpanded = !!expandedGroups[g.title];
              const nestedIds = g.incidents.map(i => i.id);
              const anyOpen = g.incidents.some(i => i.status === "open");
              const anyUnresolved = g.incidents.some(i => i.status !== "resolved");

              return (
                <div
                  key={g.title}
                  style={{
                    padding: "1rem 1.25rem",
                    borderLeft: `4px solid ${SEV[g.severity]?.dot || "#f59e0b"}`,
                    background: "rgba(255,255,255,0.02)",
                    border: "1px solid var(--border)",
                    borderRadius: "var(--r-md)",
                    display: "flex",
                    flexDirection: "column",
                    gap: "0.65rem",
                    // Stacked card deck layered shadow aesthetic!
                    boxShadow: "0 4px 0 -2px #0a0f1d, 0 4px 0 -1px var(--border), 0 8px 0 -4px #0a0f1d, 0 8px 0 -3px var(--border), var(--shadow-md)",
                    marginBottom: "10px",
                    transition: "all 0.15s ease",
                  }}
                >
                  {/* Group Header Info */}
                  <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-start" }}>
                    <div style={{ flex: 1, marginRight: "1rem" }}>
                      <div style={{ fontWeight: 700, fontSize: "0.95rem", color: "var(--text-1)", display: "flex", alignItems: "center", gap: "0.5rem" }}>
                        📁 {g.title}
                      </div>
                      <div style={{ fontSize: "0.75rem", color: "var(--text-3)", marginTop: "0.2rem" }}>
                        Last occurrence: {relTime(g.latest)}
                      </div>
                    </div>
                    <div style={{ display: "flex", gap: "0.45rem", alignItems: "center", flexShrink: 0 }}>
                      <SevBadge sev={g.severity} />
                      <StatusBadge status={g.status} />
                      <span className="badge badge-neutral" style={{ background: "var(--surface-active)" }}>{g.count} recurring events</span>
                    </div>
                  </div>

                  {/* Bulk Actions for the Group */}
                  <div style={{ display: "flex", gap: "0.45rem", alignItems: "center", borderTop: "1px solid var(--border)", paddingTop: "0.6rem", flexWrap: "wrap" }}>
                    <button
                      className="btn btn-secondary btn-sm"
                      style={{ fontSize: "0.75rem", padding: "0.25rem 0.6rem" }}
                      onClick={() => setExpandedGroups(prev => ({ ...prev, [g.title]: !prev[g.title] }))}
                    >
                      {isExpanded ? "▲ Collapse View" : `▼ Expand ${g.count} Incidents`}
                    </button>

                    <div style={{ marginLeft: "auto", display: "flex", gap: "0.4rem" }}>
                      {anyOpen && (
                        <button
                          className="btn btn-ghost btn-sm"
                          style={{ fontSize: "0.72rem", padding: "0.25rem 0.6rem" }}
                          onClick={() => handleAckAll(nestedIds)}
                        >
                          👁 Acknowledge All
                        </button>
                      )}
                      {anyUnresolved && (
                        <button
                          className="btn btn-success btn-sm"
                          style={{ fontSize: "0.72rem", padding: "0.25rem 0.6rem" }}
                          onClick={() => handleResolveAll(nestedIds)}
                        >
                          ✓ Resolve All
                        </button>
                      )}
                    </div>
                  </div>

                  {/* Inline Expanded List of Nested Incidents */}
                  {isExpanded && (
                    <div
                      style={{
                        display: "flex",
                        flexDirection: "column",
                        gap: "0.45rem",
                        padding: "0.6rem 0.75rem",
                        background: "rgba(0,0,0,0.25)",
                        border: "1px solid var(--border)",
                        borderRadius: "var(--r-sm)",
                        marginTop: "0.25rem",
                        animation: "fadeIn 0.2s ease"
                      }}
                    >
                      <div style={{ fontSize: "0.68rem", fontWeight: 700, color: "var(--text-3)", textTransform: "uppercase", letterSpacing: "0.06em", borderBottom: "1px solid var(--border)", paddingBottom: "0.3rem", marginBottom: "0.25rem" }}>
                        Grouped Occurrences
                      </div>
                      {g.incidents.map(inc => {
                        const isSelected = selectedId === inc.id;
                        return (
                          <div
                            key={inc.id}
                            onClick={() => setSelectedId(isSelected ? null : inc.id)}
                            style={{
                              display: "flex", justifyContent: "space-between", alignItems: "center",
                              padding: "0.45rem 0.65rem", borderRadius: "var(--r-xs)",
                              background: isSelected ? "var(--primary-soft)" : "transparent",
                              cursor: "pointer", border: isSelected ? "1px solid var(--primary-glow)" : "1px solid transparent"
                            }}
                          >
                            <span style={{ fontSize: "0.78rem", fontWeight: 600, fontFamily: "var(--font-mono)", color: isSelected ? "var(--primary-2)" : "var(--text-2)" }}>
                              #{inc.id.slice(-8)}
                            </span>
                            <div style={{ display: "flex", gap: "1rem", fontSize: "0.75rem", color: "var(--text-3)", alignItems: "center" }}>
                              <span>⏱ {relTime(inc.triggered_at)}</span>
                              {inc.assigned_to && <span>👤 {inc.assigned_to}</span>}
                              <StatusBadge status={inc.status} />
                            </div>
                          </div>
                        );
                      })}
                    </div>
                  )}
                </div>
              );
            })
          ) : (
            incidents.map(inc => (
              <IncidentCard
                key={inc.id}
                inc={inc}
                isSelected={selectedId === inc.id}
                onClick={() => setSelectedId(selectedId === inc.id ? null : inc.id)}
                onAck={handleAck}
                onResolve={handleResolve}
              />
            ))
          )}
        </div>

        {/* Detail panel */}
        {selectedId && (
          <div style={{ minHeight: 0, overflowY: "auto" }}>
            <DetailPanel
              incidentId={selectedId}
              token={token}
              onClose={() => setSelectedId(null)}
              onAck={async (id) => { await handleAck(id); }}
              onResolve={async (id) => { await handleResolve(id); }}
              onAssign={handleAssign}
              onComment={handleComment}
            />
          </div>
        )}
      </div>
    </div>
  );
}
