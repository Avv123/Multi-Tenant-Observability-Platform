import { useState } from "react";
import Section from "../components/Section";
import { tenantApi } from "../lib/api";

export default function LoginPage({ state, setState, onNotification }) {
  const [form, setForm] = useState({
    email: state.email,
    password: state.password || "password123",
  });

  async function handleLogin(event) {
    event.preventDefault();
    try {
      const response = await tenantApi.login(form);
      setState({
        ...state,
        token: response.token,
        email: form.email,
        password: form.password,
      });
      onNotification("Logged in successfully.");
    } catch (error) {
      onNotification(error.message, "error");
    }
  }

  return (
    <Section title="Login">
      <form className="form-grid" onSubmit={handleLogin}>
        <label>
          Email
          <input data-testid="login-email" value={form.email} onChange={(event) => setForm({ ...form, email: event.target.value })} />
        </label>
        <label>
          Password
          <input data-testid="login-password" type="password" value={form.password} onChange={(event) => setForm({ ...form, password: event.target.value })} />
        </label>
        <div className="form-actions">
          <button data-testid="login-submit" type="submit">Login</button>
        </div>
      </form>
    </Section>
  );
}
