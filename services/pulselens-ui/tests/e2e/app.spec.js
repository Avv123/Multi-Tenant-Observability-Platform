import { expect, test } from "@playwright/test";

test("bootstrap, ingest, analytics, users, and webhook flow from the UI", async ({ page }) => {
  const suffix = Date.now();
  await page.goto("/");

  await page.getByTestId("bootstrap-submit").click();
  await expect(page.locator(".banner")).toContainText("Tenant, service, API key, and session created.");

  await expect.poll(async () => {
    const state = await page.evaluate(() => JSON.parse(localStorage.getItem("pulselens-state") || "{}"));
    return Boolean(state.tenantId && state.serviceId && state.apiKey && state.token);
  }).toBeTruthy();

  await page.getByTestId("send-log").click();
  await expect(page.locator(".banner")).toContainText("log event submitted.");

  await page.getByTestId("send-metric").click();
  await expect(page.locator(".banner")).toContainText("metric event submitted.");

  await page.getByTestId("send-trace").click();
  await expect(page.locator(".banner")).toContainText("trace event submitted.");
  await expect(page.getByText("Log Severity Rollups")).toBeVisible();
  await expect(page.getByText("Metric Rollups")).toBeVisible();
  await expect(page.getByText("Trace Latency Rollups")).toBeVisible();

  await page.getByTestId("filter-severity").fill("error");
  await page.getByTestId("filter-search").fill("checkout");
  await page.getByTestId("apply-filters").click();
  await expect(page.getByText("Recent Logs")).toBeVisible();

  await page.getByTestId("saved-query-name").fill(`Errors ${suffix}`);
  await page.getByTestId("create-saved-query").click();
  await expect(page.locator(".banner")).toContainText("Saved query created.");
  await expect(page.getByText(`Errors ${suffix}`)).toBeVisible();
  await page.getByTestId("update-first-query").click();
  await expect(page.locator(".banner")).toContainText("Saved query updated from active filters.");

  await page.getByTestId("dashboard-name").fill(`Ops ${suffix}`);
  await page.getByTestId("dashboard-description").fill("UI-created dashboard");
  await page.getByTestId("create-dashboard").click();
  await expect(page.locator(".banner")).toContainText("Dashboard created.");

  await page.getByTestId("dashboard-widget-dashboard").selectOption({ label: `Ops ${suffix}` });
  await page.getByTestId("dashboard-widget-title").fill(`Severity ${suffix}`);
  await page.getByTestId("dashboard-widget-dataset").selectOption("log_severity");
  await page.getByTestId("save-dashboard-widget").click();
  await expect(page.locator(".banner")).toContainText("Dashboard widget saved.");
  await expect(page.getByText(`Severity ${suffix}`)).toBeVisible();
  await page.getByTestId("dashboard-widget-title").fill(`Latency ${suffix}`);
  await page.getByTestId("dashboard-widget-dataset").selectOption("metric_series");
  await page.getByTestId("dashboard-widget-metric").fill("checkout_latency_ms");
  await page.getByTestId("save-dashboard-widget").click();
  await expect(page.locator(".banner")).toContainText("Dashboard widget saved.");
  await expect(page.getByText(`Latency ${suffix}`)).toBeVisible();
  const severityCard = page.locator(".widget-card").filter({ has: page.getByRole("heading", { name: `Severity ${suffix}` }) });
  const latencyCard = page.locator(".widget-card").filter({ has: page.getByRole("heading", { name: `Latency ${suffix}` }) });
  await severityCard.getByRole("button", { name: "Edit" }).click();
  await expect(page.getByTestId("save-dashboard-widget")).toContainText("Update Widget");
  await page.getByTestId("dashboard-widget-title").fill(`Severity Updated ${suffix}`);
  await page.getByTestId("save-dashboard-widget").click();
  await expect(page.locator(".banner")).toContainText("Dashboard widget updated.");
  await expect(page.getByText(`Severity Updated ${suffix}`)).toBeVisible();
  const updatedSeverityCard = page.locator(".widget-card").filter({ has: page.getByRole("heading", { name: `Severity Updated ${suffix}` }) });
  await updatedSeverityCard.getByRole("button", { name: "Down" }).click();
  await expect(page.locator(".banner")).toContainText("Dashboard widget order updated.");
  await latencyCard.getByRole("button", { name: "Delete" }).click();
  await expect(page.locator(".banner")).toContainText("Dashboard widget deleted.");

  await page.getByTestId("policy-name").fill(`Policy ${suffix}`);
  await page.getByTestId("policy-open-channels").fill("webhook,email");
  await page.getByTestId("create-policy").click();
  await expect(page.locator(".banner")).toContainText("Alert policy created.");
  await expect(page.getByRole("cell", { name: `Policy ${suffix}` })).toBeVisible();

  await page.getByTestId("create-rule").click();
  await expect(page.locator(".banner")).toContainText("Alert rule created.");
  await page.getByTestId("send-log").click();
  await expect(page.locator(".banner")).toContainText("log event submitted.");

  await page.getByTestId("channel-name").fill(`UI Webhook ${suffix}`);
  await page.getByTestId("channel-type").selectOption("webhook");
  await page.getByTestId("channel-webhook-url").fill(`http://127.0.0.1:9099/webhooks/${suffix}`);
  await page.getByTestId("create-channel").click();

  await expect(page.locator(".banner")).toContainText("Notification channel created.");
  await page.getByTestId("user-name").fill(`Viewer ${suffix}`);
  await page.getByTestId("user-email").fill(`viewer-${suffix}@pulselens.local`);
  await page.getByTestId("user-password").fill("viewer-pass");
  await page.getByTestId("user-role").selectOption("viewer");
  await page.getByTestId("create-user").click();
  await expect(page.locator(".banner")).toContainText("User created.");
  await expect(page.getByRole("heading", { name: "Incidents" })).toBeVisible();
  await expect(page.getByTestId("apply-incident-filters")).toBeVisible();
  await expect(page.getByText(`viewer-${suffix}@pulselens.local`)).toBeVisible();
  await expect(page.getByText(`UI Webhook ${suffix}`)).toBeVisible();
  await expect(page.getByRole("cell", { name: "webhook", exact: true })).toBeVisible();
  await expect(page.getByText("PulseLens Local Console")).toBeVisible();
});
