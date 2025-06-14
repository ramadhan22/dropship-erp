import { useEffect, useState } from "react";
import {
  Table,
  TableHead,
  TableRow,
  TableCell,
  TableBody,
  Button,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  Alert,
} from "@mui/material";
import { listJournal, deleteJournal, createJournal } from "../api/journal";
import { listAccounts } from "../api";
import type { JournalEntry, Account } from "../types";

export default function JournalPage() {
  const [list, setList] = useState<JournalEntry[]>([]);
  const [open, setOpen] = useState(false);
  const [entryDate, setEntryDate] = useState("");
  const [description, setDescription] = useState("");
  const [lines, setLines] = useState<
    { account: string; debit: string; credit: string }[]
  >([{ account: "", debit: "", credit: "" }]);
  const [accounts, setAccounts] = useState<Account[]>([]);
  const [msg, setMsg] = useState<{
    type: "success" | "error";
    text: string;
  } | null>(null);
  const fetchData = () => listJournal().then((r) => setList(r.data));
  useEffect(() => {
    fetchData();
    listAccounts().then((r) => setAccounts(r.data));
  }, []);
  return (
    <div>
      <h2>Journal Entries</h2>
      <Button variant="contained" onClick={() => setOpen(true)} sx={{ mb: 2 }}>
        Add Journal
      </Button>
      {msg && <Alert severity={msg.type}>{msg.text}</Alert>}
      <Table size="small">
        <TableHead>
          <TableRow>
            <TableCell>ID</TableCell>
            <TableCell>Date</TableCell>
            <TableCell>Description</TableCell>
            <TableCell></TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {list.map((j) => (
            <TableRow key={j.journal_id}>
              <TableCell>{j.journal_id}</TableCell>
              <TableCell>
                {new Date(j.entry_date).toLocaleDateString()}
              </TableCell>
              <TableCell>{j.description}</TableCell>
              <TableCell>
                <Button
                  size="small"
                  onClick={() => {
                    deleteJournal(j.journal_id).then(fetchData);
                  }}
                >
                  Del
                </Button>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
      <Dialog open={open} onClose={() => setOpen(false)}>
        <DialogTitle>Add Journal</DialogTitle>
        <DialogContent
          sx={{ display: "flex", flexDirection: "column", gap: 1 }}
        >
          <TextField
            label="Date"
            type="date"
            value={entryDate}
            onChange={(e) => setEntryDate(e.target.value)}
            InputLabelProps={{ shrink: true }}
          />
          <TextField
            label="Description"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
          />
          {lines.map((l, idx) => (
            <div key={idx} style={{ display: "flex", gap: "0.5rem" }}>
              <select
                aria-label="Account"
                value={l.account}
                onChange={(e) => {
                  const arr = [...lines];
                  arr[idx].account = e.target.value;
                  setLines(arr);
                }}
                style={{ fontSize: "0.875rem" }}
              >
                <option value="">Select Account</option>
                {accounts.map((a) => (
                  <option key={a.account_id} value={String(a.account_id)}>
                    {a.account_code} - {a.account_name}
                  </option>
                ))}
              </select>
              <TextField
                label="Debit"
                value={l.debit}
                onChange={(e) => {
                  const arr = [...lines];
                  arr[idx].debit = e.target.value;
                  setLines(arr);
                }}
                size="small"
              />
              <TextField
                label="Credit"
                value={l.credit}
                onChange={(e) => {
                  const arr = [...lines];
                  arr[idx].credit = e.target.value;
                  setLines(arr);
                }}
                size="small"
              />
            </div>
          ))}
          <Button
            onClick={() =>
              setLines([...lines, { account: "", debit: "", credit: "" }])
            }
          >
            Add Line
          </Button>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOpen(false)}>Cancel</Button>
          <Button
            variant="contained"
            onClick={async () => {
              try {
                const entry = {
                  entry_date:
                    entryDate || new Date().toISOString().split("T")[0],
                  description: description || null,
                  source_type: "manual",
                  source_id: String(Date.now()),
                  shop_username: "",
                };
                const jl = lines.map((l) => {
                  if (l.debit) {
                    return {
                      account_id: Number(l.account),
                      is_debit: true,
                      amount: Number(l.debit),
                    };
                  }
                  return {
                    account_id: Number(l.account),
                    is_debit: false,
                    amount: Number(l.credit),
                  };
                });
                await createJournal({ entry, lines: jl });
                setOpen(false);
                setEntryDate("");
                setDescription("");
                setLines([{ account: "", debit: "", credit: "" }]);
                fetchData();
                setMsg({ type: "success", text: "saved" });
              } catch (e: any) {
                setMsg({
                  type: "error",
                  text: e.response?.data?.error || e.message,
                });
              }
            }}
          >
            Save
          </Button>
        </DialogActions>
      </Dialog>
    </div>
  );
}
