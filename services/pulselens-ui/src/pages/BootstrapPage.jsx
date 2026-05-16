import { useState } from "react";
import { tenantApi } from "../lib/api";

const STEPS = [
  { n: 1, label: "Workspace" },
  { n: 2, label: "Admin" },
  { n: 3, label: "Review" },
];

export default function BootstrapPage({ setState, notify }) {
  const [step, setStep] = useState(1);
  const [loading, setLoading] = useState(false);
  const [successData, setSuccessData] = useState(null);

  const [form, setForm] = useState({
    name: "", slug: "", admin_name: "", admin_email: "", admin_password: "", service_name: "web-app"
  });
  const f = (k, v) => setForm(p => ({ ...p, [k]: v }));

  const slugify = (s) => s.toLowerCase().replace(/\s+/g, "-").replace(/[^a-z0-9-]/g, "").replace(/-+/g, "-").slice(0, 40);

  const next = () => {
    if (step === 1 && (!form.name || !form.slug)) return notify("Please fill out workspace details", "error");
    if (step === 2 && (!form.admin_name || !form.admin_email || !form.admin_password))
      return notify("Please fill out all admin fields", "error");
    if (step === 2 && form.admin_password.length < 8)
      return notify("Password must be at least 8 characters", "error");
    setStep(s => s + 1);
  };
  const prev = () => setStep(s => s - 1);

  const submit = async () => {
    setLoading(true);
    try {
      let tenantId;

      // 1. Create Tenant (single call — creates tenant + admin user atomically)
      try {
        const tRes = await tenantApi.createTenant({
          name: form.name,
          slug: form.slug,
          plan: "starter",
          admin_name: form.admin_name,
          admin_email: form.admin_email,
          admin_password: form.admin_password,
        });
        tenantId = tRes?.tenant?.id || tRes?.id;
      } catch (e) {
        if (e.message?.includes("already exists")) {
          // Tenant/slug collision → try logging in with provided credentials
          notify("Workspace already exists — attempting login…", "info");
        } else {
          throw e;
        }
      }

      // 2. Login with admin credentials to get a real JWT
      const loginRes = await tenantApi.login({ email: form.admin_email, password: form.admin_password });
      const token = loginRes.token;
      if (!tenantId) tenantId = loginRes.tenant_id;

      // 3. Create default service & API key for ingestion
      let apiKey = "";
      try {
        let serviceId = "all";
        try {
          const sRes = await tenantApi.createService(tenantId, { name: form.service_name || "web-app", type: "web" }, token);
          serviceId = sRes.id || "all";
        } catch (e) { console.error("Service creation failed", e); }

        const keyRes = await tenantApi.createAPIKey({
          tenant_id: tenantId,
          service_id: serviceId,
          name: "Ingestion Key",
          scopes: ["ingest"]
        }, token);
        apiKey = keyRes.key || keyRes.api_key || keyRes.value || "";
      } catch (e) {
        console.error("API key creation failed", e);
      }

      setSuccessData({ tenantId, apiKey, token });
      setStep(4);
    } catch (e) {
      notify(e.message || "Provisioning failed", "error");
    } finally {
      setLoading(false);
    }
  };

  const handleFinish = () => {
    setState({
      token: successData.token,
      email: form.admin_email,
      tenantId: successData.tenantId,
      apiKey: successData.apiKey,
    });
    window.location.hash = "overview";
  };

  const progressPct = step >= 4 ? 100 : ((step - 1) / (STEPS.length - 1)) * 100;

  return (
    <div style={{ display: "flex", minHeight: "100vh", background: "var(--bg)" }}>
      {/* Left panel */}
      <div style={{
        width: "340px", flexShrink: 0,
        background: "linear-gradient(160deg, #0B1120 0%, #030712 100%)",
        borderRight: "1px solid var(--border)",
        display: "flex", flexDirection: "column", justifyContent: "center",
        padding: "3rem 2.5rem",
        position: "relative", overflow: "hidden"
      }}>
        <div style={{ position:"absolute", top:"-20%", right:"-30%", width:"300px", height:"300px", borderRadius:"50%", background:"radial-gradient(circle, rgba(99,102,241,0.12), transparent 60%)", pointerEvents:"none" }} />

        <div style={{ display:"flex", alignItems:"center", gap:"0.75rem", marginBottom:"2rem" }}>
          <div className="brand-icon" style={{ width:"36px", height:"36px", fontSize:"16px" }}>PL</div>
          <span className="brand-name" style={{ fontSize:"1.1rem" }}>PulseLens</span>
        </div>

        <h2 style={{ fontSize:"1.5rem", fontWeight:700, lineHeight:1.2, marginBottom:"1rem" }}>
          Provision your<br />observability workspace
        </h2>
        <p style={{ fontSize:"0.875rem", color:"var(--text-2)", lineHeight:1.6, marginBottom:"2rem" }}>
          A dedicated, isolated environment with its own ingestion pipeline, data store, and alert routing.
        </p>

        {/* Step list */}
        <div style={{ display:"flex", flexDirection:"column", gap:"1rem" }}>
          {STEPS.map(({ n, label }) => (
            <div key={n} style={{ display:"flex", alignItems:"center", gap:"0.75rem" }}>
              <div style={{
                width:"28px", height:"28px", borderRadius:"50%", flexShrink:0,
                background: step > n ? "var(--success)" : step === n ? "var(--primary)" : "var(--surface-active)",
                border: `2px solid ${step > n ? "var(--success)" : step === n ? "var(--primary)" : "var(--border)"}`,
                display:"flex", alignItems:"center", justifyContent:"center",
                fontSize:"0.75rem", fontWeight:700, color:"white",
                transition:"all 0.3s",
                boxShadow: step === n ? "0 0 14px var(--primary-glow)" : "none"
              }}>
                {step > n ? "✓" : n}
              </div>
              <span style={{ fontSize:"0.875rem", fontWeight: step === n ? 600 : 400, color: step === n ? "var(--text)" : "var(--text-2)" }}>
                {label}
              </span>
            </div>
          ))}
        </div>
      </div>

      {/* Right panel — form */}
      <div style={{ flex:1, display:"flex", alignItems:"center", justifyContent:"center", padding:"2rem" }}>
        <div style={{ width:"100%", maxWidth:"520px" }}>

          {/* Progress bar */}
          {step < 4 && (
            <div style={{ height:"3px", background:"var(--surface-active)", borderRadius:"2px", marginBottom:"2.5rem", overflow:"hidden" }}>
              <div style={{ height:"100%", width:`${progressPct}%`, background:"linear-gradient(90deg, var(--primary), var(--cyan))", borderRadius:"2px", transition:"width 0.4s" }} />
            </div>
          )}

          {/* ── Step 1: Workspace ─────────────── */}
          {step === 1 && (
            <div className="animate-fade-in">
              <h3 style={{ fontSize:"1.5rem", fontWeight:700, marginBottom:"0.4rem" }}>Name your workspace</h3>
              <p style={{ color:"var(--text-2)", marginBottom:"2rem", fontSize:"0.9rem" }}>This will be your isolated telemetry namespace.</p>

              <div className="form-group">
                <label className="form-label">Company / Project Name</label>
                <input
                  className="form-input"
                  placeholder="e.g. Acme Corp"
                  value={form.name}
                  onChange={e => {
                    const v = e.target.value;
                    f("name", v);
                    f("slug", slugify(v));
                  }}
                />
              </div>

              <div className="form-group">
                <label className="form-label">Workspace Slug</label>
                <input
                  className="form-input"
                  placeholder="acme-corp"
                  value={form.slug}
                  onChange={e => f("slug", slugify(e.target.value))}
                  style={{ fontFamily:"var(--font-mono)" }}
                />
                <p style={{ fontSize:"0.75rem", color:"var(--text-3)", marginTop:"0.3rem" }}>
                  Used for API routing: <code style={{color:"var(--cyan)"}}>POST /ingest</code> (unique per workspace)
                </p>
              </div>

              <div style={{ display:"flex", justifyContent:"flex-end", marginTop:"2rem" }}>
                <button className="btn btn-primary" onClick={next}>Continue →</button>
              </div>
            </div>
          )}

          {/* ── Step 2: Admin ─────────────────── */}
          {step === 2 && (
            <div className="animate-fade-in">
              <h3 style={{ fontSize:"1.5rem", fontWeight:700, marginBottom:"0.4rem" }}>Create admin account</h3>
              <p style={{ color:"var(--text-2)", marginBottom:"2rem", fontSize:"0.9rem" }}>This will be the primary owner of the workspace.</p>

              <div className="form-group">
                <label className="form-label">Full Name</label>
                <input className="form-input" placeholder="Jane Doe" value={form.admin_name} onChange={e => f("admin_name", e.target.value)} />
              </div>
              <div className="form-group">
                <label className="form-label">Email Address</label>
                <input className="form-input" type="email" placeholder="jane@acme.com" value={form.admin_email} onChange={e => f("admin_email", e.target.value)} />
              </div>
              <div className="form-group">
                <label className="form-label">Password <span style={{color:"var(--text-3)",fontSize:"0.75rem"}}>(min 8 chars)</span></label>
                <input className="form-input" type="password" placeholder="••••••••" value={form.admin_password} onChange={e => f("admin_password", e.target.value)} />
              </div>

              <div style={{ display:"flex", justifyContent:"space-between", marginTop:"2rem" }}>
                <button className="btn btn-ghost" onClick={prev}>← Back</button>
                <button className="btn btn-primary" onClick={next}>Continue →</button>
              </div>
            </div>
          )}

          {/* ── Step 3: Review & Provision ─────── */}
          {step === 3 && (
            <div className="animate-fade-in">
              <h3 style={{ fontSize:"1.5rem", fontWeight:700, marginBottom:"0.4rem" }}>Ready to provision</h3>
              <p style={{ color:"var(--text-2)", marginBottom:"2rem", fontSize:"0.9rem" }}>Review your configuration before creating the workspace.</p>

              <div className="panel" style={{ background:"var(--bg)", gap:"0" }}>
                {[
                  ["Workspace Name", form.name, false],
                  ["Workspace Slug", form.slug, true],
                  ["Admin Email", form.admin_email, false],
                  ["Admin Name", form.admin_name, false],
                  ["Plan", "Starter", false],
                ].map(([k, v, mono]) => (
                  <div key={k} style={{ display:"flex", justifyContent:"space-between", alignItems:"center", padding:"0.65rem 0", borderBottom:"1px solid var(--border)" }}>
                    <span style={{ fontSize:"0.85rem", color:"var(--text-2)" }}>{k}</span>
                    <span style={{ fontSize:"0.85rem", fontFamily: mono ? "var(--font-mono)" : "inherit", color: mono ? "var(--cyan)" : "var(--text)", fontWeight:600 }}>{v}</span>
                  </div>
                ))}
                <div style={{ display:"flex", alignItems:"center", gap:"0.5rem", padding:"0.65rem 0", fontSize:"0.85rem", color:"var(--text-2)" }}>
                  <span className="status-dot success" />
                  Ingestion API key will be generated automatically
                </div>
              </div>

              <div style={{ display:"flex", justifyContent:"space-between", marginTop:"2rem" }}>
                <button className="btn btn-ghost" onClick={prev} disabled={loading}>← Back</button>
                <button className="btn btn-primary btn-lg" onClick={submit} disabled={loading} style={{ minWidth:"180px", justifyContent:"center" }}>
                  {loading ? (
                    <span style={{ display:"flex", alignItems:"center", gap:"0.5rem" }}>
                      <span style={{ width:"14px", height:"14px", borderRadius:"50%", border:"2px solid rgba(255,255,255,0.3)", borderTopColor:"white", animation:"spin 0.8s linear infinite" }} />
                      Provisioning…
                    </span>
                  ) : "🚀 Provision Workspace"}
                </button>
              </div>
            </div>
          )}

          {/* ── Step 4: Success ───────────────── */}
          {step === 4 && (
            <div className="animate-fade-in" style={{ textAlign:"center" }}>
              <div style={{
                width:"72px", height:"72px", borderRadius:"50%",
                background:"var(--success-soft)", color:"var(--success)",
                display:"flex", alignItems:"center", justifyContent:"center",
                fontSize:"2.5rem", margin:"0 auto 1.5rem",
                boxShadow:"0 0 30px var(--success-glow)"
              }}>✓</div>

              <h2 style={{ fontSize:"1.75rem", fontWeight:800, marginBottom:"0.75rem" }}>Workspace Ready!</h2>
              <p style={{ color:"var(--text-2)", marginBottom:"2rem", fontSize:"0.9rem" }}>
                Your isolated telemetry pipeline is provisioned and ready to receive data.
              </p>

              {successData?.apiKey && (
                <div className="panel" style={{ background:"var(--bg)", textAlign:"left", marginBottom:"1.5rem" }}>
                  <div style={{ fontSize:"0.75rem", color:"var(--text-3)", textTransform:"uppercase", letterSpacing:"0.08em", fontWeight:700, marginBottom:"0.5rem" }}>
                    🔑 Ingestion API Key — save this now!
                  </div>
                  <div style={{
                    background:"var(--surface-active)", padding:"0.75rem 1rem",
                    borderRadius:"var(--r-sm)", fontFamily:"var(--font-mono)",
                    fontSize:"0.78rem", color:"var(--cyan)", wordBreak:"break-all",
                    border:"1px solid rgba(6,182,212,0.2)"
                  }}>
                    {successData.apiKey}
                  </div>
                  <p style={{ fontSize:"0.75rem", color:"var(--text-3)", marginTop:"0.5rem" }}>
                    Pass this as <code>X-API-Key</code> header when sending telemetry events.
                  </p>
                </div>
              )}

              <button className="btn btn-primary btn-lg" style={{ width:"100%", justifyContent:"center" }} onClick={handleFinish}>
                Enter Dashboard →
              </button>
            </div>
          )}
        </div>
      </div>

      <style>{`
        @keyframes spin { to { transform: rotate(360deg); } }
      `}</style>
    </div>
  );
}
