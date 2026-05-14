import { useEffect, useMemo, useState } from "react";
import BarChart from "../components/BarChart";
import DashboardBuilderSection from "../components/dashboard/DashboardBuilderSection";
import DataTable from "../components/DataTable";
import IncidentWorkspaceSection from "../components/incidents/IncidentWorkspaceSection";
import LineChart from "../components/LineChart";
import Section from "../components/Section";
import StatCard from "../components/StatCard";
import { alertingApi, archiveApi, ingestApi, queryApi, tenantApi } from "../lib/api";
import {
  builderFromDashboardWidget,
  builderToWidgetPayload,
  defaultDashboardBuilder,
  defaultDashboardForm,
  widgetFormDefaults,
} from "../lib/dashboardState";
import {
  buildAnalyticsFilters,
  buildDashboardPreviewAnalyticsFilters,
  buildDashboardPreviewFilters,
  buildLogFilters,
  buildMetricFilters,
  buildSavedQueryDefinition,
  buildTraceFilters,
} from "../lib/queryFilters";
import { compactTimestamp, parseJSONSafe } from "../lib/widgets";

function formatDate(value) {
  if (!value) {
    return "-";
  }
  return new Date(value).toLocaleString();
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
    widgetLogs: [],
    widgetMetrics: [],
    widgetTraces: [],
    widgetHealth: [],
    widgetLogSeverity: [],
    widgetMetricSeries: [],
    widgetTraceLatency: [],
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
  const [dashboardForm, setDashboardForm] = useState(defaultDashboardForm);
  const [dashboardBuilder, setDashboardBuilder] = useState(defaultDashboardBuilder);
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
  const [selectedIncidentID, setSelectedIncidentID] = useState("");

  const token = state.token;

  async function loadDashboard() {
    if (!token || !state.tenantId) {
      return;
    }
    setLoading(true);
    try {
      const logFilters = buildLogFilters(filterForm);
      const metricFilters = buildMetricFilters(filterForm);
      const traceFilters = buildTraceFilters(filterForm);
      const logAnalyticsFilters = buildAnalyticsFilters(filterForm, "log_severity");
      const metricAnalyticsFilters = buildAnalyticsFilters(filterForm, "metric_series");
      const traceAnalyticsFilters = buildAnalyticsFilters(filterForm, "trace_latency");
      const widgetPreviewFilters = buildDashboardPreviewFilters(filterForm);
      const widgetPreviewAnalyticsFilters = buildDashboardPreviewAnalyticsFilters(filterForm);
      const [
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
        users,
        services,
        logSeverity,
        metricSeries,
        traceLatency,
        widgetLogs,
        widgetMetrics,
        widgetTraces,
        widgetHealth,
        widgetLogSeverity,
        widgetMetricSeries,
        widgetTraceLatency,
      ] = await Promise.all([
        tenantApi.me(token),
        queryApi.overview(token),
        queryApi.platformOverview(token),
        queryApi.serviceHealth(token),
        queryApi.logsWithFilters(token, logFilters),
        queryApi.metricsWithFilters(token, metricFilters),
        queryApi.tracesWithFilters(token, traceFilters),
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
        queryApi.logSeverityWithFilters(token, logAnalyticsFilters),
        queryApi.metricSeriesWithFilters(token, metricAnalyticsFilters),
        queryApi.traceLatencyWithFilters(token, traceAnalyticsFilters),
        queryApi.logsWithFilters(token, widgetPreviewFilters),
        queryApi.metricsWithFilters(token, widgetPreviewFilters),
        queryApi.tracesWithFilters(token, widgetPreviewFilters),
        queryApi.serviceHealth(token),
        queryApi.logSeverityWithFilters(token, widgetPreviewAnalyticsFilters),
        queryApi.metricSeriesWithFilters(token, widgetPreviewAnalyticsFilters),
        queryApi.traceLatencyWithFilters(token, widgetPreviewAnalyticsFilters),
      ]);

      setData((current) => ({
        ...current,
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
        users,
        services,
        logSeverity,
        metricSeries,
        traceLatency,
        widgetLogs,
        widgetMetrics,
        widgetTraces,
        widgetHealth,
        widgetLogSeverity,
        widgetMetricSeries,
        widgetTraceLatency,
        selectedIncident: current.selectedIncident && selectedIncidentID
          ? incidents.find((incident) => incident.id === selectedIncidentID) || current.selectedIncident
          : null,
        incidentComments: current.selectedIncident && selectedIncidentID ? current.incidentComments : [],
        incidentTimeline: current.selectedIncident && selectedIncidentID ? current.incidentTimeline : [],
        incidentDeliveries: current.selectedIncident && selectedIncidentID ? current.incidentDeliveries : [],
      }));
      setRuleForm((current) => ({
        ...current,
        policy_id: current.policy_id || policies[0]?.id || "",
      }));
      setDashboardBuilder((current) => ({
        ...current,
        dashboard_id: current.dashboard_id || dashboards[0]?.id || "",
      }));
      if (selectedIncidentID) {
        const selected = incidents.find((incident) => incident.id === selectedIncidentID);
        if (!selected) {
          setSelectedIncidentID("");
          setData((current) => ({
            ...current,
            selectedIncident: null,
            incidentComments: [],
            incidentTimeline: [],
            incidentDeliveries: [],
          }));
        }
        if (selected) {
          await loadIncidentDetail(selectedIncidentID, { silent: true });
        }
      }
    } catch (error) {
      onNotification(error.message, "error");
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    loadDashboard();
  }, [token, state.tenantId]);

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
          ...buildSavedQueryDefinition(filterForm),
          service_id: filterForm.service_id || state.serviceId,
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
      const widgetPayload = builderToWidgetPayload(dashboardBuilder);
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
      setDashboardBuilder((current) => ({
        ...defaultDashboardBuilder(),
        dashboard_id: current.dashboard_id || selected.id,
      }));
      await loadDashboard();
    } catch (error) {
      onNotification(error.message, "error");
    }
  }

  async function handleUpdateDashboardDetails() {
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
        definition: buildSavedQueryDefinition(filterForm),
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
      await loadIncidentDetail(incidentId);
    } catch (error) {
      onNotification(error.message, "error");
    }
  }

  async function handleAddComment(event) {
    event.preventDefault();
    try {
      await alertingApi.addIncidentComment(token, commentForm.incidentId, { body: commentForm.body });
      await loadIncidentDetail(commentForm.incidentId);
      setCommentForm((current) => ({ ...current, body: "Investigating now." }));
      onNotification("Incident comment added.");
    } catch (error) {
      onNotification(error.message, "error");
    }
  }

  async function handleAssignIncident(event) {
    event.preventDefault();
    try {
      await alertingApi.assignIncident(token, assignForm.incidentId, { assigned_to: assignForm.assignedTo });
      await loadDashboard();
      await loadIncidentDetail(assignForm.incidentId);
      onNotification("Incident assigned.");
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

  async function loadIncidentDetail(incidentId, options = {}) {
    try {
      const [incident, comments, timeline, deliveries] = await Promise.all([
        alertingApi.getIncident(token, incidentId),
        alertingApi.listIncidentComments(token, incidentId),
        alertingApi.incidentTimeline(token, incidentId),
        alertingApi.incidentDeliveries(token, incidentId),
      ]);
      setAssignForm((current) => ({ ...current, incidentId }));
      setCommentForm((current) => ({ ...current, incidentId }));
      setSelectedIncidentID(incidentId);
      setData((current) => ({
        ...current,
        selectedIncident: incident,
        incidentComments: comments,
        incidentTimeline: timeline,
        incidentDeliveries: deliveries,
      }));
    } catch (error) {
      if (!options.silent) {
        onNotification(error.message, "error");
      }
    }
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

  const metricChart = useMemo(() => (
    [...data.metricSeries]
      .sort((left, right) => new Date(left.bucket_start) - new Date(right.bucket_start))
      .slice(-10)
      .map((row) => ({
        label: compactTimestamp(row.bucket_start),
        value: Number(row.average_value || row.last_value || 0),
      }))
  ), [data.metricSeries]);

  const traceChart = useMemo(() => (
    [...data.traceLatency]
      .sort((left, right) => new Date(left.bucket_start) - new Date(right.bucket_start))
      .slice(-10)
      .map((row) => ({
        label: compactTimestamp(row.bucket_start),
        value: Number(row.average_duration_ms || 0),
      }))
  ), [data.traceLatency]);

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
    const widgets = [];
    for (const dashboard of dashboards.slice(0, 2)) {
      const parsed = parseJSONSafe(dashboard.widgets, []);
      for (const widget of parsed) {
        widgets.push({ dashboard: dashboard.name, ...widget });
      }
    }
    return widgets;
  }, [data.dashboards]);

  const selectedDashboard = useMemo(
    () => data.dashboards.find((row) => row.id === dashboardBuilder.dashboard_id) || null,
    [data.dashboards, dashboardBuilder.dashboard_id],
  );
  const selectedDashboardWidgets = useMemo(
    () => parseJSONSafe(selectedDashboard?.widgets, []),
    [selectedDashboard],
  );
  const widgetData = useMemo(() => ({
    ...data,
    logs: data.widgetLogs,
    metrics: data.widgetMetrics,
    traces: data.widgetTraces,
    health: data.widgetHealth,
    logSeverity: data.widgetLogSeverity,
    metricSeries: data.widgetMetricSeries,
    traceLatency: data.widgetTraceLatency,
  }), [data]);

  function selectDashboardForEditing(dashboard) {
    setDashboardBuilder((current) => ({
      ...defaultDashboardBuilder(),
      dashboard_id: dashboard.id,
      filter_service_id: current.filter_service_id,
      filter_environment: current.filter_environment,
      filter_severity: current.filter_severity,
      filter_metric_name: current.filter_metric_name,
      filter_search: current.filter_search,
      filter_trace_id: current.filter_trace_id,
    }));
    setDashboardForm({
      name: dashboard.name,
      description: dashboard.description,
      default_time_range: dashboard.default_time_range || "120m",
    });
  }

  function beginEditWidget(widget) {
    setDashboardBuilder((current) => builderFromDashboardWidget(widget, {
      ...current,
      dashboard_id: current.dashboard_id,
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

  function resetWidgetEditor() {
    setDashboardBuilder((current) => ({
      ...defaultDashboardBuilder(),
      dashboard_id: current.dashboard_id,
    }));
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
            Trace ID
            <input data-testid="filter-trace" value={filterForm.trace_id} onChange={(event) => setFilterForm({ ...filterForm, trace_id: event.target.value })} />
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
            <button data-testid="create-rule" type="submit">Create Rule</button>
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

      <IncidentWorkspaceSection
        data={data}
        incidentFilters={incidentFilters}
        setIncidentFilters={setIncidentFilters}
        assignForm={assignForm}
        setAssignForm={setAssignForm}
        commentForm={commentForm}
        setCommentForm={setCommentForm}
        onApplyFilters={(event) => { event.preventDefault(); loadDashboard(); }}
        onIncidentAction={handleIncidentAction}
        onLoadIncidentDetail={loadIncidentDetail}
        onAssignIncident={handleAssignIncident}
        onAddComment={handleAddComment}
      />

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
          {(channelForm.type === "webhook" || channelForm.type === "slack_webhook") ? (
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

      <Section title="All Notification Deliveries">
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
          rows={data.deliveries}
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

      <DashboardBuilderSection
        data={data}
        widgetData={widgetData}
        dashboardForm={dashboardForm}
        setDashboardForm={setDashboardForm}
        dashboardBuilder={dashboardBuilder}
        setDashboardBuilder={setDashboardBuilder}
        selectedDashboard={selectedDashboard}
        selectedDashboardWidgets={selectedDashboardWidgets}
        onCreateDashboard={handleCreateDashboard}
        onUpdateDashboardDetails={handleUpdateDashboardDetails}
        onSaveDashboardWidget={handleSaveDashboardWidget}
        onSelectDashboard={selectDashboardForEditing}
        onBeginEditWidget={beginEditWidget}
        onDeleteWidget={handleDeleteWidget}
        onMoveWidget={moveWidget}
        onResetWidgetEditor={resetWidgetEditor}
      />

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
