import {
  Alert,
  Pagination,
  Button,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  CircularProgress,
} from "@mui/material";
import { LocalizationProvider } from "@mui/x-date-pickers";
import { DatePicker } from "@mui/x-date-pickers/DatePicker";
import { AdapterDateFns } from "@mui/x-date-pickers/AdapterDateFns";
import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import SortableTable from "./SortableTable";
import type { Column } from "./SortableTable";
import { getCurrentMonthRange } from "../utils/date";
import {
  listJenisChannels,
  listStoresByChannelName,
  listSalesProfit,
} from "../api";
import { getJournalLinesBySource } from "../api/journal";
import type {
  JenisChannel,
  Store,
  SalesProfit,
  JournalEntryWithLines,
  JournalLineDetail,
} from "../types";

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
  const [journalOpen, setJournalOpen] = useState(false);
  const [journalData, setJournalData] = useState<JournalEntryWithLines[]>([]);
  const [journalLoading, setJournalLoading] = useState(false);
  const navigate = useNavigate();

  const openJournal = (kode: string) => {
    setJournalLoading(true);
    getJournalLinesBySource(kode)
      .then((res) => {
        setJournalData(res.data);
      })
      .finally(() => {
        setJournalLoading(false);
        setJournalOpen(true);
      });
  };

  const money = (v: number) =>
    Number(v).toLocaleString("id-ID", { style: "currency", currency: "IDR" });

  const columns: Column<SalesProfit>[] = [
    {
      label: "Kode Pesanan",
      key: "kode_pesanan",
      render: (v, row) => (
        <Button size="small" onClick={() => openJournal(row.kode_pesanan)}>
          {v}
        </Button>
      ),
    },
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
    {
      label: "Dropship",
      render: (_, row) => (
        <Button
          size="small"
          onClick={() => navigate(`/dropship?order=${row.kode_pesanan}`)}
        >
          View
        </Button>
      ),
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
      <Dialog open={journalOpen} onClose={() => setJournalOpen(false)}>
        <DialogTitle>Journal Lines</DialogTitle>
        <DialogContent>
          {journalLoading ? (
            <CircularProgress />
          ) : journalData.length === 0 ? (
            <div>No journal entries found.</div>
          ) : (
            journalData.map((e) => (
              <div key={e.entry.journal_id} style={{ marginBottom: "1rem" }}>
                <div>ID: {e.entry.journal_id}</div>
                <div>
                  Date: {new Date(e.entry.entry_date).toLocaleDateString()}
                </div>
                <div>Description: {e.entry.description}</div>
                <SortableTable
                  columns={[
                    { label: "Account", key: "account_name" },
                    {
                      label: "Debit",
                      key: "amount",
                      align: "right",
                      render: (_v, r: JournalLineDetail) =>
                        r.is_debit
                          ? r.amount.toLocaleString("id-ID", {
                              style: "currency",
                              currency: "IDR",
                            })
                          : "",
                    },
                    {
                      label: "Credit",
                      key: "amount",
                      align: "right",
                      render: (_v, r: JournalLineDetail) =>
                        !r.is_debit
                          ? r.amount.toLocaleString("id-ID", {
                              style: "currency",
                              currency: "IDR",
                            })
                          : "",
                    },
                  ]}
                  data={e.lines}
                />
              </div>
            ))
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setJournalOpen(false)}>Close</Button>
        </DialogActions>
      </Dialog>
    </div>
  );
}
