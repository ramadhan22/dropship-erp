import {
  Alert,
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  Pagination,
  TextField,
} from "@mui/material";
import { LocalizationProvider } from "@mui/x-date-pickers";
import { DatePicker } from "@mui/x-date-pickers/DatePicker";
import { AdapterDateFns } from "@mui/x-date-pickers/AdapterDateFns";

import { useEffect, useState } from "react";
import { getCurrentMonthRange } from "../utils/date";
import {
  importShopeeAffiliate,
  listShopeeAffiliate,
  sumShopeeAffiliate,
} from "../api";
import SortableTable from "./SortableTable";
import type { Column } from "./SortableTable";
import type { ShopeeAffiliateSale } from "../types";

export default function ShopeeAffiliatePage() {
  const [firstOfMonth, lastOfMonth] = getCurrentMonthRange();
  const [from, setFrom] = useState(firstOfMonth);
  const [to, setTo] = useState(lastOfMonth);
  const [order, setOrder] = useState("");
  const [page, setPage] = useState(1);
  const [data, setData] = useState<ShopeeAffiliateSale[]>([]);
  const [total, setTotal] = useState(0);
  const [pageTotal, setPageTotal] = useState(0);
  const [allTotal, setAllTotal] = useState(0);
  const pageSize = 10;
  const [importOpen, setImportOpen] = useState(false);
  const [file, setFile] = useState<File | null>(null);
  const [msg, setMsg] = useState<{
    type: "success" | "error";
    text: string;
  } | null>(null);

  const columns: Column<ShopeeAffiliateSale>[] = [
    { label: "Nama Toko", key: "nama_toko" },
    { label: "Kode Pesanan", key: "kode_pesanan" },
    { label: "Status", key: "status_pesanan" },
    { label: "Status Terverifikasi", key: "status_terverifikasi" },
    {
      label: "Waktu Pesanan",
      key: "waktu_pesanan",
      render: (v) => new Date(v).toLocaleString(),
    },
    {
      label: "Waktu Pesanan Selesai",
      key: "waktu_pesanan_selesai",
      render: (v) => new Date(v).toLocaleString(),
    },
    {
      label: "Waktu Terverifikasi",
      key: "waktu_pesanan_terverifikasi",
      render: (v) => new Date(v).toLocaleString(),
    },
    { label: "Kode Produk", key: "kode_produk" },
    { label: "Nama Produk", key: "nama_produk" },
    { label: "ID Model", key: "id_model" },
    { label: "L1 Kategori", key: "l1_kategori_global" },
    { label: "L2 Kategori", key: "l2_kategori_global" },
    { label: "L3 Kategori", key: "l3_kategori_global" },
    { label: "Kode Promo", key: "kode_promo" },
    { label: "Harga", key: "harga" },
    { label: "Jumlah", key: "jumlah" },
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
        Number(v).toLocaleString("id-ID", {
          style: "currency",
          currency: "IDR",
        }),
    },
    {
      label: "Komisi Affiliate",
      key: "estimasi_komisi_affiliate_per_pesanan",
      align: "right",
      render: (v) =>
        Number(v).toLocaleString("id-ID", {
          style: "currency",
          currency: "IDR",
        }),
    },
    { label: "MCN", key: "mcn_terhubung" },
    { label: "ID Komisi Pesanan", key: "id_komisi_pesanan" },
    { label: "Partner Promo", key: "partner_promo" },
    { label: "Jenis Promo", key: "jenis_promo" },
    { label: "Nilai Pembelian", key: "nilai_pembelian" },
    { label: "Jumlah Pengembalian", key: "jumlah_pengembalian" },
    { label: "Tipe Pesanan", key: "tipe_pesanan" },
    { label: "Komisi per Produk", key: "estimasi_komisi_per_produk" },
    {
      label: "Komisi Affiliate per Produk",
      key: "estimasi_komisi_affiliate_per_produk",
    },
    {
      label: "% Komisi Affiliate",
      key: "persentase_komisi_affiliate_per_produk",
    },
    { label: "Komisi MCN per Produk", key: "estimasi_komisi_mcn_per_produk" },
    { label: "% Komisi MCN", key: "persentase_komisi_mcn_per_produk" },
    { label: "Komisi per Pesanan", key: "estimasi_komisi_per_pesanan" },
    {
      label: "Komisi Affiliate per Pesanan",
      key: "estimasi_komisi_affiliate_per_pesanan",
    },
    { label: "Komisi MCN per Pesanan", key: "estimasi_komisi_mcn_per_pesanan" },
    { label: "Catatan Produk", key: "catatan_produk" },
    { label: "Platform", key: "platform" },
    { label: "Tingkat Komisi", key: "tingkat_komisi" },
    { label: "Pengeluaran", key: "pengeluaran" },
    { label: "Status Pemotongan", key: "status_pemotongan" },
    { label: "Metode Pemotongan", key: "metode_pemotongan" },
    {
      label: "Waktu Pemotongan",
      key: "waktu_pemotongan",
      render: (v) => new Date(v).toLocaleString(),
    },
  ];

  const fetchData = async () => {
    try {
      const res = await listShopeeAffiliate({
        no_pesanan: order || undefined,
        from,
        to,
        page,
        page_size: pageSize,
      });
      setData(res.data.data);
      setTotal(res.data.total);
      const sum = res.data.data.reduce(
        (acc, cur) => acc + cur.estimasi_komisi_affiliate_per_pesanan,
        0,
      );
      setPageTotal(sum);
      const allRes = await sumShopeeAffiliate({
        no_pesanan: order || undefined,
        from,
        to,
      });
      setAllTotal(allRes.data.total_komisi_affiliate);
      setMsg(null);
    } catch (e: any) {
      setMsg({ type: "error", text: e.response?.data?.error || e.message });
    }
  };

  useEffect(() => {
    fetchData();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [order, from, to, page]);

  const handleImport = async () => {
    try {
      if (!file) return;
      const res = await importShopeeAffiliate(file);
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
      <h2>Shopee Affiliate Sales</h2>
      <Button
        variant="contained"
        onClick={() => setImportOpen(true)}
        sx={{ mb: 2 }}
      >
        Import
      </Button>

      <div style={{ display: "flex", gap: "1rem", marginBottom: "1rem" }}>
        <TextField
          label="No Pesanan"
          value={order}
          onChange={(e) => {
            setOrder(e.target.value);
            setPage(1);
          }}
          size="small"
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
      {msg && (
        <Alert severity={msg.type} sx={{ mb: 2 }}>
          {msg.text}
        </Alert>
      )}

      <div style={{ overflowX: "auto" }}>
        <SortableTable
          columns={columns}
          data={data}
          defaultSort={{ key: "waktu_pesanan", direction: "desc" }}
        />
      </div>
      <Pagination
        sx={{ mt: 2 }}
        page={page}
        count={Math.max(1, Math.ceil(total / pageSize))}
        onChange={(_, val) => setPage(val)}
      />
      <Dialog open={importOpen} onClose={() => setImportOpen(false)}>
        <DialogTitle>Import Shopee Affiliate CSV</DialogTitle>
        <DialogContent>
          <input
            type="file"
            aria-label="CSV file"
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
