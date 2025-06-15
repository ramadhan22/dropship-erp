import { useEffect, useState } from "react";
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
import { fetchPL } from "../api/pl";
import { listAllStores } from "../api";
import type { Metric, Store } from "../types";
import usePagination from "../usePagination";

export default function PLPage() {
  const [shop, setShop] = useState("");
  const [stores, setStores] = useState<Store[]>([]);
  const [period, setPeriod] = useState(new Date().toISOString().slice(0, 7));
  const [data, setData] = useState<Metric | null>(null);
  const { paginated, controls } = usePagination(
    data ? Object.entries(data) : [],
  );

  useEffect(() => {
    listAllStores().then((s) => setStores(s));
  }, []);

  useEffect(() => {
    handleFetch();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const handleFetch = async () => {
    const res = await fetchPL(shop, period);
    setData(res.data);
  };

  return (
    <div>
      <h2>Profit & Loss</h2>
      <div style={{ display: "flex", gap: "0.5rem" }}>
        <select
          aria-label="Shop"
          value={shop}
          onChange={(e) => setShop(e.target.value)}
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
            slotProps={{ textField: { size: "small" } }}
          />
        </LocalizationProvider>
        <Button onClick={handleFetch}>Fetch</Button>
      </div>
      {data && (
        <Table size="small">
          <TableHead>
            <TableRow>
              <TableCell>Metric</TableCell>
              <TableCell align="right">Value</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {paginated.map(([k, v]) => (
              <TableRow key={k}>
                <TableCell>{k}</TableCell>
                <TableCell align="right">
                  {typeof v === "number"
                    ? v.toLocaleString("id-ID", {
                        style: "currency",
                        currency: "IDR",
                      })
                    : (v as any)}
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      )}
      {controls}
    </div>
  );
}
