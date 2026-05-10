import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import App from "./App";

describe("App", () => {
  it("renders the console shell", () => {
    render(<App />);
    expect(screen.getByText("PulseLens Local Console")).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Bootstrap Workspace" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Login" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Dashboard" })).toBeInTheDocument();
  });
});
