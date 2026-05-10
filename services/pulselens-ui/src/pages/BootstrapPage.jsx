import { useState } from "react";
import Section from "../components/Section";
import { tenantApi } from "../lib/api";

function initialForm() {
  const suffix = `${Date.now()}`;
  return {
    name: "Acme Logistics",
    slug: `acme-${suffix}`,
    plan: "starter",
    ingest_quota: 1000,
    retention_days: 7,
    admin_name: "Platform Admin",
    admin_email: `admin+${suffix}@pulselens.local`,
    admin_password: "password123",
    service_name: "checkout-api",
    service_environment: "local",
  };
}

export default function BootstrapPage({ state, setState, onNotification }) {
  const [form, setForm] = useState(initialForm);
  const [loading, setLoading] = useState(false);

  async function handleBootstrap(event) {
    event.preventDefault();
    setLoading(true);
    try {
      const tenantResponse = await tenantApi.createTenant({
        name: form.name,
        slug: form.slug,
        plan: form.plan,
        ingest_quota: Number(form.ingest_quota),
        retention_days: Number(form.retention_days),
        admin_name: form.admin_name,
        admin_email: form.admin_email,
        admin_password: form.admin_password,
      });

      const loginResponse = await tenantApi.login({
        email: form.admin_email,
        password: form.admin_password,
      });

      const serviceResponse = await tenantApi.createService(
        tenantResponse.tenant.id,
        {
          name: form.service_name,
          environment: form.service_environment,
          tags: { team: "platform", source: "ui-bootstrap" },
        },
        loginResponse.token,
      );

      const keyResponse = await tenantApi.createAPIKey(
        {
          tenant_id: tenantResponse.tenant.id,
          service_id: serviceResponse.id,
          name: "frontend-bootstrap-key",
          scopes: ["ingest", "query", "admin"],
        },
        loginResponse.token,
      );

      setState({
        ...state,
        tenantId: tenantResponse.tenant.id,
        serviceId: serviceResponse.id,
        apiKey: keyResponse.key,
        token: loginResponse.token,
        email: form.admin_email,
        password: form.admin_password,
      });
      onNotification("Tenant, service, API key, and session created.");
    } catch (error) {
      onNotification(error.message, "error");
    } finally {
      setLoading(false);
    }
  }

  function updateField(key, value) {
    setForm((current) => ({ ...current, [key]: value }));
  }

  return (
    <Section title="Bootstrap Workspace">
      <form className="form-grid" onSubmit={handleBootstrap}>
        <label>
          Tenant Name
          <input data-testid="bootstrap-tenant-name" value={form.name} onChange={(event) => updateField("name", event.target.value)} />
        </label>
        <label>
          Slug
          <input data-testid="bootstrap-slug" value={form.slug} onChange={(event) => updateField("slug", event.target.value)} />
        </label>
        <label>
          Admin Name
          <input value={form.admin_name} onChange={(event) => updateField("admin_name", event.target.value)} />
        </label>
        <label>
          Admin Email
          <input data-testid="bootstrap-admin-email" value={form.admin_email} onChange={(event) => updateField("admin_email", event.target.value)} />
        </label>
        <label>
          Admin Password
          <input data-testid="bootstrap-admin-password" type="password" value={form.admin_password} onChange={(event) => updateField("admin_password", event.target.value)} />
        </label>
        <label>
          Service Name
          <input value={form.service_name} onChange={(event) => updateField("service_name", event.target.value)} />
        </label>
        <label>
          Service Environment
          <input value={form.service_environment} onChange={(event) => updateField("service_environment", event.target.value)} />
        </label>
        <label>
          Ingest Quota
          <input value={form.ingest_quota} onChange={(event) => updateField("ingest_quota", event.target.value)} />
        </label>
        <label>
          Retention Days
          <input value={form.retention_days} onChange={(event) => updateField("retention_days", event.target.value)} />
        </label>
        <div className="form-actions">
          <button data-testid="bootstrap-submit" type="submit" disabled={loading}>
            {loading ? "Bootstrapping..." : "Create Tenant"}
          </button>
        </div>
      </form>
    </Section>
  );
}
