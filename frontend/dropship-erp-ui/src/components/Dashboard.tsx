import { useCallback, useEffect, useState } from "react";
import { LineChart, Line, XAxis, YAxis, Tooltip, CartesianGrid } from "recharts";
import { fetchDashboard } from "../api";
import type { DashboardData, DashboardMetrics } from "../types";
import SummaryCard from "./SummaryCard";

// Dashboard component showing summary cards and charts
export default function Dashboard() {
  const now = new Date();
  const [orderType, setOrderType] = useState("");
  const [channel, setChannel] = useState("");
  const [store, setStore] = useState("");
  const [period, setPeriod] = useState<"Monthly" | "Yearly">("Monthly");
  const [month, setMonth] = useState(now.getMonth() + 1);
  const [year, setYear] = useState(now.getFullYear());
  const [data, setData] = useState<DashboardData | null>(null);
  const [loading, setLoading] = useState(false);

  const fetchData = useCallback(async () => {
    setLoading(true);
    try {
      const res = await fetchDashboard({
        order: orderType,
        channel,
        store,
        period,
        month,
        year,
      });
      setData(res.data);
    } catch (err) {
      // eslint-disable-next-line no-console
      console.error("dashboard fetch", err);
    } finally {
      setLoading(false);
    }
  }, [orderType, channel, store, period, month, year]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  // metrics row should render even while loading, so do not early return
  // metrics values returned from the backend dashboard endpoint
  const charts = data?.charts || {};
  // typed metrics object for summary numbers
  const metrics: DashboardMetrics = data?.summary || {};
  // global flag for metrics loading state
  const metricsLoading = loading || !data;

  return (
    <div
      style={{ padding: "1rem", backgroundColor: "#f9fafb", minHeight: "100vh" }}
    >
      {/* Filter Controls */}
      <div
        style={{ display: "flex", gap: "0.5rem" }}
      >
        <select
          style={{ border: "1px solid #ccc", padding: "0.25rem" }}
          value={orderType}
          onChange={(e) => setOrderType(e.target.value)}
        >
          <option value="">All Orders</option>
          <option value="COD">COD</option>
        </select>
        <select
          style={{ border: "1px solid #ccc", padding: "0.25rem" }}
          value={channel}
          onChange={(e) => setChannel(e.target.value)}
        >
          <option value="">All Channels</option>
          <option value="Shopee">Shopee</option>
        </select>
        <select
          style={{ border: "1px solid #ccc", padding: "0.25rem" }}
          value={store}
          onChange={(e) => setStore(e.target.value)}
        >
          <option value="">All Stores</option>
          <option value="StoreA">StoreA</option>
        </select>
        <select
          style={{ border: "1px solid #ccc", padding: "0.25rem" }}
          value={period}
          onChange={(e) => setPeriod(e.target.value as any)}
        >
          <option value="Monthly">Monthly</option>
          <option value="Yearly">Yearly</option>
        </select>
        {period === "Monthly" && (
          <select
            style={{ border: "1px solid #ccc", padding: "0.25rem" }}
            value={month}
            onChange={(e) => setMonth(Number(e.target.value))}
          >
            {Array.from({ length: 12 }, (_, i) => i + 1).map((m) => (
              <option key={m} value={m}>{m}</option>
            ))}
          </select>
        )}
        <select
          style={{ border: "1px solid #ccc", padding: "0.25rem" }}
          value={year}
          onChange={(e) => setYear(Number(e.target.value))}
        >
          {Array.from({ length: 3 }, (_, i) => now.getFullYear() - i).map((y) => (
            <option key={y} value={y}>{y}</option>
          ))}
        </select>
      </div>


      {/*
        Parent container for the summary cards. Using a responsive grid keeps
        the metrics aligned consistently across pages and matches the layout on
        the Balance Sheet page.
      */}
      {metrics && (
        <div style={{ marginTop: "1rem", display: "flex", gap: "12rem" }}>
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
      )}
      {/* Charts */}
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
          <LineChart width={700} height={200} data={charts.total_sales}>
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
          <LineChart width={700} height={200} data={charts.avg_order_value}>
            <CartesianGrid strokeDasharray="3 3" />
            <XAxis dataKey="date" />
            <YAxis />
            <Tooltip />
            <Line type="monotone" dataKey="value" stroke="#82ca9d" />
          </LineChart>
        </div>
      </div>

      {/* Additional Summary Cards */}
      <div style={{ marginTop: "1rem", display: "flex", gap: "12rem" }}>
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
          <LineChart width={1400} height={200} data={charts.number_of_orders}>
            <CartesianGrid strokeDasharray="3 3" />
            <XAxis dataKey="date" />
            <YAxis />
            <Tooltip />
            <Line type="monotone" dataKey="value" stroke="#8884d8" />
          </LineChart>
        </div>
      </div>
    </div>
  );
}
