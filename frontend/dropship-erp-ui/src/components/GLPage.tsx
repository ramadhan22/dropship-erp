import { Fragment, useEffect, useState } from "react";
import {
  Button,
  Table,
  TableHead,
  TableRow,
  TableCell,
  TableBody,
} from "@mui/material";
import { LocalizationProvider } from "@mui/x-date-pickers";
import { DatePicker } from "@mui/x-date-pickers/DatePicker";
import { AdapterDateFns } from "@mui/x-date-pickers/AdapterDateFns";
import { fetchGeneralLedger } from "../api/gl";
import { listAllStores } from "../api";
import type { Account, Store } from "../types";

interface AccountGroup {
  type: string;
  accounts: Account[];
}

function groupByType(data: Account[]): AccountGroup[] {
  const groups: Record<string, Account[]> = {};
  data.forEach((a) => {
    if (!groups[a.account_type]) {
      groups[a.account_type] = [];
    }
    groups[a.account_type].push(a);
  });
  return Object.keys(groups).map((t) => ({ type: t, accounts: groups[t] }));
}

export default function GLPage() {
  const [shop, setShop] = useState("");
  const [stores, setStores] = useState<Store[]>([]);
  const now = new Date();
  const firstOfMonth = new Date(now.getFullYear(), now.getMonth(), 1)
    .toISOString()
    .split("T")[0];
  const lastOfMonth = new Date(now.getFullYear(), now.getMonth() + 1, 0)
    .toISOString()
    .split("T")[0];
  const [from, setFrom] = useState(firstOfMonth);
  const [to, setTo] = useState(lastOfMonth);
  const [data, setData] = useState<Account[]>([]);

  const totalDebit = data
    .filter((a) => a.balance > 0)
    .reduce((sum, a) => sum + a.balance, 0);
  const totalCredit = data
    .filter((a) => a.balance < 0)
    .reduce((sum, a) => sum + -a.balance, 0);

  useEffect(() => {
    listAllStores().then((s) => setStores(s));
  }, []);

  useEffect(() => {
    handleFetch();
  }, [shop]);

  const handleFetch = async () => {
    const res = await fetchGeneralLedger(shop, from, to);
    setData(res.data);
  };

  return (
    <div>
      <h2>General Ledger</h2>
      <div style={{ display: "flex", gap: "0.5rem" }}>
        <select
          aria-label="Shop"
          value={shop}
          onChange={(e) => setShop(e.target.value)}
        >
          <option value="">Select Store</option>
          {stores.map((s) => (
            <option key={s.store_id} value={s.nama_toko}>
              {s.nama_toko}
            </option>
          ))}
        </select>
        <LocalizationProvider dateAdapter={AdapterDateFns}>
          <DatePicker
            label="From"
            format="yyyy-MM-dd"
            value={new Date(from)}
            onChange={(date) => {
              if (!date) return;
              setFrom(date.toISOString().split("T")[0]);
            }}
            slotProps={{ textField: { size: "small" } }}
          />
        </LocalizationProvider>
        <LocalizationProvider dateAdapter={AdapterDateFns}>
          <DatePicker
            label="To"
            format="yyyy-MM-dd"
            value={new Date(to)}
            onChange={(date) => {
              if (!date) return;
              setTo(date.toISOString().split("T")[0]);
            }}
            slotProps={{ textField: { size: "small" } }}
          />
        </LocalizationProvider>
        <Button onClick={handleFetch}>Fetch</Button>
      </div>
      <Table size="small">
        <TableHead>
          <TableRow>
            <TableCell>Code</TableCell>
            <TableCell>Account</TableCell>
            <TableCell align="right">Debit</TableCell>
            <TableCell align="right">Credit</TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {groupByType(data).map((grp) => (
            <Fragment key={grp.type}>
              <TableRow>
                <TableCell colSpan={4} style={{ fontWeight: "bold" }}>
                  {grp.type}
                </TableCell>
              </TableRow>
              {grp.accounts.map((a) => (
                <TableRow key={a.account_id}>
                  <TableCell>{a.account_code}</TableCell>
                  <TableCell>{a.account_name}</TableCell>
                  <TableCell align="right">
                    {a.balance > 0
                      ? a.balance.toLocaleString("id-ID", {
                          style: "currency",
                          currency: "IDR",
                        })
                      : ""}
                  </TableCell>
                  <TableCell align="right">
                    {a.balance < 0
                      ? (-a.balance).toLocaleString("id-ID", {
                          style: "currency",
                          currency: "IDR",
                        })
                      : ""}
                  </TableCell>
                </TableRow>
              ))}
            </Fragment>
          ))}
          <TableRow>
            <TableCell colSpan={2} style={{ fontWeight: "bold" }}>
              Total
            </TableCell>
            <TableCell align="right" style={{ fontWeight: "bold" }}>
              {totalDebit.toLocaleString("id-ID", {
                style: "currency",
                currency: "IDR",
              })}
            </TableCell>
            <TableCell align="right" style={{ fontWeight: "bold" }}>
              {totalCredit.toLocaleString("id-ID", {
                style: "currency",
                currency: "IDR",
              })}
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>
    </div>
  );
}
