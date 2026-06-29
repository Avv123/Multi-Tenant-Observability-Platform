import { useCallback, useEffect, useReducer, useRef, useState } from "react";
import "./styles.css";
import BootstrapPage from "./pages/BootstrapPage";
import LoginPage from "./pages/LoginPage";
import OverviewPage from "./pages/OverviewPage";
import LogsPage from "./pages/LogsPage";
import MetricsPage from "./pages/MetricsPage";
import TracesPage from "./pages/TracesPage";
import TransactionsPage from "./pages/TransactionsPage";
import ErrorsPage from "./pages/ErrorsPage";
import AlertsPage from "./pages/AlertsPage";
import IncidentsPage from "./pages/IncidentsPage";
import ServiceMapPage from "./pages/ServiceMapPage";
import PlatformPage from "./pages/PlatformPage";
import SettingsPage from "./pages/SettingsPage";
import { loadState, saveState } from "./lib/storage";
import { tenantApi } from "./lib/api";

// Decode JWT payload without a library
function jwtRole(token) {
  try {
    if (!token || token === "bootstrap-session") return "tenant";
    const payload = JSON.parse(atob(token.split(".")[1]));
    return payload.role || "tenant";
  } catch { return "tenant"; }
}

// SVG Icons
const Icons = {
  Home: () => <svg fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6" /></svg>,
  Server: () => <svg fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01" /></svg>,
  List: () => <svg fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 10h16M4 14h16M4 18h16" /></svg>,
  Chart: () => <svg fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M7 12l3-3 3 3 4-4M8 21l4-4 4 4M3 4h18M4 4h16v12a1 1 0 01-1 1H5a1 1 0 01-1-1V4z" /></svg>,
  Target: () => <svg fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 14v3m4-3v3m4-3v3M3 21h18M3 10h18M3 7l9-4 9 4M4 10h16v11H4V10z" /></svg>,
  Shield: () => <svg fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" /></svg>,
  Bell: () => <svg fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" /></svg>,
  Search: () => <svg fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" /></svg>,
  Archive: () => <svg fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 8h14M5 8a2 2 0 110-4h14a2 2 0 110 4M5 8v10a2 2 0 002 2h10a2 2 0 002-2V8m-9 4h4" /></svg>,
  Settings: () => <svg fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" /><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" /></svg>,
  Zap: () => <svg fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" /></svg>,
  Flash: () => <svg fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 2l3.09 6.26L22 9.27l-5 4.87L18.18 21 12 17.77 5.82 21 7 14.14l-5-4.87 6.91-1.01L12 2z" /></svg>,
  AlertTriangle: () => <svg fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10.29 3.86L1.82 18a2 2 0 001.71 3h16.94a2 2 0 001.71-3L13.71 3.86a2 2 0 00-3.42 0zM12 9v4m0 4h.01" /></svg>,
  Globe: () => <svg fill="none" stroke="currentColor" viewBox="0 0 24 24"><circle cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="2" /><path stroke="currentColor" strokeWidth="2" d="M2 12h20M12 2a15.3 15.3 0 014 10 15.3 15.3 0 01-4 10M12 2a15.3 15.3 0 00-4 10 15.3 15.3 0 004 10" /></svg>,
};

