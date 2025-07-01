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
import { LocalizationProvider } from "@mui/x-date-pickers";
import { DatePicker } from "@mui/x-date-pickers/DatePicker";
import { AdapterDateFns } from "@mui/x-date-pickers/AdapterDateFns";
import SortableTable from "./SortableTable";
import type { Column } from "./SortableTable";
import { useEffect, useState } from "react";
import {
  createExpense,
  listExpenses,
  deleteExpense,
  getExpense,
  updateExpense,
} from "../api/expenses";
import { getJournalLinesBySource } from "../api/journal";
import { listAccounts } from "../api";
import type { Expense, Account, JournalEntryWithLines } from "../types";
import usePagination from "../usePagination";

export default function ExpensePage() {
  const [list, setList] = useState<Expense[]>([]);
  const { paginated, controls } = usePagination(list);
  const [desc, setDesc] = useState("");
  const [asset, setAsset] = useState("");
  const [date, setDate] = useState(
    new Date().toISOString().split("T")[0]
  );
  const [lines, setLines] = useState<{ account: string; amount: string }[]>([
    { account: "", amount: "" },
  ]);
  const [accounts, setAccounts] = useState<Account[]>([]);
  const assetAccounts = accounts.filter((a) => a.account_type === "Asset");
  const expenseAccounts = accounts.filter((a) => a.account_type === "Expense");
  const [open, setOpen] = useState(false);
  const [editing, setEditing] = useState<Expense | null>(null);
  const [detailOpen, setDetailOpen] = useState(false);
  const [detailExpense, setDetailExpense] = useState<Expense | null>(null);
  const [journal, setJournal] = useState<JournalEntryWithLines[]>([]);
  const [msg, setMsg] = useState<{
    type: "success" | "error";
    text: string;
  } | null>(null);

  const fetchData = () => listExpenses().then((r) => setList(r.data.data));

  useEffect(() => {
    fetchData();
    listAccounts().then((r) => setAccounts(r.data));
  }, []);

  const handleSave = async () => {
    try {
      if (editing) {
        await updateExpense(editing.id, {
          description: desc,
          asset_account_id: Number(asset),
          lines: lines.map((l) => ({
            account_id: Number(l.account),
            amount: Number(l.amount),
          })),
          date: new Date(date).toISOString(),
        });
      } else {
        await createExpense({
          description: desc,
          asset_account_id: Number(asset),
          lines: lines.map((l) => ({
            account_id: Number(l.account),
            amount: Number(l.amount),
          })),
          date: new Date(date).toISOString(),
        });
      }
      setDesc("");
      setAsset("");
      setLines([{ account: "", amount: "" }]);
      setOpen(false);
      setEditing(null);
      fetchData();
      setMsg({ type: "success", text: "saved" });
    } catch (e: any) {
      setMsg({ type: "error", text: e.message });
    }
  };

  const columns: Column<Expense>[] = [
    {
      label: "Date",
      key: "date",
      render: (v) => new Date(v).toLocaleDateString(),
    },
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
      label: "",
      render: (_, e) => (
        <>
          <Button
            size="small"
            onClick={() => {
              setEditing(e);
              setDesc(e.description);
              setAsset(String(e.asset_account_id));
              setDate(e.date.split("T")[0]);
              setLines(e.lines.map((l) => ({ account: String(l.account_id), amount: String(l.amount) })));
              setOpen(true);
            }}
          >
            Edit
          </Button>
          <Button
            size="small"
            onClick={() => {
              setDetailExpense(e);
              getJournalLinesBySource(e.id).then((r) => {
                setJournal(r.data);
                setDetailOpen(true);
              });
            }}
          >
            Detail
          </Button>
          <Button
            size="small"
            onClick={() => {
              deleteExpense(e.id).then(fetchData);
            }}
          >
            Del
          </Button>
        </>
      ),
    },
  ];

  return (
    <div>
      <h2>Expenses</h2>
      <Button
        variant="contained"
        onClick={() => {
          setEditing(null);
          setDesc("");
          setAsset("");
          setDate(new Date().toISOString().split("T")[0]);
          setLines([{ account: "", amount: "" }]);
          setOpen(true);
        }}
        sx={{ mb: 2 }}
      >
        Add Expense
      </Button>
      {msg && <Alert severity={msg.type}>{msg.text}</Alert>}
      <SortableTable
        columns={columns}
        data={paginated}
        defaultSort={{ key: "date", direction: "desc" }}
      />
      {controls}
      <Dialog
        open={open}
        onClose={() => {
          setOpen(false);
          setEditing(null);
        }}
      >
        <DialogTitle>{editing ? "Edit Expense" : "Add Expense"}</DialogTitle>
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
        <LocalizationProvider dateAdapter={AdapterDateFns}>
          <DatePicker
            label="Date"
            format="yyyy-MM-dd"
            value={new Date(date)}
            onChange={(d) => {
              if (!d) return;
              setDate(d.toISOString().split("T")[0]);
            }}
            slotProps={{ textField: { size: "small" } }}
          />
        </LocalizationProvider>
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
          <Button
            onClick={() => {
              setOpen(false);
              setEditing(null);
            }}
          >
            Cancel
          </Button>
          <Button variant="contained" onClick={handleSave}>
            Save
          </Button>
        </DialogActions>
      </Dialog>
      <Dialog open={detailOpen} onClose={() => setDetailOpen(false)}>
        <DialogTitle>Expense Detail</DialogTitle>
        <DialogContent>
          {detailExpense && (
            <div style={{ marginBottom: "0.5rem" }}>
              <div>Description: {detailExpense.description}</div>
              <div>Amount: {detailExpense.amount}</div>
            </div>
          )}
          {journal.map((j) => (
            <div key={j.entry.journal_id} style={{ marginBottom: "0.5rem" }}>
              <div>
                <strong>Journal {j.entry.journal_id}</strong> - {new Date(j.entry.entry_date).toLocaleDateString()}
              </div>
              <ul>
                {j.lines.map((l) => (
                  <li key={l.line_id}>
                    {l.account_name} - {l.is_debit ? "D" : "C"} {l.amount}
                  </li>
                ))}
              </ul>
            </div>
          ))}
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDetailOpen(false)}>Close</Button>
        </DialogActions>
      </Dialog>
    </div>
  );
}
