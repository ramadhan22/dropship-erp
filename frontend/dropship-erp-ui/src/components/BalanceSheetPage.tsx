// File: src/components/BalanceSheetPage.tsx

import {
  Button,
  Card,
  CardContent,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableRow,
  Typography,
} from "@mui/material";
import { LocalizationProvider } from "@mui/x-date-pickers";
import { DatePicker } from "@mui/x-date-pickers/DatePicker";
import { AdapterDateFns } from "@mui/x-date-pickers/AdapterDateFns";
import { useEffect, useState } from "react";
import { fetchBalanceSheet, listAllStores } from "../api";
import type { BalanceCategory, Store, Account } from "../types";
import usePagination from "../usePagination";

function AccountTable({ accounts }: { accounts: Account[] }) {
  const { paginated, controls } = usePagination(accounts);
  return (
    <>
      <Table size="small">
        <TableHead>
          <TableRow>
            <TableCell>Code</TableCell>
            <TableCell>Name</TableCell>
            <TableCell align="right">Balance</TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {paginated.map((a) => (
            <TableRow key={a.account_id}>
              <TableCell>{a.account_code}</TableCell>
              <TableCell>{a.account_name}</TableCell>
              <TableCell align="right">
                {a.balance.toLocaleString("id-ID", {
                  style: "currency",
                  currency: "IDR",
                })}
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
      {controls}
    </>
  );
}

export default function BalanceSheetPage() {
  const [shop, setShop] = useState("");
  const [stores, setStores] = useState<Store[]>([]);
  const [period, setPeriod] = useState(
    new Date().toISOString().slice(0, 7),
  );
  const [data, setData] = useState<BalanceCategory[]>([]);

  useEffect(() => {
    listAllStores().then((s) => setStores(s));
  }, []);

  const handleFetch = async () => {
    const res = await fetchBalanceSheet(shop, period);
    setData(res.data);
  };

  return (
    <div>
      <h2>Balance Sheet</h2>
      <select
        aria-label="Shop"
        value={shop}
        onChange={(e) => setShop(e.target.value)}
        style={{ marginRight: "0.5rem" }}
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
          label="Period (YYYY-MM)"
          views={["year", "month"]}
          openTo="month"
          format="yyyy-MM"
          value={period ? new Date(period) : null}
          onChange={(date) => {
            if (!date) return;
            setPeriod(date.toISOString().slice(0, 7));
          }}
          slotProps={{ textField: { size: "small", sx: { mr: 2 }, InputLabelProps: { shrink: true } } }}
        />
      </LocalizationProvider>
      <Button variant="contained" onClick={handleFetch}>
        Fetch
      </Button>

      <div style={{ marginTop: "1rem" }}>
        {data.map((cat) => (
          <Card key={cat.category} sx={{ mb: 2 }}>
            <CardContent>
              <Typography variant="h6">
                {cat.category} (Total: {cat.total.toLocaleString("id-ID", {
                  style: "currency",
                  currency: "IDR",
                })})
              </Typography>
              <AccountTable accounts={cat.accounts} />
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  );
}
