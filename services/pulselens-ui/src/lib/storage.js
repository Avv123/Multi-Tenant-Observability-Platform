// F3: JWT token is stored in sessionStorage (clears on tab close) to reduce the
// XSS attack surface. Non-sensitive bootstrap config (tenantId, serviceId, apiKey)
// stays in localStorage since losing it on tab close would break the UX flow.

const SESSION_KEY = "pulselens-session";
const CONFIG_KEY = "pulselens-config";

const DEFAULT_CONFIG = {
  tenantId: "",
  serviceId: "",
  apiKey: "",
  email: "",
  password: "password123",
};

/** Load the non-sensitive bootstrap config from localStorage. */
export function loadConfig() {
  try {
    const raw = localStorage.getItem(CONFIG_KEY);
    return raw ? { ...DEFAULT_CONFIG, ...JSON.parse(raw) } : { ...DEFAULT_CONFIG };
  } catch {
    return { ...DEFAULT_CONFIG };
  }
}

/** Persist non-sensitive config to localStorage. */
export function saveConfig(config) {
  try {
    const { token, ...safe } = config; // never persist token here
    localStorage.setItem(CONFIG_KEY, JSON.stringify(safe));
  } catch (err) {
    console.warn("Storage access restricted, safe config not saved:", err);
  }
}

/** Read the JWT token from sessionStorage. */
export function getToken() {
  try {
    return sessionStorage.getItem(SESSION_KEY) || "";
  } catch {
    return "";
  }
}

/** Write the JWT token to sessionStorage. */
export function setToken(token) {
  try {
    if (token) {
      sessionStorage.setItem(SESSION_KEY, token);
    } else {
      sessionStorage.removeItem(SESSION_KEY);
    }
  } catch (err) {
    console.warn("Storage access restricted, session token not saved:", err);
  }
}

/** Backward-compatible loadState — combines config + live session token. */
export function loadState() {
  return { ...loadConfig(), token: getToken() };
}

/** Backward-compatible saveState — routes token to sessionStorage. */
export function saveState(state) {
  const { token, ...config } = state;
  saveConfig(config);
  setToken(token);
}