function getHash() { return window.location.hash.replace(/^#\/?/, "") || "overview"; }
function useRouter() {
  const [route, setRoute] = useState(getHash);
  useEffect(() => {
    const fn = () => setRoute(getHash());
    window.addEventListener("hashchange", fn);
    return () => window.removeEventListener("hashchange", fn);
  }, []);
  return { route, navigate: (p) => { window.location.hash = p; } };
}

let _tid = 0;
function toastReducer(s, a) {
  if (a.type === "ADD") return [...s, a.toast];
  if (a.type === "REMOVE") return s.filter(t => t.id !== a.id);
  return s;
}

const PAGE_META = {
  overview:     { label: "Overview",            icon: Icons.Home          },
  logs:         { label: "Logs",                icon: Icons.List          },
  metrics:      { label: "Metrics",             icon: Icons.Chart         },
  traces:       { label: "Traces",              icon: Icons.Target        },
  transactions:  { label: "Transactions",        icon: Icons.Flash         },
  "service-map": { label: "Service Map",          icon: Icons.Globe         },
  errors:       { label: "Error Inbox",         icon: Icons.AlertTriangle },
  alerts:       { label: "Alert Rules",         icon: Icons.Bell          },
  incidents:    { label: "Incidents",           icon: Icons.Zap           },
  platform:     { label: "Platform",            icon: Icons.Server        },
  settings:     { label: "Settings",            icon: Icons.Settings      },
  bootstrap:    { label: "Setup Wizard",        icon: Icons.Shield        },
};

function ToastStack({ toasts, dismiss }) {
  return (
    <div className="toast-stack" aria-live="polite">
      {toasts.map(t => (
        <div key={t.id} className={`toast ${t.tone}`} onClick={() => dismiss(t.id)} role="alert">
          <span style={{ flex: 1 }}>{t.message}</span>
        </div>
      ))}
    </div>
  );
}

function SideNav({ route, navigate, state, onLogout }) {
  const initials = state.email ? state.email.substring(0, 2).toUpperCase() : "PL";

  // Decode role from real JWT — fallback to email check for dev convenience
  const role = jwtRole(state.token);
  const isAdmin = role === "super_admin" || role === "platform_admin" || state.email === "admin@pulselens.io";

  const adminNav = [
    {
      group: "Platform Ops", items: [
        { path: "platform", label: "Infrastructure" },
        { path: "overview", label: "Tenant Overview" },
      ]
    },
    {
      group: "Management", items: [
        { path: "bootstrap", label: "Onboard Tenant" },
        { path: "settings", label: "Settings" },
      ]
    },
  ];

  const tenantNav = [
    {
      group: "Observability", items: [
        { path: "overview",      label: "Dashboard"         },
        { path: "logs",          label: "Logs Explorer"     },
        { path: "metrics",       label: "Metrics"           },
        { path: "traces",        label: "Distributed Traces"},
        { path: "transactions",  label: "Transactions"      },
        { path: "errors",        label: "Error Inbox"       },
      ]
    },
    {
      group: "Reliability", items: [
        { path: "incidents", label: "Active Incidents" },
        { path: "alerts",    label: "Alert Rules"      },
      ]
    },
    {
      group: "Management", items: [
        { path: "settings", label: "Settings" },
      ]
    }
  ];

  const navToRender = isAdmin ? adminNav : tenantNav;

  return (
    <nav className="app-nav">
      <div className="app-nav__brand">
        <div className="brand-icon">PL</div>
        <span className="brand-name">PulseLens</span> 
      </div>

      <div className="app-nav__scroll">
        {navToRender.map(({ group, items }) => (
          <div key={group} style={{ marginBottom: "1rem" }}>
            <div className="nav-section-title">{group}</div>
            {items.map(({ path, label }) => {
              const Icon = PAGE_META[path].icon;
              return (
                <button
                  key={path}
                  id={`nav-${path}`}
                  className={`nav-link ${route === path ? "active" : ""}`}
                  onClick={() => navigate(path)}
                >
                  <Icon />
                  {label}
                </button>
              );
            })}
          </div>
        ))}
      </div>

      <div className="app-nav__footer">
        <div className="user-pill">
          <div className="user-avatar">{initials}</div>
          <div className="user-info">
            <div className="user-name">{state.email || "Not logged in"}</div>
            <div className="user-role" style={{ color: isAdmin ? "var(--cyan)" : "var(--primary-2)" }}>
              {isAdmin ? "Platform Admin" : (state.tenantId ? state.tenantId.split("-")[0] : "Tenant")}
            </div>
          </div>
        </div>
        {state.token && (
          <button className="btn btn-ghost btn-sm" style={{ width: "100%", marginTop: "0.5rem", color: "var(--text-3)", fontSize: "0.78rem" }} onClick={onLogout}>
            Sign Out
          </button>
        )}
      </div>
    </nav>
  );
}

function RouteView({ route, state, setState, notify }) {
  const props = { state, setState, notify };
  if (!state.token && route !== "bootstrap") return <LoginPage {...props} />;
  switch (route) {
    case "bootstrap":    return <BootstrapPage    {...props} />;
    case "overview":     return <OverviewPage      {...props} />;
    case "logs":         return <LogsPage          {...props} />;
    case "metrics":      return <MetricsPage        {...props} />;
    case "traces":       return <TracesPage         {...props} />;
    case "transactions":  return <TransactionsPage  {...props} />;
    case "service-map":   return <ServiceMapPage     {...props} />;
    case "errors":       return <ErrorsPage         {...props} />;
    case "alerts":       return <AlertsPage         {...props} />;
    case "incidents":    return <IncidentsPage      {...props} />;
    case "platform":     return <PlatformPage       {...props} />;
    case "settings":     return <SettingsPage       {...props} />;
    default:             return <OverviewPage        {...props} />;
  }
}

export default function App() {
  const [state, setState] = useState(loadState);
  const [toasts, dispatch] = useReducer(toastReducer, []);
  const { route, navigate } = useRouter();
  const timers = useRef({});

  useEffect(() => { saveState(state); }, [state]);
  
  // Auto-sync API key if missing but we have a session
  useEffect(() => {
    if (state.token && state.token !== "bootstrap-session" && !state.apiKey) {
      tenantApi.listAPIKeys(state.token).then(keys => {
        const activeIngestKey = keys?.find(k => k.active && (k.scopes.includes("ingest") || k.scopes.includes("*")));
        if (activeIngestKey) {
          // Note: we can't get the raw key back for security, but rotation will fix it.
          // However, for demo/dev, we might need a way to show the key or just prompt to rotate.
          // For now, we just inform the user if we find keys.
        }
      }).catch(() => {});
    }
  }, [state.token, state.apiKey]);

  const notify = useCallback((message, tone = "info") => {
    const id = ++_tid;
    dispatch({ type: "ADD", toast: { id, message, tone } });
    timers.current[id] = setTimeout(() => {
      dispatch({ type: "REMOVE", id });
      delete timers.current[id];
    }, 5000);
  }, []);

  const dismiss = useCallback((id) => {
    clearTimeout(timers.current[id]);
    delete timers.current[id];
    dispatch({ type: "REMOVE", id });
  }, []);

  const handleLogout = () => {
    setState({ token: "", email: "", tenantId: "", apiKey: "" });
    navigate("overview");
    notify("Signed out successfully.", "info");
  };

  const isBare = route === "bootstrap" || !state.token;
  if (isBare) return (
    <>
      <RouteView route={route} state={state} setState={setState} notify={notify} />
      <ToastStack toasts={toasts} dismiss={dismiss} />
    </>
  );

  const CurrentIcon = PAGE_META[route]?.icon || PAGE_META.overview.icon;

  return (
    <div className="app-shell">
      <SideNav route={route} navigate={navigate} state={state} onLogout={handleLogout} />

      <div className="app-main">
        <header className="app-topbar">
          <div className="topbar-title">
            <CurrentIcon />
            {PAGE_META[route]?.label || "PulseLens"}
          </div>

          <div style={{ display: "flex", alignItems: "center", gap: "1rem" }}>
            <div className="form-input" style={{ width: "240px", display: "flex", alignItems: "center", gap: "0.5rem", padding: "0.4rem 0.75rem" }}>
              <span style={{ width: "16px", height: "16px", display: "flex", color: "var(--text-3)" }}><Icons.Search /></span>
              <input type="text" placeholder="Search resources..." style={{ background: "transparent", border: "none", color: "white", outline: "none", width: "100%", fontSize: "0.85rem" }} />
            </div>

            <button className="btn btn-ghost" onClick={() => navigate("bootstrap")} style={{ padding: "0.5rem" }}>
              <Icons.Shield /> Setup
            </button>
          </div>
        </header>

        <main className="app-content">
          <RouteView route={route} state={state} setState={setState} notify={notify} />
        </main>
      </div>

      <ToastStack toasts={toasts} dismiss={dismiss} />
    </div>
  );
}
