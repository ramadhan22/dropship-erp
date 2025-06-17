import { Alert, Pagination } from "@mui/material";
import { LocalizationProvider } from "@mui/x-date-pickers";
import { DatePicker } from "@mui/x-date-pickers/DatePicker";
import { AdapterDateFns } from "@mui/x-date-pickers/AdapterDateFns";
import { useEffect, useState } from "react";
import SortableTable from "./SortableTable";
import type { Column } from "./SortableTable";
import { getCurrentMonthRange } from "../utils/date";
import {
  listJenisChannels,
  listStoresByChannelName,
  listSalesProfit,
} from "../api";
import type { JenisChannel, Store, SalesProfit } from "../types";

export default function SalesProfitPage() {
  const [channels, setChannels] = useState<JenisChannel[]>([]);
  const [channel, setChannel] = useState("");
  const [stores, setStores] = useState<Store[]>([]);
  const [store, setStore] = useState("");
  const [order, setOrder] = useState("");
  const [sortKey, setSortKey] = useState<keyof SalesProfit>("tanggal_pesanan");
  const [sortDir, setSortDir] = useState<"asc" | "desc">("desc");
  const [firstOfMonth, lastOfMonth] = getCurrentMonthRange();
  const [from, setFrom] = useState(firstOfMonth);
  const [to, setTo] = useState(lastOfMonth);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [data, setData] = useState<SalesProfit[]>([]);
  const [total, setTotal] = useState(0);
  const [msg, setMsg] = useState<{
    type: "success" | "error";
    text: string;
  } | null>(null);

  const money = (v: number) =>
    Number(v).toLocaleString("id-ID", { style: "currency", currency: "IDR" });

  const columns: Column<SalesProfit>[] = [
    { label: "Kode Pesanan", key: "kode_pesanan" },
    {
      label: "Tanggal Pesanan",
      key: "tanggal_pesanan",
      render: (v) => new Date(v).toLocaleDateString("id-ID"),
    },
    { label: "Modal", key: "modal_purchase", align: "right", render: money },
    {
      label: "Jumlah Sales",
      key: "amount_sales",
      align: "right",
      render: money,
    },
    {
      label: "Biaya Mitra",
      key: "biaya_mitra_jakmall",
      align: "right",
      render: money,
    },
    {
      label: "Biaya Admin",
      key: "biaya_administrasi",
      align: "right",
      render: money,
    },
    {
      label: "Biaya Layanan",
      key: "biaya_layanan",
      align: "right",
      render: money,
    },
    {
      label: "Biaya Voucher",
      key: "biaya_voucher",
      align: "right",
      render: money,
    },
    {
      label: "Biaya Affiliate",
      key: "biaya_affiliate",
      align: "right",
      render: money,
    },
    { label: "Profit", key: "profit", align: "right", render: money },
    {
      label: "% Profit",
      key: "profit_percent",
      align: "right",
      render: (v) => `${v.toFixed(2)}%`,
    },
  ];

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
      const res = await listSalesProfit({
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
      setMsg(null);
    } catch (e: any) {
      setMsg({ type: "error", text: e.response?.data?.error || e.message });
    }
  };

  useEffect(() => {
    fetchData();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [channel, store, from, to, order, page, sortKey, sortDir, pageSize]);

  return (
    <div>
      <h2>Sales Profit</h2>
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
          type="text"
          placeholder="No Pesanan"
          value={order}
          onChange={(e) => {
            setOrder(e.target.value);
            setPage(1);
          }}
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
        <select
          aria-label="Rows"
          value={pageSize}
          onChange={(e) => {
            setPageSize(Number(e.target.value));
            setPage(1);
          }}
        >
          <option value={10}>10</option>
          <option value={25}>25</option>
          <option value={50}>50</option>
        </select>
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
          defaultSort={{ key: "tanggal_pesanan", direction: "desc" }}
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
    </div>
  );
}
