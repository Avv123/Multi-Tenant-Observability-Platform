// ArchivePage — archive stats and replay job management.
import { useState } from "react";
import { EmptyState, ErrorMessage, SectionLoader, useAsyncData } from "../lib/hooks";
import { archiveApi } from "../lib/api";

export default function ArchivePage({ state, notify }) {
  const token = state.token;

  const { data: stats, loading: statsLoading, error: statsError, refetch: refetchStats } =
    useAsyncData(() => archiveApi.stats(token), [token], { skip: !token });

  const { data: jobs, loading: jobsLoading, refetch: refetchJobs } =
    useAsyncData(() => archiveApi.listReplayJobs(token), [token], { skip: !token });

  const [replayForm, setReplayForm] = useState({ event_type: "log", window_minutes: 30 });

  async function handleCreateReplay(e) {
    e.preventDefault();
    try {
      const endTime   = new Date();
      const startTime = new Date(Date.now() - Number(replayForm.window_minutes) * 60 * 1000);
      await archiveApi.createReplayJob(token, {
        service_id: state.serviceId,
        event_type: replayForm.event_type,
        start_time: startTime.toISOString(),
        end_time:   endTime.toISOString(),
      });
      notify("Replay job created.", "success");
      refetchJobs();
    } catch (err) {
      notify(err.message, "error");
    }
  }

  return (
    <div style={{ display: "flex", flexDirection: "column", gap: "1rem" }}>
      {statsError && <ErrorMessage message={statsError} />}

      {/* Archive stats */}
      <div className="panel">
        <div className="panel__header">
          <div className="panel__title">Archive Statistics</div>
          <button id="btn-refresh-archive" className="btn btn-ghost btn-sm" onClick={refetchStats}>↺</button>
        </div>
        {statsLoading ? <SectionLoader /> : (
          stats ? (
            <div className="stat-grid">
              {Object.entries(stats).map(([k, v]) => (
                <div key={k} className="stat-card">
                  <span className="stat-card__label">{k.replace(/_/g, " ")}</span>
                  <div className="stat-card__value">{String(v)}</div>
                </div>
              ))}
            </div>
          ) : (
            <EmptyState icon="⏮" title="No archive stats" body="Archive data is populated as the processing service writes to MinIO." />
          )
        )}
      </div>

      {/* Replay jobs */}
      <div className="panel">
        <div className="panel__header">
          <div className="panel__title">Replay Jobs</div>
          <button id="btn-refresh-replay-jobs" className="btn btn-ghost btn-sm" onClick={refetchJobs}>↺</button>
        </div>
        {jobsLoading ? <SectionLoader /> : (
          jobs?.length ? (
            <div className="table-wrap">
              <table className="data-table" id="table-replay-jobs">
                <thead><tr><th>Created</th><th>Event Type</th><th>Window</th><th>Status</th></tr></thead>
                <tbody>
                  {jobs.map((job) => (
                    <tr key={job.id}>
                      <td className="text-muted text-sm">{new Date(job.created_at).toLocaleString()}</td>
                      <td><span className="badge badge--info">{job.event_type}</span></td>
                      <td className="text-xs font-mono">{job.start_time?.slice(0,16)} → {job.end_time?.slice(0,16)}</td>
                      <td><span className={`badge badge--${job.status === "completed" ? "success" : job.status === "failed" ? "danger" : "warning"}`}>{job.status}</span></td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          ) : (
            <EmptyState icon="⏮" title="No replay jobs" body="Create a replay job to re-process archived events from MinIO." />
          )
        )}
      </div>

      {/* Create replay job */}
      <div className="panel">
        <div className="panel__title" style={{ marginBottom: "0.75rem" }}>Create Replay Job</div>
        <form id="form-create-replay" onSubmit={handleCreateReplay}>
          <div className="form-grid">
            <div className="form-field">
              <label>Event Type</label>
              <select id="replay-event-type" value={replayForm.event_type}
                onChange={(e) => setReplayForm((f) => ({ ...f, event_type: e.target.value }))}>
                {["log","metric","trace","custom"].map((t) => <option key={t} value={t}>{t}</option>)}
              </select>
            </div>
            <div className="form-field">
              <label>Window (minutes)</label>
              <input id="replay-window" type="number" value={replayForm.window_minutes}
                onChange={(e) => setReplayForm((f) => ({ ...f, window_minutes: e.target.value }))} />
            </div>
          </div>
          <button id="btn-create-replay" type="submit" className="btn btn-primary" style={{ marginTop: "0.75rem" }}>
            Create Replay Job
          </button>
        </form>
      </div>
    </div>
  );
}
