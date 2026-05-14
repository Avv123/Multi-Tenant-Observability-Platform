import DataTable from "../DataTable";
import Section from "../Section";
import { normalizeWidgetFilters, widgetFormDefaults } from "../../lib/dashboardState";
import WidgetRenderer from "./WidgetRenderer";
import { parseJSONSafe } from "../../lib/widgets";

function widgetFiltersFromBuilder(builder) {
  return normalizeWidgetFilters(builder.dataset, {
    service_id: builder.filter_service_id || "",
    environment: builder.filter_environment || "",
    severity: builder.filter_severity || "",
    metric_name: builder.filter_metric_name || "",
    search: builder.filter_search || "",
    trace_id: builder.filter_trace_id || "",
  });
}

function buildDraftWidget(builder) {
  return {
    id: builder.widget_id || "draft-widget",
    title: builder.title,
    type: builder.type,
    dataset: builder.dataset,
    chart_type: builder.chart_type,
    metric: builder.metric,
    filters: widgetFiltersFromBuilder(builder),
    layout: {
      w: Number(builder.layout_w || 1),
      h: Number(builder.layout_h || 1),
    },
  };
}

export default function DashboardBuilderSection({
  data,
  widgetData,
  dashboardForm,
  setDashboardForm,
  dashboardBuilder,
  setDashboardBuilder,
  selectedDashboard,
  selectedDashboardWidgets,
  onCreateDashboard,
  onUpdateDashboardDetails,
  onSaveDashboardWidget,
  onSelectDashboard,
  onBeginEditWidget,
  onDeleteWidget,
  onMoveWidget,
  onResetWidgetEditor,
}) {
  const draftWidget = buildDraftWidget(dashboardBuilder);

  return (
    <>
      <Section title="Dashboards">
        <form className="form-grid" onSubmit={onCreateDashboard}>
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
            <select data-testid="dashboard-default-range" value={dashboardForm.default_time_range} onChange={(event) => setDashboardForm({ ...dashboardForm, default_time_range: event.target.value })}>
              <option value="30m">30m</option>
              <option value="120m">120m</option>
              <option value="24h">24h</option>
            </select>
          </label>
          <div className="form-actions">
            <button data-testid="create-dashboard" type="submit">Create Dashboard</button>
          </div>
          <div className="form-actions">
            <button data-testid="update-dashboard" type="button" onClick={onUpdateDashboardDetails}>Update Selected Dashboard</button>
          </div>
        </form>
        <DataTableDashboards dashboards={data.dashboards} onSelectDashboard={onSelectDashboard} />
      </Section>

      <Section title="Dashboard Builder">
        <div className="builder-shell">
          <form className="form-grid" onSubmit={onSaveDashboardWidget}>
            <label>
              Dashboard
              <select
                data-testid="dashboard-widget-dashboard"
                value={dashboardBuilder.dashboard_id}
                onChange={(event) => {
                  const nextDashboardID = event.target.value;
                  const selected = data.dashboards.find((dashboard) => dashboard.id === nextDashboardID);
                  if (selected) {
                    onSelectDashboard(selected);
                    return;
                  }
                  setDashboardBuilder({ ...dashboardBuilder, dashboard_id: "", widget_id: "" });
                }}
              >
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
              Widget Type
              <select
                data-testid="dashboard-widget-type"
                value={dashboardBuilder.type}
                onChange={(event) => {
                  const type = event.target.value;
                  const defaults = widgetFormDefaults(type);
                  setDashboardBuilder({
                    ...dashboardBuilder,
                    type,
                    dataset: defaults.dataset,
                    chart_type: defaults.chart_type,
                    metric: defaults.metric,
                  });
                }}
              >
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
                <option value="metrics">metrics</option>
                <option value="traces">traces</option>
              </select>
            </label>
            <label>
              Chart Type
              <select data-testid="dashboard-widget-chart-type" value={dashboardBuilder.chart_type} onChange={(event) => setDashboardBuilder({ ...dashboardBuilder, chart_type: event.target.value })}>
                <option value="bar">bar</option>
                <option value="line">line</option>
              </select>
            </label>
            <label>
              Metric Key
              <input data-testid="dashboard-widget-metric" value={dashboardBuilder.metric} onChange={(event) => setDashboardBuilder({ ...dashboardBuilder, metric: event.target.value })} />
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
              Filter Service
              <select value={dashboardBuilder.filter_service_id} onChange={(event) => setDashboardBuilder({ ...dashboardBuilder, filter_service_id: event.target.value })}>
                <option value="">all</option>
                {data.services.map((service) => (
                  <option key={service.id} value={service.id}>{service.name}</option>
                ))}
              </select>
            </label>
            <label>
              Filter Environment
              <input value={dashboardBuilder.filter_environment} onChange={(event) => setDashboardBuilder({ ...dashboardBuilder, filter_environment: event.target.value })} />
            </label>
            <label>
              Filter Severity
              <input value={dashboardBuilder.filter_severity} onChange={(event) => setDashboardBuilder({ ...dashboardBuilder, filter_severity: event.target.value })} />
            </label>
            <label>
              Filter Metric
              <input value={dashboardBuilder.filter_metric_name} onChange={(event) => setDashboardBuilder({ ...dashboardBuilder, filter_metric_name: event.target.value })} />
            </label>
            <label>
              Filter Search
              <input value={dashboardBuilder.filter_search} onChange={(event) => setDashboardBuilder({ ...dashboardBuilder, filter_search: event.target.value })} />
            </label>
            <label>
              Filter Trace
              <input value={dashboardBuilder.filter_trace_id} onChange={(event) => setDashboardBuilder({ ...dashboardBuilder, filter_trace_id: event.target.value })} />
            </label>
            <div className="form-actions">
              <button data-testid="save-dashboard-widget" type="submit">{dashboardBuilder.widget_id ? "Update Widget" : "Add Widget To Dashboard"}</button>
            </div>
            <div className="form-actions">
              <button data-testid="reset-dashboard-widget" type="button" onClick={onResetWidgetEditor}>Reset Widget Editor</button>
            </div>
          </form>

          <div className="widget-card">
            <h3>Draft Preview</h3>
            <WidgetRenderer widget={draftWidget} data={widgetData} overview={data.overview} />
          </div>
        </div>
      </Section>

      <Section title="Dashboard Detail">
        {selectedDashboard ? (
          <>
            <div className="key-value-grid">
              <span>Dashboard</span><code>{selectedDashboard.name}</code>
              <span>Description</span><code>{selectedDashboard.description || "-"}</code>
              <span>Default Range</span><code>{selectedDashboard.default_time_range || "120m"}</code>
              <span>Layout</span><code>{JSON.stringify(parseJSONSafe(selectedDashboard.layout, { columns: 2 }))}</code>
            </div>
            <div className="widget-grid">
              {selectedDashboardWidgets.map((widget) => (
                <div className="widget-card" key={widget.id}>
                  <div className="widget-card__header">
                    <h3>{widget.title || widget.dataset || widget.type}</h3>
                    <div className="button-row">
                      <button type="button" data-testid={`edit-widget-${widget.id}`} onClick={() => onBeginEditWidget(widget)}>Edit</button>
                      <button type="button" data-testid={`move-widget-up-${widget.id}`} onClick={() => onMoveWidget(widget.id, "up")}>Up</button>
                      <button type="button" data-testid={`move-widget-down-${widget.id}`} onClick={() => onMoveWidget(widget.id, "down")}>Down</button>
                      <button type="button" data-testid={`delete-widget-${widget.id}`} onClick={() => onDeleteWidget(widget.id)}>Delete</button>
                    </div>
                  </div>
                  <WidgetRenderer widget={widget} data={widgetData} overview={data.overview} />
                </div>
              ))}
            </div>
          </>
        ) : (
          <p>Select a dashboard to edit widgets and review the saved layout.</p>
        )}
      </Section>
    </>
  );
}

function DataTableDashboards({ dashboards, onSelectDashboard }) {
  return (
    <DataTable
      columns={[
        { key: "name", label: "Name" },
        { key: "description", label: "Description" },
        { key: "default_time_range", label: "Time Range" },
        { key: "created_by", label: "Created By" },
        { key: "edit", label: "Edit", render: (row) => <button type="button" data-testid={`select-dashboard-${row.id}`} onClick={() => onSelectDashboard(row)}>Select</button> },
      ]}
      rows={dashboards}
    />
  );
}
