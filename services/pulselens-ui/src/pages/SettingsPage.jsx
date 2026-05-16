// SettingsPage — users, API keys, audit logs with side-sheet drawer for user creation
import { useState } from "react";
import { EmptyState, SectionLoader, useAsyncData } from "../lib/hooks";
import { tenantApi } from "../lib/api";
import Drawer from "../components/Drawer";
import Tooltip from "../components/Tooltip";

const ROLES = [
  { value: "viewer",       label: "Viewer",       color: "badge-info",    desc: "Read-only access to all telemetry, dashboards, and incidents. Cannot modify alert rules or create users." },
  { value: "tenant_admin", label: "Admin",         color: "badge-warning", desc: "Full tenant control: create users, manage API keys, configure alert rules and notification channels." },
];

export default function SettingsPage({ state, notify }) {
  const token    = state.token;
  const tenantId = state.tenantId;

  const { data: users,     loading: usersLoading,  refetch: refetchUsers  } = useAsyncData(() => tenantApi.listUsers(tenantId, token),    [token, tenantId], { skip: !token || !tenantId });
  const { data: apiKeys,   loading: keysLoading,   refetch: refetchKeys   } = useAsyncData(() => tenantApi.listAPIKeys(token),             [token],           { skip: !token });
  const { data: auditLogs, loading: auditLoading                           } = useAsyncData(() => tenantApi.listAuditLogs(tenantId, token), [token, tenantId], { skip: !token || !tenantId });

  const [drawerOpen, setDrawerOpen] = useState(false);
  const [userForm, setUserForm] = useState({ name: "", email: "", password: "", role: "viewer" });
  const [creating, setCreating] = useState(false);

  const uf = (k, v) => setUserForm(p => ({ ...p, [k]: v }));

  async function handleCreateUser(e) {
    e?.preventDefault();
    if (!userForm.name || !userForm.email || !userForm.password) return notify("All fields required", "error");
    setCreating(true);
    try {
      await tenantApi.createUser(tenantId, userForm, token);
      notify(`User ${userForm.email} created successfully.`, "success");
      setUserForm({ name: "", email: "", password: "", role: "viewer" });
      setDrawerOpen(false);
      refetchUsers();
    } catch (err) {
      notify(err.message, "error");
    } finally {
      setCreating(false);
    }
  }

  async function handleRevokeKey(keyId) {
    try {
      await tenantApi.revokeAPIKey(keyId, token);
      notify("API key revoked — cache invalidated immediately.", "success");
      refetchKeys();
    } catch (err) { notify(err.message, "error"); }
  }

  async function handleRotateKey(keyId) {
    try {
      await tenantApi.rotateAPIKey(keyId, {}, token);
      notify("API key rotated. Old key is now invalid.", "success");
      refetchKeys();
    } catch (err) { notify(err.message, "error"); }
  }

  return (
    <div style={{ display: "flex", flexDirection: "column", gap: "1.25rem" }}>
      {/* Page header */}
      <div>
        <h1 style={{ fontSize: "1.4rem", fontWeight: 700, marginBottom: "0.2rem" }}>Settings</h1>
        <p style={{ color: "var(--text-2)", fontSize: "0.875rem" }}>Manage team members, API keys, and audit history.</p>
      </div>

      {/* ── Users ─────────────────────────────────────────── */}
      <div className="panel">
        <div className="panel-header">
          <div>
            <div className="panel-title">Team Members</div>
            <div className="panel-desc">Users who have access to this workspace</div>
          </div>
          <div style={{ display: "flex", gap: "0.5rem" }}>
            <button className="btn btn-secondary btn-sm" onClick={refetchUsers}>↺ Refresh</button>
            <button className="btn btn-primary btn-sm" id="btn-invite-user" onClick={() => setDrawerOpen(true)}>
              + Invite User
            </button>
          </div>
        </div>

        {usersLoading ? <SectionLoader /> : (
          users?.length ? (
            <div className="table-wrap">
              <table className="data-table" id="table-users">
                <thead><tr><th>Name</th><th>Email</th><th>Role</th><th>Joined</th></tr></thead>
                <tbody>
                  {users.map(u => (
                    <tr key={u.id}>
                      <td style={{ fontWeight: 500 }}>{u.name}</td>
                      <td style={{ fontFamily: "var(--font-mono)", fontSize: "0.8rem", color: "var(--text-2)" }}>{u.email}</td>
                      <td>
                        <span className={`badge ${ROLES.find(r => r.value === u.role)?.color || "badge-neutral"}`}>
                          {ROLES.find(r => r.value === u.role)?.label || u.role}
                        </span>
                      </td>
                      <td style={{ color: "var(--text-2)", fontSize: "0.8rem" }}>{new Date(u.created_at).toLocaleDateString()}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          ) : (
            <EmptyState icon="◎" title="No team members yet"
              body="Invite colleagues to share access to this tenant workspace."
              action={<button className="btn btn-primary btn-sm" onClick={() => setDrawerOpen(true)}>+ Invite User</button>}
            />
          )
        )}
      </div>

      {/* ── API Keys ─────────────────────────────────────── */}
      <div className="panel">
        <div className="panel-header">
          <div>
            <div className="panel-title">
              API Keys
              <Tooltip text="API keys authenticate your services when sending telemetry to PulseLens. Each key is scoped (ingest / query) and cached in Redis for fast validation." />
            </div>
            <div className="panel-desc">Used to authenticate telemetry ingestion and queries</div>
          </div>
          <button className="btn btn-secondary btn-sm" onClick={refetchKeys}>↺ Refresh</button>
        </div>

        {keysLoading ? <SectionLoader /> : (
          apiKeys?.length ? (
            <div className="table-wrap">
              <table className="data-table" id="table-api-keys">
                <thead>
                  <tr>
                    <th>Name</th><th>Prefix</th><th>Scopes</th><th>Status</th><th>Last Used</th>
                    <th>
                      Actions
                      <Tooltip text="Rotate: create a new secret, old key stops working immediately. Revoke: permanently deactivate — cannot be undone." position="left" />
                    </th>
                  </tr>
                </thead>
                <tbody>
                  {apiKeys.map(k => (
                    <tr key={k.id}>
                      <td style={{ fontWeight: 500 }}>{k.name}</td>
                      <td style={{ fontFamily: "var(--font-mono)", fontSize: "0.78rem", color: "var(--text-2)" }}>{k.key_prefix}…</td>
                      <td style={{ fontSize: "0.75rem" }}>{Array.isArray(k.scopes) ? k.scopes.join(", ") : k.scopes || "—"}</td>
                      <td><span className={`badge ${k.active ? "badge-success" : "badge-neutral"}`}>{k.active ? "active" : "revoked"}</span></td>
                      <td style={{ color: "var(--text-2)", fontSize: "0.8rem" }}>{k.last_used_at ? new Date(k.last_used_at).toLocaleString() : "never"}</td>
                      <td>
                        {k.active && (
                          <div style={{ display: "flex", gap: "0.35rem" }}>
                            <Tooltip text="Generates a new secret. The old key is invalidated and purged from the Redis cache immediately.">
                              <button id={`btn-rotate-${k.id}`} className="btn btn-secondary btn-sm" onClick={() => handleRotateKey(k.id)}>Rotate</button>
                            </Tooltip>
                            <Tooltip text="Permanently deactivates this key. Any service using it will immediately start receiving 401 errors." position="left">
                              <button id={`btn-revoke-${k.id}`} className="btn btn-sm" style={{ background: "var(--danger-soft)", color: "var(--danger)", border: "1px solid rgba(239,68,68,0.25)" }} onClick={() => handleRevokeKey(k.id)}>Revoke</button>
                            </Tooltip>
                          </div>
                        )}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          ) : (
            <EmptyState icon="⊞" title="No API keys" body="API keys are created during the workspace setup. Use the Setup wizard to generate one." />
          )
        )}
      </div>

      {/* ── Audit Log ────────────────────────────────────── */}
      <div className="panel">
        <div className="panel-header">
          <div>
            <div className="panel-title">
              Audit Log
              <Tooltip text="Every user and system action on this tenant is recorded here — key creation, user changes, rule updates, and API calls." />
            </div>
            <div className="panel-desc">Immutable record of all tenant operations</div>
          </div>
        </div>

        {auditLoading ? <SectionLoader /> : (
          auditLogs?.length ? (
            <div className="table-wrap">
              <table className="data-table" id="table-audit-logs">
                <thead><tr><th>Time</th><th>Actor</th><th>Action</th><th>Resource</th></tr></thead>
                <tbody>
                  {auditLogs.slice(0, 50).map(log => (
                    <tr key={log.id}>
                      <td style={{ color: "var(--text-2)", fontSize: "0.8rem" }}>{new Date(log.created_at).toLocaleString()}</td>
                      <td><span className={`badge ${log.actor_type === "user" ? "badge-info" : "badge-neutral"}`}>{log.actor_type}</span></td>
                      <td style={{ fontFamily: "var(--font-mono)", fontSize: "0.78rem" }}>{log.action}</td>
                      <td style={{ fontSize: "0.78rem", color: "var(--text-2)" }}>{log.resource_type} / {log.resource_id?.slice(0, 20)}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          ) : (
            <EmptyState icon="◑" title="No audit logs yet" body="All user and system operations will appear here." />
          )
        )}
      </div>

      {/* ── Create User Drawer ───────────────────────────── */}
      <Drawer
        open={drawerOpen}
        onClose={() => setDrawerOpen(false)}
        title="Invite Team Member"
        description="Create a new user account with access to this tenant workspace."
      >
        <form onSubmit={handleCreateUser} style={{ display: "flex", flexDirection: "column", gap: "1.25rem" }}>
          <div className="form-group">
            <label className="form-label">Full Name</label>
            <input className="form-input" placeholder="Jane Doe" value={userForm.name} onChange={e => uf("name", e.target.value)} required />
          </div>

          <div className="form-group">
            <label className="form-label">Email Address</label>
            <input className="form-input" type="email" placeholder="jane@company.com" value={userForm.email} onChange={e => uf("email", e.target.value)} required />
          </div>

          <div className="form-group">
            <label className="form-label">Temporary Password</label>
            <input className="form-input" type="password" placeholder="min 8 characters" value={userForm.password} onChange={e => uf("password", e.target.value)} required />
            <p style={{ fontSize: "0.75rem", color: "var(--text-3)", marginTop: "0.25rem" }}>The user can change this after first login.</p>
          </div>

          <div className="form-group">
            <label className="form-label">
              Role
              <Tooltip text="Choose the level of access this user should have." />
            </label>
            <div style={{ display: "flex", flexDirection: "column", gap: "0.5rem", marginTop: "0.25rem" }}>
              {ROLES.map(role => (
                <label key={role.value} style={{
                  display: "flex", alignItems: "flex-start", gap: "0.75rem",
                  padding: "0.75rem", borderRadius: "var(--r-sm)", cursor: "pointer",
                  border: `1.5px solid ${userForm.role === role.value ? "var(--primary)" : "var(--border)"}`,
                  background: userForm.role === role.value ? "var(--primary-soft)" : "var(--surface-active)",
                  transition: "all 0.15s",
                }}>
                  <input type="radio" name="role" value={role.value} checked={userForm.role === role.value} onChange={() => uf("role", role.value)} style={{ marginTop: "2px", accentColor: "var(--primary)" }} />
                  <div>
                    <div style={{ fontSize: "0.875rem", fontWeight: 600, display: "flex", alignItems: "center", gap: "0.5rem" }}>
                      {role.label} <span className={`badge ${role.color}`} style={{ fontSize: "0.65rem" }}>{role.value}</span>
                    </div>
                    <div style={{ fontSize: "0.78rem", color: "var(--text-2)", marginTop: "0.2rem", lineHeight: 1.5 }}>{role.desc}</div>
                  </div>
                </label>
              ))}
            </div>
          </div>

          <div style={{ display: "flex", gap: "0.65rem", paddingTop: "0.5rem", borderTop: "1px solid var(--border)" }}>
            <button type="button" className="btn btn-ghost" style={{ flex: 1 }} onClick={() => setDrawerOpen(false)}>Cancel</button>
            <button type="submit" className="btn btn-primary" id="btn-create-user" style={{ flex: 2, justifyContent: "center" }} disabled={creating}>
              {creating ? "Creating…" : "Create User"}
            </button>
          </div>
        </form>
      </Drawer>
    </div>
  );
}
