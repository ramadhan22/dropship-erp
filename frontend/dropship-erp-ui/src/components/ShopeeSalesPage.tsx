import {
  Alert,
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  Pagination,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableRow,
  TextField,
} from "@mui/material";
import { useEffect, useState } from "react";
import {
  importShopee,
  listJenisChannels,
  listStores,
  listShopeeSettled,
} from "../api";
import type { JenisChannel, Store, ShopeeSettled } from "../types";

export default function ShopeeSalesPage() {
  const [channels, setChannels] = useState<JenisChannel[]>([]);
  const [channelId, setChannelId] = useState<number | "">("");
  const [stores, setStores] = useState<Store[]>([]);
  const [store, setStore] = useState("");
  const [date, setDate] = useState("");
  const [month, setMonth] = useState("");
  const [year, setYear] = useState("");
  const [page, setPage] = useState(1);
  const [data, setData] = useState<ShopeeSettled[]>([]);
  const [total, setTotal] = useState(0);
  const pageSize = 10;

  const [importOpen, setImportOpen] = useState(false);
  const [file, setFile] = useState<File | null>(null);
  const [msg, setMsg] = useState<{ type: "success" | "error"; text: string } | null>(
    null,
  );

  useEffect(() => {
    listJenisChannels().then((res) => setChannels(res.data));
  }, []);

  useEffect(() => {
    if (channelId !== "") {
      listStores(Number(channelId)).then((res) => setStores(res.data ?? []));
    } else {
      setStores([]);
    }
  }, [channelId]);

  const fetchData = async () => {
    try {
      const channelName =
        channelId === ""
          ? undefined
          : channels.find((c) => c.jenis_channel_id === Number(channelId))?.
              jenis_channel;
      const res = await listShopeeSettled({
        channel: channelName,
        store,
        date,
        month,
        year,
        page,
        page_size: pageSize,
      });
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
  }, [channelId, store, date, month, year, page]);

  const handleImport = async () => {
    try {
      if (!file) return;
      const res = await importShopee(file);
      setMsg({
        type: "success",
        text: `Imported ${res.data.inserted} rows successfully!`,
      });
      setFile(null);
      setImportOpen(false);
      fetchData();
    } catch (e: any) {
      setMsg({ type: "error", text: e.response?.data?.error || e.message });
    }
  };

  return (
    <div>
      <h2>Shopee Sales</h2>
      <Button variant="contained" onClick={() => setImportOpen(true)} sx={{ mb: 2 }}>
        Import
      </Button>
      <div style={{ display: "flex", gap: "1rem", marginBottom: "1rem" }}>
        <select
          aria-label="Channel"
          value={channelId}
          onChange={(e) => setChannelId(e.target.value as any)}
        >
          <option value="">All Channels</option>
          {channels.map((c) => (
            <option key={c.jenis_channel_id} value={c.jenis_channel_id}>
              {c.jenis_channel}
            </option>
          ))}
        </select>
        <select aria-label="Store" value={store} onChange={(e) => setStore(e.target.value)}>
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
          onChange={(e) => setDate(e.target.value)}
          size="small"
          InputLabelProps={{ shrink: true }}
        />
        <TextField
          label="Month"
          value={month}
          onChange={(e) => setMonth(e.target.value)}
          size="small"
        />
        <TextField
          label="Year"
          value={year}
          onChange={(e) => setYear(e.target.value)}
          size="small"
        />
      </div>
      {msg && (
        <Alert severity={msg.type} sx={{ mb: 2 }}>
          {msg.text}
        </Alert>
      )}
      <Table size="small">
        <TableHead>
          <TableRow>
            <TableCell>No Pesanan</TableCell>
            <TableCell>Nama Toko</TableCell>
            <TableCell>Tanggal Dana Dilepaskan</TableCell>
            <TableCell>Total Penerimaan</TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {data.map((d, i) => (
            <TableRow key={i}>
              <TableCell>{d.no_pesanan}</TableCell>
              <TableCell>{d.nama_toko}</TableCell>
              <TableCell>{d.tanggal_dana_dilepaskan}</TableCell>
              <TableCell>{d.total_penerimaan}</TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
      <Pagination
        sx={{ mt: 2 }}
        page={page}
        count={Math.max(1, Math.ceil(total / pageSize))}
        onChange={(_, val) => setPage(val)}
      />
      <Dialog open={importOpen} onClose={() => setImportOpen(false)}>
        <DialogTitle>Import Shopee XLSX</DialogTitle>
        <DialogContent>
          <input
            type="file"
            aria-label="XLSX file"
            onChange={(e) => setFile(e.target.files?.[0] || null)}
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setImportOpen(false)}>Cancel</Button>
          <Button variant="contained" onClick={handleImport}>
            Import
          </Button>
        </DialogActions>
      </Dialog>
    </div>
  );
}
