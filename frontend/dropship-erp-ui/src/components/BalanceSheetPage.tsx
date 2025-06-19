// File: src/components/BalanceSheetPage.tsx

import { Button, Card, CardContent, Typography } from "@mui/material";
import { LocalizationProvider } from "@mui/x-date-pickers";
import { DatePicker } from "@mui/x-date-pickers/DatePicker";
import { AdapterDateFns } from "@mui/x-date-pickers/AdapterDateFns";
import { useEffect, useState } from "react";
import { fetchBalanceSheet, listAllStores } from "../api";
import { fetchProfitLoss } from "../api/pl";
import type { BalanceCategory, Store, Account } from "../types";
import usePagination from "../usePagination";
import SortableTable from "./SortableTable";
import type { Column } from "./SortableTable";

function AccountTable({ accounts }: { accounts: Account[] }) {
  const { paginated, controls } = usePagination(accounts);
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
    <>
      <SortableTable columns={columns} data={paginated} />
      {controls}
    </>
  );
}

export default function BalanceSheetPage() {
  const [shop, setShop] = useState("");
  const [stores, setStores] = useState<Store[]>([]);
  const [period, setPeriod] = useState(new Date().toISOString().slice(0, 7));
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
    const [bsRes, plRes] = await Promise.all([
      fetchBalanceSheet(shop, period),
      fetchProfitLoss({
        type: "Yearly",
        year: Number(period.slice(0, 4)),
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
      <LocalizationProvider dateAdapter={AdapterDateFns}>
        <DatePicker
          label="Period"
          views={["year", "month"]}
          openTo="month"
          format="yyyy-MM"
          value={new Date(period)}
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

      <div style={{ marginTop: "1rem", display: "flex", gap: "1rem" }}>
        <div style={{ flex: 1 }}>
          <Card sx={{ mb: 2 }}>
            <CardContent>
              <Typography variant="h6">Assets</Typography>
              <AccountTable accounts={assetCat?.accounts ?? []} />
            </CardContent>
          </Card>
          <Typography variant="subtitle1" sx={{ fontWeight: "bold", textAlign: "right" }}>
            Total Assets: {format(assetCat?.total ?? 0)}
          </Typography>
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
          <Typography variant="subtitle1" sx={{ fontWeight: "bold", textAlign: "right" }}>
            Total Liabilities + Equity:{" "}
            {format((liabilityCat?.total ?? 0) + equityTotal)}
          </Typography>
        </div>
      </div>
    </div>
  );
}
