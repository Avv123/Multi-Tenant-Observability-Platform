import { useEffect, useMemo, useState } from "react";
import BarChart from "../components/BarChart";
import DataTable from "../components/DataTable";
import LineChart from "../components/LineChart";
import Section from "../components/Section";
import StatCard from "../components/StatCard";
import { alertingApi, archiveApi, ingestApi, queryApi, tenantApi } from "../lib/api";

function formatDate(value) {
  if (!value) {
    return "-";
  }
  return new Date(value).toLocaleString();
}

function compactTimestamp(value) {
  if (!value) {
    return "-";
  }
  const date = new Date(value);
  return `${date.getHours().toString().padStart(2, "0")}:${date.getMinutes().toString().padStart(2, "0")}`;
}

function parseJSONSafe(value, fallback) {
  if (!value) {
    return fallback;
  }
  if (typeof value !== "string") {
    return value;
  }
  try {
    return JSON.parse(value);
  } catch {
    return fallback;
  }
}

function filterPayload(filterForm) {
  const payload = { limit: 20 };
  if (filterForm.service_id) payload.service_id = filterForm.service_id;
  if (filterForm.environment) payload.environment = filterForm.environment;
  if (filterForm.severity) payload.severity = filterForm.severity;
  if (filterForm.metric_name) payload.metric_name = filterForm.metric_name;
  if (filterForm.search) payload.search = filterForm.search;
  if (filterForm.trace_id) payload.trace_id = filterForm.trace_id;
  if (filterForm.lookback_minutes) {
    payload.start_time = new Date(Date.now() - Number(filterForm.lookback_minutes) * 60 * 1000).toISOString();
  }
  return payload;
}

function widgetDataset(widget, data) {
  switch (widget.dataset) {
    case "logs":
      return data.logs;
    case "metrics":
      return data.metrics;
    case "traces":
      return data.traces;
    case "service_health":
      return data.health;
    case "log_severity":
      return data.logSeverity;
    case "metric_series":
      return data.metricSeries;
    case "trace_latency":
      return data.traceLatency;
    default:
      return [];
  }
}

function sampleEvents() {
  return {
    log: [
      {
        event_type: "log",
        severity: "error",
        payload: {
          message: `checkout failure at ${new Date().toISOString()}`,
          logger: "frontend-playground",
        },
      },
    ],
    metric: [
      {
        event_type: "metric",
        payload: {
          metric_name: "checkout_latency_ms",
          value: 180,
          unit: "ms",
        },
      },
    ],
    trace: [
      {
        event_type: "trace",
        trace_id: `trace-${Date.now()}`,
        payload: {
          span_id: "root-span",
          parent_span_id: "",
          operation: "checkout",
          status: "error",
          start_time: new Date(Date.now() - 120).toISOString(),
          end_time: new Date().toISOString(),
        },
      },
    ],
  };
}

