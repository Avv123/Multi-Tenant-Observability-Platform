// IncidentsPage — incident management workspace.
// F1: extracted. F4: per-section loading. F9: empty states.

import { useState } from "react";
import { EmptyState, ErrorMessage, SectionLoader, useAsyncData } from "../lib/hooks";
import { alertingApi } from "../lib/api";

function formatDate(v) { return v ? new Date(v).toLocaleString() : "—"; }

export default function IncidentsPage({ state, notify }) {
  const token = state.token;

  const [incidentFilters, setIncidentFilters] = useState({ status: "", assigned_to: "", severity: "" });
  const [selectedId, setSelectedId] = useState("");
  const [commentBody, setCommentBody] = useState("Investigating now.");
  const [assignTo, setAssignTo] = useState("");

  const { data: incidents, loading, error, refetch } = useAsyncData(
    () => alertingApi.listIncidents(token, incidentFilters),
    [token, JSON.stringify(incidentFilters)],
    { skip: !token },
  );

  const { data: detail, loading: detailLoading, refetch: refetchDetail } = useAsyncData(
    () => selectedId ? alertingApi.getIncident(token, selectedId) : Promise.resolve(null),
    [token, selectedId],
    { skip: !token || !selectedId },
  );

  const { data: timeline, loading: timelineLoading } = useAsyncData(
    () => selectedId ? alertingApi.incidentTimeline(token, selectedId) : Promise.resolve([]),
    [token, selectedId],
    { skip: !token || !selectedId },
  );

  const { data: comments, loading: commentsLoading, refetch: refetchComments } = useAsyncData(
    () => selectedId ? alertingApi.listIncidentComments(token, selectedId) : Promise.resolve([]),
    [token, selectedId],
    { skip: !token || !selectedId },
  );

  async function handleAck(id) {
    try {
      await alertingApi.acknowledgeIncident(token, id);
      notify("Incident acknowledged.", "success");
      refetch();
      if (selectedId === id) refetchDetail();
    } catch (err) { notify(err.message, "error"); }
  }

  async function handleResolve(id) {
    try {
      await alertingApi.resolveIncident(token, id);
      notify("Incident resolved.", "success");
      refetch();
      if (selectedId === id) refetchDetail();
    } catch (err) { notify(err.message, "error"); }
  }

  async function handleAddComment(e) {
    e.preventDefault();
    try {
      await alertingApi.addIncidentComment(token, selectedId, { body: commentBody });
      notify("Comment added.", "success");
      setCommentBody("Investigating now.");
      refetchComments();
    } catch (err) { notify(err.message, "error"); }
  }

  async function handleAssign(e) {
    e.preventDefault();
    try {
      await alertingApi.assignIncident(token, selectedId, { assigned_to: assignTo });
      notify("Incident assigned.", "success");
      refetchDetail();
    } catch (err) { notify(err.message, "error"); }
  }

  return (
    <div style={{ display:"flex", flexDirection:"column", gap:"1rem" }}>
      {error && (
        <div style={{ padding:"0.875rem", background:"var(--danger-soft)", border:"1px solid rgba(239,68,68,0.25)", borderRadius:"var(--r-md)", color:"var(--danger)", fontSize:"0.875rem" }}>⚠ {error}</div>
      )}

      {/* Filter bar */}
      <div className="panel">
        <div className="panel-header">
          <div>
            <div className="panel-title">Incidents</div>
            <div className="panel-desc">Track, acknowledge and resolve triggered alerts.</div>
          </div>
          <button id="btn-refresh-incidents" className="btn btn-secondary btn-sm" onClick={refetch}>↺ Refresh</button>
        </div>
        <div className="form-grid" style={{ marginBottom:"1rem" }}>
          <div className="form-field">
            <label>Status</label>
            <select id="incident-filter-status" value={incidentFilters.status}
              onChange={(e) => setIncidentFilters((f) => ({ ...f, status: e.target.value }))}>
              <option value="">All</option>
              {["open","acknowledged","resolved"].map((s) => <option key={s} value={s}>{s}</option>)}
            </select>
          </div>
          <div className="form-field">
            <label>Severity</label>
            <select id="incident-filter-severity" value={incidentFilters.severity}
              onChange={(e) => setIncidentFilters((f) => ({ ...f, severity: e.target.value }))}>
              <option value="">All</option>
              {["warning","critical","error"].map((s) => <option key={s} value={s}>{s}</option>)}
            </select>
          </div>
        </div>

        {loading ? <SectionLoader /> : (
          incidents?.length ? (
            <div className="table-wrap">
              <table className="data-table" id="table-incidents">
                <thead>
                  <tr><th>Triggered</th><th>Title</th><th>Severity</th><th>Status</th><th>Assigned</th><th>Actions</th></tr>
                </thead>
                <tbody>
                  {incidents.map((inc) => (
                    <tr
                      key={inc.id}
                      onClick={() => setSelectedId(inc.id)}
                      style={{ cursor:"pointer", background: selectedId === inc.id ? "var(--primary-soft)" : undefined }}
                    >
                      <td className="text-muted text-sm">{formatDate(inc.triggered_at)}</td>
                      <td>{inc.title || inc.summary}</td>
                      <td>
                        <span className={`badge ${inc.severity==="critical"||inc.severity==="error" ? "badge-danger" : "badge-warning"}`}>
                          {inc.severity}
                        </span>
                      </td>
                      <td>
                        <span className={`badge ${inc.status==="open" ? "badge-danger" : inc.status==="acknowledged" ? "badge-warning" : "badge-success"}`}>
                          {inc.status}
                        </span>
                      </td>
                      <td className="text-muted text-sm">{inc.assigned_to || "—"}</td>
                      <td>
                        <div className="button-row" onClick={(e) => e.stopPropagation()}>
                          {inc.status === "open" && (
                            <button id={`btn-ack-${inc.id}`} className="btn btn-ghost btn-sm" onClick={() => handleAck(inc.id)}>Ack</button>
                          )}
                          {inc.status !== "resolved" && (
                            <button id={`btn-resolve-${inc.id}`} className="btn btn-success btn-sm" onClick={() => handleResolve(inc.id)}>Resolve</button>
                          )}
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          ) : (
            <EmptyState
              icon="⚠"
              title="No incidents"
              body="No incidents match the current filters. Create an alert rule on the Alerts page to begin monitoring."
            />
          )
        )}
      </div>

      {/* Detail panel */}
      {selectedId && (
        <div className="page-grid-2">
          <div className="panel">
            <div className="panel-title" style={{ marginBottom:"0.75rem" }}>Incident Detail</div>
            {detailLoading ? <SectionLoader /> : detail ? (
              <div style={{ display: "flex", flexDirection: "column", gap: "0.5rem", fontSize: "0.83rem" }}>
                {[
                  ["ID",       detail.id],
                  ["Title",    detail.title],
                  ["Summary",  detail.summary],
                  ["Severity", detail.severity],
                  ["Status",   detail.status],
                  ["Rule ID",  detail.alert_rule_id],
                  ["Value",    `${detail.observed_value} (threshold: ${detail.threshold})`],
                  ["Triggered",detail.triggered_at ? new Date(detail.triggered_at).toLocaleString() : "—"],
                  ["Assigned", detail.assigned_to || "unassigned"],
                ].map(([k, v]) => (
                  <div key={k} style={{ display: "flex", gap: "0.75rem" }}>
                    <span className="text-muted" style={{ width: "90px", flexShrink: 0 }}>{k}</span>
                    <span>{v}</span>
                  </div>
                ))}
              </div>
            ) : null}

            {/* Assign form */}
            <form id="form-assign-incident" onSubmit={handleAssign} style={{ marginTop: "1rem" }}>
              <div className="form-field">
                <label>Assign To</label>
                <input id="assign-to-input" value={assignTo} onChange={(e) => setAssignTo(e.target.value)} placeholder="user-id or email" />
              </div>
              <button id="btn-assign-incident" type="submit" className="btn btn-ghost btn-sm" style={{ marginTop: "0.5rem" }}>
                Assign
              </button>
            </form>
          </div>

          <div className="panel">
            {/* Comments */}
            <div className="panel-title" style={{ marginBottom:"0.75rem" }}>Comments</div>
            {commentsLoading ? <SectionLoader /> : (
              comments?.length ? (
                <div style={{ display: "flex", flexDirection: "column", gap: "0.5rem", marginBottom: "1rem" }}>
                  {comments.map((c) => <div style={{ background:"var(--surface)", padding:"0.6rem 0.85rem", borderRadius:"var(--r-sm)", fontSize:"0.82rem", border:"1px solid var(--border)" }}>
                      <div className="text-muted text-xs">{formatDate(c.created_at)}</div>
                      <div>{c.body}</div>
                    </div>
                  )}
                </div>
              ) : (
                <p className="text-muted text-sm" style={{ marginBottom: "0.75rem" }}>No comments yet.</p>
              )
            )}

            <form id="form-add-comment" onSubmit={handleAddComment}>
              <div className="form-field">
                <label>Add Comment</label>
                <textarea
                  id="comment-body-input"
                  rows={3}
                  value={commentBody}
                  onChange={(e) => setCommentBody(e.target.value)}
                />
              </div>
              <button id="btn-add-comment" type="submit" className="btn btn-primary btn-sm" style={{ marginTop: "0.5rem" }}>
                Submit Comment
              </button>
            </form>

            {/* Timeline */}
            <div className="panel-title" style={{ margin:"1rem 0 0.5rem" }}>Timeline</div>
            {timelineLoading ? <SectionLoader /> : (
              timeline?.length ? (
                <div style={{ display: "flex", flexDirection: "column", gap: "0.35rem" }}>
                  {timeline.map((ev) => (
                    <div key={ev.id} style={{ fontSize: "0.78rem", display: "flex", gap: "0.5rem" }}>
                      <span className="text-muted">{formatDate(ev.created_at)}</span>
                      <span className="badge badge-info">{ev.event_type}</span>
                      <span>{ev.summary}</span>
                    </div>
                  ))}
                </div>
              ) : (
                <p className="text-subtle text-sm">No timeline events.</p>
              )
            )}
          </div>
        </div>
      )}
    </div>
  );
}
