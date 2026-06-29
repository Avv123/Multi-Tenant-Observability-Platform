const BASE_URLS = {
  tenant: "http://localhost:8081",
  ingest: "http://localhost:8082",
  query: "http://localhost:8084",
  alerting: "http://localhost:8085",
};

function buildQueryString(filters = {}) {
  const query = new URLSearchParams();
  Object.entries(filters).forEach(([key, value]) => {
    if (value === undefined || value === null || value === "") {
      return;
    }
    query.set(key, String(value));
  });
  const stringified = query.toString();
  return stringified ? `?${stringified}` : "";
}

async function request(base, path, options = {}) {
  const response = await fetch(`${BASE_URLS[base]}${path}`, {
    headers: {
      "Content-Type": "application/json",
      ...(options.headers || {}),
    },
    ...options,
  });

  const raw = await response.text();
  let payload;
  try {
    payload = JSON.parse(raw);
  } catch (error) {
    throw new Error(`${base}${path} returned invalid JSON: ${raw.slice(0, 160)}`);
  }
  if (!response.ok || payload?.is_success === false) {
    throw new Error(payload?.message || `request failed: ${response.status}`);
  }

  return payload.data;
}

export const tenantApi = {
  createTenant(body) {
    return request("tenant", "/internal/api/v1/tenants", {
      method: "POST",
      headers: { "X-Internal-Token": "pulselens-internal-token" },
      body: JSON.stringify(body),
    });
  },
  createService(tenantId, body, token) {
    return request("tenant", `/admin/api/v1/tenants/${tenantId}/services`, {
      method: "POST",
      headers: { Authorization: `Bearer ${token}` },
      body: JSON.stringify(body),
    });
  },
  createAPIKey(body, token) {
    return request("tenant", "/admin/api/v1/api-keys", {
      method: "POST",
      headers: { Authorization: `Bearer ${token}` },
      body: JSON.stringify(body),
    });
  },
  listAPIKeys(token) {
    return request("tenant", "/admin/api/v1/api-keys", {
      headers: { Authorization: `Bearer ${token}` },
    });
  },
  rotateAPIKey(keyId, body, token) {
    return request("tenant", `/admin/api/v1/api-keys/${keyId}/rotate`, {
      method: "POST",
      headers: { Authorization: `Bearer ${token}` },
      body: JSON.stringify(body || {}),
    });
  },
  revokeAPIKey(keyId, token) {
    return request("tenant", `/admin/api/v1/api-keys/${keyId}/revoke`, {
      method: "POST",
      headers: { Authorization: `Bearer ${token}` },
      body: JSON.stringify({}),
    });
  },
  login(body) {
    return request("tenant", "/api/v1/auth/login", {
      method: "POST",
      body: JSON.stringify(body),
    });
  },
  me(token) {
    return request("tenant", "/api/v1/auth/me", {
      headers: { Authorization: `Bearer ${token}` },
    });
  },
  listAuditLogs(tenantId, token) {
    return request("tenant", `/admin/api/v1/tenants/${tenantId}/audit-logs`, {
      headers: { Authorization: `Bearer ${token}` },
    });
  },
  listServices(tenantId, token) {
    return request("tenant", `/admin/api/v1/tenants/${tenantId}/services`, {
      headers: { Authorization: `Bearer ${token}` },
    });
  },
  listUsers(tenantId, token) {
    return request("tenant", `/admin/api/v1/tenants/${tenantId}/users`, {
      headers: { Authorization: `Bearer ${token}` },
    });
  },
  createUser(tenantId, body, token) {
    return request("tenant", `/admin/api/v1/tenants/${tenantId}/users`, {
      method: "POST",
      headers: { Authorization: `Bearer ${token}` },
      body: JSON.stringify(body),
    });
  },
};

export const ingestApi = {
  ingest(apiKey, events) {
    return request("ingest", "/api/v1/ingest", {
      method: "POST",
      headers: { "X-API-Key": apiKey },
      body: JSON.stringify({ events }),
    });
  },
};

