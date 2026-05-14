import DataTable from "../DataTable";
import Section from "../Section";

function formatDate(value) {
  if (!value) {
    return "-";
  }
  return new Date(value).toLocaleString();
}

function parseMetadata(value) {
  if (!value) {
    return {};
  }
  if (typeof value === "object") {
    return value;
  }
  try {
    return JSON.parse(value);
  } catch {
    return { raw: value };
  }
}

function formatMetadata(value) {
  const parsed = parseMetadata(value);
  const entries = Object.entries(parsed);
  if (!entries.length) {
    return "{}";
  }
  return entries
    .map(([key, item]) => `${key}=${typeof item === "object" ? JSON.stringify(item) : String(item)}`)
    .join(", ");
}

export default function IncidentWorkspaceSection({
  data,
  incidentFilters,
  setIncidentFilters,
  assignForm,
  setAssignForm,
  commentForm,
  setCommentForm,
  onApplyFilters,
  onIncidentAction,
  onLoadIncidentDetail,
  onAssignIncident,
  onAddComment,
}) {
  const selectedIncidentID = data.selectedIncident?.id || assignForm.incidentId || commentForm.incidentId || "";

  return (
    <>
      <Section title="Incidents">
        <form className="form-grid" onSubmit={onApplyFilters}>
          <label>
            Status
            <select data-testid="incident-filter-status" value={incidentFilters.status} onChange={(event) => setIncidentFilters({ ...incidentFilters, status: event.target.value })}>
              <option value="">all</option>
              <option value="open">open</option>
              <option value="acknowledged">acknowledged</option>
              <option value="resolved">resolved</option>
            </select>
          </label>
          <label>
            Assigned To
            <input data-testid="incident-filter-assigned" value={incidentFilters.assigned_to} onChange={(event) => setIncidentFilters({ ...incidentFilters, assigned_to: event.target.value })} />
          </label>
          <label>
            Service
            <select data-testid="incident-filter-service" value={incidentFilters.service_id} onChange={(event) => setIncidentFilters({ ...incidentFilters, service_id: event.target.value })}>
              <option value="">all</option>
              {data.services.map((service) => (
                <option key={service.id} value={service.id}>{service.name}</option>
              ))}
            </select>
          </label>
          <label>
            Severity
            <input data-testid="incident-filter-severity" value={incidentFilters.severity} onChange={(event) => setIncidentFilters({ ...incidentFilters, severity: event.target.value })} />
          </label>
          <div className="form-actions">
            <button data-testid="apply-incident-filters" type="submit">Apply Incident Filters</button>
          </div>
        </form>
        <DataTable
          columns={[
            { key: "title", label: "Title" },
            { key: "severity", label: "Severity" },
            { key: "status", label: "Status" },
            { key: "assigned_to", label: "Assigned To" },
            { key: "escalation_level", label: "Escalation" },
            { key: "observed_value", label: "Observed" },
            { key: "threshold", label: "Threshold" },
            { key: "triggered_at", label: "Triggered", render: (row) => formatDate(row.triggered_at) },
            {
              key: "actions",
              label: "Actions",
              render: (row) => (
                <div className="button-row">
                  <button type="button" data-testid={`ack-incident-${row.id}`} onClick={() => onIncidentAction("ack", row.id)}>Ack</button>
                  <button type="button" data-testid={`resolve-incident-${row.id}`} onClick={() => onIncidentAction("resolve", row.id)}>Resolve</button>
                  <button type="button" data-testid={`detail-incident-${row.id}`} onClick={() => onLoadIncidentDetail(row.id)}>Details</button>
                </div>
              ),
            },
          ]}
          rows={data.incidents}
        />
      </Section>

      <Section title="Incident Detail">
        {data.selectedIncident ? (
          <div className="incident-layout">
            <div className="widget-card">
              <h3>Summary</h3>
              <div className="key-value-grid">
                <span>Incident</span><code>{data.selectedIncident.id}</code>
                <span>Status</span><code>{data.selectedIncident.status}</code>
                <span>Severity</span><code>{data.selectedIncident.severity || "-"}</code>
                <span>Assigned To</span><code>{data.selectedIncident.assigned_to || "-"}</code>
                <span>Escalation Count</span><code>{data.selectedIncident.escalation_count ?? 0}</code>
                <span>Last Escalated</span><code>{formatDate(data.selectedIncident.last_escalated_at)}</code>
                <span>Next Escalation</span><code>{formatDate(data.selectedIncident.next_escalation_at)}</code>
              </div>
              <div className="button-row">
                <button type="button" onClick={() => onIncidentAction("ack", data.selectedIncident.id)}>Acknowledge</button>
                <button type="button" onClick={() => onIncidentAction("resolve", data.selectedIncident.id)}>Resolve</button>
              </div>
            </div>

            <div className="widget-card">
              <h3>Assignment And Comments</h3>
              <form className="form-grid" onSubmit={onAssignIncident}>
                <label>
                  Incident ID
                  <input data-testid="assign-incident-id" value={selectedIncidentID} readOnly />
                </label>
                <label>
                  Assign To
                  <input data-testid="assign-incident-user" value={assignForm.assignedTo} onChange={(event) => setAssignForm({ ...assignForm, assignedTo: event.target.value })} />
                </label>
                <div className="form-actions">
                  <button data-testid="assign-incident-submit" type="submit">Assign Incident</button>
                </div>
              </form>

              <form className="form-grid" onSubmit={onAddComment}>
                <label>
                  Incident ID
                  <input data-testid="comment-incident-id" value={selectedIncidentID} readOnly />
                </label>
                <label>
                  Comment
                  <input data-testid="comment-incident-body" value={commentForm.body} onChange={(event) => setCommentForm({ ...commentForm, body: event.target.value })} />
                </label>
                <div className="form-actions">
                  <button data-testid="comment-incident-submit" type="submit">Add Comment</button>
                </div>
              </form>
            </div>

            <div className="widget-card">
              <h3>Timeline</h3>
              <DataTable
                columns={[
                  { key: "event_type", label: "Event" },
                  { key: "summary", label: "Summary" },
                  { key: "actor_id", label: "Actor" },
                  { key: "metadata", label: "Metadata", render: (row) => <code>{formatMetadata(row.metadata)}</code> },
                  { key: "created_at", label: "Created", render: (row) => formatDate(row.created_at) },
                ]}
                rows={data.incidentTimeline}
              />
            </div>

            <div className="widget-card">
              <h3>Comments</h3>
              <DataTable
                columns={[
                  { key: "author_id", label: "Author" },
                  { key: "body", label: "Comment" },
                  { key: "created_at", label: "Created", render: (row) => formatDate(row.created_at) },
                ]}
                rows={data.incidentComments}
              />
            </div>

            <div className="widget-card">
              <h3>Delivery History</h3>
              <DataTable
                columns={[
                  { key: "event_type", label: "Event" },
                  { key: "status", label: "Status" },
                  { key: "attempt_count", label: "Attempts" },
                  { key: "channel_id", label: "Channel" },
                  { key: "response", label: "Response" },
                  { key: "delivered_at", label: "Delivered", render: (row) => formatDate(row.delivered_at) },
                ]}
                rows={data.incidentDeliveries}
              />
            </div>
          </div>
        ) : (
          <p>Select an incident to inspect timeline, comments, assignments, and deliveries.</p>
        )}
      </Section>
    </>
  );
}
