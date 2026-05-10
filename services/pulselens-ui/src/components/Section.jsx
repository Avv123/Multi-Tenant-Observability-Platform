export default function Section({ title, actions, children }) {
  return (
    <section className="panel">
      <div className="panel__header">
        <h2>{title}</h2>
        <div className="panel__actions">{actions}</div>
      </div>
      {children}
    </section>
  );
}
