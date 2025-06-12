import {
  Alert,
  Button,
  TextField,
  Table,
  TableHead,
  TableRow,
  TableCell,
  TableBody,
} from "@mui/material";
import { useEffect, useState } from "react";
import { createExpense, listExpenses, deleteExpense } from "../api/expenses";
import type { Expense } from "../types";

export default function ExpensePage() {
  const [list, setList] = useState<Expense[]>([]);
  const [desc, setDesc] = useState("");
  const [amount, setAmount] = useState("");
  const [account, setAccount] = useState("");
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
      fetchData();
      setMsg({ type: "success", text: "saved" });
    } catch (e: any) {
      setMsg({ type: "error", text: e.message });
    }
  };

  return (
    <div>
      <h2>Expenses</h2>
      <div style={{ display: "flex", gap: "0.5rem", marginBottom: "1rem" }}>
        <TextField
          label="Description"
          value={desc}
          onChange={(e) => setDesc(e.target.value)}
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
        <Button variant="contained" onClick={handleCreate}>
          Add
        </Button>
      </div>
      {msg && <Alert severity={msg.type}>{msg.text}</Alert>}
      <Table size="small">
        <TableHead>
          <TableRow>
            <TableCell>Description</TableCell>
            <TableCell>Amount</TableCell>
            <TableCell></TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {list.map((e) => (
            <TableRow key={e.id}>
              <TableCell>{e.description}</TableCell>
              <TableCell>{e.amount}</TableCell>
              <TableCell>
                <Button
                  size="small"
                  onClick={() => {
                    deleteExpense(e.id).then(fetchData);
                  }}
                >
                  Del
                </Button>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  );
}
