import {
  Alert,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
} from "@mui/material";
import { LocalizationProvider } from "@mui/x-date-pickers";
import { DatePicker } from "@mui/x-date-pickers/DatePicker";
import { AdapterDateFns } from "@mui/x-date-pickers/AdapterDateFns";
import { useEffect, useState } from "react";
import { getCurrentMonthRange } from "../utils/date";
import { listShopeeAdjustments } from "../api/shopeeAdjustments";
import { getJournalLinesBySource } from "../api/journal";
import {
  listJenisChannels,
  listStoresByChannelName,
  fetchDailyPurchaseTotals,
  fetchMonthlyPurchaseTotals,
  fetchTopProducts,
  fetchCancelledSummary,
  getShopeeSettleDetail,
} from "../api";
import type {
  JenisChannel,
  Store,
  ProductSales,
  ShopeeAdjustment,
  JournalLineDetail,
  ShopeeSettled,
  CancelledSummary,
} from "../types";
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
  const [period, setPeriod] = useState<"Daily" | "Monthly">("Daily");
  const [data, setData] = useState<{ date: string; total: number }[]>([]);
  const [countData, setCountData] = useState<{ date: string; count: number }[]>(
    [],
  );
  const [totalRevenue, setTotalRevenue] = useState(0);
  const [totalOrders, setTotalOrders] = useState(0);
  const [topProducts, setTopProducts] = useState<ProductSales[]>([]);
  const [cancelSummary, setCancelSummary] = useState<CancelledSummary>({
    count: 0,
    biaya_mitra: 0,
  });
  const [adjustments, setAdjustments] = useState<ShopeeAdjustment[]>([]);
  const [lines, setLines] = useState<JournalLineDetail[]>([]);
  const [detailOpen, setDetailOpen] = useState(false);
  const [detail, setDetail] = useState<{ data: ShopeeSettled; dropship_total: number } | null>(null);
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
      const res =
        period === "Daily"
          ? await fetchDailyPurchaseTotals({
              channel: channel || undefined,
              store,
              from,
              to,
            })
          : await fetchMonthlyPurchaseTotals({
              channel: channel || undefined,
              store,
              from,
              to,
            });
      const arr = (res.data as any[]).sort((a, b) => {
        const da = period === "Daily" ? a.date : a.month;
        const db = period === "Daily" ? b.date : b.month;
        return da < db ? -1 : 1;
      });
      setData(
        arr.map((d: any) => ({
          date: period === "Daily" ? d.date : d.month,
          total: d.total,
        })),
      );
      setCountData(
        arr.map((d: any) => ({
          date: period === "Daily" ? d.date : d.month,
          count: d.count,
        })),
      );
      setTotalRevenue(arr.reduce((sum, d) => sum + d.total, 0));
      setTotalOrders(arr.reduce((sum, d) => sum + d.count, 0));
      const topRes = await fetchTopProducts({
        channel: channel || undefined,
        store,
        from,
        to,
        limit: 5,
      });
      setTopProducts(topRes.data);
      const cancelRes = await fetchCancelledSummary({
        channel: channel || undefined,
        store,
        from,
        to,
      });
      setCancelSummary(cancelRes.data);
      const adjRes = await listShopeeAdjustments({ from, to });
      setAdjustments(adjRes.data);
      setMsg(null);
    } catch (e: any) {
      setMsg({ type: "error", text: e.response?.data?.error || e.message });
    }
  };

  useEffect(() => {
    fetchData();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [channel, store, from, to, period]);

  const openDetail = async (a: ShopeeAdjustment) => {
    const id = `${a.no_pesanan}-${a.tanggal_penyesuaian.replace(/-/g, "")}`;
    try {
      const [linesRes, detailRes] = await Promise.all([
        getJournalLinesBySource(id),
        getShopeeSettleDetail(a.no_pesanan),
      ]);
      if (linesRes.data.length > 0) {
        setLines(linesRes.data[0].lines);
      } else {
        setLines([]);
      }
      setDetail(detailRes.data);
      setDetailOpen(true);
    } catch (e: any) {
      setMsg({ type: "error", text: e.response?.data?.error || e.message });
    }
  };

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
        <select
          aria-label="Period"
          value={period}
          onChange={(e) => setPeriod(e.target.value as any)}
        >
          <option value="Daily">Daily</option>
          <option value="Monthly">Monthly</option>
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
        })}{" "}
        | <strong>Total Orders:</strong> {totalOrders}
      </div>
      <div style={{ marginBottom: "1rem" }}>
        <strong>Cancelled Orders:</strong> {cancelSummary.count} |{" "}
        <strong>Biaya Mitra:</strong>{" "}
        {cancelSummary.biaya_mitra.toLocaleString("id-ID", {
          style: "currency",
          currency: "IDR",
        })}
      </div>
      <h3>Total Sales by Amount</h3>
      <LineChart
        width={600}
        height={300}
        data={data}
        style={{ marginBottom: "1rem" }}
      >
        <CartesianGrid strokeDasharray="3 3" />
        <XAxis dataKey="date" />
        <YAxis
          tickFormatter={(v) =>
            v.toLocaleString("id-ID", {
              style: "currency",
              currency: "IDR",
              maximumFractionDigits: 0,
            })
          }
        />
        <Tooltip
          formatter={(v: number) =>
            v.toLocaleString("id-ID", {
              style: "currency",
              currency: "IDR",
              maximumFractionDigits: 0,
            })
          }
        />
        <Line type="monotone" dataKey="total" stroke="#8884d8" />
      </LineChart>
      <h3>Total Sales by Quantity</h3>
      <BarChart width={600} height={300} data={countData}>
        <CartesianGrid strokeDasharray="3 3" />
        <XAxis dataKey="date" />
        <YAxis tickFormatter={(v) => v.toLocaleString("id-ID")}/> 
        <Tooltip formatter={(v: number) => v.toLocaleString("id-ID")} />
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
      {adjustments.length > 0 && (
        <div style={{ marginTop: "1rem" }}>
          <h3>Order Adjustments</h3>
          <table style={{ width: "100%", borderCollapse: "collapse" }}>
            <thead>
              <tr>
                <th>Date</th>
                <th>Type</th>
                <th>Amount</th>
                <th>Order</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              {adjustments.map((a) => (
                <tr key={a.id}>
                  <td>{a.tanggal_penyesuaian}</td>
                  <td>{a.tipe_penyesuaian}</td>
                  <td style={{ textAlign: "right" }}>
                    {a.biaya_penyesuaian.toLocaleString("id-ID", {
                      style: "currency",
                      currency: "IDR",
                    })}
                  </td>
                  <td>{a.no_pesanan}</td>
                  <td>
                    <button onClick={() => openDetail(a)}>View</button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
      <Dialog
        open={detailOpen}
        onClose={() => {
          setDetailOpen(false);
          setDetail(null);
        }}
      >
        <DialogTitle>Order Detail & Journal</DialogTitle>
        <DialogContent>
          {detail && (
            <div style={{ marginBottom: "1rem" }}>
              <div>
                <strong>Order:</strong> {detail.data.no_pesanan}
              </div>
              <div>
                <strong>Harga Asli:</strong>{" "}
                {detail.data.harga_asli_produk.toLocaleString("id-ID", {
                  style: "currency",
                  currency: "IDR",
                })}
              </div>
              <div>
                <strong>Diskon Produk:</strong>{" "}
                {detail.data.total_diskon_produk.toLocaleString("id-ID", {
                  style: "currency",
                  currency: "IDR",
                })}
              </div>
              <div>
                <strong>Dropship Total:</strong>{" "}
                {detail.dropship_total.toLocaleString("id-ID", {
                  style: "currency",
                  currency: "IDR",
                })}
              </div>
            </div>
          )}
          <h4>Journal Lines</h4>
          <table style={{ width: "100%", borderCollapse: "collapse" }}>
            <thead>
              <tr>
                <th>Account</th>
                <th>Debit</th>
                <th>Credit</th>
              </tr>
            </thead>
            <tbody>
              {lines.map((l, idx) => (
                <tr key={idx}>
                  <td>{l.account_name}</td>
                  <td style={{ textAlign: "right" }}>
                    {l.is_debit
                      ? l.amount.toLocaleString("id-ID", {
                          style: "currency",
                          currency: "IDR",
                        })
                      : ""}
                  </td>
                  <td style={{ textAlign: "right" }}>
                    {!l.is_debit
                      ? l.amount.toLocaleString("id-ID", {
                          style: "currency",
                          currency: "IDR",
                        })
                      : ""}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </DialogContent>
        <DialogActions>
          <Button
            onClick={() => {
              setDetailOpen(false);
              setDetail(null);
            }}
          >
            Close
          </Button>
        </DialogActions>
      </Dialog>
    </div>
  );
}
