import { Alert, Button, Dialog, DialogActions, DialogContent, DialogTitle, Pagination } from "@mui/material";
import { LocalizationProvider } from "@mui/x-date-pickers";
import { DatePicker } from "@mui/x-date-pickers/DatePicker";
import { AdapterDateFns } from "@mui/x-date-pickers/AdapterDateFns";
import { useEffect, useState } from "react";
import { importShopeeAffiliate, listShopeeAffiliate } from "../api";
import SortableTable from "./SortableTable";
import type { Column } from "./SortableTable";
import type { ShopeeAffiliateSale } from "../types";

export default function ShopeeAffiliatePage() {
  const [period, setPeriod] = useState(() => new Date().toISOString().slice(0, 7));
  const [page, setPage] = useState(1);
  const [data, setData] = useState<ShopeeAffiliateSale[]>([]);
  const [total, setTotal] = useState(0);
  const pageSize = 10;
  const [importOpen, setImportOpen] = useState(false);
  const [file, setFile] = useState<File | null>(null);
  const [msg, setMsg] = useState<{ type: "success" | "error"; text: string } | null>(null);

  const columns: Column<ShopeeAffiliateSale>[] = [
    { label: "Kode Pesanan", key: "kode_pesanan" },
    { label: "Status", key: "status_pesanan" },
    { label: "Nama Affiliate", key: "nama_affiliate" },
    { label: "Username", key: "username_affiliate" },
    {
      label: "Waktu Pesanan",
      key: "waktu_pesanan",
      render: (v) => new Date(v).toLocaleDateString("id-ID"),
    },
    {
      label: "Nilai Pembelian",
      key: "nilai_pembelian",
      align: "right",
      render: (v) =>
        Number(v).toLocaleString("id-ID", { style: "currency", currency: "IDR" }),
    },
    {
      label: "Komisi Affiliate",
      key: "estimasi_komisi_affiliate_per_pesanan",
      align: "right",
      render: (v) =>
        Number(v).toLocaleString("id-ID", { style: "currency", currency: "IDR" }),
    },
  ];

  const fetchData = async () => {
    try {
      const [year, month] = period.split("-");
      const res = await listShopeeAffiliate({ month, year, page, page_size: pageSize });
      setData(res.data.data);
      setTotal(res.data.total);
      setMsg(null);
    } catch (e: any) {
      setMsg({ type: "error", text: e.response?.data?.error || e.message });
    }
  };

  useEffect(() => {
    fetchData();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [period, page]);

  const handleImport = async () => {
    try {
      if (!file) return;
      const res = await importShopeeAffiliate(file);
      setMsg({ type: "success", text: `Imported ${res.data.inserted} rows successfully!` });
      setFile(null);
      setImportOpen(false);
      fetchData();
    } catch (e: any) {
      setMsg({ type: "error", text: e.response?.data?.error || e.message });
    }
  };

  return (
    <div>
      <h2>Shopee Affiliate Sales</h2>
      <Button variant="contained" onClick={() => setImportOpen(true)} sx={{ mb: 2 }}>
        Import
      </Button>
      <LocalizationProvider dateAdapter={AdapterDateFns}>
        <DatePicker
          label="Period"
          views={["year", "month"]}
          openTo="month"
          format="yyyy-MM"
          value={new Date(period)}
          onChange={date => {
            if (!date) return;
            setPeriod(date.toISOString().slice(0, 7));
            setPage(1);
          }}
          slotProps={{ textField: { size: "small", sx: { mb: 2, ml: 1 }, InputLabelProps: { shrink: true } } }}
        />
      </LocalizationProvider>
      {msg && <Alert severity={msg.type} sx={{ mb: 2 }}>{msg.text}</Alert>}
      <div style={{ overflowX: "auto" }}>
        <SortableTable columns={columns} data={data} />
      </div>
      <Pagination sx={{ mt: 2 }} page={page} count={Math.max(1, Math.ceil(total / pageSize))} onChange={(_, val) => setPage(val)} />
      <Dialog open={importOpen} onClose={() => setImportOpen(false)}>
        <DialogTitle>Import Shopee Affiliate CSV</DialogTitle>
        <DialogContent>
          <input type="file" aria-label="CSV file" onChange={e => setFile(e.target.files?.[0] || null)} />
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setImportOpen(false)}>Cancel</Button>
          <Button variant="contained" onClick={handleImport}>Import</Button>
        </DialogActions>
      </Dialog>
    </div>
  );
}
