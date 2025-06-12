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
  listStoresByChannelName,
  getDropshipPurchaseDetails,
} from "../api";
import type {
  DropshipPurchase,
  JenisChannel,
  Store,
  DropshipPurchaseDetail,
} from "../types";

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
  const [page, setPage] = useState(1);
  const pageSize = 10;

  const [channels, setChannels] = useState<JenisChannel[]>([]);
  const [stores, setStores] = useState<Store[]>([]);
  const [data, setData] = useState<DropshipPurchase[]>([]);
  const [total, setTotal] = useState(0);
  const [detailOpen, setDetailOpen] = useState(false);
  const [details, setDetails] = useState<DropshipPurchaseDetail[]>([]);
  const [selected, setSelected] = useState<DropshipPurchase | null>(null);

  useEffect(() => {
    listJenisChannels().then((res) => setChannels(res.data));
  }, []);

  useEffect(() => {
    if (channel) {
      listStoresByChannelName(channel).then((res) => setStores(res.data ?? []));
    } else {
      setStores([]);
    }
  }, [channel]);

  const fetchData = async () => {
    const res = await listDropshipPurchases({
      channel,
      store,
      date,
      page,
      page_size: pageSize,
    });
    setData(res.data.data);
    setTotal(res.data.total);
  };

  useEffect(() => {
    fetchData();
  }, [channel, store, date, page]);

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
          label="Date"
          type="date"
          value={date}
          onChange={(e) => {
            setDate(e.target.value);
            setPage(1);
          }}
          size="small"
          InputLabelProps={{ shrink: true }}
        />
      </div>

      <Dialog open={detailOpen} onClose={() => setDetailOpen(false)}>
        <DialogTitle>Purchase Detail</DialogTitle>
        <DialogContent>
          {selected && (
            <div style={{ marginBottom: "1rem" }}>
              <div>Kode Pesanan: {selected.kode_pesanan}</div>
              <div>Nama Toko: {selected.nama_toko}</div>
            </div>
          )}
          <Table size="small">
            <TableHead>
              <TableRow>
                <TableCell>SKU</TableCell>
                <TableCell>Nama Produk</TableCell>
                <TableCell>Qty</TableCell>
                <TableCell>Harga</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {details.map((dt) => (
                <TableRow key={dt.id}>
                  <TableCell>{dt.sku}</TableCell>
                  <TableCell>{dt.nama_produk}</TableCell>
                  <TableCell>{dt.qty}</TableCell>
                  <TableCell>
                    {dt.total_harga_produk.toLocaleString("id-ID", {
                      style: "currency",
                      currency: "IDR",
                    })}
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDetailOpen(false)}>Close</Button>
        </DialogActions>
      </Dialog>
      <Table size="small">
        <TableHead>
          <TableRow>
            <TableCell>Order Code</TableCell>
            <TableCell>Store</TableCell>
            <TableCell>Channel</TableCell>
            <TableCell>Date</TableCell>
          <TableCell>Total</TableCell>
          <TableCell>Action</TableCell>
        </TableRow>
        </TableHead>
        <TableBody>
          {data.map((d) => (
            <TableRow key={d.kode_pesanan}>
              <TableCell>{d.kode_pesanan}</TableCell>
              <TableCell>{d.nama_toko}</TableCell>
              <TableCell>{d.jenis_channel}</TableCell>
              <TableCell>
                {new Date(d.waktu_pesanan_terbuat).toLocaleDateString("id-ID")}
              </TableCell>
              <TableCell>
                {d.total_transaksi.toLocaleString("id-ID", {
                  style: "currency",
                  currency: "IDR",
                })}
              </TableCell>
              <TableCell>
                <Button
                  size="small"
                  onClick={async () => {
                    const res = await getDropshipPurchaseDetails(d.kode_pesanan);
                    setDetails(res.data);
                    setSelected(d);
                    setDetailOpen(true);
                  }}
                >
                  Detail
                </Button>
              </TableCell>
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
