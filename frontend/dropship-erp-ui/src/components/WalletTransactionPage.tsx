import { useEffect, useState } from "react";
import { Alert } from "@mui/material";
import SortableTable from "./SortableTable";
import type { Column } from "./SortableTable";
import { listAllStores } from "../api";
import { listWalletTransactions } from "../api/wallet";
import type { Store, WalletTransaction } from "../types";

export default function WalletTransactionPage() {
  const [store, setStore] = useState("");
  const [stores, setStores] = useState<Store[]>([]);
  const [list, setList] = useState<WalletTransaction[]>([]);
  const [hasNext, setHasNext] = useState(false);
  const [pageNo, setPageNo] = useState(0);
  const [pageSize, setPageSize] = useState(40);
  const [createFrom, setCreateFrom] = useState("");
  const [createTo, setCreateTo] = useState("");
  const [walletType, setWalletType] = useState("");
  const [transactionType, setTransactionType] = useState("");
  const [moneyFlow, setMoneyFlow] = useState("");
  const [txnTabType, setTxnTabType] = useState("");
  const [msg, setMsg] = useState<string | null>(null);

  useEffect(() => {
    listAllStores().then((r) => setStores(r));
  }, []);

  const columns: Column<WalletTransaction>[] = [
    { label: "Create Time", key: "create_time" },
    { label: "Status", key: "status" },
    { label: "Transaction Type", key: "transaction_type" },
    { label: "Money Flow", key: "money_flow" },
    { label: "Amount", key: "amount", align: "right" },
    { label: "Current Balance", key: "current_balance", align: "right" },
    { label: "Order SN", key: "order_sn" },
  ];

  const fetchData = async (page: number) => {
    try {
      const res = await listWalletTransactions({
        store,
        page_no: page,
        page_size: pageSize,
        create_time_from: createFrom ? Number(createFrom) : undefined,
        create_time_to: createTo ? Number(createTo) : undefined,
        wallet_type: walletType || undefined,
        transaction_type: transactionType || undefined,
        money_flow: moneyFlow || undefined,
        transaction_tab_type: txnTabType || undefined,
      });
      setList(res.data.data);
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
        <input
          type="number"
          placeholder="Page Size"
          value={pageSize}
          onChange={(e) => setPageSize(Number(e.target.value))}
          style={{ width: "6rem" }}
        />
        <input
          type="text"
          placeholder="Create From ts"
          value={createFrom}
          onChange={(e) => setCreateFrom(e.target.value)}
        />
        <input
          type="text"
          placeholder="Create To ts"
          value={createTo}
          onChange={(e) => setCreateTo(e.target.value)}
        />
        <input
          type="text"
          placeholder="Wallet Type"
          value={walletType}
          onChange={(e) => setWalletType(e.target.value)}
        />
        <input
          type="text"
          placeholder="Transaction Type"
          value={transactionType}
          onChange={(e) => setTransactionType(e.target.value)}
        />
        <input
          type="text"
          placeholder="Money Flow"
          value={moneyFlow}
          onChange={(e) => setMoneyFlow(e.target.value)}
        />
        <input
          type="text"
          placeholder="Transaction Tab"
          value={txnTabType}
          onChange={(e) => setTxnTabType(e.target.value)}
        />
        <button disabled={!store} onClick={() => fetchData(0)}>
          Fetch
        </button>
      </div>
      {msg && <Alert severity="error">{msg}</Alert>}
      {list.length > 0 && (
        <div style={{ overflowX: "auto" }}>
          <SortableTable columns={columns} data={list} />
          <div style={{ marginTop: "1rem" }}>
            <button
              disabled={pageNo === 0}
              onClick={() => fetchData(pageNo - 1)}
            >
              Prev
            </button>
            <span style={{ margin: "0 1rem" }}>Page {pageNo + 1}</span>
            <button disabled={!hasNext} onClick={() => fetchData(pageNo + 1)}>
              Next
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
