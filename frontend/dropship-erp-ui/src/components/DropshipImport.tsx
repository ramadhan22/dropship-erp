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
import SortableTable, { Column } from "./SortableTable";
import { useEffect, useState } from "react";
import {
  importDropship,
  listDropshipPurchases,
  sumDropshipPurchases,
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
  const [files, setFiles] = useState<FileList | null>(null);
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

  const columns: Column<DropshipPurchase>[] = [
    { label: "Kode Pesanan", key: "kode_pesanan" },
    { label: "Kode Transaksi", key: "kode_transaksi" },
    {
      label: "Waktu Pesanan Terbuat",
      key: "waktu_pesanan_terbuat",
      render: (v) => new Date(v).toLocaleDateString("id-ID"),
    },
    { label: "Status Pesanan", key: "status_pesanan_terakhir" },
    { label: "Biaya Lainnya", key: "biaya_lainnya" },
    { label: "Biaya Mitra Jakmall", key: "biaya_mitra_jakmall" },
    { label: "Total Transaksi", key: "total_transaksi" },
    { label: "Dibuat Oleh", key: "dibuat_oleh" },
    { label: "Channel", key: "jenis_channel" },
    { label: "Nama Toko", key: "nama_toko" },
    { label: "Kode Invoice Channel", key: "kode_invoice_channel" },
    { label: "Gudang Pengiriman", key: "gudang_pengiriman" },
    { label: "Jenis Ekspedisi", key: "jenis_ekspedisi" },
    { label: "Cashless", key: "cashless" },
    { label: "Nomor Resi", key: "nomor_resi" },
    {
      label: "Waktu Pengiriman",
      key: "waktu_pengiriman",
      render: (v) => new Date(v).toLocaleDateString("id-ID"),
    },
    { label: "Provinsi", key: "provinsi" },
    { label: "Kota", key: "kota" },
    {
      label: "Action",
      render: (_, row) => (
        <Button
          size="small"
          onClick={async () => {
            const res = await getDropshipPurchaseDetails(row.kode_pesanan);
            setDetails(res.data);
            setSelected(row);
            setDetailOpen(true);
          }}
        >
          Detail
        </Button>
      ),
    },
  ];

  const [channels, setChannels] = useState<JenisChannel[]>([]);
  const [stores, setStores] = useState<Store[]>([]);
  const [data, setData] = useState<DropshipPurchase[]>([]);
  const [total, setTotal] = useState(0);
  const [pageTotal, setPageTotal] = useState(0);
  const [allTotal, setAllTotal] = useState(0);
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
      month,
      year,
      page,
      page_size: pageSize,
    });
    setData(res.data.data);
    setTotal(res.data.total);
    const sum = res.data.data.reduce(
      (acc, cur) => acc + cur.total_transaksi,
      0,
    );
    setPageTotal(sum);
    const totalRes = await sumDropshipPurchases({
      channel,
      store,
      date,
      month,
      year,
    });
    setAllTotal(totalRes.data.total);
  };

  useEffect(() => {
    fetchData();
  }, [channel, store, date, month, year, page]);

  const handleSubmit = async () => {
    try {
      if (!files || files.length === 0) return;
      let inserted = 0;
      for (const f of Array.from(files)) {
        const res = await importDropship(f);
        inserted += res.data.inserted;
      }
      setMsg({
        type: "success",
        text: `Imported ${inserted} rows from ${files.length} files successfully!`,
      });
      setFiles(null);
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
        <TextField
          label="Month"
          type="number"
          value={month}
          onChange={(e) => {
            setMonth(e.target.value);
            setPage(1);
          }}
          size="small"
          sx={{ width: 100 }}
        />
        <TextField
          label="Year"
          type="number"
          value={year}
          onChange={(e) => {
            setYear(e.target.value);
            setPage(1);
          }}
          size="small"
          sx={{ width: 100 }}
        />
      </div>
      <div style={{ marginBottom: "0.5rem" }}>
        <strong>Page Total:</strong>{" "}
        {pageTotal.toLocaleString("id-ID", { style: "currency", currency: "IDR" })}
        {" | "}
        <strong>Total Rows:</strong> {total}
        {" | "}
        <strong>All Total:</strong>{" "}
        {allTotal.toLocaleString("id-ID", { style: "currency", currency: "IDR" })}
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
      <div style={{ overflowX: "auto" }}>
        <SortableTable columns={columns} data={data} />
      </div>
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
            multiple
            aria-label="CSV files"
            onChange={(e) => setFiles(e.target.files)}
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
