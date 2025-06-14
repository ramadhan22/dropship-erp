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
import SortableTable from "./SortableTable";
import type { Column } from "./SortableTable";
import { useEffect, useState } from "react";
import {
  importShopee,
  listJenisChannels,
  listStoresByChannelName,
  listShopeeSettled,
  sumShopeeSettled,
} from "../api";
import type { JenisChannel, Store, ShopeeSettled } from "../types";

export default function ShopeeSalesPage() {
  const [channels, setChannels] = useState<JenisChannel[]>([]);
  const [channel, setChannel] = useState("");
  const [stores, setStores] = useState<Store[]>([]);
  const [store, setStore] = useState("");
  const [date, setDate] = useState("");
  const [month, setMonth] = useState("");
  const [year, setYear] = useState("");
  const [page, setPage] = useState(1);
  const [data, setData] = useState<ShopeeSettled[]>([]);
  const [total, setTotal] = useState(0);
  const [pageTotal, setPageTotal] = useState(0);
  const [allTotal, setAllTotal] = useState(0);
  const pageSize = 10;

  const columns: Column<ShopeeSettled>[] = [
    { label: "Nama Toko", key: "nama_toko" },
    { label: "No Pesanan", key: "no_pesanan" },
    { label: "No Pengajuan", key: "no_pengajuan" },
    { label: "Username Pembeli", key: "username_pembeli" },
    {
      label: "Waktu Pesanan Dibuat",
      key: "waktu_pesanan_dibuat",
      render: (v) => new Date(v).toLocaleDateString("id-ID"),
    },
    { label: "Metode Pembayaran", key: "metode_pembayaran_pembeli" },
    {
      label: "Tanggal Dana Dilepaskan",
      key: "tanggal_dana_dilepaskan",
      render: (v) => new Date(v).toLocaleDateString("id-ID"),
    },
    { label: "Harga Asli Produk", key: "harga_asli_produk" },
    { label: "Total Diskon Produk", key: "total_diskon_produk" },
    {
      label: "Jumlah Pengembalian Dana",
      key: "jumlah_pengembalian_dana_ke_pembeli",
    },
    { label: "Diskon Produk Shopee", key: "diskon_produk_dari_shopee" },
    {
      label: "Diskon Voucher Penjual",
      key: "diskon_voucher_ditanggung_penjual",
    },
    { label: "Cashback Koin Penjual", key: "cashback_koin_ditanggung_penjual" },
    { label: "Ongkir Dibayar Pembeli", key: "ongkir_dibayar_pembeli" },
    {
      label: "Diskon Ongkir Jasa Kirim",
      key: "diskon_ongkir_ditanggung_jasa_kirim",
    },
    { label: "Gratis Ongkir Shopee", key: "gratis_ongkir_dari_shopee" },
    {
      label: "Ongkir Diteruskan Shopee",
      key: "ongkir_yang_diteruskan_oleh_shopee_ke_jasa_kirim",
    },
    {
      label: "Ongkos Kirim Pengembalian",
      key: "ongkos_kirim_pengembalian_barang",
    },
    { label: "Pengembalian Biaya Kirim", key: "pengembalian_biaya_kirim" },
    { label: "Biaya Komisi AMS", key: "biaya_komisi_ams" },
    { label: "Biaya Administrasi", key: "biaya_administrasi" },
    { label: "Biaya Layanan (+PPN)", key: "biaya_layanan_termasuk_ppn_11" },
    { label: "Premi", key: "premi" },
    { label: "Biaya Program", key: "biaya_program" },
    { label: "Biaya Kartu Kredit", key: "biaya_kartu_kredit" },
    { label: "Biaya Kampanye", key: "biaya_kampanye" },
    { label: "Bea Masuk/PPN/PPh", key: "bea_masuk_ppn_pph" },
    { label: "Total Penghasilan", key: "total_penghasilan" },
    { label: "Kompensasi", key: "kompensasi" },
    {
      label: "Promo Gratis Ongkir Dari Penjual",
      key: "promo_gratis_ongkir_dari_penjual",
    },
    { label: "Jasa Kirim", key: "jasa_kirim" },
    { label: "Nama Kurir", key: "nama_kurir" },
    { label: "Pengembalian Dana", key: "pengembalian_dana_ke_pembeli" },
    {
      label: "Pro-rata Koin Refund",
      key: "pro_rata_koin_yang_ditukarkan_untuk_pengembalian_barang",
    },
    {
      label: "Pro-rata Voucher Refund",
      key: "pro_rata_voucher_shopee_untuk_pengembalian_barang",
    },
    {
      label: "Promo Bank Returns",
      key: "pro_rated_bank_payment_channel_promotion_for_returns",
    },
    {
      label: "Promo Shopee Returns",
      key: "pro_rated_shopee_payment_channel_promotion_for_returns",
    },
  ];

  const [importOpen, setImportOpen] = useState(false);
  const [file, setFile] = useState<File | null>(null);
  const [msg, setMsg] = useState<{
    type: "success" | "error";
    text: string;
  } | null>(null);

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
    try {
      const res = await listShopeeSettled({
        channel: channel || undefined,
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
        (acc, cur) => acc + cur.total_penghasilan,
        0,
      );
      setPageTotal(sum);
      const totalRes = await sumShopeeSettled({
        channel: channel || undefined,
        store,
        date,
        month,
        year,
      });
      setAllTotal(totalRes.data.total);
      setMsg(null);
    } catch (e: any) {
      setMsg({ type: "error", text: e.response?.data?.error || e.message });
    }
  };

  useEffect(() => {
    fetchData();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [channel, store, date, month, year, page]);

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
      <Button
        variant="contained"
        onClick={() => setImportOpen(true)}
        sx={{ mb: 2 }}
      >
        Import
      </Button>
      <div style={{ display: "flex", gap: "1rem", marginBottom: "1rem" }}>
        <select
          aria-label="Channel"
          value={channel}
          onChange={(e) => setChannel(e.target.value)}
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
          onChange={(e) => setStore(e.target.value)}
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
          onChange={(e) => setDate(e.target.value)}
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
      {msg && (
        <Alert severity={msg.type} sx={{ mb: 2 }}>
          {msg.text}
        </Alert>
      )}
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
      <div style={{ overflowX: "auto" }}>
        <SortableTable columns={columns} data={data} />
      </div>
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
