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
import {
  listJournal,
  deleteJournal,
  createJournal,
  getJournalLines,
} from "../api/journal";
import { listAccounts } from "../api";
import type { JournalEntry, Account, JournalLineDetail } from "../types";
import usePagination from "../usePagination";
import { getCurrentMonthRange } from "../utils/date";

export default function JournalPage() {
  const [list, setList] = useState<JournalEntry[]>([]);
  const { paginated, controls } = usePagination(list);
  const [firstOfMonth, lastOfMonth] = getCurrentMonthRange();
  const [from, setFrom] = useState(firstOfMonth);
  const [to, setTo] = useState(lastOfMonth);
  const [search, setSearch] = useState("");
  const [open, setOpen] = useState(false);
  const [entryDate, setEntryDate] = useState(
    () => new Date().toISOString().split("T")[0],
  );
  const [description, setDescription] = useState("");
  const [lines, setLines] = useState<
    { account: string; debit: string; credit: string }[]
  >([{ account: "", debit: "", credit: "" }]);
  const [accounts, setAccounts] = useState<Account[]>([]);
  const [msg, setMsg] = useState<{
    type: "success" | "error";
    text: string;
  } | null>(null);
  const [detailOpen, setDetailOpen] = useState(false);
  const [detailLines, setDetailLines] = useState<JournalLineDetail[]>([]);
  const { paginated: linesPage, controls: lineControls } =
    usePagination(detailLines);
  const [detailEntry, setDetailEntry] = useState<JournalEntry | null>(null);
  const totalDebit = detailLines.reduce(
    (sum, l) => (l.is_debit ? sum + l.amount : sum),
    0,
  );
  const totalCredit = detailLines.reduce(
    (sum, l) => (!l.is_debit ? sum + l.amount : sum),
    0,
  );
  const lineColumns: Column<JournalLineDetail>[] = [
    { label: "Account", key: "account_name" },
    {
      label: "Debit",
      key: "amount",
      align: "right",
      render: (_, row) =>
        row.is_debit
          ? row.amount.toLocaleString("id-ID", {
              style: "currency",
              currency: "IDR",
            })
          : "",
    },
    {
      label: "Credit",
      key: "amount",
      align: "right",
      render: (_, row) =>
        !row.is_debit
          ? row.amount.toLocaleString("id-ID", {
              style: "currency",
              currency: "IDR",
            })
          : "",
    },
  ];
  const columns: Column<JournalEntry>[] = [
    { label: "ID", key: "journal_id" },
    {
      label: "Date",
      key: "entry_date",
      render: (v) => new Date(v).toLocaleDateString(),
    },
    { label: "Description", key: "description" },
    {
      label: "",
      render: (_, j) => (
        <>
          <Button
            size="small"
              onClick={() => {
                setDetailEntry(j);
                getJournalLines(j.journal_id).then((r) => {
                  setDetailLines(r.data as any);
                  setDetailOpen(true);
                });
              }}
          >
            Detail
          </Button>
          <Button
            size="small"
            onClick={() => {
              deleteJournal(j.journal_id).then(fetchData);
            }}
          >
            Del
          </Button>
        </>
      ),
    },
  ];
  const fetchData = () =>
    listJournal({
      from: from || undefined,
      to: to || undefined,
      q: search || undefined,
    }).then((r) => setList(r.data));
  useEffect(() => {
    fetchData();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [from, to, search]);
  useEffect(() => {
    listAccounts().then((r) => setAccounts(r.data));
  }, []);
  return (
    <div>
      <h2>Journal Entries</h2>
      <Button variant="contained" onClick={() => setOpen(true)} sx={{ mb: 2 }}>
        Add Journal
      </Button>
      <div style={{ display: "flex", gap: "0.5rem", marginBottom: "1rem" }}>
        <TextField
          label="Search"
          size="small"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
        />
        <TextField
          label="From"
          type="date"
          size="small"
          value={from}
          InputLabelProps={{ shrink: true }}
          onChange={(e) => setFrom(e.target.value)}
        />
        <TextField
          label="To"
          type="date"
          size="small"
          value={to}
          InputLabelProps={{ shrink: true }}
          onChange={(e) => setTo(e.target.value)}
        />
      </div>
      {msg && <Alert severity={msg.type}>{msg.text}</Alert>}
      <SortableTable
        columns={columns}
        data={paginated}
        defaultSort={{ key: "entry_date", direction: "desc" }}
      />
      {controls}
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
      <Dialog open={detailOpen} onClose={() => setDetailOpen(false)}>
        <DialogTitle>Journal Lines</DialogTitle>
        <DialogContent>
          {detailEntry && (
            <div style={{ marginBottom: "0.5rem" }}>
              <div>ID: {detailEntry.journal_id}</div>
              <div>
                Date: {new Date(detailEntry.entry_date).toLocaleDateString()}
              </div>
              <div>Description: {detailEntry.description}</div>
            </div>
          )}
          <SortableTable columns={lineColumns} data={linesPage} />
          <div style={{ marginTop: "0.5rem", fontWeight: "bold" }}>
            Total Debit:{" "}
            {totalDebit.toLocaleString("id-ID", {
              style: "currency",
              currency: "IDR",
            })}
            {" | Total Credit: "}
            {totalCredit.toLocaleString("id-ID", {
              style: "currency",
              currency: "IDR",
            })}
          </div>
          {lineControls}
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDetailOpen(false)}>Close</Button>
        </DialogActions>
      </Dialog>
    </div>
  );
}
