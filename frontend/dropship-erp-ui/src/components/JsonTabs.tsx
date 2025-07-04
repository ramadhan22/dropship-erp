import { Box, Tabs, Tab } from "@mui/material";
import { useState } from "react";

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
          <pre style={{ margin: 0, whiteSpace: "pre-wrap" }}>
            {JSON.stringify(item, null, 2)}
          </pre>
        </Box>
      ))}
    </Box>
  );
}
