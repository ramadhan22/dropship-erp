import {
  Alert,
  Button,
  TextField,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
} from "@mui/material";
import SortableTable, { Column } from "./SortableTable";
import { useEffect, useState } from "react";
import { createExpense, listExpenses, deleteExpense } from "../api/expenses";
import type { Expense } from "../types";
import usePagination from "../usePagination";

export default function ExpensePage() {
  const [list, setList] = useState<Expense[]>([]);
  const { paginated, controls } = usePagination(list);
  const [desc, setDesc] = useState("");
  const [amount, setAmount] = useState("");
  const [account, setAccount] = useState("");
  const [open, setOpen] = useState(false);
  const [msg, setMsg] = useState<{
    type: "success" | "error";
    text: string;
  } | null>(null);

  const fetchData = () => listExpenses().then((r) => setList(r.data));

  useEffect(() => {
    fetchData();
  }, []);

  const handleCreate = async () => {
    try {
      await createExpense({
        description: desc,
        amount: Number(amount),
        account_id: Number(account),
        date: new Date().toISOString(),
      });
      setDesc("");
      setAmount("");
      setAccount("");
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
        Number(v).toLocaleString("id-ID", { style: "currency", currency: "IDR" }),
    },
    { label: "Account", key: "account_id" },
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
        <DialogContent sx={{ display: "flex", flexDirection: "column", gap: 1 }}>
          <TextField
            label="Description"
            value={desc}
            onChange={(e) => setDesc(e.target.value)}
            autoFocus
          />
          <TextField
            label="Amount"
            value={amount}
            onChange={(e) => setAmount(e.target.value)}
          />
          <TextField
            label="Account"
            value={account}
            onChange={(e) => setAccount(e.target.value)}
          />
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
