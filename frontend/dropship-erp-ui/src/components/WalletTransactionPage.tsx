import { useEffect, useState } from "react";
import { Alert } from "@mui/material";
import { LocalizationProvider } from "@mui/x-date-pickers";
import { DateTimePicker } from "@mui/x-date-pickers/DateTimePicker";
import { AdapterDateFns } from "@mui/x-date-pickers/AdapterDateFns";
import SortableTable from "./SortableTable";
import type { Column } from "./SortableTable";
import { listAllStores } from "../api";
import { listWalletTransactions } from "../api/wallet";
import type { Store, WalletTransaction } from "../types";

const TRANSACTION_TYPES = [
  { value: 101, label: "ESCROW_VERIFIED_ADD" },
  { value: 102, label: "ESCROW_VERIFIED_MINUS" },
  { value: 201, label: "WITHDRAWAL_CREATED" },
  { value: 202, label: "WITHDRAWAL_COMPLETED" },
  { value: 203, label: "WITHDRAWAL_CANCELLED" },
  { value: 401, label: "ADJUSTMENT_ADD" },
  { value: 402, label: "ADJUSTMENT_MINUS" },
  { value: 404, label: "FBS_ADJUSTMENT_ADD" },
  { value: 405, label: "FBS_ADJUSTMENT_MINUS" },
  { value: 406, label: "ADJUSTMENT_CENTER_ADD" },
  { value: 407, label: "ADJUSTMENT_CENTER_DEDUCT" },
  { value: 408, label: "FSF_COST_PASSING_DEDUCT" },
  { value: 409, label: "PERCEPTION_VAT_TAX_DEDUCT" },
];

const MONEY_FLOW = ["MONEY_IN", "MONEY_OUT"] as const;

const TXN_TAB_TYPES = [
  "wallet_order_income",
  "wallet_adjustment_filter",
  "wallet_wallet_payment",
  "wallet_refund_from_order",
  "wallet_withdrawals",
  "fast_escrow_repayment",
  "fast_pay",
  "seller_loan",
  "corporate_loan",
  "pix_transactions_filter",
  "open_finance_transactions_filter",
];

export default function WalletTransactionPage() {
  const [store, setStore] = useState("");
  const [stores, setStores] = useState<Store[]>([]);
  const [list, setList] = useState<WalletTransaction[]>([]);
  const [hasNext, setHasNext] = useState(false);
  const [pageNo, setPageNo] = useState(0);
  const [pageSize, setPageSize] = useState(25);
  const [createFrom, setCreateFrom] = useState<Date | null>(null);
  const [createTo, setCreateTo] = useState<Date | null>(null);
  const [walletType, setWalletType] = useState("");
  const [transactionType, setTransactionType] = useState("");
  const [moneyFlow, setMoneyFlow] = useState("");
  const [txnTabType, setTxnTabType] = useState("");
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
    { label: "Status", key: "status" },
    { label: "Transaction Type", key: "transaction_type" },
    { label: "Money Flow", key: "money_flow" },
    { label: "Amount", key: "amount", align: "right" },
    { label: "Current Balance", key: "current_balance", align: "right" },
    { label: "Order SN", key: "order_sn" },
  ];

  const fetchData = async (page: number, replace = true) => {
    try {
      const res = await listWalletTransactions({
        store,
        page_no: page,
        page_size: pageSize,
        create_time_from: createFrom
          ? Math.floor(createFrom.getTime() / 1000)
          : undefined,
        create_time_to: createTo ? Math.floor(createTo.getTime() / 1000) : undefined,
        wallet_type: walletType || undefined,
        transaction_type: transactionType || undefined,
        money_flow: moneyFlow || undefined,
        transaction_tab_type: txnTabType || undefined,
      });
      setList((prev) => (replace ? res.data.data : [...prev, ...res.data.data]));
      setHasNext(res.data.has_next_page);
      setPageNo(page);
      setMsg(null);
    } catch (e: any) {
      setMsg(e.response?.data?.error || e.message);
    }
  };

  return (
    <div>
      <h2>Wallet Transactions</h2>
      <div style={{ display: "flex", gap: "0.5rem", marginBottom: "1rem", flexWrap: "wrap" }}>
        <select aria-label="Store" value={store} onChange={(e) => setStore(e.target.value)}>
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
        <input
          type="text"
          placeholder="Wallet Type"
          value={walletType}
          onChange={(e) => setWalletType(e.target.value)}
        />
        <select
          aria-label="Transaction Type"
          value={transactionType}
          onChange={(e) => setTransactionType(e.target.value)}
        >
          <option value="">Txn Type</option>
          {TRANSACTION_TYPES.map((t) => (
            <option key={t.value} value={String(t.value)}>
              {t.label}
            </option>
          ))}
        </select>
        <select
          aria-label="Money Flow"
          value={moneyFlow}
          onChange={(e) => setMoneyFlow(e.target.value)}
        >
          <option value="">Money Flow</option>
          {MONEY_FLOW.map((m) => (
            <option key={m} value={m}>
              {m}
            </option>
          ))}
        </select>
        <select
          aria-label="Transaction Tab"
          value={txnTabType}
          onChange={(e) => setTxnTabType(e.target.value)}
        >
          <option value="">Tab Type</option>
          {TXN_TAB_TYPES.map((t) => (
            <option key={t} value={t}>
              {t}
            </option>
          ))}
        </select>
        <button disabled={!store} onClick={() => fetchData(0)}>
          Fetch
        </button>
      </div>
      {msg && <Alert severity="error">{msg}</Alert>}
      {list.length > 0 && (
        <div style={{ overflowX: "auto" }}>
          <SortableTable columns={columns} data={list} />
          <div style={{ marginTop: "1rem" }}>
            <button disabled={!hasNext} onClick={() => fetchData(pageNo + 1, false)}>
              More
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
