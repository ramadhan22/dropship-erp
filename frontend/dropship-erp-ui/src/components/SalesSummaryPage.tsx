import {
  Alert,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
} from "@mui/material";
import { useCallback, useEffect, useState } from "react";
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
  fetchDashboard,
} from "../api";
import type {
  JenisChannel,
  Store,
  ProductSales,
  ShopeeAdjustment,
  JournalLineDetail,
  ShopeeSettled,
  CancelledSummary,
  DashboardData,
  DashboardMetrics,
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
import SummaryCard from "./SummaryCard";

export default function SalesSummaryPage() {
  const now = new Date();
  const [orderType, setOrderType] = useState("");
  const [channels, setChannels] = useState<JenisChannel[]>([]);
  const [channel, setChannel] = useState("");
  const [stores, setStores] = useState<Store[]>([]);
  const [store, setStore] = useState("");
  const [period, setPeriod] = useState<"Monthly" | "Yearly">("Monthly");
  const [month, setMonth] = useState(now.getMonth() + 1);
  const [year, setYear] = useState(now.getFullYear());
  const [data, setData] = useState<{ date: string; total: number }[]>([]);
  const [countData, setCountData] = useState<{ date: string; count: number }[]>(
    [],
  );
  const [totalRevenue, setTotalRevenue] = useState(0);
  const [totalOrders, setTotalOrders] = useState(0);
  const [dashboardData, setDashboardData] = useState<DashboardData | null>(null);
  const [dashboardLoading, setDashboardLoading] = useState(false);
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

  const getRange = () => {
    if (period === "Monthly") {
      const fromDate = new Date(year, month - 1, 1);
      const toDate = new Date(year, month, 0);
      return {
        from: fromDate.toISOString().split("T")[0],
        to: toDate.toISOString().split("T")[0],
      };
    }
    const fromDate = new Date(year, 0, 1);
    const toDate = new Date(year, 11, 31);
    return {
      from: fromDate.toISOString().split("T")[0],
      to: toDate.toISOString().split("T")[0],
    };
  };

  const fetchDashboardData = useCallback(async () => {
    setDashboardLoading(true);
    try {
      const res = await fetchDashboard({
        order: orderType,
        channel,
        store,
        period,
        month,
        year,
      });
      setDashboardData(res.data);
    } catch (err) {
       
      console.error("dashboard fetch", err);
    } finally {
      setDashboardLoading(false);
    }
  }, [orderType, channel, store, period, month, year]);

  const fetchData = async () => {
    const { from, to } = getRange();
    try {
      const res =
        period === "Monthly"
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
        const da = period === "Monthly" ? a.date : a.month;
        const db = period === "Monthly" ? b.date : b.month;
        return da < db ? -1 : 1;
      });
      setData(
        arr.map((d: any) => ({
          date: period === "Monthly" ? d.date : d.month,
          total: d.total,
        })),
      );
      setCountData(
        arr.map((d: any) => ({
          date: period === "Monthly" ? d.date : d.month,
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

  const charts = dashboardData?.charts || {};
  const metrics: DashboardMetrics = dashboardData?.summary || {};
  const metricsLoading = dashboardLoading || !dashboardData;

  useEffect(() => {
    fetchData();
    fetchDashboardData();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [channel, store, period, month, year, orderType, fetchDashboardData]);

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
          value={orderType}
          onChange={(e) => setOrderType(e.target.value)}
        >
          <option value="">All Orders</option>
          <option value="COD">COD</option>
        </select>
        <select
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
          value={period}
          onChange={(e) => setPeriod(e.target.value as any)}
        >
          <option value="Monthly">Monthly</option>
          <option value="Yearly">Yearly</option>
        </select>
        {period === "Monthly" && (
          <select
            value={month}
            onChange={(e) => setMonth(Number(e.target.value))}
          >
            {Array.from({ length: 12 }, (_, i) => i + 1).map((m) => (
              <option key={m} value={m}>
                {m}
              </option>
            ))}
          </select>
        )}
        <select value={year} onChange={(e) => setYear(Number(e.target.value))}>
          {Array.from({ length: 3 }, (_, i) => now.getFullYear() - i).map((y) => (
            <option key={y} value={y}>
              {y}
            </option>
          ))}
        </select>
      </div>
      {msg && (
        <Alert severity={msg.type} sx={{ mb: 2 }}>
          {msg.text}
        </Alert>
      )}
      <div style={{ marginTop: "1rem", display: "flex", gap: "1rem" }}>
        <SummaryCard
          label="Total Revenue"
          value={totalRevenue}
          loading={false}
        />
        <SummaryCard
          label="Total Orders"
          value={totalOrders}
          loading={false}
        />
        <SummaryCard
          label="Cancelled Orders"
          value={cancelSummary.count}
          loading={false}
        />
        <SummaryCard
          label="Biaya Mitra"
          value={cancelSummary.biaya_mitra}
          loading={false}
        />
      </div>
      <div style={{ marginTop: "1rem", display: "flex", gap: "1rem" }}>
        <div style={{ flex: 1 }}>
          <h3>Total Sales by Amount</h3>
          <LineChart width={600} height={300} data={data}>
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
        </div>
        <div style={{ flex: 1 }}>
          <h3>Total Sales by Quantity</h3>
          <BarChart width={600} height={300} data={countData}>
            <CartesianGrid strokeDasharray="3 3" />
            <XAxis dataKey="date" />
            <YAxis tickFormatter={(v) => v.toLocaleString("id-ID")} />
            <Tooltip formatter={(v: number) => v.toLocaleString("id-ID")} />
            <Bar dataKey="count" fill="#82ca9d" />
          </BarChart>
        </div>
      </div>
      {dashboardData && (
        <>
          <div style={{ marginTop: "1rem", display: "flex", gap: "1rem" }}>
            <SummaryCard
              label="TOTAL ORDERS"
              value={metrics.total_orders?.value}
              change={metrics.total_orders?.change}
              loading={metricsLoading}
            />
            <SummaryCard
              label="AVERAGE ORDER VALUE"
              value={metrics.avg_order_value?.value}
              change={metrics.avg_order_value?.change}
              loading={metricsLoading}
            />
            <SummaryCard
              label="TOTAL CANCELLED ORDERS"
              value={metrics.total_cancelled?.value}
              change={metrics.total_cancelled?.change}
              loading={metricsLoading}
            />
            <SummaryCard
              label="TOTAL CUSTOMERS"
              value={metrics.total_customers?.value}
              change={metrics.total_customers?.change}
              loading={metricsLoading}
            />
          </div>
          <div style={{ marginTop: "1rem", display: "flex", gap: "1rem" }}>
            <div
              style={{
                backgroundColor: "#fff",
                borderRadius: "0.75rem",
                boxShadow:
                  "0 1px 3px 0 rgba(0,0,0,0.1), 0 1px 2px 0 rgba(0,0,0,0.06)",
                padding: "1rem",
                height: "16rem",
              }}
            >
              <h3
                style={{
                  fontSize: "0.875rem",
                  textTransform: "uppercase",
                  color: "#9ca3af",
                  marginBottom: "0.25rem",
                }}
              >
                Total Sales
              </h3>
              <LineChart width={650} height={200} data={charts.total_sales}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="date" />
                <YAxis />
                <Tooltip />
                <Line type="monotone" dataKey="value" stroke="#8884d8" />
              </LineChart>
            </div>
            <div
              style={{
                backgroundColor: "#fff",
                borderRadius: "0.75rem",
                boxShadow:
                  "0 1px 3px 0 rgba(0,0,0,0.1), 0 1px 2px 0 rgba(0,0,0,0.06)",
                padding: "1rem",
                height: "16rem",
              }}
            >
              <h3
                style={{
                  fontSize: "0.875rem",
                  textTransform: "uppercase",
                  color: "#9ca3af",
                  marginBottom: "0.25rem",
                }}
              >
                Average Order Value
              </h3>
              <LineChart width={650} height={200} data={charts.avg_order_value}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="date" />
                <YAxis />
                <Tooltip />
                <Line type="monotone" dataKey="value" stroke="#82ca9d" />
              </LineChart>
            </div>
          </div>
          <div style={{ marginTop: "1rem", display: "flex", gap: "1rem" }}>
            <SummaryCard
              label="Total Price"
              value={metrics.total_price?.value}
              change={metrics.total_price?.change}
              loading={metricsLoading}
            />
            <SummaryCard
              label="Total Discounts"
              value={metrics.total_discounts?.value}
              change={metrics.total_discounts?.change}
              loading={metricsLoading}
            />
            <SummaryCard
              label="Total Net Profit"
              value={metrics.total_net_profit?.value}
              change={metrics.total_net_profit?.change}
              loading={metricsLoading}
            />
            <SummaryCard
              label="Outstanding Amount"
              value={metrics.outstanding_amount?.value}
              change={metrics.outstanding_amount?.change}
              loading={metricsLoading}
            />
          </div>
          <div
            style={{
              display: "grid",
              gridTemplateColumns: "repeat(2, minmax(0, 1fr))",
              gap: "1.5rem",
              marginTop: "2rem",
            }}
          >
            <div
              style={{
                backgroundColor: "#fff",
                borderRadius: "0.75rem",
                boxShadow:
                  "0 1px 3px 0 rgba(0,0,0,0.1), 0 1px 2px 0 rgba(0,0,0,0.06)",
                padding: "1rem",
                height: "16rem",
              }}
            >
              <h3
                style={{
                  fontSize: "0.875rem",
                  textTransform: "uppercase",
                  color: "#9ca3af",
                  marginBottom: "0.25rem",
                }}
              >
                Number of Orders
              </h3>
              <LineChart width={650} height={200} data={charts.number_of_orders}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="date" />
                <YAxis />
                <Tooltip />
                <Line type="monotone" dataKey="value" stroke="#8884d8" />
              </LineChart>
            </div>
          </div>
        </>
      )}
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
