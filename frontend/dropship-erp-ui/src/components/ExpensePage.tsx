import {
  Alert,
  Button,
  TextField,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Autocomplete,
} from "@mui/material";
import SortableTable from "./SortableTable";
import type { Column } from "./SortableTable";
import { useEffect, useState } from "react";
import { createExpense, listExpenses, deleteExpense } from "../api/expenses";
import { listAccounts } from "../api";
import type { Expense, Account } from "../types";
import usePagination from "../usePagination";

export default function ExpensePage() {
  const [list, setList] = useState<Expense[]>([]);
  const { paginated, controls } = usePagination(list);
  const [desc, setDesc] = useState("");
  const [asset, setAsset] = useState("");
  const [lines, setLines] = useState<{ account: string; amount: string }[]>([
    { account: "", amount: "" },
  ]);
  const [accounts, setAccounts] = useState<Account[]>([]);
  const assetAccounts = accounts.filter((a) => a.account_type === "asset");
  const expenseAccounts = accounts.filter((a) => a.account_type === "expense");
  const [open, setOpen] = useState(false);
  const [msg, setMsg] = useState<{
    type: "success" | "error";
    text: string;
  } | null>(null);

  const fetchData = () => listExpenses().then((r) => setList(r.data.data));

  useEffect(() => {
    fetchData();
    listAccounts().then((r) => setAccounts(r.data));
  }, []);

  const handleCreate = async () => {
    try {
      await createExpense({
        description: desc,
        asset_account_id: Number(asset),
        lines: lines.map((l) => ({
          account_id: Number(l.account),
          amount: Number(l.amount),
        })),
        date: new Date().toISOString(),
      });
      setDesc("");
      setAsset("");
      setLines([{ account: "", amount: "" }]);
      setOpen(false);
      fetchData();
      setMsg({ type: "success", text: "saved" });
    } catch (e: any) {
      setMsg({ type: "error", text: e.message });
    }
  };

  const columns: Column<Expense>[] = [
    { label: "Description", key: "description" },
    {
      label: "Amount",
      key: "amount",
      align: "right",
      render: (v) =>
        Number(v).toLocaleString("id-ID", {
          style: "currency",
          currency: "IDR",
        }),
    },
    { label: "Asset", key: "asset_account_id" },
    {
      label: "Lines",
      render: (_, e) =>
        e.lines.map((l) => `${l.account_id}:${l.amount}`).join(", "),
    },
    {
      label: "",
      render: (_, e) => (
        <Button
          size="small"
          onClick={() => {
            deleteExpense(e.id).then(fetchData);
          }}
        >
          Del
        </Button>
      ),
    },
  ];

  return (
    <div>
      <h2>Expenses</h2>
      <Button variant="contained" onClick={() => setOpen(true)} sx={{ mb: 2 }}>
        Add Expense
      </Button>
      {msg && <Alert severity={msg.type}>{msg.text}</Alert>}
      <SortableTable columns={columns} data={paginated} />
      {controls}
      <Dialog open={open} onClose={() => setOpen(false)}>
        <DialogTitle>Add Expense</DialogTitle>
        <DialogContent
          sx={{ display: "flex", flexDirection: "column", gap: 1 }}
        >
          <TextField
            label="Description"
            value={desc}
            onChange={(e) => setDesc(e.target.value)}
            autoFocus
          />
          <Autocomplete
            options={assetAccounts}
            getOptionLabel={(a) => `${a.account_code} - ${a.account_name}`}
            isOptionEqualToValue={(o, v) => o.account_id === v.account_id}
            value={
              assetAccounts.find((a) => String(a.account_id) === asset) || null
            }
            onChange={(_, v) => setAsset(v ? String(v.account_id) : "")}
            renderInput={(params) => (
              <TextField {...params} label="Asset Account" />
            )}
            size="small"
          />
          {lines.map((ln, idx) => (
            <div key={idx} style={{ display: "flex", gap: 4 }}>
              <Autocomplete
                options={expenseAccounts}
                getOptionLabel={(a) => `${a.account_code} - ${a.account_name}`}
                isOptionEqualToValue={(o, v) => o.account_id === v.account_id}
                value={
                  expenseAccounts.find((a) => String(a.account_id) === ln.account) ||
                  null
                }
                onChange={(_, v) => {
                  const n = [...lines];
                  n[idx].account = v ? String(v.account_id) : "";
                  setLines(n);
                }}
                renderInput={(params) => (
                  <TextField {...params} label="Expense Account" size="small" />
                )}
                sx={{ width: 220 }}
              />
              <TextField
                label="Amount"
                value={ln.amount}
                onChange={(e) => {
                  const n = [...lines];
                  n[idx].amount = e.target.value;
                  setLines(n);
                }}
                size="small"
              />
            </div>
          ))}
          <Button
            onClick={() => setLines([...lines, { account: "", amount: "" }])}
          >
            Add Line
          </Button>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOpen(false)}>Cancel</Button>
          <Button variant="contained" onClick={handleCreate}>
            Save
          </Button>
        </DialogActions>
      </Dialog>
    </div>
  );
}
