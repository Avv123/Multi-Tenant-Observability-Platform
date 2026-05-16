import { useState } from "react";
import { tenantApi } from "../lib/api";

export default function LoginPage({ setState, notify }) {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!email || !password) { notify("Please fill in both fields", "error"); return; }
    setLoading(true);
    try {
      const res = await tenantApi.login({ email, password });
      setState({
        token: res.token,
        email,
        tenantId: res.tenant_id || "",
        apiKey: res.api_key || ""
      });
      notify("Authenticated successfully", "success");
    } catch (err) {
      notify(err.message || "Invalid credentials", "error");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div style={{ display: "flex", height: "100vh", width: "100vw", overflow: "hidden" }}>
      {/* ── Left: Brand / Marketing ─────────────────────── */}
      <div style={{
        flex: 1,
        background: "linear-gradient(160deg, #060D1F 0%, #030712 100%)",
        position: "relative",
        display: "flex",
        flexDirection: "column",
        justifyContent: "center",
        padding: "4rem",
        overflow: "hidden"
      }}>
        {/* Glow blobs */}
        <div style={{ position:"absolute", top:"-15%", left:"-10%", width:"500px", height:"500px", borderRadius:"50%", background:"radial-gradient(circle, rgba(99,102,241,0.12) 0%, transparent 60%)", pointerEvents:"none" }} />
        <div style={{ position:"absolute", bottom:"-15%", right:"-10%", width:"400px", height:"400px", borderRadius:"50%", background:"radial-gradient(circle, rgba(6,182,212,0.09) 0%, transparent 60%)", pointerEvents:"none" }} />

        {/* Grid dots overlay */}
        <div style={{ position:"absolute", inset:0, backgroundImage:"radial-gradient(rgba(255,255,255,0.035) 1px, transparent 1px)", backgroundSize:"30px 30px", pointerEvents:"none" }} />

        <div style={{ position:"relative", zIndex:1, maxWidth:"560px" }}>
          <div style={{ display:"inline-flex", alignItems:"center", gap:"0.9rem", marginBottom:"2.5rem" }}>
            <div className="brand-icon" style={{ width:"44px", height:"44px", fontSize:"20px" }}>PL</div>
            <span className="brand-name" style={{ fontSize:"1.5rem" }}>PulseLens</span>
          </div>

          <h1 style={{ fontSize:"3rem", fontWeight:800, lineHeight:1.1, marginBottom:"1.25rem", letterSpacing:"-0.03em" }}>
            The Modern<br/>
            <span style={{ background:"linear-gradient(90deg, #818CF8 0%, #22D3EE 100%)", WebkitBackgroundClip:"text", WebkitTextFillColor:"transparent" }}>
              Observability
            </span><br/>
            Platform
          </h1>

          <p style={{ fontSize:"1.1rem", color:"var(--text-2)", lineHeight:1.65, marginBottom:"2.5rem" }}>
            Unified metrics, logs, traces & incident management — built for multi-tenant SaaS at any scale.
          </p>

          <div className="grid-2" style={{ gap:"1rem" }}>
            {[
              { icon:"⚡", head:"Real-time ingestion", body:"Kafka-backed pipeline with sub-second latency." },
              { icon:"🔐", head:"Tenant isolation", body:"Strict data segregation & quota enforcement built-in." },
              { icon:"📊", head:"Rich analytics", body:"ClickHouse powers instant query across billions of events." },
              { icon:"🔔", head:"Smart alerts", body:"Rule-based firing with incident lifecycle management." },
            ].map(({ icon, head, body }) => (
              <div key={head} style={{ padding:"1rem", borderRadius:"var(--r-md)", background:"rgba(255,255,255,0.025)", border:"1px solid rgba(255,255,255,0.07)" }}>
                <div style={{ fontSize:"1.4rem", marginBottom:"0.5rem" }}>{icon}</div>
                <div style={{ fontWeight:600, fontSize:"0.9rem", marginBottom:"0.25rem" }}>{head}</div>
                <div style={{ fontSize:"0.8rem", color:"var(--text-2)", lineHeight:1.5 }}>{body}</div>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* ── Right: Login Form ────────────────────────────── */}
      <div style={{
        width:"460px",
        background:"var(--bg)",
        borderLeft:"1px solid var(--border)",
        display:"flex",
        flexDirection:"column",
        justifyContent:"center",
        padding:"3.5rem",
        position:"relative"
      }}>
        <div style={{ marginBottom:"2.5rem" }}>
          <h2 style={{ fontSize:"1.6rem", fontWeight:700, marginBottom:"0.5rem" }}>Welcome back</h2>
          <p style={{ color:"var(--text-2)", fontSize:"0.9rem" }}>Sign in to your PulseLens workspace.</p>
        </div>

        <form onSubmit={handleSubmit} noValidate>
          <div className="form-group">
            <label className="form-label">Email Address</label>
            <input
              id="login-email"
              type="email"
              className="form-input"
              placeholder="you@company.com"
              value={email}
              onChange={e => setEmail(e.target.value)}
              autoComplete="username"
              required
            />
          </div>

          <div className="form-group">
            <label className="form-label" style={{ display:"flex", justifyContent:"space-between" }}>
              Password
              <span style={{ color:"var(--primary)", fontSize:"0.78rem", cursor:"pointer" }}>Forgot?</span>
            </label>
            <input
              id="login-password"
              type="password"
              className="form-input"
              placeholder="••••••••"
              value={password}
              onChange={e => setPassword(e.target.value)}
              autoComplete="current-password"
              required
            />
          </div>

          <button
            id="btn-login"
            type="submit"
            className="btn btn-primary btn-lg"
            style={{ width:"100%", marginTop:"0.75rem", justifyContent:"center" }}
            disabled={loading}
          >
            {loading ? "Signing in…" : "Sign In →"}
          </button>
        </form>

        <div style={{ marginTop:"2.5rem", textAlign:"center" }}>
          <p style={{ color:"var(--text-3)", fontSize:"0.85rem" }}>
            New to PulseLens?{" "}
            <a href="#bootstrap" style={{ color:"var(--primary)", fontWeight:600, textDecoration:"none" }}>
              Set up your first tenant
            </a>
          </p>
        </div>

        {/* Demo credentials hint */}
        <div style={{
          marginTop:"1.5rem", padding:"0.875rem 1rem",
          background:"rgba(99,102,241,0.07)", border:"1px solid rgba(99,102,241,0.18)",
          borderRadius:"var(--r-sm)", fontSize:"0.8rem"
        }}>
          <div style={{ color:"var(--primary-2)", fontWeight:600, marginBottom:"0.3rem" }}>🔑 Demo Credentials</div>
          <div style={{ color:"var(--text-2)", fontFamily:"var(--font-mono)", fontSize:"0.75rem" }}>
            <div>Email: demo@pulselens.io</div>
            <div>Password: pulselens123</div>
          </div>
        </div>


        {/* Decorative corner glow */}
        <div style={{ position:"absolute", top:0, right:0, width:"200px", height:"200px", background:"radial-gradient(circle at top right, rgba(99,102,241,0.07) 0%, transparent 70%)", pointerEvents:"none" }} />
      </div>
    </div>
  );
}
