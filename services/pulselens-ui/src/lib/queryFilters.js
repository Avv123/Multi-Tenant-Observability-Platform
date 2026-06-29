function withTimeBounds(filterForm, payload) {
  if (filterForm.lookback_minutes) {
    payload.start_time = new Date(Date.now() - Number(filterForm.lookback_minutes) * 60 * 1000).toISOString();
  }
  return payload;
}

function withLimit(payload, limit) {
  return {
    limit,
    ...payload,
  };
}

export function buildLogFilters(filterForm) {
  return withTimeBounds(filterForm, withLimit({
    service_name: filterForm.service_name || filterForm.service_id || undefined,
    environment: filterForm.environment || undefined,
    severity: filterForm.severity || undefined,
    search: filterForm.search || undefined,
    trace_id: filterForm.trace_id || undefined,
  }, 20));
}

export function buildMetricFilters(filterForm) {
  return withTimeBounds(filterForm, withLimit({
    service_name: filterForm.service_name || filterForm.service_id || undefined,
    environment: filterForm.environment || undefined,
    metric_name: filterForm.metric_name || undefined,
  }, 20));
}

export function buildTraceFilters(filterForm) {
  return withTimeBounds(filterForm, withLimit({
    service_name: filterForm.service_name || filterForm.service_id || undefined,
    environment: filterForm.environment || undefined,
    trace_id: filterForm.trace_id || undefined,
  }, 20));
}

export function buildAnalyticsFilters(filterForm, datasetType) {
  const base = withLimit({
    service_id: filterForm.service_id || undefined,
    environment: filterForm.environment || undefined,
  }, 20);
  if (datasetType === "log_severity" && filterForm.severity) {
    base.severity = filterForm.severity;
  }
  if (datasetType === "metric_series" && filterForm.metric_name) {
    base.metric_name = filterForm.metric_name;
  }
  return withTimeBounds(filterForm, base);
}

export function buildSavedQueryDefinition(filterForm) {
  return {
    service_id: filterForm.service_id || undefined,
    environment: filterForm.environment || undefined,
    severity: filterForm.severity || undefined,
    metric_name: filterForm.metric_name || undefined,
    search: filterForm.search || undefined,
    trace_id: filterForm.trace_id || undefined,
    start_time: filterForm.lookback_minutes
      ? new Date(Date.now() - Number(filterForm.lookback_minutes) * 60 * 1000).toISOString()
      : undefined,
    limit: 20,
  };
}

export function buildDashboardPreviewFilters(filterForm) {
  return withTimeBounds(filterForm, withLimit({}, 100));
}

export function buildDashboardPreviewAnalyticsFilters(filterForm) {
  return withTimeBounds(filterForm, withLimit({}, 100));
}