export const queryApi = {
  overview(token) {
    return request("query", "/api/v1/overview", {
      headers: { Authorization: `Bearer ${token}` },
    });
  },
  logs(token, query = "") {
    return request("query", `/api/v1/logs${query}`, {
      headers: { Authorization: `Bearer ${token}` },
    });
  },
  logsWithFilters(token, filters = {}) {
    return request("query", `/api/v1/logs${buildQueryString(filters)}`, {
      headers: { Authorization: `Bearer ${token}` },
    });
  },
  metricsWithFilters(token, filters = {}) {
    return request("query", `/api/v1/metrics${buildQueryString(filters)}`, {
      headers: { Authorization: `Bearer ${token}` },
    });
  },
  tracesWithFilters(token, filters = {}) {
    return request("query", `/api/v1/traces${buildQueryString(filters)}`, {
      headers: { Authorization: `Bearer ${token}` },
    });
  },
  traceDetail(token, traceId) {
    return request("query", `/api/v1/traces/${traceId}`, {
      headers: { Authorization: `Bearer ${token}` },
    });
  },
  transactions(token, filters = {}) {
    return request("query", `/api/v1/transactions${buildQueryString(filters)}`, {
      headers: { Authorization: `Bearer ${token}` },
    });
  },
  errorGroups(token, filters = {}) {
    return request("query", `/api/v1/errors/groups${buildQueryString(filters)}`, {
      headers: { Authorization: `Bearer ${token}` },
    });
  },
  correlatedLogs(token, traceId, lookback = 120) {
    return request("query", `/api/v1/logs${buildQueryString({ trace_id: traceId, lookback_minutes: lookback, limit: 50 })}`, {
      headers: { Authorization: `Bearer ${token}` },
    });
  },
  serviceMap(token, lookbackMinutes = 60) {
    return request("query", `/api/v1/service-map${buildQueryString({ lookback_minutes: lookbackMinutes })}`, {
      headers: { Authorization: `Bearer ${token}` },
    });
  },
  serviceHealth(token) {
    return request("query", "/api/v1/services/health?limit=20", {
      headers: { Authorization: `Bearer ${token}` },
    });
  },
  listSavedQueries(token) {
    return request("query", "/api/v1/saved-queries", {
      headers: { Authorization: `Bearer ${token}` },
    });
  },
  createSavedQuery(token, body) {
    return request("query", "/api/v1/saved-queries", {
      method: "POST",
      headers: { Authorization: `Bearer ${token}` },
      body: JSON.stringify(body),
    });
  },
  updateSavedQuery(token, queryId, body) {
    return request("query", `/api/v1/saved-queries/${queryId}`, {
      method: "PATCH",
      headers: { Authorization: `Bearer ${token}` },
      body: JSON.stringify(body),
    });
  },
  listDashboards(token) {
    return request("query", "/api/v1/dashboards", {
      headers: { Authorization: `Bearer ${token}` },
    });
  },
  createDashboard(token, body) {
    return request("query", "/api/v1/dashboards", {
      method: "POST",
      headers: { Authorization: `Bearer ${token}` },
      body: JSON.stringify(body),
    });
  },
  updateDashboard(token, dashboardId, body) {
    return request("query", `/api/v1/dashboards/${dashboardId}`, {
      method: "PUT",
      headers: { Authorization: `Bearer ${token}` },
      body: JSON.stringify(body),
    });
  },
  updateDashboardWidget(token, dashboardId, widgetId, body) {
    return request("query", `/api/v1/dashboards/${dashboardId}/widgets/${widgetId}`, {
      method: "PATCH",
      headers: { Authorization: `Bearer ${token}` },
      body: JSON.stringify(body),
    });
  },
  deleteDashboardWidget(token, dashboardId, widgetId) {
    return request("query", `/api/v1/dashboards/${dashboardId}/widgets/${widgetId}`, {
      method: "DELETE",
      headers: { Authorization: `Bearer ${token}` },
    });
  },
  platformOverview(token) {
    return request("query", "/api/v1/platform/overview?limit=10", {
      headers: { Authorization: `Bearer ${token}` },
    });
  },
  platformRuntime(token) {
    return request("query", "/api/v1/platform/runtime", {
      headers: { Authorization: `Bearer ${token}` },
    });
  },
  platformBackpressure(token) {
    return request("query", "/api/v1/platform/backpressure", {
      headers: { Authorization: `Bearer ${token}` },
    });
  },
  cleanupRuns(token) {
    return request("query", "/api/v1/platform/cleanup-runs?limit=10", {
      headers: { Authorization: `Bearer ${token}` },
    });
  },
  dependencies(token) {
    return request("query", "/api/v1/platform/dependencies", {
      headers: { Authorization: `Bearer ${token}` },
    });
  },
  kafkaLag(token) {
    return request("query", "/api/v1/platform/kafka-lag", {
      headers: { Authorization: `Bearer ${token}` },
    });
  },
  logSeverity(token) {
    return request("query", "/api/v1/analytics/log-severity?limit=20", {
      headers: { Authorization: `Bearer ${token}` },
    });
  },
  logSeverityWithFilters(token, filters = {}) {
    const merged = { limit: 20, ...filters };
    return request("query", `/api/v1/analytics/log-severity${buildQueryString(merged)}`, {
      headers: { Authorization: `Bearer ${token}` },
    });
  },
  metricSeries(token, metricName = "checkout_latency_ms") {
    return request("query", `/api/v1/analytics/metric-series?limit=20&metric_name=${encodeURIComponent(metricName)}`, {
      headers: { Authorization: `Bearer ${token}` },
    });
  },
  metricSeriesWithFilters(token, filters = {}) {
    const merged = { limit: 20, ...filters };
    return request("query", `/api/v1/analytics/metric-series${buildQueryString(merged)}`, {
      headers: { Authorization: `Bearer ${token}` },
    });
  },
  traceLatency(token) {
    return request("query", "/api/v1/analytics/trace-latency?limit=20", {
      headers: { Authorization: `Bearer ${token}` },
    });
  },
  traceLatencyWithFilters(token, filters = {}) {
    const merged = { limit: 20, ...filters };
    return request("query", `/api/v1/analytics/trace-latency${buildQueryString(merged)}`, {
      headers: { Authorization: `Bearer ${token}` },
    });
  },
};

