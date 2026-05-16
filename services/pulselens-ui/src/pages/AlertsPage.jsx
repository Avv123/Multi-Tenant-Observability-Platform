// AlertsPage — alert rules, channels, and policies with side-sheet drawers
import { useState } from "react";
import { EmptyState, SectionLoader, useAsyncData } from "../lib/hooks";
import { alertingApi } from "../lib/api";
import Drawer from "../components/Drawer";
import Tooltip from "../components/Tooltip";

const CHANNEL_TYPES = [
  { value: "webhook",       icon: "⚡", label: "Webhook",       desc: "POST JSON payload to any HTTP endpoint." },
  { value: "slack_webhook", icon: "💬", label: "Slack",          desc: "Send a formatted message to a Slack channel via Incoming Webhook." },
  { value: "email",         icon: "✉️",  label: "Email",          desc: "Send an email to one or more recipients." },
  { value: "stdout",        icon: "📋", label: "Console (log)", desc: "Print to the alerting-service stdout — useful for testing." },
];

const SIGNAL_TYPES   = ["log", "metric", "trace"];
const SEVERITIES     = ["debug", "info", "warn", "error"];
const AGGREGATIONS   = ["count", "avg", "sum", "min", "max"];
const COMPARATORS    = [">=", ">", "<=", "<", "=="];

function Field({ label, tip, children }) {
  return (
    <div className="form-group" style={{ marginBottom: 0 }}>
      <label className="form-label" style={{ display: "flex", alignItems: "center" }}>
        {label}
        {tip && <Tooltip text={tip} />}
      </label>
      {children}
    </div>
  );
}

