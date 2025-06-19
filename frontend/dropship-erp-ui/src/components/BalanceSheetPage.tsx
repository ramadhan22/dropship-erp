// File: src/components/BalanceSheetPage.tsx

import {
  Button,
  Card,
  CardContent,
  Typography,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
} from "@mui/material";
import { useEffect, useState } from "react";
import { fetchBalanceSheet, listAllStores } from "../api";
import { fetchProfitLoss } from "../api/pl";
import type { BalanceCategory, Store, Account } from "../types";
import SortableTable from "./SortableTable";
import type { Column } from "./SortableTable";

function AccountTable({ accounts }: { accounts: Account[] }) {
  const columns: Column<Account>[] = [
    { label: "Code", key: "account_code" },
    { label: "Name", key: "account_name" },
    {
      label: "Balance",
      key: "balance",
      align: "right",
      render: (v) =>
        Number(v).toLocaleString("id-ID", { style: "currency", currency: "IDR" }),
    },
  ];
  return (
    <SortableTable columns={columns} data={accounts} />
  );
}

export default function BalanceSheetPage() {
  const [shop, setShop] = useState("");
  const [stores, setStores] = useState<Store[]>([]);
  const now = new Date();
  const [periodType, setPeriodType] = useState<"Monthly" | "Yearly">("Monthly");
  const [month, setMonth] = useState(now.getMonth() + 1);
  const [year, setYear] = useState(now.getFullYear());

  const months = [
    "Jan",
    "Feb",
    "Mar",
    "Apr",
    "May",
    "Jun",
    "Jul",
    "Aug",
    "Sep",
    "Oct",
    "Nov",
    "Dec",
  ];
  const years = [2023, 2024, 2025];
  const [data, setData] = useState<BalanceCategory[]>([]);
  const [netProfit, setNetProfit] = useState(0);

  useEffect(() => {
    listAllStores().then((s) => setStores(s));
  }, []);

  useEffect(() => {
    handleFetch();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const handleFetch = async () => {
    const periodStr =
      periodType === "Monthly"
        ? `${year}-${String(month).padStart(2, "0")}`
        : `${year}-12`;
    const [bsRes, plRes] = await Promise.all([
      fetchBalanceSheet(shop, periodStr),
      fetchProfitLoss({
        type: periodType,
        month: periodType === "Monthly" ? month : undefined,
        year,
        store: shop,
      }),
    ]);
    setData(bsRes.data);
    setNetProfit(plRes.data.labaRugiBersih.amount);
  };

  const assetCat = data.find((c) => c.category === "Assets");
  const liabilityCat = data.find((c) => c.category === "Liabilities");
  const equityCat = data.find((c) => c.category === "Equity");
  const profitName = "Laba/Rugi Tahun Berjalan";
  const equityAccounts = (() => {
    if (!equityCat) return [] as Account[];
    const list = equityCat.accounts.map((a) =>
      a.account_name === profitName ? { ...a, balance: netProfit } : a,
    );
    if (!list.some((a) => a.account_name === profitName)) {
      list.push({
        account_id: 0,
        account_code: "3.3",
        account_name: profitName,
        account_type: "Equity",
        parent_id: null,
        balance: netProfit,
      });
    }
    return list;
  })();
  const equityTotal = equityAccounts.reduce((sum, a) => sum + a.balance, 0);
  const format = (n: number) =>
    n.toLocaleString("id-ID", { style: "currency", currency: "IDR" });

  return (
    <div>
      <h2>Balance Sheet</h2>
      <select
        aria-label="Shop"
        value={shop}
        onChange={(e) => setShop(e.target.value)}
        style={{ marginRight: "0.5rem" }}
      >
        <option value="">All Stores</option>
        {stores.map((s) => (
          <option key={s.store_id} value={s.nama_toko}>
            {s.nama_toko}
          </option>
        ))}
      </select>
      <Typography sx={{ minWidth: 70 }}>Periode:</Typography>
        <FormControl size="small" sx={{ mr: 1 }}>
          <InputLabel id="type-label">Type</InputLabel>
          <Select
            labelId="type-label"
            value={periodType}
            label="Type"
            onChange={(e) => setPeriodType(e.target.value as any)}
          >
            <MenuItem value="Monthly">Monthly</MenuItem>
            <MenuItem value="Yearly">Yearly</MenuItem>
          </Select>
        </FormControl>
        {periodType === "Monthly" && (
          <FormControl size="small" sx={{ mr: 1 }}>
            <InputLabel id="month-label">Month</InputLabel>
            <Select
              labelId="month-label"
              value={month}
              label="Month"
              onChange={(e) => setMonth(Number(e.target.value))}
            >
              {months.map((m, idx) => (
                <MenuItem key={m} value={idx + 1}>
                  {m}
                </MenuItem>
              ))}
            </Select>
          </FormControl>
        )}
        <FormControl size="small" sx={{ mr: 1 }}>
          <InputLabel id="year-label">Year</InputLabel>
          <Select
            labelId="year-label"
            value={year}
            label="Year"
            onChange={(e) => setYear(Number(e.target.value))}
          >
            {years.map((y) => (
              <MenuItem key={y} value={y}>
                {y}
              </MenuItem>
            ))}
          </Select>
        </FormControl>
      <Button variant="contained" onClick={handleFetch}>
        Fetch
      </Button>

      <div style={{ marginTop: "1rem", display: "flex", gap: "1rem" }}>
        <div style={{ flex: 1 }}>
          <Card sx={{ mb: 2 }}>
            <CardContent>
              <Typography variant="h6">Assets</Typography>
              <AccountTable accounts={assetCat?.accounts ?? []} />
            </CardContent>
          </Card>
        </div>
        <div style={{ flex: 1 }}>
          <Card sx={{ mb: 2 }}>
            <CardContent>
              <Typography variant="h6">Liabilities</Typography>
              <AccountTable accounts={liabilityCat?.accounts ?? []} />
            </CardContent>
          </Card>
          <Card sx={{ mb: 2 }}>
            <CardContent>
              <Typography variant="h6">Equity</Typography>
              <AccountTable accounts={equityAccounts} />
            </CardContent>
          </Card>
        </div>
      </div>
      <div style={{ display: "flex", marginTop: "0.5rem" }}>
        <Typography variant="subtitle1" sx={{ flex: 1, fontWeight: "bold", textAlign: "right" }}>
          Total Assets: {format(assetCat?.total ?? 0)}
        </Typography>
        <Typography variant="subtitle1" sx={{ flex: 1, fontWeight: "bold", textAlign: "right" }}>
          Total Liabilities + Equity: {format((liabilityCat?.total ?? 0) + equityTotal)}
        </Typography>
      </div>
    </div>
  );
}
