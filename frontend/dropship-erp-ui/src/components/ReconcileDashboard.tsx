import { useEffect, useState } from "react";
import { Button, Checkbox, FormControlLabel } from "@mui/material";
import SortableTable, { Column } from "./SortableTable";
import { listCandidates, bulkReconcile } from "../api/reconcile";
import { listAllStores } from "../api";
import type { ReconcileCandidate, Store } from "../types";
import usePagination from "../usePagination";

export default function ReconcileDashboard() {
  const [shop, setShop] = useState("");
  const [stores, setStores] = useState<Store[]>([]);
  const [data, setData] = useState<ReconcileCandidate[]>([]);
  const [diffOnly, setDiffOnly] = useState(false);
  const { paginated, controls } = usePagination(
    diffOnly ? data.filter((d) => d.no_pesanan) : data,
  );

  useEffect(() => {
    listAllStores().then((s) => setStores(s));
  }, []);

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

  const columns: Column<ReconcileCandidate>[] = [
    { label: "Kode Pesanan", key: "kode_pesanan" },
    { label: "Kode Invoice Channel", key: "kode_invoice_channel" },
    { label: "Status", key: "status_pesanan_terakhir" },
    { label: "No Pesanan Shopee", key: "no_pesanan" },
  ];


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
      <SortableTable columns={columns} data={paginated} />
      {controls}
    </div>
  );
}
