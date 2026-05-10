export function loadState() {
  const raw = localStorage.getItem("pulselens-state");
  if (!raw) {
    return {
      tenantId: "",
      serviceId: "",
      apiKey: "",
      token: "",
      email: "",
      password: "password123",
    };
  }
  try {
    return JSON.parse(raw);
  } catch {
    return {
      tenantId: "",
      serviceId: "",
      apiKey: "",
      token: "",
      email: "",
      password: "password123",
    };
  }
}

export function saveState(nextState) {
  localStorage.setItem("pulselens-state", JSON.stringify(nextState));
}
