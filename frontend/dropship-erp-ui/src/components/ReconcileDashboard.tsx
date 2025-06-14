import { useEffect, useState } from "react";
import {
  Button,
  Table,
  TableHead,
  TableRow,
  TableCell,
  TableBody,
  TextField,
  Checkbox,
  FormControlLabel,
} from "@mui/material";
import { listCandidates, bulkReconcile } from "../api/reconcile";
import type { ReconcileCandidate } from "../types";

export default function ReconcileDashboard() {
  const [shop, setShop] = useState("");
  const [data, setData] = useState<ReconcileCandidate[]>([]);
  const [diffOnly, setDiffOnly] = useState(false);

  const fetchData = () => {
    listCandidates(shop).then((r) => setData(r.data));
  };

  useEffect(() => {
    fetchData();
  }, [shop]);

  const handleBulk = async () => {
    const pairs = data
      .filter((d) => d.no_pesanan)
      .map((d) => [d.kode_invoice_channel, d.no_pesanan!] as [string, string]);
    if (pairs.length) await bulkReconcile(pairs, shop);
  };

  const displayData = diffOnly ? data.filter((d) => d.no_pesanan) : data;

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
      <FormControlLabel
        control={
          <Checkbox
            checked={diffOnly}
            onChange={(e) => setDiffOnly(e.target.checked)}
          />
        }
        label="Status mismatch only"
      />
      <Table size="small">
        <TableHead>
          <TableRow>
            <TableCell>Kode Pesanan</TableCell>
            <TableCell>Kode Invoice Channel</TableCell>
            <TableCell>Status</TableCell>
            <TableCell>No Pesanan Shopee</TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {displayData.map((r) => (
            <TableRow key={r.kode_pesanan}>
              <TableCell>{r.kode_pesanan}</TableCell>
              <TableCell>{r.kode_invoice_channel}</TableCell>
              <TableCell>{r.status_pesanan_terakhir}</TableCell>
              <TableCell>{r.no_pesanan || "-"}</TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  );
}
