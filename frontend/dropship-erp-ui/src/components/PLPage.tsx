import { useState } from "react";
import {
  Button,
  TextField,
  Table,
  TableHead,
  TableRow,
  TableCell,
  TableBody,
} from "@mui/material";
import { fetchPL } from "../api/pl";
import type { Metric } from "../types";

export default function PLPage() {
  const [shop, setShop] = useState("");
  const [period, setPeriod] = useState("");
  const [data, setData] = useState<Metric | null>(null);

  const handleFetch = async () => {
    const res = await fetchPL(shop, period);
    setData(res.data);
  };

  return (
    <div>
      <h2>Profit & Loss</h2>
      <div style={{ display: "flex", gap: "0.5rem" }}>
        <TextField
          label="Shop"
          value={shop}
          onChange={(e) => setShop(e.target.value)}
        />
        <TextField
          label="Period"
          value={period}
          onChange={(e) => setPeriod(e.target.value)}
        />
        <Button onClick={handleFetch}>Fetch</Button>
      </div>
      {data && (
        <Table size="small">
          <TableHead>
            <TableRow>
              <TableCell>Metric</TableCell>
              <TableCell>Value</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {Object.entries(data).map(([k, v]) => (
              <TableRow key={k}>
                <TableCell>{k}</TableCell>
                <TableCell>{v as any}</TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      )}
    </div>
  );
}
