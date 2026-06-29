// SettingsPage — users, API keys, audit logs with side-sheet drawer for user creation
import { useState } from "react";
import { EmptyState, SectionLoader, useAsyncData } from "../lib/hooks";
import { tenantApi } from "../lib/api";
import Drawer from "../components/Drawer";
import Tooltip from "../components/Tooltip";

const ROLES = [
  { value: "viewer",       label: "Viewer", color: "badge-info",    desc: "Read-only access to all telemetry, dashboards, and incidents. Cannot modify alert rules or create users." },
  { value: "tenant_admin", label: "Admin",  color: "badge-warning", desc: "Full tenant control: create users, manage API keys, configure alert rules and notification channels." },
];

export default function SettingsPage({ state, notify, setState }) {
  const token    = state.token;
  const tenantId = state.tenantId;

  const { data: users,     loading: usersLoading,  refetch: refetchUsers  } = useAsyncData(() => tenantApi.listUsers(tenantId, token),    [token, tenantId], { skip: !token || !tenantId });
  const { data: apiKeys,   loading: keysLoading,   refetch: refetchKeys   } = useAsyncData(() => tenantApi.listAPIKeys(token),             [token],           { skip: !token });
  const { data: auditLogs, loading: auditLoading                           } = useAsyncData(() => tenantApi.listAuditLogs(tenantId, token), [token, tenantId], { skip: !token || !tenantId });
  const { data: services,  loading: servicesLoading, refetch: refetchServices } = useAsyncData(() => tenantApi.listServices(tenantId, token), [token, tenantId], { skip: !token || !tenantId });

  const [drawerOpen, setDrawerOpen] = useState(false);
  const [keyDrawerOpen, setKeyDrawerOpen] = useState(false);
  const [serviceDrawerOpen, setServiceDrawerOpen] = useState(false);
  const [userForm,   setUserForm]   = useState({ name: "", email: "", password: "", role: "viewer" });
  const [keyForm,    setKeyForm]    = useState({ name: "", scope: "ingest", serviceId: "" });
  const [serviceForm, setServiceForm] = useState({ name: "", environment: "production" });
  const [creating,   setCreating]   = useState(false);
  const [newlyCreatedKey, setNewlyCreatedKey] = useState("");
  const uf = (k, v) => setUserForm(p => ({ ...p, [k]: v }));
  const kf = (k, v) => setKeyForm(p => ({ ...p, [k]: v }));
  const sf = (k, v) => setServiceForm(p => ({ ...p, [k]: v }));

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
    } catch (err) { notify(err.message, "error"); }
    finally { setCreating(false); }
  }

  async function handleRevokeKey(keyId) {
    try { await tenantApi.revokeAPIKey(keyId, token); notify("API key revoked — cache invalidated immediately.", "success"); refetchKeys(); }
    catch (err) { notify(err.message, "error"); }
  }
  async function handleCreateKey(e) {
    e?.preventDefault();
    if (!keyForm.name) return notify("Key name required", "error");
    setCreating(true);
    try {
      let svcId = keyForm.serviceId;
      
      // If no service selected, use the first one or create one
      if (!svcId) {
        if (services?.length) {
          svcId = services[0].id;
        } else {
          // Create a default service if none exist
          const newSvc = await tenantApi.createService(tenantId, { name: "default-service", environment: "production" }, token);
          svcId = newSvc.id;
        }
      }

      const keyRes = await tenantApi.createAPIKey({
        tenant_id: tenantId,
        service_id: svcId,
        name: keyForm.name,
        scopes: [keyForm.scope]
      }, token);
      
      const rawKey = keyRes.key || keyRes.api_key || keyRes.value;
      if ((keyForm.scope === "ingest" || keyForm.scope === "*") && rawKey) {
        setState(prev => ({ ...prev, apiKey: rawKey }));
      }
      notify(`Key "${keyForm.name}" created for service "${svcId}".`, "success");
      setKeyForm({ name: "", scope: "ingest", serviceId: "" });
      setNewlyCreatedKey(rawKey);
      refetchKeys();
    } catch (err) { notify(err.message, "error"); }
    finally { setCreating(false); }
  }
  async function handleCreateService(e) {
    e?.preventDefault();
    if (!serviceForm.name || !serviceForm.environment) return notify("All fields required", "error");
    setCreating(true);
    try {
      await tenantApi.createService(tenantId, { name: serviceForm.name, environment: serviceForm.environment }, token);
      notify(`Service "${serviceForm.name}" registered successfully.`, "success");
      setServiceForm({ name: "", environment: "production" });
      setServiceDrawerOpen(false);
      refetchServices();
    } catch (err) { notify(err.message, "error"); }
    finally { setCreating(false); }
  }
  async function handleRotateKey(keyId, keyScopes) {
    try {
      const newKey = await tenantApi.rotateAPIKey(keyId, {}, token);
      // If this was an ingest-scoped key, update global state so Test Log/Trace works immediately
      const scopes = Array.isArray(keyScopes) ? keyScopes : [keyScopes];
      const rawKey = newKey?.key || newKey?.api_key || newKey?.value;
      if (scopes.some(s => s === "ingest" || s === "*") && rawKey) {
        setState(prev => ({ ...prev, apiKey: rawKey }));
      }
      notify("API key rotated. Old key is now invalid. New key is active.", "success");
      refetchKeys();
    } catch (err) { notify(err.message, "error"); }
  }

  return (
    <div style={{ display: "flex", flexDirection: "column", gap: "1.25rem" }}>
      <div>
        <h1 style={{ fontSize: "1.4rem", fontWeight: 700, marginBottom: "0.2rem" }}>Settings</h1>
        <p style={{ color: "var(--text-2)", fontSize: "0.875rem" }}>Manage team members, API keys, and audit history.</p>
      </div>

      {/* ── Team Members ─────────────────────────── */}
      <div className="panel">
        <div className="panel-header">
          <div>
            <div className="panel-title">Team Members</div>
            <div className="panel-desc">Users who have access to this workspace</div>
          </div>
          <div style={{ display: "flex", gap: "0.5rem" }}>
            <button className="btn btn-secondary btn-sm" onClick={refetchUsers}>↺</button>
            <button className="btn btn-primary btn-sm" id="btn-invite-user" onClick={() => setDrawerOpen(true)}>+ Invite User</button>
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
                      <td><span className={`badge ${ROLES.find(r => r.value === u.role)?.color || "badge-neutral"}`}>{ROLES.find(r => r.value === u.role)?.label || u.role}</span></td>
                      <td style={{ color: "var(--text-2)", fontSize: "0.8rem" }}>{new Date(u.created_at).toLocaleDateString()}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          ) : (
            <EmptyState icon="◎" title="No team members yet" body="Invite colleagues to share access."
              action={<button className="btn btn-primary btn-sm" onClick={() => setDrawerOpen(true)}>+ Invite User</button>} />
          )
        )}
      </div>

      {/* ── Registered Services ─────────────────── */}
      <div className="panel">
        <div className="panel-header">
          <div>
            <div className="panel-title">Registered Services <Tooltip text="Services represent your applications (microservices, databases, collectors) that send telemetry data. API keys are linked to services for dynamic scoping and tracking." /></div>
            <div className="panel-desc">Applications linked to this workspace</div>
          </div>
          <div style={{ display: "flex", gap: "0.5rem" }}>
            <button className="btn btn-secondary btn-sm" onClick={refetchServices}>↺</button>
            <button className="btn btn-primary btn-sm" id="btn-register-service" onClick={() => setServiceDrawerOpen(true)}>+ Register Service</button>
          </div>
        </div>
        {servicesLoading ? <SectionLoader /> : (
          services?.length ? (
            <div className="table-wrap">
              <table className="data-table" id="table-services">
                <thead><tr><th>Name</th><th>Environment</th><th>Service ID</th><th>Actions</th></tr></thead>
                <tbody>
                  {services.map(s => (
                    <tr key={s.id}>
                      <td style={{ fontWeight: 500 }}>{s.name}</td>
                      <td><span className={`badge ${s.environment === "production" ? "badge-warning" : "badge-info"}`}>{s.environment}</span></td>
                      <td style={{ fontFamily: "var(--font-mono)", fontSize: "0.8rem", color: "var(--text-2)" }}>{s.id}</td>
                      <td style={{ color: "var(--text-3)", fontSize: "0.8rem" }}>—</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          ) : (
            <EmptyState icon="⊞" title="No services registered yet" body="Register your first application service to start ingesting logs and traces."
              action={<button className="btn btn-primary btn-sm" onClick={() => setServiceDrawerOpen(true)}>+ Register Service</button>} />
          )
        )}
      </div>

      {/* ── API Keys ─────────────────────────────── */}
      <div className="panel">
        <div className="panel-header">
          <div>
            <div className="panel-title">API Keys <Tooltip text="API keys authenticate your services when sending telemetry. Each key is scoped (ingest/query) and validated via Redis cache for low latency." /></div>
            <div className="panel-desc">Authenticate telemetry ingestion and queries</div>
          </div>
          <div style={{ display: "flex", gap: "0.5rem", alignItems: "center" }}>
            <button className="btn btn-secondary btn-sm" onClick={refetchKeys} title="Refresh">↺</button>
            <button 
              className="btn btn-primary btn-sm" 
              id="btn-generate-key-header" 
              onClick={() => setKeyDrawerOpen(true)}
              style={{ fontWeight: 600 }}
            >
              + Generate Key
            </button>
          </div>
        </div>
        {keysLoading ? <SectionLoader /> : (
          apiKeys?.length ? (
            <div className="table-wrap">
              <table className="data-table" id="table-api-keys">
                <thead><tr><th>Name</th><th>Prefix</th><th>Scopes</th><th>Status</th><th>Last Used</th><th>Actions <Tooltip text="Rotate: new secret, old key invalidated instantly. Revoke: permanent — services get 401 immediately." position="left" /></th></tr></thead>
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
                            <Tooltip text="Generates a new secret. Old key purged from Redis cache immediately.">
                              <button id={`btn-rotate-${k.id}`} className="btn btn-secondary btn-sm" onClick={() => handleRotateKey(k.id, k.scopes)}>Rotate</button>
                            </Tooltip>
                            <Tooltip text="Permanently deactivates. Services using this key will get 401 errors immediately." position="left">
                              <button id={`btn-revoke-${k.id}`} className="btn btn-sm" style={{ background:"var(--danger-soft)",color:"var(--danger)",border:"1px solid rgba(239,68,68,0.25)" }} onClick={() => handleRevokeKey(k.id)}>Revoke</button>
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
            <EmptyState 
              icon="⊞" 
              title="No API keys" 
              body="Create a new API key to authenticate your services for log and trace ingestion." 
              action={
                <button className="btn btn-primary btn-sm" onClick={() => setKeyDrawerOpen(true)}>
                  + Generate First Key
                </button>
              }
            />
          )
        )}
      </div>

      {/* ── Audit Log ────────────────────────────── */}
      <div className="panel">
        <div className="panel-header">
          <div>
            <div className="panel-title">Audit Log <Tooltip text="Every user and system action on this tenant is recorded here — key creation, user changes, rule updates, and API calls." /></div>
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
                      <td style={{ color:"var(--text-2)",fontSize:"0.8rem" }}>{new Date(log.created_at).toLocaleString()}</td>
                      <td><span className={`badge ${log.actor_type === "user" ? "badge-info" : "badge-neutral"}`}>{log.actor_type}</span></td>
                      <td style={{ fontFamily:"var(--font-mono)",fontSize:"0.78rem" }}>{log.action}</td>
                      <td style={{ fontSize:"0.78rem",color:"var(--text-2)" }}>{log.resource_type} / {log.resource_id?.slice(0, 20)}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          ) : <EmptyState icon="◑" title="No audit logs yet" body="All user and system operations will appear here." />
        )}
      </div>

      {/* ── Create User Drawer ───────────────────── */}
      <Drawer open={drawerOpen} onClose={() => setDrawerOpen(false)} title="Invite Team Member" description="Create a new user account with access to this tenant workspace.">
        <form onSubmit={handleCreateUser} style={{ display:"flex",flexDirection:"column",gap:"1.25rem" }}>
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
            <p style={{ fontSize:"0.75rem",color:"var(--text-3)",marginTop:"0.25rem" }}>The user can change this after first login.</p>
          </div>
          <div className="form-group">
            <label className="form-label">Role <Tooltip text="Choose the level of access this user should have." /></label>
            <div style={{ display:"flex",flexDirection:"column",gap:"0.5rem",marginTop:"0.25rem" }}>
              {ROLES.map(role => (
                <label key={role.value} style={{
                  display:"flex",alignItems:"flex-start",gap:"0.75rem",padding:"0.75rem",
                  borderRadius:"var(--r-sm)",cursor:"pointer",
                  border:`1.5px solid ${userForm.role === role.value ? "var(--primary)" : "var(--border)"}`,
                  background: userForm.role === role.value ? "var(--primary-soft)" : "var(--surface-active)",
                  transition:"all 0.15s",
                }}>
                  <input type="radio" name="role" value={role.value} checked={userForm.role === role.value} onChange={() => uf("role", role.value)} style={{ marginTop:"2px",accentColor:"var(--primary)" }} />
                  <div>
                    <div style={{ fontSize:"0.875rem",fontWeight:600,display:"flex",alignItems:"center",gap:"0.5rem" }}>
                      {role.label} <span className={`badge ${role.color}`} style={{ fontSize:"0.65rem" }}>{role.value}</span>
                    </div>
                    <div style={{ fontSize:"0.78rem",color:"var(--text-2)",marginTop:"0.2rem",lineHeight:1.5 }}>{role.desc}</div>
                  </div>
                </label>
              ))}
            </div>
          </div>
          <div style={{ display:"flex",gap:"0.65rem",paddingTop:"0.5rem",borderTop:"1px solid var(--border)" }}>
            <button type="button" className="btn btn-ghost" style={{ flex:1 }} onClick={() => setDrawerOpen(false)}>Cancel</button>
            <button type="submit" className="btn btn-primary" id="btn-create-user" style={{ flex:2,justifyContent:"center" }} disabled={creating}>
              {creating ? "Creating…" : "Create User"}
            </button>
          </div>
        </form>
      </Drawer>

      {/* ── Create API Key Drawer ────────────────── */}
      <Drawer open={keyDrawerOpen} onClose={() => { setKeyDrawerOpen(false); setNewlyCreatedKey(""); }} title="Generate API Key" description={newlyCreatedKey ? "Please copy your new API key secret." : "Create a new credential for telemetry ingestion or querying."}>
        {newlyCreatedKey ? (
          <div style={{ display:"flex", flexDirection:"column", gap:"1.25rem" }}>
            <div style={{ padding:"1rem", borderRadius:"var(--r-sm)", background:"var(--success-soft)", border:"1px solid rgba(16,185,129,0.2)", color:"var(--success)", fontSize:"0.875rem", lineHeight:1.5 }}>
              <strong>API Key Generated!</strong>
              <p style={{ marginTop:"0.25rem", color:"var(--text-2)", fontSize:"0.78rem" }}>Copy this key now. For security reasons, it cannot be shown again.</p>
            </div>
            <div className="form-group">
              <label className="form-label">API Key Secret</label>
              <div style={{ display:"flex", gap:"0.5rem" }}>
                <input className="form-input" readOnly value={newlyCreatedKey} style={{ fontFamily:"var(--font-mono)", fontSize:"0.82rem", background:"rgba(0,0,0,0.2)", flex: 1 }} />
                <button type="button" className="btn btn-primary" onClick={() => {
                  navigator.clipboard.writeText(newlyCreatedKey);
                  notify("API Key copied to clipboard!", "success");
                }}>Copy</button>
              </div>
            </div>
            <div style={{ display:"flex", paddingTop:"0.5rem", borderTop:"1px solid var(--border)" }}>
              <button type="button" className="btn btn-primary" style={{ width:"100%", justifyContent:"center" }} onClick={() => {
                setNewlyCreatedKey("");
                setKeyDrawerOpen(false);
              }}>Done</button>
            </div>
          </div>
        ) : (
          <form onSubmit={handleCreateKey} style={{ display:"flex",flexDirection:"column",gap:"1.25rem" }}>
            <div className="form-group">
              <label className="form-label">Key Name</label>
              <input className="form-input" placeholder="e.g. Production Ingest" value={keyForm.name} onChange={e => kf("name", e.target.value)} required />
            </div>
            <div className="form-group">
              <label className="form-label">Scope <Tooltip text="Ingest: can send logs/traces. Query: can access the query API." /></label>
              <select className="form-input" value={keyForm.scope} onChange={e => kf("scope", e.target.value)} style={{ appearance:"none",background:"var(--surface-active) url(\"data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' fill='none' stroke='%2394A3B8' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'%3E%3Cpath d='M3 4.5l3 3 3-3'/%3E%3C/svg%3E\") no-repeat right 0.75rem center" }}>
                <option value="ingest">Ingest (Logs & Traces)</option>
                <option value="query">Query (Dashboards & CLI)</option>
                <option value="*">Full Access (*)</option>
              </select>
            </div>
            <div className="form-group">
              <label className="form-label">Service <Tooltip text="Link this key to a specific service for better data categorization." /></label>
              <select className="form-input" value={keyForm.serviceId} onChange={e => kf("serviceId", e.target.value)} style={{ appearance:"none",background:"var(--surface-active) url(\"data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' fill='none' stroke='%2394A3B8' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'%3E%3Cpath d='M3 4.5l3 3 3-3'/%3E%3C/svg%3E\") no-repeat right 0.75rem center" }}>
                <option value="">-- Select Service (Optional) --</option>
                {services?.map(s => (
                  <option key={s.id} value={s.id}>{s.name} ({s.environment})</option>
                ))}
              </select>
              <p style={{ fontSize: "0.75rem", color: "var(--text-3)", marginTop: "0.25rem" }}>
                If no service is selected, the first available service will be used.
              </p>
            </div>
            <div style={{ padding:"1rem", borderRadius:"var(--r-sm)", background:"var(--info-soft)", border:"1px solid rgba(59,130,246,0.2)", fontSize:"0.78rem", color:"var(--info)", lineHeight:1.5 }}>
              <strong>Security Note:</strong> New keys are cached in Redis for high-performance validation. After creation, the full secret will be displayed once.
            </div>
            <div style={{ display:"flex",gap:"0.65rem",paddingTop:"0.5rem",borderTop:"1px solid var(--border)" }}>
              <button type="button" className="btn btn-ghost" style={{ flex:1 }} onClick={() => setKeyDrawerOpen(false)}>Cancel</button>
              <button type="submit" className="btn btn-primary" id="btn-save-key" style={{ flex:2,justifyContent:"center" }} disabled={creating}>
                {creating ? "Generating…" : "Generate Key"}
              </button>
            </div>
          </form>
        )}
      </Drawer>
      {/* ── Register Service Drawer ──────────────── */}
      <Drawer open={serviceDrawerOpen} onClose={() => setServiceDrawerOpen(false)} title="Register Service" description="Register a new application service under this tenant workspace.">
        <form onSubmit={handleCreateService} style={{ display:"flex",flexDirection:"column",gap:"1.25rem" }}>
          <div className="form-group">
            <label className="form-label">Service Name</label>
            <input className="form-input" placeholder="e.g. payment-gateway" value={serviceForm.name} onChange={e => sf("name", e.target.value)} required />
          </div>
          <div className="form-group">
            <label className="form-label">Environment</label>
            <select className="form-input" value={serviceForm.environment} onChange={e => sf("environment", e.target.value)} style={{ appearance:"none",background:"var(--surface-active) url(\"data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' fill='none' stroke='%2394A3B8' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'%3E%3Cpath d='M3 4.5l3 3 3-3'/%3E%3C/svg%3E\") no-repeat right 0.75rem center" }}>
              <option value="production">Production</option>
              <option value="staging">Staging</option>
              <option value="development">Development</option>
              <option value="testing">Testing</option>
            </select>
          </div>
          <div style={{ display:"flex",gap:"0.65rem",paddingTop:"0.5rem",borderTop:"1px solid var(--border)" }}>
            <button type="button" className="btn btn-ghost" style={{ flex:1 }} onClick={() => setServiceDrawerOpen(false)}>Cancel</button>
            <button type="submit" className="btn btn-primary" id="btn-save-service" style={{ flex:2,justifyContent:"center" }} disabled={creating}>
              {creating ? "Registering…" : "Register Service"}
            </button>
          </div>
        </form>
      </Drawer>
    </div>
  );
}
