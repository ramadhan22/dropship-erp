import { useEffect, useState } from "react";
import { Alert, Button } from "@mui/material";
import { LocalizationProvider } from "@mui/x-date-pickers";
import { DateTimePicker } from "@mui/x-date-pickers/DateTimePicker";
import { AdapterDateFns } from "@mui/x-date-pickers/AdapterDateFns";
import SortableTable from "./SortableTable";
import type { Column } from "./SortableTable";
import { listAllStores } from "../api";
import { listAdsTopups, createAdsTopupJournal } from "../api/adsTopup";
import type { Store, WalletTransaction } from "../types";

export default function AdsTopupPage() {
  const [store, setStore] = useState("");
  const [stores, setStores] = useState<Store[]>([]);
  const [list, setList] = useState<WalletTransaction[]>([]);
  const [hasNext, setHasNext] = useState(false);
  const [pageNo, setPageNo] = useState(0);
  const [pageSize, setPageSize] = useState(25);
  const [createFrom, setCreateFrom] = useState<Date | null>(null);
  const [createTo, setCreateTo] = useState<Date | null>(null);
  const [msg, setMsg] = useState<string | null>(null);

  useEffect(() => {
    listAllStores().then((r) => setStores(r));
  }, []);

  const columns: Column<WalletTransaction>[] = [
    {
      label: "Create Time",
      key: "create_time",
      render: (v) => new Date(v * 1000).toLocaleString(),
    },
    { label: "Amount", key: "amount", align: "right" },
    { label: "Current Balance", key: "current_balance", align: "right" },
    {
      label: "Journal",
      render: (_, row) => (
        <Button size="small" onClick={() => handleJournal(row)}>
          Insert
        </Button>
      ),
    },
  ];

  const fetchData = async (page: number, replace = true) => {
    try {
      const res = await listAdsTopups({
        store,
        page_no: page,
        page_size: pageSize,
        create_time_from: createFrom
          ? Math.floor(createFrom.getTime() / 1000)
          : undefined,
        create_time_to: createTo
          ? Math.floor(createTo.getTime() / 1000)
          : undefined,
      });
      setList((prev) =>
        replace ? res.data.data : [...prev, ...res.data.data],
      );
      setHasNext(res.data.has_next_page);
      setPageNo(page);
      setMsg(null);
    } catch (e: any) {
      setMsg(e.response?.data?.error || e.message);
    }
  };

  const handleJournal = async (tx: WalletTransaction) => {
    try {
      await createAdsTopupJournal({
        store,
        transaction_id: tx.transaction_id,
        create_time: tx.create_time,
        amount: tx.amount,
      });
      setMsg("journaled");
    } catch (e: any) {
      setMsg(e.response?.data?.error || e.message);
    }
  };

  return (
    <div>
      <h2>Ads Topup</h2>
      <div
        style={{
          display: "flex",
          gap: "0.5rem",
          marginBottom: "1rem",
          flexWrap: "wrap",
        }}
      >
        <select
          aria-label="Store"
          value={store}
          onChange={(e) => setStore(e.target.value)}
        >
          <option value="">Select Store</option>
          {stores.map((s) => (
            <option key={s.store_id} value={s.nama_toko}>
              {s.nama_toko}
            </option>
          ))}
        </select>
        <select
          aria-label="Page Size"
          value={pageSize}
          onChange={(e) => setPageSize(Number(e.target.value))}
        >
          {[25, 50, 100].map((n) => (
            <option key={n} value={n}>
              {n}
            </option>
          ))}
        </select>
        <LocalizationProvider dateAdapter={AdapterDateFns}>
          <DateTimePicker
            label="Create From"
            value={createFrom}
            onChange={(d) => setCreateFrom(d)}
            slotProps={{ textField: { size: "small" } }}
          />
        </LocalizationProvider>
        <LocalizationProvider dateAdapter={AdapterDateFns}>
          <DateTimePicker
            label="Create To"
            value={createTo}
            onChange={(d) => setCreateTo(d)}
            slotProps={{ textField: { size: "small" } }}
          />
        </LocalizationProvider>
        <Button disabled={!store} onClick={() => fetchData(0)}>
          Fetch
        </Button>
      </div>
      {msg && <Alert severity="info">{msg}</Alert>}
      {list.length > 0 && (
        <div style={{ overflowX: "auto" }}>
          <SortableTable columns={columns} data={list} />
          <div style={{ marginTop: "1rem" }}>
            <Button
              disabled={!hasNext}
              onClick={() => fetchData(pageNo + 1, false)}
            >
              More
            </Button>
          </div>
        </div>
      )}
    </div>
  );
}
