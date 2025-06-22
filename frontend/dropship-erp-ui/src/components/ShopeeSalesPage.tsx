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
  TableRow,
} from "@mui/material";
import { LocalizationProvider } from "@mui/x-date-pickers";
import { DatePicker } from "@mui/x-date-pickers/DatePicker";
import { AdapterDateFns } from "@mui/x-date-pickers/AdapterDateFns";
import SortableTable from "./SortableTable";
import type { Column } from "./SortableTable";
import { useNavigate } from "react-router-dom";
import { useEffect, useState } from "react";
import { getCurrentMonthRange } from "../utils/date";
import {
  importShopee,
  confirmShopeeSettle,
  listJenisChannels,
  listStoresByChannelName,
  listShopeeSettled,
  sumShopeeSettled,
} from "../api";
import type {
  JenisChannel,
  Store,
  ShopeeSettled,
  ShopeeSettledSummary,
} from "../types";

export default function ShopeeSalesPage() {
  const [channels, setChannels] = useState<JenisChannel[]>([]);
  const [channel, setChannel] = useState("");
  const [stores, setStores] = useState<Store[]>([]);
  const [store, setStore] = useState("");
  const [order, setOrder] = useState("");
  const [sortKey, setSortKey] = useState<keyof ShopeeSettled>(
    "waktu_pesanan_dibuat",
  );
  const [sortDir, setSortDir] = useState<"asc" | "desc">("desc");
  const [firstOfMonth, lastOfMonth] = getCurrentMonthRange();
  const [from, setFrom] = useState(firstOfMonth);
  const [to, setTo] = useState(lastOfMonth);
  const [page, setPage] = useState(1);
  const [data, setData] = useState<ShopeeSettled[]>([]);
  const [total, setTotal] = useState(0);
  const [pageTotal, setPageTotal] = useState(0);
  const [allTotal, setAllTotal] = useState(0);
  const [pageSummary, setPageSummary] = useState<ShopeeSettledSummary | null>(
    null,
  );
  const [allSummary, setAllSummary] = useState<ShopeeSettledSummary | null>(
    null,
  );
  const [settling, setSettling] = useState<string | null>(null);
  const pageSize = 10;
  const navigate = useNavigate();

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
    {
      label: "Harga Asli Produk",
      key: "harga_asli_produk",
      align: "right",
      render: (v) =>
        Number(v).toLocaleString("id-ID", {
          style: "currency",
          currency: "IDR",
        }),
    },
    {
      label: "Total Diskon Produk",
      key: "total_diskon_produk",
      align: "right",
      render: (v) =>
        Number(v).toLocaleString("id-ID", {
          style: "currency",
          currency: "IDR",
        }),
    },
    {
      label: "Jumlah Pengembalian Dana",
      key: "jumlah_pengembalian_dana_ke_pembeli",
      align: "right",
      render: (v) =>
        Number(v).toLocaleString("id-ID", {
          style: "currency",
          currency: "IDR",
        }),
    },
    {
      label: "Diskon Produk Shopee",
      key: "diskon_produk_dari_shopee",
      align: "right",
      render: (v) =>
        Number(v).toLocaleString("id-ID", {
          style: "currency",
          currency: "IDR",
        }),
    },
    {
      label: "Diskon Voucher Penjual",
      key: "diskon_voucher_ditanggung_penjual",
      align: "right",
      render: (v) =>
        Number(v).toLocaleString("id-ID", {
          style: "currency",
          currency: "IDR",
        }),
    },
    {
      label: "Cashback Koin Penjual",
      key: "cashback_koin_ditanggung_penjual",
      align: "right",
      render: (v) =>
        Number(v).toLocaleString("id-ID", {
          style: "currency",
          currency: "IDR",
        }),
    },
    {
      label: "Ongkir Dibayar Pembeli",
      key: "ongkir_dibayar_pembeli",
      align: "right",
      render: (v) =>
        Number(v).toLocaleString("id-ID", {
          style: "currency",
          currency: "IDR",
        }),
    },
    {
      label: "Diskon Ongkir Jasa Kirim",
      key: "diskon_ongkir_ditanggung_jasa_kirim",
      align: "right",
      render: (v) =>
        Number(v).toLocaleString("id-ID", {
          style: "currency",
          currency: "IDR",
        }),
    },
    {
      label: "Gratis Ongkir Shopee",
      key: "gratis_ongkir_dari_shopee",
      align: "right",
      render: (v) =>
        Number(v).toLocaleString("id-ID", {
          style: "currency",
          currency: "IDR",
        }),
    },
    {
      label: "Ongkir Diteruskan Shopee",
      key: "ongkir_yang_diteruskan_oleh_shopee_ke_jasa_kirim",
      align: "right",
      render: (v) =>
        Number(v).toLocaleString("id-ID", {
          style: "currency",
          currency: "IDR",
        }),
    },
    {
      label: "Ongkos Kirim Pengembalian",
      key: "ongkos_kirim_pengembalian_barang",
      align: "right",
      render: (v) =>
        Number(v).toLocaleString("id-ID", {
          style: "currency",
          currency: "IDR",
        }),
    },
    {
      label: "Pengembalian Biaya Kirim",
      key: "pengembalian_biaya_kirim",
      align: "right",
      render: (v) =>
        Number(v).toLocaleString("id-ID", {
          style: "currency",
          currency: "IDR",
        }),
    },
    {
      label: "Biaya Komisi AMS",
      key: "biaya_komisi_ams",
      align: "right",
      render: (v) =>
        Number(v).toLocaleString("id-ID", {
          style: "currency",
          currency: "IDR",
        }),
    },
    {
      label: "Biaya Administrasi",
      key: "biaya_administrasi",
      align: "right",
      render: (v) =>
        Number(v).toLocaleString("id-ID", {
          style: "currency",
          currency: "IDR",
        }),
    },
    {
      label: "Biaya Layanan (+PPN)",
      key: "biaya_layanan_termasuk_ppn_11",
      align: "right",
      render: (v) =>
        Number(v).toLocaleString("id-ID", {
          style: "currency",
          currency: "IDR",
        }),
    },
    {
      label: "Premi",
      key: "premi",
      align: "right",
      render: (v) =>
        Number(v).toLocaleString("id-ID", {
          style: "currency",
          currency: "IDR",
        }),
    },
    {
      label: "Biaya Program",
      key: "biaya_program",
      align: "right",
      render: (v) =>
        Number(v).toLocaleString("id-ID", {
          style: "currency",
          currency: "IDR",
        }),
    },
    {
      label: "Biaya Kartu Kredit",
      key: "biaya_kartu_kredit",
      align: "right",
      render: (v) =>
        Number(v).toLocaleString("id-ID", {
          style: "currency",
          currency: "IDR",
        }),
    },
    {
      label: "Biaya Kampanye",
      key: "biaya_kampanye",
      align: "right",
      render: (v) =>
        Number(v).toLocaleString("id-ID", {
          style: "currency",
          currency: "IDR",
        }),
    },
    {
      label: "Bea Masuk/PPN/PPh",
      key: "bea_masuk_ppn_pph",
      align: "right",
      render: (v) =>
        Number(v).toLocaleString("id-ID", {
          style: "currency",
          currency: "IDR",
        }),
    },
    {
      label: "Total Penghasilan",
      key: "total_penghasilan",
      align: "right",
      render: (v) =>
        Number(v).toLocaleString("id-ID", {
          style: "currency",
          currency: "IDR",
        }),
    },
    {
      label: "Kompensasi",
      key: "kompensasi",
      align: "right",
      render: (v) =>
        Number(v).toLocaleString("id-ID", {
          style: "currency",
          currency: "IDR",
        }),
    },
    {
      label: "Promo Gratis Ongkir Dari Penjual",
      key: "promo_gratis_ongkir_dari_penjual",
      align: "right",
      render: (v) =>
        Number(v).toLocaleString("id-ID", {
          style: "currency",
          currency: "IDR",
        }),
    },
    { label: "Jasa Kirim", key: "jasa_kirim" },
    { label: "Nama Kurir", key: "nama_kurir" },
    {
      label: "Pengembalian Dana",
      key: "pengembalian_dana_ke_pembeli",
      align: "right",
      render: (v) =>
        Number(v).toLocaleString("id-ID", {
          style: "currency",
          currency: "IDR",
        }),
    },
    {
      label: "Pro-rata Koin Refund",
      key: "pro_rata_koin_yang_ditukarkan_untuk_pengembalian_barang",
      align: "right",
      render: (v) =>
        Number(v).toLocaleString("id-ID", {
          style: "currency",
          currency: "IDR",
        }),
    },
    {
      label: "Pro-rata Voucher Refund",
      key: "pro_rata_voucher_shopee_untuk_pengembalian_barang",
      align: "right",
      render: (v) =>
        Number(v).toLocaleString("id-ID", {
          style: "currency",
          currency: "IDR",
        }),
    },
    {
      label: "Promo Bank Returns",
      key: "pro_rated_bank_payment_channel_promotion_for_returns",
      align: "right",
      render: (v) =>
        Number(v).toLocaleString("id-ID", {
          style: "currency",
          currency: "IDR",
        }),
    },
    {
      label: "Promo Shopee Returns",
      key: "pro_rated_shopee_payment_channel_promotion_for_returns",
      align: "right",
      render: (v) =>
        Number(v).toLocaleString("id-ID", {
          style: "currency",
          currency: "IDR",
        }),
    },
    {
      label: "Dropship",
      render: (_, row) => (
        <Button
          size="small"
          onClick={() => navigate(`/dropship?order=${row.no_pesanan}`)}
        >
          View
        </Button>
      ),
    },
    {
      label: "Mismatch",
      key: "is_data_mismatch",
      align: "center",
      render: (v) => (v ? "⚠️" : ""),
    },
    {
      label: "Actions",
      key: "actions",
      render: (_, row) => {
        if (row.is_settled_confirmed) return "✔️";
        if (row.is_data_mismatch) return "";
        return (
          <Button
            size="small"
            variant="contained"
            onClick={() => handleConfirm(row.no_pesanan)}
            disabled={settling === row.no_pesanan}
          >
            {settling === row.no_pesanan ? "..." : "Confirm Settle"}
          </Button>
        );
      },
    },
  ];

  const [importOpen, setImportOpen] = useState(false);
  const [files, setFiles] = useState<File[]>([]);
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
        from,
        to,
        order,
        sort: sortKey as string,
        dir: sortDir,
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
      const pageSum: ShopeeSettledSummary = {
        harga_asli_produk: 0,
        total_diskon_produk: 0,
        gmv: 0,
        diskon_voucher_ditanggung_penjual: 0,
        biaya_administrasi: 0,
        biaya_layanan_termasuk_ppn_11: 0,
        total_penghasilan: sum,
      };
      res.data.data.forEach((cur) => {
        pageSum.harga_asli_produk += cur.harga_asli_produk;
        pageSum.total_diskon_produk += cur.total_diskon_produk;
        pageSum.diskon_voucher_ditanggung_penjual +=
          cur.diskon_voucher_ditanggung_penjual;
        pageSum.biaya_administrasi += cur.biaya_administrasi;
        pageSum.biaya_layanan_termasuk_ppn_11 +=
          cur.biaya_layanan_termasuk_ppn_11;
      });
      pageSum.gmv = pageSum.harga_asli_produk - pageSum.total_diskon_produk;
      setPageSummary(pageSum);
      const totalRes = await sumShopeeSettled({
        channel: channel || undefined,
        store,
        from,
        to,
      });
      setAllTotal(totalRes.data.total_penghasilan);
      setAllSummary(totalRes.data);
      setMsg(null);
    } catch (e: any) {
      setMsg({ type: "error", text: e.response?.data?.error || e.message });
    }
  };

  const handleConfirm = async (sn: string) => {
    setSettling(sn);
    try {
      await confirmShopeeSettle(sn);
      fetchData();
    } catch (e: any) {
      setMsg({ type: "error", text: e.response?.data?.error || e.message });
    } finally {
      setSettling(null);
    }
  };

  useEffect(() => {
    fetchData();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [channel, store, from, to, page, order, sortKey, sortDir]);

  const handleImport = async () => {
    try {
      if (files.length === 0) return;
      const res = await importShopee(files);
      setMsg({
        type: "success",
        text: `Imported ${res.data.inserted} rows successfully!`,
      });
      setFiles([]);
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
      {pageSummary && (
        <Table size="small" sx={{ mb: 1, maxWidth: 600 }}>
          <caption
            style={{ captionSide: "top", textAlign: "left", fontWeight: 600 }}
          >
            Page Summary
          </caption>
          <TableBody>
            <TableRow>
              <TableCell>Harga Asli Produk</TableCell>
              <TableCell align="right">
                {pageSummary.harga_asli_produk.toLocaleString("id-ID", {
                  style: "currency",
                  currency: "IDR",
                })}
              </TableCell>
            </TableRow>
            <TableRow>
              <TableCell>Total Diskon Produk</TableCell>
              <TableCell align="right">
                {pageSummary.total_diskon_produk.toLocaleString("id-ID", {
                  style: "currency",
                  currency: "IDR",
                })}
              </TableCell>
            </TableRow>
            <TableRow>
              <TableCell>GMV</TableCell>
              <TableCell align="right">
                {pageSummary.gmv.toLocaleString("id-ID", {
                  style: "currency",
                  currency: "IDR",
                })}
              </TableCell>
            </TableRow>
            <TableRow>
              <TableCell>Diskon Voucher Penjual</TableCell>
              <TableCell align="right">
                {pageSummary.diskon_voucher_ditanggung_penjual.toLocaleString(
                  "id-ID",
                  {
                    style: "currency",
                    currency: "IDR",
                  },
                )}
              </TableCell>
            </TableRow>
            <TableRow>
              <TableCell>Biaya Administrasi</TableCell>
              <TableCell align="right">
                {pageSummary.biaya_administrasi.toLocaleString("id-ID", {
                  style: "currency",
                  currency: "IDR",
                })}
              </TableCell>
            </TableRow>
            <TableRow>
              <TableCell>Biaya Layanan (+PPN)</TableCell>
              <TableCell align="right">
                {pageSummary.biaya_layanan_termasuk_ppn_11.toLocaleString(
                  "id-ID",
                  {
                    style: "currency",
                    currency: "IDR",
                  },
                )}
              </TableCell>
            </TableRow>
            <TableRow>
              <TableCell>Total Penghasilan</TableCell>
              <TableCell align="right">
                {pageSummary.total_penghasilan.toLocaleString("id-ID", {
                  style: "currency",
                  currency: "IDR",
                })}
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      )}
      {allSummary && (
        <Table size="small" sx={{ mb: 1, maxWidth: 600 }}>
          <caption
            style={{ captionSide: "top", textAlign: "left", fontWeight: 600 }}
          >
            All Summary
          </caption>
          <TableBody>
            <TableRow>
              <TableCell>Harga Asli Produk</TableCell>
              <TableCell align="right">
                {allSummary.harga_asli_produk.toLocaleString("id-ID", {
                  style: "currency",
                  currency: "IDR",
                })}
              </TableCell>
            </TableRow>
            <TableRow>
              <TableCell>Total Diskon Produk</TableCell>
              <TableCell align="right">
                {allSummary.total_diskon_produk.toLocaleString("id-ID", {
                  style: "currency",
                  currency: "IDR",
                })}
              </TableCell>
            </TableRow>
            <TableRow>
              <TableCell>GMV</TableCell>
              <TableCell align="right">
                {allSummary.gmv.toLocaleString("id-ID", {
                  style: "currency",
                  currency: "IDR",
                })}
              </TableCell>
            </TableRow>
            <TableRow>
              <TableCell>Diskon Voucher Penjual</TableCell>
              <TableCell align="right">
                {allSummary.diskon_voucher_ditanggung_penjual.toLocaleString(
                  "id-ID",
                  {
                    style: "currency",
                    currency: "IDR",
                  },
                )}
              </TableCell>
            </TableRow>
            <TableRow>
              <TableCell>Biaya Administrasi</TableCell>
              <TableCell align="right">
                {allSummary.biaya_administrasi.toLocaleString("id-ID", {
                  style: "currency",
                  currency: "IDR",
                })}
              </TableCell>
            </TableRow>
            <TableRow>
              <TableCell>Biaya Layanan (+PPN)</TableCell>
              <TableCell align="right">
                {allSummary.biaya_layanan_termasuk_ppn_11.toLocaleString(
                  "id-ID",
                  {
                    style: "currency",
                    currency: "IDR",
                  },
                )}
              </TableCell>
            </TableRow>
            <TableRow>
              <TableCell>Total Penghasilan</TableCell>
              <TableCell align="right">
                {allSummary.total_penghasilan.toLocaleString("id-ID", {
                  style: "currency",
                  currency: "IDR",
                })}
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      )}
      <div style={{ overflowX: "auto" }}>
        <SortableTable
          columns={columns}
          data={data}
          defaultSort={{ key: "waktu_pesanan_dibuat", direction: "desc" }}
          onSortChange={(k, d) => {
            setSortKey(k);
            setSortDir(d);
          }}
        />
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
            multiple
            aria-label="XLSX file"
            onChange={(e) => setFiles(Array.from(e.target.files || []))}
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