export default function AlertsPage({ state, notify }) {
  const token = state.token;

  const { data: rules,    loading: rulesLoading,    refetch: refetchRules    } = useAsyncData(() => alertingApi.listRules(token),               [token], { skip: !token });
  const { data: policies, loading: policiesLoading, refetch: refetchPolicies } = useAsyncData(() => alertingApi.listPolicies(token),             [token], { skip: !token });
  const { data: channels, loading: channelsLoading, refetch: refetchChannels } = useAsyncData(() => alertingApi.listNotificationChannels(token), [token], { skip: !token });

  const firstPolicyId = policies?.[0]?.id || "";

  // ── Rule drawer ──────────────────────────────────────────
  const [ruleDrawer, setRuleDrawer] = useState(false);
  const [creatingRule, setCreatingRule] = useState(false);
  const [ruleForm, setRuleForm] = useState({
    name: "", signal_type: "log", severity: "error",
    aggregation: "count", comparator: ">=", threshold: 1,
    window_minutes: 10, cooldown_minutes: 5, metric_name: "", policy_id: "",
  });
  const rf = (k, v) => setRuleForm(p => ({ ...p, [k]: v }));

  async function handleCreateRule(e) {
    e?.preventDefault();
    if (!ruleForm.name) return notify("Rule name is required", "error");
    setCreatingRule(true);
    try {
      await alertingApi.createRule(token, {
        ...ruleForm,
        threshold: Number(ruleForm.threshold),
        window_minutes: Number(ruleForm.window_minutes),
        cooldown_minutes: Number(ruleForm.cooldown_minutes),
        service_id: state.serviceId,
        policy_id: ruleForm.policy_id || firstPolicyId,
      });
      notify("Alert rule created.", "success");
      setRuleForm({ name: "", signal_type: "log", severity: "error", aggregation: "count", comparator: ">=", threshold: 1, window_minutes: 10, cooldown_minutes: 5, metric_name: "", policy_id: "" });
      setRuleDrawer(false);
      refetchRules();
    } catch (err) { notify(err.message, "error"); }
    finally { setCreatingRule(false); }
  }

  // ── Channel drawer ───────────────────────────────────────
  const [channelDrawer, setChannelDrawer] = useState(false);
  const [creatingChannel, setCreatingChannel] = useState(false);
  const [channelForm, setChannelForm] = useState({ name: "", type: "webhook", webhook_url: "", webhook_method: "POST", email_to: "" });
  const cf = (k, v) => setChannelForm(p => ({ ...p, [k]: v }));

  async function handleCreateChannel(e) {
    e?.preventDefault();
    if (!channelForm.name) return notify("Channel name is required", "error");
    setCreatingChannel(true);
    try {
      const config =
        channelForm.type === "webhook" || channelForm.type === "slack_webhook"
          ? { url: channelForm.webhook_url, method: channelForm.webhook_method, headers: { "X-PulseLens-Source": "ui" }, timeout_seconds: 5 }
          : channelForm.type === "email"
            ? { to: channelForm.email_to.split(",").map(v => v.trim()).filter(Boolean), subject_prefix: "[PulseLens]" }
            : { destination: "local-log" };
      await alertingApi.createNotificationChannel(token, { name: channelForm.name, type: channelForm.type, config });
      notify("Notification channel created.", "success");
      setChannelForm({ name: "", type: "webhook", webhook_url: "", webhook_method: "POST", email_to: "" });
      setChannelDrawer(false);
      refetchChannels();
    } catch (err) { notify(err.message, "error"); }
    finally { setCreatingChannel(false); }
  }

  return (
    <div style={{ display: "flex", flexDirection: "column", gap: "1.25rem" }}>
      {/* Page header */}
      <div>
        <h1 style={{ fontSize: "1.4rem", fontWeight: 700, marginBottom: "0.2rem" }}>Alerts</h1>
        <p style={{ color: "var(--text-2)", fontSize: "0.875rem" }}>Configure detection rules and notification channels. Rules are evaluated every 30s.</p>
      </div>

      {/* ── Alert Rules ─────────────────────────────── */}
      <div className="panel">
        <div className="panel-header">
          <div>
            <div className="panel-title">
              Alert Rules
              <Tooltip text="Rules continuously evaluate your signal data (logs, metrics, traces). When the threshold is crossed, an incident is created and notifications are dispatched." />
            </div>
            <div className="panel-desc">Evaluated every 30s by a concurrent goroutine pool</div>
          </div>
          <div style={{ display: "flex", gap: "0.5rem" }}>
            <button className="btn btn-secondary btn-sm" onClick={refetchRules}>↺ Refresh</button>
            <button className="btn btn-primary btn-sm" id="btn-new-rule" onClick={() => setRuleDrawer(true)}>+ New Rule</button>
          </div>
        </div>

        {rulesLoading ? <SectionLoader /> : (
          rules?.length ? (
            <div className="table-wrap">
              <table className="data-table" id="table-rules">
                <thead><tr><th>Name</th><th>Signal</th><th>Condition</th><th>Window</th><th>Cooldown</th><th>Status</th></tr></thead>
                <tbody>
                  {rules.map(r => (
                    <tr key={r.id}>
                      <td style={{ fontWeight: 500 }}>{r.name}</td>
                      <td><span className="badge badge-info">{r.signal_type}</span></td>
                      <td style={{ fontFamily: "var(--font-mono)", fontSize: "0.78rem" }}>{r.aggregation} {r.comparator} {r.threshold}</td>
                      <td>{r.window_minutes}m</td>
                      <td>{r.cooldown_minutes}m</td>
                      <td><span className={`badge ${r.active ? "badge-success" : "badge-neutral"}`}>{r.active ? "active" : "inactive"}</span></td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          ) : (
            <EmptyState icon="◆" title="No alert rules"
              body="Create your first rule to start monitoring signal thresholds."
              action={<button className="btn btn-primary btn-sm" onClick={() => setRuleDrawer(true)}>+ New Rule</button>}
            />
          )
        )}
      </div>

      {/* ── Notification Channels ───────────────────── */}
      <div className="panel">
        <div className="panel-header">
          <div>
            <div className="panel-title">
              Notification Channels
              <Tooltip text="Channels define where alerts are delivered — webhooks (Slack, PagerDuty, custom), email, or local logging. A channel must be attached to a Policy to receive incidents." />
            </div>
            <div className="panel-desc">Webhook, Slack, email, or console destinations</div>
          </div>
          <div style={{ display: "flex", gap: "0.5rem" }}>
            <button className="btn btn-secondary btn-sm" onClick={refetchChannels}>↺ Refresh</button>
            <button className="btn btn-primary btn-sm" id="btn-add-channel" onClick={() => setChannelDrawer(true)}>+ Add Channel</button>
          </div>
        </div>

        {channelsLoading ? <SectionLoader /> : (
          channels?.length ? (
            <div className="table-wrap">
              <table className="data-table" id="table-channels">
                <thead><tr><th>Name</th><th>Type</th><th>Status</th></tr></thead>
                <tbody>
                  {channels.map(c => (
                    <tr key={c.id}>
                      <td style={{ fontWeight: 500 }}>{c.name}</td>
                      <td>
                        <span className="badge badge-info">
                          {CHANNEL_TYPES.find(t => t.value === c.type)?.icon} {c.type}
                        </span>
                      </td>
                      <td><span className={`badge ${c.active ? "badge-success" : "badge-neutral"}`}>{c.active ? "active" : "inactive"}</span></td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          ) : (
            <EmptyState icon="◈" title="No channels"
              body="Add a webhook, Slack, or email channel to receive incident notifications."
              action={<button className="btn btn-primary btn-sm" onClick={() => setChannelDrawer(true)}>+ Add Channel</button>}
            />
          )
        )}
      </div>

      {/* ── Policies ────────────────────────────────── */}
      <div className="panel">
        <div className="panel-header">
          <div>
            <div className="panel-title">
              Alert Policies
              <Tooltip text="Policies govern how incidents are escalated and re-notified. A policy links an alert rule to one or more channels and defines retry intervals and maximum escalation steps." />
            </div>
            <div className="panel-desc">Control escalation, retry, and delivery behaviour</div>
          </div>
          <button className="btn btn-secondary btn-sm" onClick={refetchPolicies}>↺ Refresh</button>
        </div>

        {policiesLoading ? <SectionLoader /> : (
          policies?.length ? (
            <div className="table-wrap">
              <table className="data-table" id="table-policies">
                <thead><tr><th>Name</th><th>Max Attempts</th><th>Escalations</th><th>Backoff</th></tr></thead>
                <tbody>
                  {policies.map(p => (
                    <tr key={p.id}>
                      <td style={{ fontWeight: 500 }}>{p.name}</td>
                      <td>{p.max_delivery_attempts}</td>
                      <td>{p.max_escalations}</td>
                      <td>{p.delivery_backoff_millis}ms</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          ) : (
            <EmptyState icon="◧" title="No policies" body="Create a policy to control escalation and retry behaviour." />
          )
        )}
      </div>

      {/* ═══ Create Rule Drawer ═══════════════════════════════ */}
      <Drawer
        open={ruleDrawer}
        onClose={() => setRuleDrawer(false)}
        title="New Alert Rule"
        description="Define when PulseLens should create an incident based on your telemetry signals."
        width="480px"
      >
        <form onSubmit={handleCreateRule} style={{ display: "flex", flexDirection: "column", gap: "1.25rem" }}>
          {/* What to watch */}
          <div>
            <div style={{ fontSize: "0.7rem", fontWeight: 700, color: "var(--text-3)", textTransform: "uppercase", letterSpacing: "0.07em", marginBottom: "0.75rem" }}>What to watch</div>
            <div style={{ display: "flex", flexDirection: "column", gap: "0.85rem" }}>
              <Field label="Rule Name">
                <input className="form-input" placeholder="e.g. Checkout Error Burst" value={ruleForm.name} onChange={e => rf("name", e.target.value)} required />
              </Field>
              <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: "0.75rem" }}>
                <Field label="Signal Type" tip="Logs = log events, Metric = numeric gauge/counter, Trace = distributed spans.">
                  <select className="form-input" value={ruleForm.signal_type} onChange={e => rf("signal_type", e.target.value)}>
                    {SIGNAL_TYPES.map(o => <option key={o} value={o}>{o}</option>)}
                  </select>
                </Field>
                <Field label="Severity" tip="Only log/trace events matching this severity level count towards the threshold.">
                  <select className="form-input" value={ruleForm.severity} onChange={e => rf("severity", e.target.value)}>
                    {SEVERITIES.map(o => <option key={o} value={o}>{o}</option>)}
                  </select>
                </Field>
              </div>
              {ruleForm.signal_type === "metric" && (
                <Field label="Metric Name" tip="The exact metric_name field to watch (e.g. http_request_duration_ms).">
                  <input className="form-input" placeholder="e.g. http_request_duration_ms" value={ruleForm.metric_name} onChange={e => rf("metric_name", e.target.value)} />
                </Field>
              )}
            </div>
          </div>

          {/* When to fire */}
          <div>
            <div style={{ fontSize: "0.7rem", fontWeight: 700, color: "var(--text-3)", textTransform: "uppercase", letterSpacing: "0.07em", marginBottom: "0.75rem" }}>When to fire</div>
            <div style={{ display: "flex", flexDirection: "column", gap: "0.85rem" }}>
              <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr 1fr", gap: "0.75rem" }}>
                <Field label="Aggregation" tip="How events are counted in the evaluation window. 'count' = number of matching events; 'avg' = mean value of a metric field.">
                  <select className="form-input" value={ruleForm.aggregation} onChange={e => rf("aggregation", e.target.value)}>
                    {AGGREGATIONS.map(o => <option key={o} value={o}>{o}</option>)}
                  </select>
                </Field>
                <Field label="Comparator" tip="The operator that compares the aggregated value against the threshold (e.g. '>=' fires when count ≥ threshold).">
                  <select className="form-input" value={ruleForm.comparator} onChange={e => rf("comparator", e.target.value)}>
                    {COMPARATORS.map(o => <option key={o} value={o}>{o}</option>)}
                  </select>
                </Field>
                <Field label="Threshold" tip="The numeric value that triggers the rule when crossed by the aggregated result.">
                  <input className="form-input" type="number" value={ruleForm.threshold} onChange={e => rf("threshold", e.target.value)} />
                </Field>
              </div>
              <Field label="Evaluation Window (min)" tip="The rolling time window evaluated each cycle. A 10-minute window with count >= 5 means '5 events in the last 10 minutes'.">
                <input className="form-input" type="number" min="1" value={ruleForm.window_minutes} onChange={e => rf("window_minutes", e.target.value)} />
              </Field>
            </div>
          </div>

          {/* Recovery */}
          <div>
            <div style={{ fontSize: "0.7rem", fontWeight: 700, color: "var(--text-3)", textTransform: "uppercase", letterSpacing: "0.07em", marginBottom: "0.75rem" }}>Recovery</div>
            <Field label="Cooldown (min)" tip="Minimum time between repeated firings of the same rule. Prevents alert storms for sustained issues — e.g. a 5-minute cooldown means you get at most one incident per 5 minutes.">
              <input className="form-input" type="number" min="1" value={ruleForm.cooldown_minutes} onChange={e => rf("cooldown_minutes", e.target.value)} />
            </Field>
          </div>

          <div style={{ display: "flex", gap: "0.65rem", paddingTop: "0.5rem", borderTop: "1px solid var(--border)" }}>
            <button type="button" className="btn btn-ghost" style={{ flex: 1 }} onClick={() => setRuleDrawer(false)}>Cancel</button>
            <button type="submit" className="btn btn-primary" id="btn-create-rule" style={{ flex: 2, justifyContent: "center" }} disabled={creatingRule}>
              {creatingRule ? "Creating…" : "Create Rule"}
            </button>
          </div>
        </form>
      </Drawer>

      {/* ═══ Add Channel Drawer ══════════════════════════════ */}
      <Drawer
        open={channelDrawer}
        onClose={() => setChannelDrawer(false)}
        title="Add Notification Channel"
        description="Choose how PulseLens delivers incident notifications."
        width="460px"
      >
        <form onSubmit={handleCreateChannel} style={{ display: "flex", flexDirection: "column", gap: "1.25rem" }}>
          <Field label="Channel Name">
            <input className="form-input" placeholder="e.g. Ops Webhook" value={channelForm.name} onChange={e => cf("name", e.target.value)} required />
          </Field>

          {/* Type selector */}
          <div className="form-group" style={{ marginBottom: 0 }}>
            <label className="form-label">Channel Type</label>
            <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: "0.5rem", marginTop: "0.25rem" }}>
              {CHANNEL_TYPES.map(t => (
                <label key={t.value} style={{
                  display: "flex", alignItems: "flex-start", gap: "0.6rem",
                  padding: "0.75rem", borderRadius: "var(--r-sm)", cursor: "pointer",
                  border: `1.5px solid ${channelForm.type === t.value ? "var(--primary)" : "var(--border)"}`,
                  background: channelForm.type === t.value ? "var(--primary-soft)" : "var(--surface-active)",
                  transition: "all 0.15s",
                }}>
                  <input type="radio" name="channel_type" value={t.value} checked={channelForm.type === t.value} onChange={() => cf("type", t.value)} style={{ marginTop: "2px", accentColor: "var(--primary)" }} />
                  <div>
                    <div style={{ fontSize: "0.85rem", fontWeight: 600 }}>{t.icon} {t.label}</div>
                    <div style={{ fontSize: "0.72rem", color: "var(--text-2)", marginTop: "0.15rem", lineHeight: 1.45 }}>{t.desc}</div>
                  </div>
                </label>
              ))}
            </div>
          </div>

          {/* Dynamic config fields */}
          {(channelForm.type === "webhook" || channelForm.type === "slack_webhook") && (
            <div style={{ display: "flex", flexDirection: "column", gap: "0.75rem" }}>
              <Field label="Webhook URL" tip={channelForm.type === "slack_webhook" ? "Use the Incoming Webhook URL from your Slack App configuration." : "The full URL including protocol (https://...)."}>
                <input className="form-input" type="url" placeholder="https://hooks.example.com/..." value={channelForm.webhook_url} onChange={e => cf("webhook_url", e.target.value)} />
              </Field>
              {channelForm.type === "webhook" && (
                <Field label="HTTP Method">
                  <select className="form-input" value={channelForm.webhook_method} onChange={e => cf("webhook_method", e.target.value)}>
                    {["POST", "PUT"].map(m => <option key={m} value={m}>{m}</option>)}
                  </select>
                </Field>
              )}
            </div>
          )}

          {channelForm.type === "email" && (
            <Field label="Recipients" tip="Comma-separated list of email addresses (e.g. ops@company.com, alerts@company.com).">
              <input className="form-input" type="text" placeholder="ops@company.com, alerts@company.com" value={channelForm.email_to} onChange={e => cf("email_to", e.target.value)} />
            </Field>
          )}

          {channelForm.type === "stdout" && (
            <div style={{ padding: "0.75rem", background: "var(--surface-active)", borderRadius: "var(--r-sm)", border: "1px solid var(--border)", fontSize: "0.8rem", color: "var(--text-2)" }}>
              ℹ️ Console channels write incident payloads to the <code style={{ background: "rgba(255,255,255,0.06)", padding: "1px 4px", borderRadius: "3px" }}>pulselens-alerting-service</code> stdout. No configuration needed.
            </div>
          )}

          <div style={{ display: "flex", gap: "0.65rem", paddingTop: "0.5rem", borderTop: "1px solid var(--border)" }}>
            <button type="button" className="btn btn-ghost" style={{ flex: 1 }} onClick={() => setChannelDrawer(false)}>Cancel</button>
            <button type="submit" className="btn btn-primary" id="btn-create-channel" style={{ flex: 2, justifyContent: "center" }} disabled={creatingChannel}>
              {creatingChannel ? "Creating…" : "Add Channel"}
            </button>
          </div>
        </form>
      </Drawer>
    </div>
  );
}
