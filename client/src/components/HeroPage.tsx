import { useEffect, useState } from "react";
import { getSettings } from "../lib/api";
import type { UserSettings } from "../types";

const FEATURES = [
  { icon: "⎇", label: "Repo comprehension" },
  { icon: "◈", label: "HCL generation" },
  { icon: "✓", label: "Terraform validate" },
  { icon: "⎇", label: "Auto PR" },
];

const INTEGRATIONS: { key: keyof Pick<UserSettings, "github_token_set" | "atlassian_token_set">; label: string }[] = [
  { key: "github_token_set", label: "GitHub" },
  { key: "atlassian_token_set", label: "Jira" },
];

interface Props {
  onNavigate?: (page: string) => void;
}

export function HeroPage({ onNavigate }: Props) {
  const [settings, setSettings] = useState<UserSettings | null>(null);

  useEffect(() => {
    getSettings().then(setSettings).catch(() => {});
  }, []);

  return (
    <div style={{ animation: "slideUp .25s ease" }}>
      {/* Hero */}
      <div style={{ paddingTop: 24, paddingBottom: 48 }}>
        <h1 style={{
          fontSize: 32, fontWeight: 700, color: "var(--text)",
          letterSpacing: "-.03em", lineHeight: 1.2, marginBottom: 14,
        }}>
          Generate Terraform<br />infrastructure, instantly.
        </h1>

        <p style={{
          fontSize: "var(--text-base)", color: "var(--text-3)",
          lineHeight: 1.75, maxWidth: 460, marginBottom: 28,
        }}>
          Describe what you need. tf-agent scans your repos for existing patterns,
          writes HCL, runs{" "}
          <code style={{ fontFamily: "var(--font-mono)", fontSize: 13, color: "var(--text-2)", background: "var(--surface-2)", padding: "1px 5px", borderRadius: 4 }}>
            terraform validate
          </code>
          , and opens a pull request — autonomously.
        </p>

      </div>

      {/* Feature tags */}
      <div style={{
        borderTop: "1px solid var(--border)",
        paddingTop: 28,
        display: "flex", flexWrap: "wrap", gap: 8,
      }}>
        {FEATURES.map((f) => (
          <span key={f.label} style={{
            display: "inline-flex", alignItems: "center", gap: 5,
            fontSize: "var(--text-xs)", fontWeight: 500,
            color: "var(--text-2)", background: "var(--surface)",
            border: "1px solid var(--border)", borderRadius: 20,
            padding: "5px 12px", boxShadow: "0 1px 2px rgba(0,0,0,.04)",
          }}>
            <span style={{ fontSize: 10, color: "var(--accent)" }}>{f.icon}</span>
            {f.label}
          </span>
        ))}
        {settings && INTEGRATIONS.map(({ key, label }) => {
          const configured = settings[key];
          return (
            <span
              key={label}
              title={configured ? `${label} token configured` : `${label} token not set — click to configure`}
              onClick={!configured && onNavigate ? () => onNavigate("settings") : undefined}
              style={{
                display: "inline-flex", alignItems: "center", gap: 6,
                fontSize: "var(--text-xs)", fontWeight: 500,
                color: configured ? "#2d6a4f" : "var(--text-3)",
                background: configured ? "#d8f3dc" : "var(--surface)",
                border: `1px solid ${configured ? "#95d5b2" : "var(--border)"}`,
                borderRadius: 20, padding: "5px 12px",
                boxShadow: "0 1px 2px rgba(0,0,0,.04)",
                cursor: !configured && onNavigate ? "pointer" : "default",
              }}>
              <span style={{
                width: 6, height: 6, borderRadius: "50%",
                background: configured ? "#52b788" : "#ccc",
                flexShrink: 0,
              }} />
              {label}
            </span>
          );
        })}
      </div>
    </div>
  );
}
