import { useEffect, useState } from "react";
import {
  Button,
  Table,
  TableHead,
  TableRow,
  TableCell,
  TableBody,
  TextField,
} from "@mui/material";
import { listUnmatched, bulkReconcile } from "../api/reconcile";
import type { ReconciledTransaction } from "../types";

export default function ReconcileDashboard() {
  const [shop, setShop] = useState("");
  const [data, setData] = useState<ReconciledTransaction[]>([]);

  const fetchData = () => {
    if (shop) listUnmatched(shop).then((r) => setData(r.data));
  };

  useEffect(() => {
    if (shop) fetchData();
  }, [shop]);

  const handleBulk = async () => {
    const pairs = data.map(
      (d) => [d.dropship_id!, d.shopee_id!] as [string, string],
    );
    await bulkReconcile(pairs, shop);
  };

  return (
    <div>
      <h2>Reconcile Dashboard</h2>
      <TextField
        label="Shop"
        value={shop}
        onChange={(e) => setShop(e.target.value)}
      />
      <Button onClick={fetchData}>Refresh</Button>
      <Button onClick={handleBulk}>Bulk</Button>
      <Table size="small">
        <TableHead>
          <TableRow>
            <TableCell>ID</TableCell>
            <TableCell>Status</TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {data.map((r) => (
            <TableRow key={r.id}>
              <TableCell>{r.id}</TableCell>
              <TableCell>{r.status}</TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  );
}
