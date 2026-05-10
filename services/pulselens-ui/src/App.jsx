import { useEffect, useState } from "react";
import BootstrapPage from "./pages/BootstrapPage";
import DashboardPage from "./pages/DashboardPage";
import LoginPage from "./pages/LoginPage";
import { loadState, saveState } from "./lib/storage";

export default function App() {
  const [state, setState] = useState(loadState);
  const [notification, setNotification] = useState({ message: "", tone: "info" });

  useEffect(() => {
    saveState(state);
  }, [state]);

  function onNotification(message, tone = "info") {
    setNotification({ message, tone });
  }

  return (
    <div className="app-shell">
      <header className="app-header">
        <div>
          <p className="eyebrow">Multi-Tenant Observability Platform</p>
          <h1>PulseLens Local Console</h1>
        </div>
        <button onClick={() => setState(loadState())}>Reload Local State</button>
      </header>

      {notification.message ? (
        <div className={`banner banner--${notification.tone}`}>{notification.message}</div>
      ) : null}

      <main className="layout">
        <BootstrapPage state={state} setState={setState} onNotification={onNotification} />
        <LoginPage state={state} setState={setState} onNotification={onNotification} />
        <DashboardPage state={state} onNotification={onNotification} />
      </main>
    </div>
  );
}
