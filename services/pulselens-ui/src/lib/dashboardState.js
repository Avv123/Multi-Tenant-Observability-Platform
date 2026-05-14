export function defaultDashboardForm() {
  return {
    name: "Operations Overview",
    description: "Starter dashboard for PulseLens",
    default_time_range: "120m",
  };
}

export function defaultDashboardBuilder() {
  return {
    dashboard_id: "",
    widget_id: "",
    title: "Error Trend",
    type: "chart",
    dataset: "log_severity",
    chart_type: "bar",
    metric: "",
    layout_w: "1",
    layout_h: "1",
    filter_service_id: "",
    filter_environment: "",
    filter_severity: "error",
    filter_metric_name: "",
    filter_search: "checkout",
    filter_trace_id: "",
  };
}

export function widgetFormDefaults(type) {
  if (type === "stat") {
    return {
      dataset: "service_health",
      chart_type: "bar",
      metric: "log_count",
    };
  }
  if (type === "table") {
    return {
      dataset: "logs",
      chart_type: "bar",
      metric: "",
    };
  }
  return {
    dataset: "log_severity",
    chart_type: "bar",
    metric: "",
  };
}

export function normalizeWidgetFilters(dataset, filters = {}) {
  const normalized = {
    service_id: filters.service_id || "",
    environment: filters.environment || "",
  };

  switch (dataset) {
    case "logs":
    case "log_severity":
      normalized.severity = filters.severity || "";
      normalized.search = filters.search || "";
      normalized.trace_id = filters.trace_id || "";
      break;
    case "metrics":
    case "metric_series":
      normalized.metric_name = filters.metric_name || "";
      break;
    case "traces":
    case "trace_latency":
      normalized.trace_id = filters.trace_id || "";
      break;
    default:
      break;
  }

  return normalized;
}

export function builderFromDashboardWidget(widget, current = defaultDashboardBuilder()) {
  const filters = normalizeWidgetFilters(widget.dataset, widget.filters || {});
  return {
    ...current,
    widget_id: widget.id || "",
    title: widget.title || "",
    type: widget.type || "chart",
    dataset: widget.dataset || "log_severity",
    chart_type: widget.chart_type || "bar",
    metric: widget.metric || "",
    layout_w: String(widget.layout?.w || 1),
    layout_h: String(widget.layout?.h || 1),
    filter_service_id: filters.service_id || "",
    filter_environment: filters.environment || "",
    filter_severity: filters.severity || "",
    filter_metric_name: filters.metric_name || "",
    filter_search: filters.search || "",
    filter_trace_id: filters.trace_id || "",
  };
}

export function builderToWidgetPayload(builder) {
  const filters = normalizeWidgetFilters(builder.dataset, {
    service_id: builder.filter_service_id,
    environment: builder.filter_environment,
    severity: builder.filter_severity,
    metric_name: builder.filter_metric_name,
    search: builder.filter_search,
    trace_id: builder.filter_trace_id,
  });
  return {
    title: builder.title,
    type: builder.type,
    dataset: builder.dataset,
    chart_type: builder.chart_type,
    metric: builder.metric,
    filters,
    layout: {
      w: Number(builder.layout_w || 1),
      h: Number(builder.layout_h || 1),
    },
  };
}