export const alertingApi = {
  listRules(token) {
    return request("alerting", "/api/v1/alert-rules", {
      headers: { Authorization: `Bearer ${token}` },
    });
  },
  createRule(token, body) {
    return request("alerting", "/api/v1/alert-rules", {
      method: "POST",
      headers: { Authorization: `Bearer ${token}` },
      body: JSON.stringify(body),
    });
  },
  listPolicies(token) {
    return request("alerting", "/api/v1/alert-policies", {
      headers: { Authorization: `Bearer ${token}` },
    });
  },
  createPolicy(token, body) {
    return request("alerting", "/api/v1/alert-policies", {
      method: "POST",
      headers: { Authorization: `Bearer ${token}` },
      body: JSON.stringify(body),
    });
  },
  updatePolicy(token, policyId, body) {
    return request("alerting", `/api/v1/alert-policies/${policyId}`, {
      method: "PATCH",
      headers: { Authorization: `Bearer ${token}` },
      body: JSON.stringify(body),
    });
  },
  listIncidents(token, filters = {}) {
    return request("alerting", `/api/v1/incidents${buildQueryString(filters)}`, {
      headers: { Authorization: `Bearer ${token}` },
    });
  },
  getIncident(token, incidentId) {
    return request("alerting", `/api/v1/incidents/${incidentId}`, {
      headers: { Authorization: `Bearer ${token}` },
    });
  },
  incidentTimeline(token, incidentId) {
    return request("alerting", `/api/v1/incidents/${incidentId}/timeline`, {
      headers: { Authorization: `Bearer ${token}` },
    });
  },
  incidentDeliveries(token, incidentId) {
    return request("alerting", `/api/v1/incidents/${incidentId}/deliveries`, {
      headers: { Authorization: `Bearer ${token}` },
    });
  },
  acknowledgeIncident(token, incidentId) {
    return request("alerting", `/api/v1/incidents/${incidentId}/acknowledge`, {
      method: "POST",
      headers: { Authorization: `Bearer ${token}` },
    });
  },
  resolveIncident(token, incidentId) {
    return request("alerting", `/api/v1/incidents/${incidentId}/resolve`, {
      method: "POST",
      headers: { Authorization: `Bearer ${token}` },
    });
  },
  assignIncident(token, incidentId, body) {
    return request("alerting", `/api/v1/incidents/${incidentId}/assign`, {
      method: "POST",
      headers: { Authorization: `Bearer ${token}` },
      body: JSON.stringify(body),
    });
  },
  listIncidentComments(token, incidentId) {
    return request("alerting", `/api/v1/incidents/${incidentId}/comments`, {
      headers: { Authorization: `Bearer ${token}` },
    });
  },
  addIncidentComment(token, incidentId, body) {
    return request("alerting", `/api/v1/incidents/${incidentId}/comments`, {
      method: "POST",
      headers: { Authorization: `Bearer ${token}` },
      body: JSON.stringify(body),
    });
  },
  listNotificationChannels(token) {
    return request("alerting", "/api/v1/notification-channels", {
      headers: { Authorization: `Bearer ${token}` },
    });
  },
  createNotificationChannel(token, body) {
    return request("alerting", "/api/v1/notification-channels", {
      method: "POST",
      headers: { Authorization: `Bearer ${token}` },
      body: JSON.stringify(body),
    });
  },
  listNotificationDeliveries(token) {
    return request("alerting", "/api/v1/notification-deliveries", {
      headers: { Authorization: `Bearer ${token}` },
    });
  },
};

