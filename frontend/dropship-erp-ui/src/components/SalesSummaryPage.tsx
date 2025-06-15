import { Alert } from "@mui/material";
import { LocalizationProvider } from "@mui/x-date-pickers";
import { DatePicker } from "@mui/x-date-pickers/DatePicker";
import { AdapterDateFns } from "@mui/x-date-pickers/AdapterDateFns";
import { useEffect, useState } from "react";
import { getCurrentMonthRange } from "../utils/date";
import {
  listJenisChannels,
  listStoresByChannelName,
  listShopeeSettled,
  fetchTopProducts,
} from "../api";
import type { JenisChannel, Store, ProductSales } from "../types";
import {
  LineChart,
  Line,
  CartesianGrid,
  XAxis,
  YAxis,
  Tooltip,
  BarChart,
  Bar,
} from "recharts";

export default function SalesSummaryPage() {
  const [channels, setChannels] = useState<JenisChannel[]>([]);
  const [channel, setChannel] = useState("");
  const [stores, setStores] = useState<Store[]>([]);
  const [store, setStore] = useState("");
  const [firstOfMonth, lastOfMonth] = getCurrentMonthRange();
  const [from, setFrom] = useState(firstOfMonth);
  const [to, setTo] = useState(lastOfMonth);
  const [data, setData] = useState<{ date: string; total: number }[]>([]);
  const [countData, setCountData] = useState<{ date: string; count: number }[]>(
    [],
  );
  const [totalRevenue, setTotalRevenue] = useState(0);
  const [totalOrders, setTotalOrders] = useState(0);
  const [topProducts, setTopProducts] = useState<ProductSales[]>([]);
  const [msg, setMsg] = useState<{
    type: "success" | "error";
    text: string;
  } | null>(null);

  useEffect(() => {
    listJenisChannels().then((res) => setChannels(res.data));
  }, []);

  useEffect(() => {
    if (channel) {
      listStoresByChannelName(channel).then((res) => setStores(res.data ?? []));
    } else {
      setStores([]);
    }
  }, [channel]);

  const fetchData = async () => {
    try {
      const res = await listShopeeSettled({
        channel: channel || undefined,
        store,
        from,
        to,
        page: 1,
        page_size: 1000,
      });
      const amountMap = new Map<string, number>();
      const countMap = new Map<string, number>();
      let totalAmt = 0;
      res.data.data.forEach((d) => {
        const dateStr =
          (d as any).waktu_pesanan_dibuat ?? (d as any).tanggal_dana_dilepaskan;
        const key = new Date(dateStr).toISOString().split("T")[0];
        amountMap.set(key, (amountMap.get(key) || 0) + d.total_penghasilan);
        countMap.set(key, (countMap.get(key) || 0) + 1);
        totalAmt += d.total_penghasilan;
      });
      const arr = Array.from(amountMap.entries()).sort((a, b) =>
        a[0] < b[0] ? -1 : 1,
      );
      const arrCount = Array.from(countMap.entries()).sort((a, b) =>
        a[0] < b[0] ? -1 : 1,
      );
      setData(arr.map(([date, total]) => ({ date, total })));
      setCountData(arrCount.map(([date, count]) => ({ date, count })));
      setTotalRevenue(totalAmt);
      setTotalOrders(res.data.data.length);
      const topRes = await fetchTopProducts({
        channel: channel || undefined,
        store,
        from,
        to,
        limit: 5,
      });
      setTopProducts(topRes.data);
      setMsg(null);
    } catch (e: any) {
      setMsg({ type: "error", text: e.response?.data?.error || e.message });
    }
  };

  useEffect(() => {
    fetchData();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [channel, store, from, to]);

  return (
    <div>
      <h2>Sales Summary</h2>
      <div style={{ display: "flex", gap: "0.5rem", marginBottom: "1rem" }}>
        <select
          aria-label="Channel"
          value={channel}
          onChange={(e) => setChannel(e.target.value)}
        >
          <option value="">All Channels</option>
          {channels.map((c) => (
            <option key={c.jenis_channel_id} value={c.jenis_channel}>
              {c.jenis_channel}
            </option>
          ))}
        </select>
        <select
          aria-label="Store"
          value={store}
          onChange={(e) => setStore(e.target.value)}
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
            label="From"
            format="yyyy-MM-dd"
            value={new Date(from)}
            onChange={(d) => {
              if (!d) return;
              setFrom(d.toISOString().split("T")[0]);
            }}
            slotProps={{ textField: { size: "small" } }}
          />
        </LocalizationProvider>
        <LocalizationProvider dateAdapter={AdapterDateFns}>
          <DatePicker
            label="To"
            format="yyyy-MM-dd"
            value={new Date(to)}
            onChange={(d) => {
              if (!d) return;
              setTo(d.toISOString().split("T")[0]);
            }}
            slotProps={{ textField: { size: "small" } }}
          />
        </LocalizationProvider>
      </div>
      {msg && (
        <Alert severity={msg.type} sx={{ mb: 2 }}>
          {msg.text}
        </Alert>
      )}
      <div style={{ marginBottom: "1rem" }}>
        <strong>Total Revenue:</strong>{" "}
        {totalRevenue.toLocaleString("id-ID", {
          style: "currency",
          currency: "IDR",
        })}{" "}|{" "}
        <strong>Total Orders:</strong> {totalOrders}
      </div>
      <LineChart
        width={600}
        height={300}
        data={data}
        style={{ marginBottom: "1rem" }}
      >
        <CartesianGrid strokeDasharray="3 3" />
        <XAxis dataKey="date" />
        <YAxis />
        <Tooltip />
        <Line type="monotone" dataKey="total" stroke="#8884d8" />
      </LineChart>
      <BarChart width={600} height={300} data={countData}>
        <CartesianGrid strokeDasharray="3 3" />
        <XAxis dataKey="date" />
        <YAxis />
        <Tooltip />
        <Bar dataKey="count" fill="#82ca9d" />
      </BarChart>
      {topProducts.length > 0 && (
        <div style={{ marginTop: "1rem" }}>
          <h3>Top Products</h3>
          <table style={{ width: "100%", borderCollapse: "collapse" }}>
            <thead>
              <tr>
                <th style={{ textAlign: "left" }}>Product</th>
                <th style={{ textAlign: "right" }}>Qty</th>
                <th style={{ textAlign: "right" }}>Value</th>
              </tr>
            </thead>
            <tbody>
              {topProducts.map((p) => (
                <tr key={p.nama_produk}>
                  <td>{p.nama_produk}</td>
                  <td style={{ textAlign: "right" }}>{p.total_qty}</td>
                  <td style={{ textAlign: "right" }}>
                    {p.total_value.toLocaleString("id-ID", {
                      style: "currency",
                      currency: "IDR",
                    })}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