export default function DashboardPage({ state, onNotification }) {
  const [loading, setLoading] = useState(false);
  const [data, setData] = useState({
    me: null,
    overview: null,
    platformOverview: null,
    health: [],
    logs: [],
    metrics: [],
    traces: [],
    rules: [],
    incidents: [],
    auditLogs: [],
    savedQueries: [],
    dashboards: [],
    channels: [],
    deliveries: [],
    replayJobs: [],
    replayStats: null,
    incidentComments: [],
    incidentTimeline: [],
    incidentDeliveries: [],
    selectedIncident: null,
    users: [],
    services: [],
    policies: [],
    logSeverity: [],
    metricSeries: [],
    traceLatency: [],
  });
  const [filterForm, setFilterForm] = useState({
    service_id: "",
    environment: "",
    severity: "",
    metric_name: "checkout_latency_ms",
    search: "",
    trace_id: "",
    lookback_minutes: "120",
  });
  const [ruleForm, setRuleForm] = useState({
    name: "Checkout Error Burst",
    signal_type: "log",
    severity: "error",
    aggregation: "count",
    comparator: ">=",
    threshold: 1,
    window_minutes: 10,
    cooldown_minutes: 5,
    metric_name: "",
    policy_id: "",
  });
  const [savedQueryForm, setSavedQueryForm] = useState({
    name: "Errors in checkout-api",
    query_type: "logs",
  });
  const [dashboardForm, setDashboardForm] = useState({
    name: "Operations Overview",
    description: "Starter dashboard for PulseLens",
    default_time_range: "120m",
  });
  const [dashboardBuilder, setDashboardBuilder] = useState({
    dashboard_id: "",
    widget_id: "",
    title: "Error Trend",
    type: "chart",
    dataset: "log_severity",
    chart_type: "bar",
    metric: "",
    layout_w: "1",
    layout_h: "1",
  });
  const [channelForm, setChannelForm] = useState({
    name: "Ops Webhook Channel",
    type: "webhook",
    webhook_url: "http://127.0.0.1:9099/webhooks/incidents",
    webhook_method: "POST",
    email_to: "ops@pulselens.local",
  });
  const [replayForm, setReplayForm] = useState({
    event_type: "log",
    window_minutes: 30,
  });
  const [commentForm, setCommentForm] = useState({
    incidentId: "",
    body: "Investigating now.",
  });
  const [assignForm, setAssignForm] = useState({
    incidentId: "",
    assignedTo: "",
  });
  const [incidentFilters, setIncidentFilters] = useState({
    status: "",
    assigned_to: "",
    service_id: "",
    severity: "",
  });
  const [userForm, setUserForm] = useState({
    name: "Viewer User",
    email: `viewer-${Date.now()}@pulselens.local`,
    password: "viewer-pass",
    role: "viewer",
  });
  const [policyForm, setPolicyForm] = useState({
    name: "Default Policy",
    description: "Escalate after repeated failures",
    max_delivery_attempts: 3,
    delivery_backoff_millis: 200,
    escalation_interval_minutes: 5,
    max_escalations: 3,
    repeat_notification_minutes: 5,
    open_channel_types: "webhook,slack_webhook,email",
    ack_channel_types: "webhook",
    resolve_channel_types: "webhook",
    escalation_channel_types: "slack_webhook,webhook",
  });

  const token = state.token;

  async function loadDashboard() {
    if (!token || !state.tenantId) {
      return;
    }
    setLoading(true);
    try {
      const filters = filterPayload(filterForm);
      const [me, overview, platformOverview, health, logs, metrics, traces, rules, policies, incidents, auditLogs, savedQueries, dashboards, channels, deliveries, replayJobs, replayStats, users, services, logSeverity, metricSeries, traceLatency] = await Promise.all([
        tenantApi.me(token),
        queryApi.overview(token),
        queryApi.platformOverview(token),
        queryApi.serviceHealth(token),
        queryApi.logsWithFilters(token, filters),
        queryApi.metricsWithFilters(token, filters),
        queryApi.tracesWithFilters(token, filters),
        alertingApi.listRules(token),
        alertingApi.listPolicies(token),
        alertingApi.listIncidents(token, incidentFilters),
        tenantApi.listAuditLogs(state.tenantId, token),
        queryApi.listSavedQueries(token),
        queryApi.listDashboards(token),
        alertingApi.listNotificationChannels(token),
        alertingApi.listNotificationDeliveries(token),
        archiveApi.listReplayJobs(token),
        archiveApi.stats(token),
        tenantApi.listUsers(state.tenantId, token),
        tenantApi.listServices(state.tenantId, token),
        queryApi.logSeverityWithFilters(token, filters),
        queryApi.metricSeriesWithFilters(token, filters),
        queryApi.traceLatencyWithFilters(token, filters),
      ]);
      setData({
        me,
        overview,
        platformOverview,
        health,
        logs,
        metrics,
        traces,
        rules,
        policies,
        incidents,
        auditLogs,
        savedQueries,
        dashboards,
        channels,
        deliveries,
        replayJobs,
        replayStats,
        incidentComments: [],
        incidentTimeline: [],
        incidentDeliveries: [],
        selectedIncident: null,
        users,
        services,
        logSeverity,
        metricSeries,
        traceLatency,
      });
      setRuleForm((current) => ({
        ...current,
        policy_id: current.policy_id || policies[0]?.id || "",
      }));
      setDashboardBuilder((current) => ({
        ...current,
        dashboard_id: current.dashboard_id || dashboards[0]?.id || "",
        widget_id: "",
      }));
    } catch (error) {
      onNotification(error.message, "error");
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    loadDashboard();
  }, [token, state.tenantId]);

  useEffect(() => {
    if (token && state.tenantId) {
      loadDashboard();
    }
  }, [filterForm.lookback_minutes, filterForm.service_id, filterForm.environment, filterForm.severity, filterForm.metric_name, filterForm.search, filterForm.trace_id, incidentFilters.status, incidentFilters.assigned_to, incidentFilters.service_id, incidentFilters.severity]);

  async function handleIngest(kind) {
    try {
      await ingestApi.ingest(state.apiKey, sampleEvents()[kind]);
      onNotification(`${kind} event submitted.`);
      await loadDashboard();
    } catch (error) {
      onNotification(error.message, "error");
    }
  }

  async function handleCreateRule(event) {
    event.preventDefault();
    try {
      await alertingApi.createRule(token, {
        ...ruleForm,
        threshold: Number(ruleForm.threshold),
        window_minutes: Number(ruleForm.window_minutes),
        cooldown_minutes: Number(ruleForm.cooldown_minutes),
        service_id: state.serviceId,
      });
      onNotification("Alert rule created.");
      await loadDashboard();
    } catch (error) {
      onNotification(error.message, "error");
    }
  }

  async function handleCreatePolicy(event) {
    event.preventDefault();
    try {
      await alertingApi.createPolicy(token, {
        ...policyForm,
        max_delivery_attempts: Number(policyForm.max_delivery_attempts),
        delivery_backoff_millis: Number(policyForm.delivery_backoff_millis),
        escalation_interval_minutes: Number(policyForm.escalation_interval_minutes),
        max_escalations: Number(policyForm.max_escalations),
        repeat_notification_minutes: Number(policyForm.repeat_notification_minutes),
        open_channel_types: policyForm.open_channel_types.split(",").map((value) => value.trim()).filter(Boolean),
        ack_channel_types: policyForm.ack_channel_types.split(",").map((value) => value.trim()).filter(Boolean),
        resolve_channel_types: policyForm.resolve_channel_types.split(",").map((value) => value.trim()).filter(Boolean),
        escalation_channel_types: policyForm.escalation_channel_types.split(",").map((value) => value.trim()).filter(Boolean),
      });
      onNotification("Alert policy created.");
      await loadDashboard();
    } catch (error) {
      onNotification(error.message, "error");
    }
  }

  async function handleCreateSavedQuery(event) {
    event.preventDefault();
    try {
      await queryApi.createSavedQuery(token, {
        name: savedQueryForm.name,
        query_type: savedQueryForm.query_type,
        definition: {
          service_id: state.serviceId,
          severity: "error",
          limit: 20,
        },
      });
      onNotification("Saved query created.");
      await loadDashboard();
    } catch (error) {
      onNotification(error.message, "error");
    }
  }

  async function handleCreateDashboard(event) {
    event.preventDefault();
    try {
      await queryApi.createDashboard(token, {
        name: dashboardForm.name,
        description: dashboardForm.description,
        default_time_range: dashboardForm.default_time_range,
        layout: { columns: 2 },
        widgets: [
          { id: `widget-${Date.now()}-1`, type: "stat", metric: "log_count", title: "Log Count", layout: { w: 1, h: 1 }, filters: {} },
          { id: `widget-${Date.now()}-2`, type: "stat", metric: "trace_count", title: "Trace Count", layout: { w: 1, h: 1 }, filters: {} },
          { id: `widget-${Date.now()}-3`, type: "table", dataset: "service_health", title: "Service Health", layout: { w: 2, h: 1 }, filters: {} },
        ],
      });
      onNotification("Dashboard created.");
      await loadDashboard();
    } catch (error) {
      onNotification(error.message, "error");
    }
  }

  async function handleSaveDashboardWidget(event) {
    event.preventDefault();
    try {
      const selected = data.dashboards.find((row) => row.id === dashboardBuilder.dashboard_id);
      if (!selected) {
        throw new Error("select a dashboard first");
      }
      const widgetPayload = {
        title: dashboardBuilder.title,
        type: dashboardBuilder.type,
        dataset: dashboardBuilder.dataset,
        chart_type: dashboardBuilder.chart_type,
        metric: dashboardBuilder.metric,
        filters: filterPayload(filterForm),
        layout: {
          w: Number(dashboardBuilder.layout_w || 1),
          h: Number(dashboardBuilder.layout_h || 1),
        },
      };
      if (dashboardBuilder.widget_id) {
        await queryApi.updateDashboardWidget(token, selected.id, dashboardBuilder.widget_id, widgetPayload);
        onNotification("Dashboard widget updated.");
      } else {
        const widgets = parseJSONSafe(selected.widgets, []);
        widgets.push({
          id: `widget-${Date.now()}`,
          ...widgetPayload,
        });
        await queryApi.updateDashboard(token, selected.id, {
          name: selected.name,
          description: selected.description,
          default_time_range: selected.default_time_range || "120m",
          layout: parseJSONSafe(selected.layout, { columns: 2 }),
          widgets,
        });
        onNotification("Dashboard widget saved.");
      }
      await loadDashboard();
    } catch (error) {
      onNotification(error.message, "error");
    }
  }

  async function handleUpdateDashboardDetails(event) {
    event.preventDefault();
    try {
      const selected = data.dashboards.find((row) => row.id === dashboardBuilder.dashboard_id);
      if (!selected) {
        throw new Error("select a dashboard first");
      }
      await queryApi.updateDashboard(token, selected.id, {
        name: dashboardForm.name,
        description: dashboardForm.description,
        default_time_range: dashboardForm.default_time_range,
        layout: parseJSONSafe(selected.layout, { columns: 2 }),
        widgets: parseJSONSafe(selected.widgets, []),
      });
      onNotification("Dashboard updated.");
      await loadDashboard();
    } catch (error) {
      onNotification(error.message, "error");
    }
  }

  async function handleSaveCurrentFiltersAsQuery(event) {
    event.preventDefault();
    try {
      if (!data.savedQueries.length) {
        throw new Error("create a saved query first");
      }
      const selected = data.savedQueries[0];
      await queryApi.updateSavedQuery(token, selected.id, {
        name: selected.name,
        query_type: selected.query_type,
        definition: filterPayload(filterForm),
      });
      onNotification("Saved query updated from active filters.");
      await loadDashboard();
    } catch (error) {
      onNotification(error.message, "error");
    }
  }

  function handleApplySavedQuery(savedQuery) {
    const definition = parseJSONSafe(savedQuery.definition, {});
    setFilterForm((current) => ({
      ...current,
      service_id: definition.service_id || "",
      environment: definition.environment || "",
      severity: definition.severity || "",
      metric_name: definition.metric_name || current.metric_name,
      search: definition.search || "",
      trace_id: definition.trace_id || "",
    }));
  }

  async function handleCreateChannel(event) {
    event.preventDefault();
    try {
      const config = channelForm.type === "webhook"
        ? {
            url: channelForm.webhook_url,
            method: channelForm.webhook_method,
            headers: { "X-PulseLens-Source": "ui" },
            timeout_seconds: 5,
          }
        : channelForm.type === "slack_webhook"
          ? {
              url: channelForm.webhook_url,
              method: "POST",
              headers: { "X-PulseLens-Source": "ui" },
              timeout_seconds: 5,
            }
          : channelForm.type === "email"
            ? {
                to: channelForm.email_to.split(",").map((value) => value.trim()).filter(Boolean),
                subject_prefix: "[PulseLens UI]",
              }
            : { destination: "local-log" };
      await alertingApi.createNotificationChannel(token, {
        name: channelForm.name,
        type: channelForm.type,
        config,
      });
      onNotification("Notification channel created.");
      await loadDashboard();
    } catch (error) {
      onNotification(error.message, "error");
    }
  }

  async function handleCreateReplayJob(event) {
    event.preventDefault();
    try {
      const endTime = new Date();
      const startTime = new Date(Date.now() - Number(replayForm.window_minutes) * 60 * 1000);
      await archiveApi.createReplayJob(token, {
        service_id: state.serviceId,
        event_type: replayForm.event_type,
        start_time: startTime.toISOString(),
        end_time: endTime.toISOString(),
      });
      onNotification("Replay job created.");
      await loadDashboard();
    } catch (error) {
      onNotification(error.message, "error");
    }
  }

  async function handleIncidentAction(action, incidentId) {
    try {
      if (action === "ack") {
        await alertingApi.acknowledgeIncident(token, incidentId);
      } else {
        await alertingApi.resolveIncident(token, incidentId);
      }
      onNotification(`Incident ${action === "ack" ? "acknowledged" : "resolved"}.`);
      await loadDashboard();
    } catch (error) {
      onNotification(error.message, "error");
    }
  }

  async function handleAddComment(event) {
    event.preventDefault();
    try {
      await alertingApi.addIncidentComment(token, commentForm.incidentId, { body: commentForm.body });
      const comments = await alertingApi.listIncidentComments(token, commentForm.incidentId);
      setData((current) => ({ ...current, incidentComments: comments }));
      onNotification("Incident comment added.");
    } catch (error) {
      onNotification(error.message, "error");
    }
  }

  async function handleAssignIncident(event) {
    event.preventDefault();
    try {
      await alertingApi.assignIncident(token, assignForm.incidentId, { assigned_to: assignForm.assignedTo });
      onNotification("Incident assigned.");
      await loadDashboard();
    } catch (error) {
      onNotification(error.message, "error");
    }
  }

  async function handleCreateUser(event) {
    event.preventDefault();
    try {
      await tenantApi.createUser(state.tenantId, userForm, token);
      onNotification("User created.");
      await loadDashboard();
    } catch (error) {
      onNotification(error.message, "error");
    }
  }

  async function loadIncidentDetail(incidentId) {
    try {
      const [incident, comments, timeline, deliveries] = await Promise.all([
        alertingApi.getIncident(token, incidentId),
        alertingApi.listIncidentComments(token, incidentId),
        alertingApi.incidentTimeline(token, incidentId),
        alertingApi.incidentDeliveries(token, incidentId),
      ]);
      setAssignForm((current) => ({ ...current, incidentId }));
      setCommentForm((current) => ({ ...current, incidentId }));
      setData((current) => ({
        ...current,
        selectedIncident: incident,
        incidentComments: comments,
        incidentTimeline: timeline,
        incidentDeliveries: deliveries,
      }));
    } catch (error) {
      onNotification(error.message, "error");
    }
  }

  async function loadIncidentComments(incidentId) {
    await loadIncidentDetail(incidentId);
  }

  const stats = useMemo(() => {
    if (!data.overview) {
      return [];
    }
    return [
      { label: "Logs", value: data.overview.log_count },
      { label: "Metrics", value: data.overview.metric_count },
      { label: "Trace Spans", value: data.overview.trace_count },
      { label: "Services", value: data.overview.service_count },
      { label: "Open Incidents", value: data.incidents.filter((incident) => incident.status === "open").length, tone: "warning" },
      { label: "Runtime Nodes", value: data.platformOverview?.runtime?.length ?? 0 },
    ];
  }, [data]);

  const severityChart = useMemo(() => {
    const totals = new Map();
    for (const row of data.logSeverity) {
      totals.set(row.severity, (totals.get(row.severity) || 0) + Number(row.event_count || 0));
    }
    return Array.from(totals.entries()).map(([label, value]) => ({ label, value }));
  }, [data.logSeverity]);

  const metricChart = useMemo(() => {
    return [...data.metricSeries]
      .sort((left, right) => new Date(left.bucket_start) - new Date(right.bucket_start))
      .slice(-10)
      .map((row) => ({
        label: compactTimestamp(row.bucket_start),
        value: Number(row.average_value || row.last_value || 0),
      }));
  }, [data.metricSeries]);

  const traceChart = useMemo(() => {
    return [...data.traceLatency]
      .sort((left, right) => new Date(left.bucket_start) - new Date(right.bucket_start))
      .slice(-10)
      .map((row) => ({
        label: compactTimestamp(row.bucket_start),
        value: Number(row.average_duration_ms || 0),
      }));
  }, [data.traceLatency]);

  const dependencyChart = useMemo(() => {
    const dependencies = data.platformOverview?.dependencies || [];
    const healthy = dependencies.filter((row) => row.status === "healthy").length;
    const unhealthy = dependencies.length - healthy;
    return [
      { label: "healthy", value: healthy, color: "#34d399" },
      { label: "down", value: unhealthy, color: "#f87171" },
    ];
  }, [data.platformOverview]);

  const dashboardWidgets = useMemo(() => {
    const dashboards = data.dashboards || [];
    if (!dashboards.length) {
      return [];
    }
    const widgets = [];
    for (const dashboard of dashboards.slice(0, 2)) {
      try {
        const parsed = typeof dashboard.widgets === "string" ? JSON.parse(dashboard.widgets) : dashboard.widgets;
        for (const widget of parsed || []) {
          widgets.push({ dashboard: dashboard.name, ...widget });
        }
      } catch {
        widgets.push({ dashboard: dashboard.name, type: "unknown" });
      }
    }
    return widgets;
  }, [data.dashboards]);

  const selectedDashboard = useMemo(() => data.dashboards.find((row) => row.id === dashboardBuilder.dashboard_id) || null, [data.dashboards, dashboardBuilder.dashboard_id]);
  const selectedDashboardWidgets = useMemo(() => parseJSONSafe(selectedDashboard?.widgets, []), [selectedDashboard]);

  function selectDashboardForEditing(dashboard) {
    setDashboardBuilder((current) => ({ ...current, dashboard_id: dashboard.id, widget_id: "" }));
    setDashboardForm({
      name: dashboard.name,
      description: dashboard.description,
      default_time_range: dashboard.default_time_range || "120m",
    });
  }

  function beginEditWidget(widget) {
    setDashboardBuilder((current) => ({
      ...current,
      widget_id: widget.id || "",
      title: widget.title || "",
      type: widget.type || "chart",
      dataset: widget.dataset || "log_severity",
      chart_type: widget.chart_type || "bar",
      metric: widget.metric || "",
      layout_w: String(widget.layout?.w || 1),
      layout_h: String(widget.layout?.h || 1),
    }));
  }

  async function handleDeleteWidget(widgetId) {
    try {
      if (!selectedDashboard) {
        throw new Error("select a dashboard first");
      }
      await queryApi.deleteDashboardWidget(token, selectedDashboard.id, widgetId);
      onNotification("Dashboard widget deleted.");
      await loadDashboard();
    } catch (error) {
      onNotification(error.message, "error");
    }
  }

  async function moveWidget(widgetId, direction) {
    try {
      if (!selectedDashboard) {
        throw new Error("select a dashboard first");
      }
      const widgets = [...selectedDashboardWidgets];
      const index = widgets.findIndex((row) => row.id === widgetId);
      if (index < 0) {
        return;
      }
      const targetIndex = direction === "up" ? index - 1 : index + 1;
      if (targetIndex < 0 || targetIndex >= widgets.length) {
        return;
      }
      [widgets[index], widgets[targetIndex]] = [widgets[targetIndex], widgets[index]];
      await queryApi.updateDashboard(token, selectedDashboard.id, {
        name: selectedDashboard.name,
        description: selectedDashboard.description,
        default_time_range: selectedDashboard.default_time_range || "120m",
        layout: parseJSONSafe(selectedDashboard.layout, { columns: 2 }),
        widgets,
      });
      onNotification("Dashboard widget order updated.");
      await loadDashboard();
    } catch (error) {
      onNotification(error.message, "error");
    }
  }

  if (!token) {
    return (
      <Section title="Dashboard">
        <p>Login first to load tenant data.</p>
      </Section>
    );
  }

  return (
    <div className="dashboard-grid">
      <Section title="Workspace Context" actions={<button onClick={loadDashboard}>{loading ? "Refreshing..." : "Refresh"}</button>}>
        <div className="key-value-grid">
          <span>Tenant ID</span><code>{state.tenantId || "-"}</code>
          <span>Service ID</span><code>{state.serviceId || "-"}</code>
          <span>User</span><code>{data.me?.email || state.email || "-"}</code>
          <span>Role</span><code>{data.me?.role || "-"}</code>
          <span>API Key Prefix</span><code>{state.apiKey ? state.apiKey.slice(0, 16) : "-"}</code>
        </div>
      </Section>

      <Section title="Platform Summary">
        <div className="stat-grid">
          {stats.map((item) => (
            <StatCard key={item.label} label={item.label} value={item.value} tone={item.tone} />
          ))}
        </div>
      </Section>

      <Section title="Query Filters">
        <form className="form-grid" onSubmit={(event) => { event.preventDefault(); loadDashboard(); }}>
          <label>
            Service
            <select data-testid="filter-service" value={filterForm.service_id} onChange={(event) => setFilterForm({ ...filterForm, service_id: event.target.value })}>
              <option value="">all</option>
              {data.services.map((service) => (
                <option key={service.id} value={service.id}>{service.name}</option>
              ))}
            </select>
          </label>
          <label>
            Environment
            <input data-testid="filter-environment" value={filterForm.environment} onChange={(event) => setFilterForm({ ...filterForm, environment: event.target.value })} />
          </label>
          <label>
            Severity
            <input data-testid="filter-severity" value={filterForm.severity} onChange={(event) => setFilterForm({ ...filterForm, severity: event.target.value })} />
          </label>
          <label>
            Metric
            <input data-testid="filter-metric" value={filterForm.metric_name} onChange={(event) => setFilterForm({ ...filterForm, metric_name: event.target.value })} />
          </label>
          <label>
            Search
            <input data-testid="filter-search" value={filterForm.search} onChange={(event) => setFilterForm({ ...filterForm, search: event.target.value })} />
          </label>
          <label>
            Lookback Minutes
            <input data-testid="filter-lookback" value={filterForm.lookback_minutes} onChange={(event) => setFilterForm({ ...filterForm, lookback_minutes: event.target.value })} />
          </label>
          <div className="form-actions">
            <button data-testid="apply-filters" type="submit">Apply Filters</button>
          </div>
        </form>
      </Section>

      <Section title="Live Widgets">
        <div className="widget-grid">
          <div className="widget-card">
            <h3>Log Severity</h3>
            <BarChart data={severityChart} />
          </div>
          <div className="widget-card">
            <h3>Metric Trend</h3>
            <LineChart data={metricChart} color="#34d399" />
          </div>
          <div className="widget-card">
            <h3>Trace Latency</h3>
            <LineChart data={traceChart} color="#f59e0b" />
          </div>
          <div className="widget-card">
            <h3>Dependency Health</h3>
            <BarChart data={dependencyChart} />
          </div>
        </div>
      </Section>

      <Section title="Saved Dashboard Widgets">
        <DataTable
          columns={[
            { key: "dashboard", label: "Dashboard" },
            { key: "type", label: "Widget Type" },
            { key: "metric", label: "Metric" },
            { key: "dataset", label: "Dataset" },
          ]}
          rows={dashboardWidgets}
        />
      </Section>

      <Section title="Platform Runtime">
        <DataTable
          columns={[
            { key: "service_name", label: "Service" },
            { key: "mode", label: "Mode" },
            { key: "status", label: "Status" },
            { key: "port", label: "Port" },
            { key: "last_seen_at", label: "Last Seen", render: (row) => formatDate(row.last_seen_at) },
          ]}
          rows={data.platformOverview?.runtime || []}
        />
      </Section>

      <Section title="Queue Backpressure">
        <DataTable
          columns={[
            { key: "queue", label: "Queue" },
            { key: "pending", label: "Pending" },
            { key: "threshold", label: "Threshold" },
            { key: "overloaded", label: "Overloaded", render: (row) => (row.overloaded ? "yes" : "no") },
          ]}
          rows={data.platformOverview?.backpressure || []}
        />
      </Section>

      <Section title="Dependency Health">
        <DataTable
          columns={[
            { key: "name", label: "Dependency" },
            { key: "type", label: "Type" },
            { key: "status", label: "Status" },
            { key: "message", label: "Message" },
            { key: "checked_at", label: "Checked", render: (row) => formatDate(row.checked_at) },
          ]}
          rows={data.platformOverview?.dependencies || []}
        />
      </Section>

      <Section title="Kafka Lag">
        <DataTable
          columns={[
            { key: "topic", label: "Topic" },
            { key: "partition", label: "Partition" },
            { key: "group_id", label: "Group" },
            { key: "current_offset", label: "Current" },
            { key: "latest_offset", label: "Latest" },
            { key: "lag", label: "Lag" },
            { key: "member_assigned", label: "Assigned", render: (row) => (row.member_assigned ? "yes" : "no") },
          ]}
          rows={data.platformOverview?.kafka_lag || []}
        />
      </Section>

      <Section title="Retention Cleanup">
        <DataTable
          columns={[
            { key: "status", label: "Status" },
            { key: "telemetry_deleted", label: "Telemetry Deleted" },
            { key: "archive_deleted", label: "Archive Deleted" },
            { key: "file_delete_errors", label: "File Errors" },
            { key: "started_at", label: "Started", render: (row) => formatDate(row.started_at) },
            { key: "completed_at", label: "Completed", render: (row) => formatDate(row.completed_at) },
          ]}
          rows={data.platformOverview?.cleanup_runs || []}
        />
      </Section>

      <Section title="Ingest Playground">
        <div className="button-row">
          <button data-testid="send-log" onClick={() => handleIngest("log")} disabled={!state.apiKey}>Send Log</button>
          <button data-testid="send-metric" onClick={() => handleIngest("metric")} disabled={!state.apiKey}>Send Metric</button>
          <button data-testid="send-trace" onClick={() => handleIngest("trace")} disabled={!state.apiKey}>Send Trace</button>
        </div>
      </Section>

      <Section title="Alert Rules">
        <form className="form-grid" onSubmit={handleCreateRule}>
          <label>
            Rule Name
            <input value={ruleForm.name} onChange={(event) => setRuleForm({ ...ruleForm, name: event.target.value })} />
          </label>
          <label>
            Signal Type
            <select value={ruleForm.signal_type} onChange={(event) => setRuleForm({ ...ruleForm, signal_type: event.target.value })}>
              <option value="log">log</option>
              <option value="metric">metric</option>
              <option value="trace">trace</option>
            </select>
          </label>
          <label>
            Severity
            <input value={ruleForm.severity} onChange={(event) => setRuleForm({ ...ruleForm, severity: event.target.value })} />
          </label>
          <label>
            Aggregation
            <select value={ruleForm.aggregation} onChange={(event) => setRuleForm({ ...ruleForm, aggregation: event.target.value })}>
              <option value="count">count</option>
              <option value="avg">avg</option>
              <option value="max">max</option>
              <option value="sum">sum</option>
            </select>
          </label>
          <label>
            Comparator
            <select value={ruleForm.comparator} onChange={(event) => setRuleForm({ ...ruleForm, comparator: event.target.value })}>
              <option value=">">{">"}</option>
              <option value=">=">{">="}</option>
              <option value="<">{"<"}</option>
              <option value="<=">{"<="}</option>
            </select>
          </label>
          <label>
            Threshold
            <input value={ruleForm.threshold} onChange={(event) => setRuleForm({ ...ruleForm, threshold: event.target.value })} />
          </label>
          <label>
            Window Minutes
            <input value={ruleForm.window_minutes} onChange={(event) => setRuleForm({ ...ruleForm, window_minutes: event.target.value })} />
          </label>
          <label>
            Cooldown Minutes
            <input value={ruleForm.cooldown_minutes} onChange={(event) => setRuleForm({ ...ruleForm, cooldown_minutes: event.target.value })} />
          </label>
          <label>
            Metric Name
            <input value={ruleForm.metric_name} onChange={(event) => setRuleForm({ ...ruleForm, metric_name: event.target.value })} />
          </label>
          <label>
            Policy
            <select value={ruleForm.policy_id} onChange={(event) => setRuleForm({ ...ruleForm, policy_id: event.target.value })}>
              <option value="">select</option>
              {data.policies.map((policy) => (
                <option key={policy.id} value={policy.id}>{policy.name}</option>
              ))}
            </select>
          </label>
          <div className="form-actions">
            <button type="submit">Create Rule</button>
          </div>
        </form>
        <DataTable
          columns={[
            { key: "name", label: "Name" },
            { key: "signal_type", label: "Signal" },
            { key: "comparator", label: "Comparator" },
            { key: "threshold", label: "Threshold" },
            { key: "window_minutes", label: "Window" },
            { key: "policy_id", label: "Policy" },
            { key: "active", label: "Active", render: (row) => (row.active ? "yes" : "no") },
          ]}
          rows={data.rules}
        />
      </Section>

      <Section title="Alert Policies">
        <form className="form-grid" onSubmit={handleCreatePolicy}>
          <label>
            Name
            <input data-testid="policy-name" value={policyForm.name} onChange={(event) => setPolicyForm({ ...policyForm, name: event.target.value })} />
          </label>
          <label>
            Max Delivery Attempts
            <input data-testid="policy-attempts" value={policyForm.max_delivery_attempts} onChange={(event) => setPolicyForm({ ...policyForm, max_delivery_attempts: event.target.value })} />
          </label>
          <label>
            Backoff ms
            <input data-testid="policy-backoff" value={policyForm.delivery_backoff_millis} onChange={(event) => setPolicyForm({ ...policyForm, delivery_backoff_millis: event.target.value })} />
          </label>
          <label>
            Escalation Minutes
            <input data-testid="policy-escalation" value={policyForm.escalation_interval_minutes} onChange={(event) => setPolicyForm({ ...policyForm, escalation_interval_minutes: event.target.value })} />
          </label>
          <label>
            Max Escalations
            <input data-testid="policy-max-escalations" value={policyForm.max_escalations} onChange={(event) => setPolicyForm({ ...policyForm, max_escalations: event.target.value })} />
          </label>
          <label>
            Open Channels
            <input data-testid="policy-open-channels" value={policyForm.open_channel_types} onChange={(event) => setPolicyForm({ ...policyForm, open_channel_types: event.target.value })} />
          </label>
          <div className="form-actions">
            <button data-testid="create-policy" type="submit">Create Policy</button>
          </div>
        </form>
        <DataTable
          columns={[
            { key: "name", label: "Name" },
            { key: "max_delivery_attempts", label: "Attempts" },
            { key: "delivery_backoff_millis", label: "Backoff ms" },
            { key: "escalation_interval_minutes", label: "Escalation Min" },
            { key: "max_escalations", label: "Max Esc" },
          ]}
          rows={data.policies}
        />
      </Section>

      <Section title="Incidents">
        <form className="form-grid" onSubmit={(event) => { event.preventDefault(); loadDashboard(); }}>
          <label>
            Status
            <select value={incidentFilters.status} onChange={(event) => setIncidentFilters({ ...incidentFilters, status: event.target.value })}>
              <option value="">all</option>
              <option value="open">open</option>
              <option value="acknowledged">acknowledged</option>
              <option value="resolved">resolved</option>
            </select>
          </label>
          <label>
            Assigned To
            <input value={incidentFilters.assigned_to} onChange={(event) => setIncidentFilters({ ...incidentFilters, assigned_to: event.target.value })} />
          </label>
          <label>
            Service
            <select value={incidentFilters.service_id} onChange={(event) => setIncidentFilters({ ...incidentFilters, service_id: event.target.value })}>
              <option value="">all</option>
              {data.services.map((service) => (
                <option key={service.id} value={service.id}>{service.name}</option>
              ))}
            </select>
          </label>
          <label>
            Severity
            <input value={incidentFilters.severity} onChange={(event) => setIncidentFilters({ ...incidentFilters, severity: event.target.value })} />
          </label>
          <div className="form-actions">
            <button type="submit">Apply Incident Filters</button>
          </div>
        </form>
        <DataTable
          columns={[
            { key: "title", label: "Title" },
            { key: "severity", label: "Severity" },
            { key: "status", label: "Status" },
            { key: "assigned_to", label: "Assigned To" },
            { key: "escalation_level", label: "Escalation" },
            { key: "observed_value", label: "Observed" },
            { key: "threshold", label: "Threshold" },
            { key: "triggered_at", label: "Triggered", render: (row) => formatDate(row.triggered_at) },
            { key: "resolved_at", label: "Resolved", render: (row) => formatDate(row.resolved_at) },
            {
              key: "actions",
              label: "Actions",
              render: (row) => (
                <div className="button-row">
                  <button onClick={() => handleIncidentAction("ack", row.id)}>Ack</button>
                  <button onClick={() => handleIncidentAction("resolve", row.id)}>Resolve</button>
                  <button onClick={() => loadIncidentDetail(row.id)}>Details</button>
                </div>
              ),
            },
          ]}
          rows={data.incidents}
        />
      </Section>

      <Section title="Incident Details">
        {data.selectedIncident ? (
          <div className="key-value-grid">
            <span>Incident</span><code>{data.selectedIncident.id}</code>
            <span>Status</span><code>{data.selectedIncident.status}</code>
            <span>Severity</span><code>{data.selectedIncident.severity || "-"}</code>
            <span>Assigned</span><code>{data.selectedIncident.assigned_to || "-"}</code>
            <span>Escalation Count</span><code>{data.selectedIncident.escalation_count ?? 0}</code>
          </div>
        ) : (
          <p>Select an incident to inspect timeline, comments, and deliveries.</p>
        )}
        <DataTable
          columns={[
            { key: "event_type", label: "Event" },
            { key: "summary", label: "Summary" },
            { key: "actor_id", label: "Actor" },
            { key: "created_at", label: "Created", render: (row) => formatDate(row.created_at) },
          ]}
          rows={data.incidentTimeline}
        />
      </Section>

      <Section title="Incident Comments">
        <form className="form-grid" onSubmit={handleAssignIncident}>
          <label>
            Incident ID
            <input data-testid="assign-incident-id" value={assignForm.incidentId} onChange={(event) => setAssignForm({ ...assignForm, incidentId: event.target.value })} />
          </label>
          <label>
            Assign To
            <input data-testid="assign-incident-user" value={assignForm.assignedTo} onChange={(event) => setAssignForm({ ...assignForm, assignedTo: event.target.value })} />
          </label>
          <div className="form-actions">
            <button data-testid="assign-incident-submit" type="submit">Assign Incident</button>
          </div>
        </form>
        <form className="form-grid" onSubmit={handleAddComment}>
          <label>
            Incident ID
            <input value={commentForm.incidentId} onChange={(event) => setCommentForm({ ...commentForm, incidentId: event.target.value })} />
          </label>
          <label>
            Comment
            <input value={commentForm.body} onChange={(event) => setCommentForm({ ...commentForm, body: event.target.value })} />
          </label>
          <div className="form-actions">
            <button type="submit">Add Comment</button>
          </div>
        </form>
        <DataTable
          columns={[
            { key: "author_id", label: "Author" },
            { key: "body", label: "Comment" },
            { key: "created_at", label: "Created", render: (row) => formatDate(row.created_at) },
          ]}
          rows={data.incidentComments}
        />
      </Section>

      <Section title="Notification Channels">
        <form className="form-grid" onSubmit={handleCreateChannel}>
          <label>
            Channel Name
            <input data-testid="channel-name" value={channelForm.name} onChange={(event) => setChannelForm({ ...channelForm, name: event.target.value })} />
          </label>
          <label>
            Channel Type
            <select data-testid="channel-type" value={channelForm.type} onChange={(event) => setChannelForm({ ...channelForm, type: event.target.value })}>
              <option value="log">log</option>
              <option value="webhook">webhook</option>
              <option value="slack_webhook">slack_webhook</option>
              <option value="email">email</option>
            </select>
          </label>
          {channelForm.type === "webhook" || channelForm.type === "slack_webhook" ? (
            <>
              <label>
                Webhook URL
                <input data-testid="channel-webhook-url" value={channelForm.webhook_url} onChange={(event) => setChannelForm({ ...channelForm, webhook_url: event.target.value })} />
              </label>
              <label>
                Method
                <select data-testid="channel-webhook-method" value={channelForm.webhook_method} onChange={(event) => setChannelForm({ ...channelForm, webhook_method: event.target.value })}>
                  <option value="POST">POST</option>
                  <option value="PUT">PUT</option>
                </select>
              </label>
            </>
          ) : null}
          {channelForm.type === "email" ? (
            <label>
              Recipients
              <input data-testid="channel-email-to" value={channelForm.email_to} onChange={(event) => setChannelForm({ ...channelForm, email_to: event.target.value })} />
            </label>
          ) : null}
          <div className="form-actions">
            <button data-testid="create-channel" type="submit">Create Channel</button>
          </div>
        </form>
        <DataTable
          columns={[
            { key: "name", label: "Name" },
            { key: "type", label: "Type" },
            { key: "active", label: "Active", render: (row) => (row.active ? "yes" : "no") },
          ]}
          rows={data.channels}
        />
      </Section>

      <Section title="Notification Deliveries">
        <DataTable
          columns={[
            { key: "event_type", label: "Event" },
            { key: "status", label: "Status" },
            { key: "attempt_count", label: "Attempts" },
            { key: "channel_id", label: "Channel" },
            { key: "incident_id", label: "Incident" },
            { key: "response", label: "Response" },
            { key: "delivered_at", label: "Delivered", render: (row) => formatDate(row.delivered_at) },
          ]}
          rows={data.selectedIncident ? data.incidentDeliveries : data.deliveries}
        />
      </Section>

      <Section title="Service Health">
        <DataTable
          columns={[
            { key: "service_name", label: "Service" },
            { key: "environment", label: "Env" },
            { key: "event_count", label: "Events" },
            { key: "error_log_count", label: "Errors" },
            { key: "critical_log_count", label: "Critical" },
            { key: "health_status", label: "Health" },
            { key: "last_event_at", label: "Last Event", render: (row) => formatDate(row.last_event_at) },
          ]}
          rows={data.health}
        />
      </Section>

      <Section title="Log Severity Rollups">
        <DataTable
          columns={[
            { key: "bucket_start", label: "Bucket", render: (row) => formatDate(row.bucket_start) },
            { key: "service_name", label: "Service" },
            { key: "severity", label: "Severity" },
            { key: "event_count", label: "Count" },
          ]}
          rows={data.logSeverity}
        />
      </Section>

      <Section title="Metric Rollups">
        <DataTable
          columns={[
            { key: "bucket_start", label: "Bucket", render: (row) => formatDate(row.bucket_start) },
            { key: "metric_name", label: "Metric" },
            { key: "sample_count", label: "Samples" },
            { key: "average_value", label: "Avg" },
            { key: "max_value", label: "Max" },
          ]}
          rows={data.metricSeries}
        />
      </Section>

      <Section title="Trace Latency Rollups">
        <DataTable
          columns={[
            { key: "bucket_start", label: "Bucket", render: (row) => formatDate(row.bucket_start) },
            { key: "service_name", label: "Service" },
            { key: "operation", label: "Operation" },
            { key: "span_count", label: "Spans" },
            { key: "average_duration_ms", label: "Avg ms" },
            { key: "max_duration_ms", label: "Max ms" },
          ]}
          rows={data.traceLatency}
        />
      </Section>

      <Section title="Recent Logs">
        <DataTable
          columns={[
            { key: "service_name", label: "Service" },
            { key: "severity", label: "Severity" },
            { key: "message", label: "Message" },
            { key: "trace_id", label: "Trace ID" },
            { key: "occurred_at", label: "Occurred", render: (row) => formatDate(row.occurred_at) },
          ]}
          rows={data.logs}
        />
      </Section>

      <Section title="Recent Metrics">
        <DataTable
          columns={[
            { key: "service_name", label: "Service" },
            { key: "metric_name", label: "Metric" },
            { key: "value", label: "Value" },
            { key: "occurred_at", label: "Occurred", render: (row) => formatDate(row.occurred_at) },
          ]}
          rows={data.metrics}
        />
      </Section>

      <Section title="Traces">
        <DataTable
          columns={[
            { key: "trace_id", label: "Trace ID" },
            { key: "service_name", label: "Service" },
            { key: "span_count", label: "Spans" },
            { key: "last_seen_at", label: "Last Seen", render: (row) => formatDate(row.last_seen_at) },
          ]}
          rows={data.traces}
        />
      </Section>

      <Section title="Audit Log">
        <DataTable
          columns={[
            { key: "action", label: "Action" },
            { key: "resource_type", label: "Resource" },
            { key: "resource_id", label: "Resource ID" },
            { key: "actor_user_id", label: "Actor" },
            { key: "created_at", label: "Created", render: (row) => formatDate(row.created_at) },
          ]}
          rows={data.auditLogs}
        />
      </Section>

      <Section title="Users">
        <form className="form-grid" onSubmit={handleCreateUser}>
          <label>
            Name
            <input data-testid="user-name" value={userForm.name} onChange={(event) => setUserForm({ ...userForm, name: event.target.value })} />
          </label>
          <label>
            Email
            <input data-testid="user-email" value={userForm.email} onChange={(event) => setUserForm({ ...userForm, email: event.target.value })} />
          </label>
          <label>
            Password
            <input data-testid="user-password" value={userForm.password} onChange={(event) => setUserForm({ ...userForm, password: event.target.value })} />
          </label>
          <label>
            Role
            <select data-testid="user-role" value={userForm.role} onChange={(event) => setUserForm({ ...userForm, role: event.target.value })}>
              <option value="viewer">viewer</option>
              <option value="operator">operator</option>
              <option value="alert_manager">alert_manager</option>
              <option value="service_owner">service_owner</option>
            </select>
          </label>
          <div className="form-actions">
            <button data-testid="create-user" type="submit">Create User</button>
          </div>
        </form>
        <DataTable
          columns={[
            { key: "name", label: "Name" },
            { key: "email", label: "Email" },
            { key: "role", label: "Role" },
          ]}
          rows={data.users}
        />
      </Section>

      <Section title="Saved Queries">
        <form className="form-grid" onSubmit={handleCreateSavedQuery}>
          <label>
            Query Name
            <input data-testid="saved-query-name" value={savedQueryForm.name} onChange={(event) => setSavedQueryForm({ ...savedQueryForm, name: event.target.value })} />
          </label>
          <label>
            Query Type
            <select data-testid="saved-query-type" value={savedQueryForm.query_type} onChange={(event) => setSavedQueryForm({ ...savedQueryForm, query_type: event.target.value })}>
              <option value="logs">logs</option>
              <option value="metrics">metrics</option>
              <option value="traces">traces</option>
            </select>
          </label>
          <div className="form-actions">
            <button data-testid="create-saved-query" type="submit">Save Query</button>
          </div>
          <div className="form-actions">
            <button data-testid="update-first-query" type="button" onClick={handleSaveCurrentFiltersAsQuery}>Update First Query From Filters</button>
          </div>
        </form>
        <DataTable
          columns={[
            { key: "name", label: "Name" },
            { key: "query_type", label: "Type" },
            { key: "created_by", label: "Created By" },
            { key: "apply", label: "Apply", render: (row) => <button data-testid={`apply-query-${row.id}`} onClick={() => handleApplySavedQuery(row)}>Use</button> },
          ]}
          rows={data.savedQueries}
        />
      </Section>

      <Section title="Dashboards">
        <form className="form-grid" onSubmit={handleCreateDashboard}>
          <label>
            Dashboard Name
            <input data-testid="dashboard-name" value={dashboardForm.name} onChange={(event) => setDashboardForm({ ...dashboardForm, name: event.target.value })} />
          </label>
          <label>
            Description
            <input data-testid="dashboard-description" value={dashboardForm.description} onChange={(event) => setDashboardForm({ ...dashboardForm, description: event.target.value })} />
          </label>
          <label>
            Default Time Range
            <select value={dashboardForm.default_time_range} onChange={(event) => setDashboardForm({ ...dashboardForm, default_time_range: event.target.value })}>
              <option value="30m">30m</option>
              <option value="120m">120m</option>
              <option value="24h">24h</option>
            </select>
          </label>
          <div className="form-actions">
            <button data-testid="create-dashboard" type="submit">Create Dashboard</button>
          </div>
          <div className="form-actions">
            <button type="button" onClick={handleUpdateDashboardDetails}>Update Selected Dashboard</button>
          </div>
        </form>
        <form className="form-grid" onSubmit={handleSaveDashboardWidget}>
          <label>
            Dashboard
            <select data-testid="dashboard-widget-dashboard" value={dashboardBuilder.dashboard_id} onChange={(event) => setDashboardBuilder({ ...dashboardBuilder, dashboard_id: event.target.value })}>
              <option value="">select</option>
              {data.dashboards.map((dashboard) => (
                <option key={dashboard.id} value={dashboard.id}>{dashboard.name}</option>
              ))}
            </select>
          </label>
          <label>
            Widget Title
            <input data-testid="dashboard-widget-title" value={dashboardBuilder.title} onChange={(event) => setDashboardBuilder({ ...dashboardBuilder, title: event.target.value })} />
          </label>
          <label>
            Width
            <input value={dashboardBuilder.layout_w} onChange={(event) => setDashboardBuilder({ ...dashboardBuilder, layout_w: event.target.value })} />
          </label>
          <label>
            Height
            <input value={dashboardBuilder.layout_h} onChange={(event) => setDashboardBuilder({ ...dashboardBuilder, layout_h: event.target.value })} />
          </label>
          <label>
            Widget Type
            <select data-testid="dashboard-widget-type" value={dashboardBuilder.type} onChange={(event) => setDashboardBuilder({ ...dashboardBuilder, type: event.target.value })}>
              <option value="chart">chart</option>
              <option value="table">table</option>
              <option value="stat">stat</option>
            </select>
          </label>
          <label>
            Dataset
            <select data-testid="dashboard-widget-dataset" value={dashboardBuilder.dataset} onChange={(event) => setDashboardBuilder({ ...dashboardBuilder, dataset: event.target.value })}>
              <option value="log_severity">log_severity</option>
              <option value="metric_series">metric_series</option>
              <option value="trace_latency">trace_latency</option>
              <option value="service_health">service_health</option>
              <option value="logs">logs</option>
            </select>
          </label>
          <label>
            Chart Type
            <select data-testid="dashboard-widget-chart-type" value={dashboardBuilder.chart_type} onChange={(event) => setDashboardBuilder({ ...dashboardBuilder, chart_type: event.target.value })}>
              <option value="bar">bar</option>
              <option value="line">line</option>
            </select>
          </label>
          <div className="form-actions">
            <button data-testid="save-dashboard-widget" type="submit">{dashboardBuilder.widget_id ? "Update Widget" : "Add Widget To Dashboard"}</button>
          </div>
        </form>
        <DataTable
          columns={[
            { key: "name", label: "Name" },
            { key: "description", label: "Description" },
            { key: "default_time_range", label: "Time Range" },
            { key: "created_by", label: "Created By" },
            { key: "edit", label: "Edit", render: (row) => <button onClick={() => selectDashboardForEditing(row)}>Select</button> },
          ]}
          rows={data.dashboards}
        />
      </Section>

      <Section title="Dashboard Builder Preview">
        {selectedDashboard ? (
          <div className="widget-grid">
            {selectedDashboardWidgets.map((widget, index) => {
              const rows = widgetDataset(widget, data);
              return (
                <div className="widget-card" key={`${widget.title || widget.dataset}-${index}`}>
                  <h3>{widget.title || widget.dataset || widget.type}</h3>
                  <div className="button-row">
                    <button onClick={() => beginEditWidget(widget)}>Edit</button>
                    <button onClick={() => moveWidget(widget.id, "up")}>Up</button>
                    <button onClick={() => moveWidget(widget.id, "down")}>Down</button>
                    <button onClick={() => handleDeleteWidget(widget.id)}>Delete</button>
                  </div>
                  {widget.type === "stat" ? (
                    <StatCard label={widget.metric || widget.title || "value"} value={data.overview?.[widget.metric] ?? rows.length} />
                  ) : widget.type === "table" ? (
                    <DataTable columns={Object.keys(rows[0] || {}).slice(0, 4).map((key) => ({ key, label: key }))} rows={rows.slice(0, 5)} />
                  ) : widget.chart_type === "line" ? (
                    <LineChart
                      data={rows.slice(0, 10).map((row) => ({
                        label: compactTimestamp(row.bucket_start || row.occurred_at || row.last_seen_at),
                        value: Number(row.average_value || row.average_duration_ms || row.event_count || row.value || 0),
                      }))}
                      color="#34d399"
                    />
                  ) : (
                    <BarChart
                      data={rows.slice(0, 10).map((row) => ({
                        label: row.severity || row.metric_name || row.service_name || compactTimestamp(row.bucket_start || row.occurred_at),
                        value: Number(row.event_count || row.average_value || row.average_duration_ms || row.value || 0),
                      }))}
                    />
                  )}
                </div>
              );
            })}
          </div>
        ) : (
          <p>Select a dashboard to preview its widgets.</p>
        )}
      </Section>

      <Section title="Archive And Replay">
        <div className="key-value-grid">
          <span>Replay Jobs</span><code>{data.replayStats?.replay_job_count ?? 0}</code>
        </div>
        <form className="form-grid" onSubmit={handleCreateReplayJob}>
          <label>
            Event Type
            <select value={replayForm.event_type} onChange={(event) => setReplayForm({ ...replayForm, event_type: event.target.value })}>
              <option value="log">log</option>
              <option value="metric">metric</option>
              <option value="trace">trace</option>
              <option value="custom">custom</option>
            </select>
          </label>
          <label>
            Window Minutes
            <input value={replayForm.window_minutes} onChange={(event) => setReplayForm({ ...replayForm, window_minutes: event.target.value })} />
          </label>
          <div className="form-actions">
            <button type="submit">Create Replay Job</button>
          </div>
        </form>
        <DataTable
          columns={[
            { key: "event_type", label: "Event Type" },
            { key: "status", label: "Status" },
            { key: "replay_count", label: "Replay Count" },
            { key: "start_time", label: "Start", render: (row) => formatDate(row.start_time) },
            { key: "end_time", label: "End", render: (row) => formatDate(row.end_time) },
          ]}
          rows={data.replayJobs}
        />
      </Section>
    </div>
  );
}
