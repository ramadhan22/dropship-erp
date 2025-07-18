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
  Pagination,
} from "@mui/material";
import { LocalizationProvider } from "@mui/x-date-pickers";
import { DatePicker } from "@mui/x-date-pickers/DatePicker";
import { AdapterDateFns } from "@mui/x-date-pickers/AdapterDateFns";
import SortableTable from "./SortableTable";
import type { Column } from "./SortableTable";
import { useEffect, useState, useCallback } from "react";
import { useSearchParams } from "react-router-dom";
import { getCurrentMonthRange } from "../utils/date";
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
  const [importChannel, setImportChannel] = useState("");
  const [searchParams] = useSearchParams();

  const [channel, setChannel] = useState("");
  const [store, setStore] = useState("");
  const [order, setOrder] = useState("");
  const [sortKey, setSortKey] = useState<keyof DropshipPurchase>(
    "waktu_pesanan_terbuat",
  );
  const [sortDir, setSortDir] = useState<"asc" | "desc">("desc");
  const [firstOfMonth, lastOfMonth] = getCurrentMonthRange();
  const [from, setFrom] = useState(firstOfMonth);
  const [to, setTo] = useState(lastOfMonth);
  const [page, setPage] = useState(1);
  const pageSizeOptions = [10, 20, 50, 100, 250, 500, 1000];
  const [pageSize, setPageSize] = useState(20);

  const columns: Column<DropshipPurchase>[] = [
    { label: "Kode Pesanan", key: "kode_pesanan" },
    { label: "Kode Transaksi", key: "kode_transaksi" },
    {
      label: "Waktu Pesanan Terbuat",
      key: "waktu_pesanan_terbuat",
      render: (v) => new Date(v).toLocaleDateString("id-ID"),
    },
    { label: "Status Pesanan", key: "status_pesanan_terakhir" },
    {
      label: "Biaya Lainnya",
      key: "biaya_lainnya",
      align: "right",
      render: (v) =>
        Number(v).toLocaleString("id-ID", {
          style: "currency",
          currency: "IDR",
        }),
    },
    {
      label: "Biaya Mitra Jakmall",
      key: "biaya_mitra_jakmall",
      align: "right",
      render: (v) =>
        Number(v).toLocaleString("id-ID", {
          style: "currency",
          currency: "IDR",
        }),
    },
    {
      label: "Total Transaksi",
      key: "total_transaksi",
      align: "right",
      render: (v) =>
        Number(v).toLocaleString("id-ID", {
          style: "currency",
          currency: "IDR",
        }),
    },
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

  const fetchData = useCallback(async (ord?: string) => {
    const res = await listDropshipPurchases({
      channel,
      store,
      from,
      to,
      order: ord ?? order,
      sort: sortKey as string,
      dir: sortDir,
      page,
      page_size: pageSize,
    });
    setData(res.data.data);
    setTotal(res.data.total);
    const pages = Math.max(1, Math.ceil(res.data.total / pageSize));
    if (page > pages) {
      setPage(pages);
    }
    const sum = res.data.data.reduce(
      (acc, cur) => acc + cur.total_transaksi,
      0,
    );
    setPageTotal(sum);
    const totalRes = await sumDropshipPurchases({
      channel,
      store,
      from,
      to,
    });
    setAllTotal(totalRes.data.total);
  }, [channel, store, from, to, order, sortKey, sortDir, page, pageSize]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  useEffect(() => {
    const ord = searchParams.get("order");
    if (ord) {
      setOrder(ord);
      fetchData(ord);
    }
  }, [searchParams, fetchData]);

  const handleSubmit = async () => {
    try {
      if (!files || files.length === 0) return;
      const res = await importDropship(
        Array.from(files),
        importChannel || undefined,
      );
      setMsg({
        type: "success",
        text: `Queued ${res.data.queued} files successfully!`,
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
        <input
          aria-label="Search Order"
          placeholder="Order No"
          value={order}
          onChange={(e) => {
            setOrder(e.target.value);
            setPage(1);
          }}
          style={{ height: "2rem" }}
        />
        <LocalizationProvider dateAdapter={AdapterDateFns}>
          <DatePicker
            label="From"
            format="yyyy-MM-dd"
            value={new Date(from)}
            onChange={(date) => {
              if (!date) return;
              setFrom(date.toISOString().split("T")[0]);
              setPage(1);
            }}
            slotProps={{ textField: { size: "small" } }}
          />
        </LocalizationProvider>
        <LocalizationProvider dateAdapter={AdapterDateFns}>
          <DatePicker
            label="To"
            format="yyyy-MM-dd"
            value={new Date(to)}
            onChange={(date) => {
              if (!date) return;
              setTo(date.toISOString().split("T")[0]);
              setPage(1);
            }}
            slotProps={{ textField: { size: "small" } }}
          />
        </LocalizationProvider>
      </div>
      <div style={{ marginBottom: "0.5rem" }}>
        <strong>Page Total:</strong>{" "}
        {pageTotal.toLocaleString("id-ID", {
          style: "currency",
          currency: "IDR",
        })}
        {" | "}
        <strong>Total Rows:</strong> {total}
        {" | "}
        <strong>All Total:</strong>{" "}
        {allTotal.toLocaleString("id-ID", {
          style: "currency",
          currency: "IDR",
        })}
      </div>

      <Dialog open={detailOpen} onClose={() => setDetailOpen(false)}>
        <DialogTitle>Purchase Detail</DialogTitle>
        <DialogContent>
          {selected && (
            <div style={{ marginBottom: "1rem" }}>
              <div>
                <strong>Kode Pesanan:</strong> {selected.kode_pesanan}
              </div>
              <div>
                <strong>Kode Transaksi:</strong> {selected.kode_transaksi}
              </div>
              <div>
                <strong>Nama Toko:</strong> {selected.nama_toko}
              </div>
              <div>
                <strong>Status:</strong> {selected.status_pesanan_terakhir}
              </div>
              <div>
                <strong>Waktu Pesanan:</strong>{" "}
                {new Date(selected.waktu_pesanan_terbuat).toLocaleString(
                  "id-ID",
                )}
              </div>
              <div>
                <strong>Total Transaksi:</strong>{" "}
                {selected.total_transaksi.toLocaleString("id-ID", {
                  style: "currency",
                  currency: "IDR",
                })}
              </div>
            </div>
          )}
          <Table size="small">
          <TableHead>
            <TableRow>
              <TableCell>SKU</TableCell>
              <TableCell>Nama Produk</TableCell>
              <TableCell>Qty</TableCell>
              <TableCell>Harga</TableCell>
              <TableCell>Harga Channel</TableCell>
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
                <TableCell>
                  {dt.total_harga_produk_channel.toLocaleString("id-ID", {
                    style: "currency",
                    currency: "IDR",
                  })}
                </TableCell>
              </TableRow>
            ))}
            <TableRow>
              <TableCell colSpan={3} sx={{ textAlign: "right", fontWeight: "bold" }}>
                Total
              </TableCell>
              <TableCell sx={{ fontWeight: "bold" }}>
                {details
                  .reduce((acc, cur) => acc + cur.total_harga_produk, 0)
                  .toLocaleString("id-ID", {
                    style: "currency",
                    currency: "IDR",
                  })}
              </TableCell>
              <TableCell sx={{ fontWeight: "bold" }}>
                {details
                  .reduce((acc, cur) => acc + cur.total_harga_produk_channel, 0)
                  .toLocaleString("id-ID", {
                    style: "currency",
                    currency: "IDR",
                  })}
              </TableCell>
            </TableRow>
          </TableBody>
          </Table>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDetailOpen(false)}>Close</Button>
        </DialogActions>
      </Dialog>
      <div style={{ overflowX: "auto" }}>
        <SortableTable
          columns={columns}
          data={data}
          defaultSort={{ key: "waktu_pesanan_terbuat", direction: "desc" }}
          onSortChange={(k, d) => {
            setSortKey(k);
            setSortDir(d);
          }}
        />
      </div>
      <div
        style={{
          marginTop: "1rem",
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
        }}
      >
        <div>Total: {total}</div>
        <div style={{ display: "flex", alignItems: "center", gap: "0.5rem" }}>
          <select
            value={pageSize}
            onChange={(e) => {
              setPageSize(Number(e.target.value));
              setPage(1);
            }}
          >
            {pageSizeOptions.map((n) => (
              <option key={n} value={n}>
                {n}
              </option>
            ))}
          </select>
          <Pagination
            page={page}
            count={Math.max(1, Math.ceil(total / pageSize))}
            onChange={(_, val) => setPage(val)}
          />
        </div>
      </div>

      <Dialog open={open} onClose={() => setOpen(false)}>
        <DialogTitle>Import Dropship CSV</DialogTitle>
        <DialogContent>
          <select
            aria-label="Import Channel"
            value={importChannel}
            onChange={(e) => setImportChannel(e.target.value)}
            style={{ display: "block", marginBottom: "0.5rem" }}
          >
            <option value="">All</option>
            <option value="Shopee">Shopee</option>
            <option value="Tokopedia">Tokopedia</option>
            <option value="Tiktok Seller Center">Tiktok Seller Center</option>
          </select>
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
