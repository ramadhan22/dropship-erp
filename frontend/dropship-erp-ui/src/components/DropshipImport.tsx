import {
  Alert,
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableRow,
  TextField,
} from "@mui/material";
import { useEffect, useState } from "react";
import {
  importDropship,
  listDropshipPurchases,
  listJenisChannels,
  listStores,
} from "../api";
import type { DropshipPurchase, JenisChannel, Store } from "../types";

export default function DropshipImport() {
  const [file, setFile] = useState<File | null>(null);
  const [msg, setMsg] = useState<{
    type: "success" | "error";
    text: string;
  } | null>(null);
  const [open, setOpen] = useState(false);

  const [channel, setChannel] = useState("");
  const [store, setStore] = useState("");
  const [date, setDate] = useState("");
  const [month, setMonth] = useState("");
  const [year, setYear] = useState("");
  const [page, setPage] = useState(1);
  const pageSize = 10;

  const [channels, setChannels] = useState<JenisChannel[]>([]);
  const [stores, setStores] = useState<Store[]>([]);
  const [data, setData] = useState<DropshipPurchase[]>([]);
  const [total, setTotal] = useState(0);

  useEffect(() => {
    listJenisChannels().then((res) => setChannels(res.data));
  }, []);

  useEffect(() => {
    if (channel) {
      listStores(Number(channel)).then((res) => setStores(res.data ?? []));
    } else {
      setStores([]);
    }
  }, [channel]);

  const fetchData = async () => {
    const res = await listDropshipPurchases({
      channel,
      store,
      date,
      month,
      year,
      page,
      page_size: pageSize,
    });
    setData(res.data.data);
    setTotal(res.data.total);
  };

  useEffect(() => {
    fetchData();
  }, [channel, store, date, month, year, page]);

  const handleSubmit = async () => {
    try {
      if (!file) return;
      const res = await importDropship(file);
      setMsg({
        type: "success",
        text: `Imported ${res.data.inserted} rows successfully!`,
      });
      setFile(null);
      fetchData();
    } catch (e: any) {
      setMsg({ type: "error", text: e.response?.data?.error || e.message });
    }
  };

  return (
    <div>
      <h2>Jakmall Purchases</h2>
      <Button variant="contained" onClick={() => setOpen(true)} sx={{ mb: 2 }}>
        Import
      </Button>
      <div style={{ display: "flex", gap: "0.5rem", marginBottom: "1rem" }}>
        <select
          aria-label="Channel"
          value={channel}
          onChange={(e) => {
            setChannel(e.target.value);
            setPage(1);
          }}
        >
          <option value="">All Channels</option>
          {channels.map((c) => (
            <option key={c.jenis_channel_id} value={c.jenis_channel}>
              {c.jenis_channel}
            </option>
          ))}
        </select>
        <select
          aria-label="Store"
          value={store}
          onChange={(e) => {
            setStore(e.target.value);
            setPage(1);
          }}
        >
          <option value="">All Stores</option>
          {stores.map((s) => (
            <option key={s.store_id} value={s.nama_toko}>
              {s.nama_toko}
            </option>
          ))}
        </select>
        <TextField
          label="Date (YYYY-MM-DD)"
          value={date}
          onChange={(e) => {
            setDate(e.target.value);
            setPage(1);
          }}
          size="small"
        />
        <TextField
          label="Month"
          value={month}
          onChange={(e) => {
            setMonth(e.target.value);
            setPage(1);
          }}
          size="small"
        />
        <TextField
          label="Year"
          value={year}
          onChange={(e) => {
            setYear(e.target.value);
            setPage(1);
          }}
          size="small"
        />
      </div>
      <Table size="small">
        <TableHead>
          <TableRow>
            <TableCell>Order Code</TableCell>
            <TableCell>Store</TableCell>
            <TableCell>Channel</TableCell>
            <TableCell>Date</TableCell>
            <TableCell>Total</TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {data.map((d) => (
            <TableRow key={d.kode_pesanan}>
              <TableCell>{d.kode_pesanan}</TableCell>
              <TableCell>{d.nama_toko}</TableCell>
              <TableCell>{d.jenis_channel}</TableCell>
              <TableCell>{d.waktu_pesanan_terbuat}</TableCell>
              <TableCell>{d.total_transaksi}</TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
      <div style={{ marginTop: "0.5rem" }}>
        <Button
          variant="outlined"
          disabled={page === 1}
          onClick={() => setPage((p) => p - 1)}
          sx={{ mr: 1 }}
        >
          Prev
        </Button>
        <Button
          variant="outlined"
          disabled={page * pageSize >= total}
          onClick={() => setPage((p) => p + 1)}
        >
          Next
        </Button>
      </div>

      <Dialog open={open} onClose={() => setOpen(false)}>
        <DialogTitle>Import Dropship CSV</DialogTitle>
        <DialogContent>
          <input
            type="file"
            aria-label="CSV file"
            onChange={(e) => setFile(e.target.files?.[0] || null)}
          />
          {msg && (
            <Alert severity={msg.type} sx={{ mt: 2 }}>
              {msg.text}
            </Alert>
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={handleSubmit}>Import</Button>
        </DialogActions>
      </Dialog>
    </div>
  );
}
