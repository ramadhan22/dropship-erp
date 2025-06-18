import { useEffect, useState } from "react";
import {
  Button,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  Alert,
} from "@mui/material";
import SortableTable from "./SortableTable";
import type { Column } from "./SortableTable";
import { listAssetAccounts, adjustAssetBalance } from "../api/assetAccounts";
import { listAccounts } from "../api";
import type { AssetAccountBalance, Account } from "../types";
import usePagination from "../usePagination";

export default function AssetAccountPage() {
  const [list, setList] = useState<AssetAccountBalance[]>([]);
  const { paginated, controls } = usePagination(list);
  const [accounts, setAccounts] = useState<Account[]>([]);
  const [msg, setMsg] = useState<{ type: "success" | "error"; text: string } | null>(null);
  const [open, setOpen] = useState(false);
  const [selected, setSelected] = useState<AssetAccountBalance | null>(null);
  const [balance, setBalance] = useState("0");

  const fetchData = () => listAssetAccounts().then((r) => setList(r.data));

  useEffect(() => {
    fetchData();
    listAccounts().then((r) => setAccounts(r.data));
  }, []);

  const columns: Column<AssetAccountBalance>[] = [
    { label: "ID", key: "asset_id" },
    {
      label: "Account",
      render: (_, row) => {
        const acc = accounts.find((a) => a.account_id === row.account_id);
        return acc ? `${acc.account_code} - ${acc.account_name}` : row.account_id;
      },
    },
    {
      label: "Balance",
      key: "balance",
      align: "right",
      render: (v) =>
        Number(v).toLocaleString("id-ID", { style: "currency", currency: "IDR" }),
    },
    {
      label: "",
      render: (_, row) => (
        <Button
          size="small"
          onClick={() => {
            setSelected(row);
            setBalance(String(row.balance));
            setOpen(true);
          }}
        >
          Adjust
        </Button>
      ),
    },
  ];

  const handleSave = async () => {
    if (!selected) return;
    try {
      await adjustAssetBalance(selected.asset_id, Number(balance));
      setOpen(false);
      setSelected(null);
      setBalance("0");
      fetchData();
      setMsg({ type: "success", text: "adjusted" });
    } catch (e: any) {
      setMsg({ type: "error", text: e.response?.data?.error || e.message });
    }
  };

  return (
    <div>
      <h2>Assets</h2>
      {msg && <Alert severity={msg.type}>{msg.text}</Alert>}
      <SortableTable columns={columns} data={paginated} />
      {controls}
      <Dialog open={open} onClose={() => setOpen(false)}>
        <DialogTitle>Adjust Balance</DialogTitle>
        <DialogContent sx={{ display: "flex", flexDirection: "column", gap: 1 }}>
          <TextField
            label="Balance"
            value={balance}
            onChange={(e) => setBalance(e.target.value)}
            type="number"
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOpen(false)}>Cancel</Button>
          <Button variant="contained" onClick={handleSave}>
            Save
          </Button>
        </DialogActions>
      </Dialog>
    </div>
  );
}
