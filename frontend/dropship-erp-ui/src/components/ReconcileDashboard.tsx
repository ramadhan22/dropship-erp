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
import { listCandidates, bulkReconcile } from "../api/reconcile";
import type { ReconcileCandidate } from "../types";

export default function ReconcileDashboard() {
  const [shop, setShop] = useState("");
  const [data, setData] = useState<ReconcileCandidate[]>([]);

  const fetchData = () => {
    if (shop) listCandidates(shop).then((r) => setData(r.data));
  };

  useEffect(() => {
    if (shop) fetchData();
  }, [shop]);

  const handleBulk = async () => {
    const pairs = data
      .filter((d) => d.no_pesanan)
      .map((d) => [d.kode_pesanan, d.no_pesanan!] as [string, string]);
    if (pairs.length) await bulkReconcile(pairs, shop);
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
            <TableCell>Kode Pesanan</TableCell>
            <TableCell>Status</TableCell>
            <TableCell>No Pesanan Shopee</TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {data.map((r) => (
            <TableRow key={r.kode_pesanan}>
              <TableCell>{r.kode_pesanan}</TableCell>
              <TableCell>{r.status_pesanan_terakhir}</TableCell>
              <TableCell>{r.no_pesanan || "-"}</TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  );
}
