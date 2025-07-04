import { Box, Tabs, Tab } from "@mui/material";
import { useState } from "react";

function formatLabel(label: string): string {
  return label
    .replace(/_/g, " ")
    .replace(/\b\w/g, (c) => c.toUpperCase());
}

interface JsonTabsProps {
  items: any[];
}

export default function JsonTabs({ items }: JsonTabsProps) {
  const [index, setIndex] = useState(0);

  return (
    <Box>
      <Tabs
        value={index}
        onChange={(_, v) => setIndex(v)}
        variant="scrollable"
        sx={{ mb: 1 }}
      >
        {items.map((_, idx) => (
          <Tab key={idx} label={`Item ${idx + 1}`} />
        ))}
      </Tabs>
      {items.map((item, idx) => (
        <Box key={idx} hidden={index !== idx}>
          {typeof item === "object" && item !== null && !Array.isArray(item) ? (
            <table style={{ width: "100%", borderCollapse: "collapse" }}>
              <tbody>
                {Object.entries(item).map(([k, v]) => (
                  <tr key={k}>
                    <td
                      style={{
                        fontWeight: "bold",
                        verticalAlign: "top",
                        paddingRight: "0.5rem",
                      }}
                    >
                      {formatLabel(k)}
                    </td>
                    <td>
                      {typeof v === "object" && v !== null ? (
                        <pre style={{ margin: 0, whiteSpace: "pre-wrap" }}>
                          {JSON.stringify(v, null, 2)}
                        </pre>
                      ) : (
                        String(v)
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          ) : (
            <pre style={{ margin: 0, whiteSpace: "pre-wrap" }}>
              {JSON.stringify(item, null, 2)}
            </pre>
          )}
        </Box>
      ))}
    </Box>
  );
}
