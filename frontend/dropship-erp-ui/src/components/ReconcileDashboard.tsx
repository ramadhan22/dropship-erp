import { useEffect, useState } from "react";
import {
  Button,
  Table,
  TableHead,
  TableRow,
  TableCell,
  TableBody,
  Checkbox,
  FormControlLabel,
} from "@mui/material";
import { listCandidates, bulkReconcile } from "../api/reconcile";
import { listAllStores } from "../api";
import type { ReconcileCandidate, Store } from "../types";

export default function ReconcileDashboard() {
  const [shop, setShop] = useState("");
  const [stores, setStores] = useState<Store[]>([]);
  const [data, setData] = useState<ReconcileCandidate[]>([]);
  const [diffOnly, setDiffOnly] = useState(false);

  useEffect(() => {
    listAllStores().then((s) => setStores(s));
  }, []);

  const fetchData = () => {
    if (shop) listCandidates(shop).then((r) => setData(r.data));
  };

  useEffect(() => {
    if (shop) fetchData();
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
      <select
        aria-label="Shop"
        value={shop}
        onChange={(e) => setShop(e.target.value)}
        style={{ marginRight: "0.5rem" }}
      >
        <option value="">Select Store</option>
        {stores.map((s) => (
          <option key={s.store_id} value={s.nama_toko}>
            {s.nama_toko}
          </option>
        ))}
      </select>
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
