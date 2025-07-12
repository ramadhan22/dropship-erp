import React from "react";

export interface SummaryCardProps {
  label: string;
  value?: number;
  change?: number;
  loading: boolean;
}

export default function SummaryCard({
  label,
  value,
  change,
  loading,
}: SummaryCardProps) {
  return (
    <div
      style={{
        backgroundColor: "#fff",
        borderRadius: "0.75rem",
        boxShadow: "0 1px 3px 0 rgba(0,0,0,0.1), 0 1px 2px 0 rgba(0,0,0,0.06)",
        padding: "1.5rem",
        display: "flex",
        flexDirection: "column",
        minWidth: "335px",
      }}
      aria-label={`${label} card`}
    >
      <div
        style={{
          fontSize: "0.75rem",
          fontWeight: 600,
          color: "#9ca3af",
          textTransform: "uppercase",
          marginBottom: "0.5rem",
        }}
      >
        {label}
      </div>
      {loading || value === undefined || value === null ? (
        <>
          <div
            style={{
              backgroundColor: "#e5e7eb",
              borderRadius: "0.25rem",
              width: "4rem",
              height: "1.5rem",
              animation: "pulse 2s cubic-bezier(0.4,0,0.6,1) infinite",
              marginBottom: "0.5rem",
            }}
          />
          <div
            style={{
              backgroundColor: "#e5e7eb",
              borderRadius: "0.25rem",
              width: "4rem",
              height: "1.5rem",
              animation: "pulse 2s cubic-bezier(0.4,0,0.6,1) infinite",
              marginTop: "0.5rem",
            }}
          />
        </>
      ) : (
        <>
          <div
            style={{
              fontSize: "1.5rem",
              fontWeight: 700,
              color: "#111827",
            }}
          >
            {value.toLocaleString()}
          </div>
          <div
            style={{
              marginTop: "0.5rem",
              display: "flex",
              flexDirection: "row",
              alignItems: "center",
              textAlign: "start",
            }}
          >
            {change && change > 0 && (
              <span style={{ color: "#16a34a", marginRight: "0.25rem" }}>▲</span>
            )}
            {change && change < 0 && (
              <span style={{ color: "#dc2626", marginRight: "0.25rem" }}>▼</span>
            )}
            {typeof change === "number" && change !== 0 ? (
              <span style={{ color: change > 0 ? "#16a34a" : "#dc2626" }}>
                {Math.abs(change * 100).toFixed(1)}%
              </span>
            ) : (
              <span style={{ color: "#9ca3af" }}>—</span>
            )}
          </div>
        </>
      )}
    </div>
  );
}
