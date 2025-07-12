import { useEffect, useState } from "react";
import { LineChart, Line, XAxis, YAxis, Tooltip, CartesianGrid } from "recharts";
import { fetchDashboard } from "../api";
import type { DashboardData } from "../types";

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

  useEffect(() => {
    const fetchData = async () => {
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
      } finally {
        setLoading(false);
      }
    };
    fetchData();
  }, [orderType, channel, store, period, month, year]);

  if (loading || !data) {
    return <div className="p-4">Loading...</div>;
  }

  const SummaryCard = ({ label, value, change }: { label: string; value: number; change: number }) => (
    <div className="bg-white shadow rounded p-4 flex-1">
      <div className="text-sm text-gray-500">{label}</div>
      <div className="text-xl font-semibold">{value.toLocaleString()}</div>
      <div className={change >= 0 ? "text-green-600" : "text-red-600"}>
        {change >= 0 ? "▲" : "▼"} {Math.abs(change * 100).toFixed(1)}%
      </div>
    </div>
  );

  const charts = data.charts;
  const s = data.summary as any;

  return (
    <div className="p-4 space-y-6">
      {/* Filter Controls */}
      <div className="flex gap-2">
        <select className="border p-1" value={orderType} onChange={(e) => setOrderType(e.target.value)}>
          <option value="">All Orders</option>
          <option value="COD">COD</option>
        </select>
        <select className="border p-1" value={channel} onChange={(e) => setChannel(e.target.value)}>
          <option value="">All Channels</option>
          <option value="Shopee">Shopee</option>
        </select>
        <select className="border p-1" value={store} onChange={(e) => setStore(e.target.value)}>
          <option value="">All Stores</option>
          <option value="StoreA">StoreA</option>
        </select>
        <select className="border p-1" value={period} onChange={(e) => setPeriod(e.target.value as any)}>
          <option value="Monthly">Monthly</option>
          <option value="Yearly">Yearly</option>
        </select>
        {period === "Monthly" && (
          <select className="border p-1" value={month} onChange={(e) => setMonth(Number(e.target.value))}>
            {Array.from({ length: 12 }, (_, i) => i + 1).map((m) => (
              <option key={m} value={m}>{m}</option>
            ))}
          </select>
        )}
        <select className="border p-1" value={year} onChange={(e) => setYear(Number(e.target.value))}>
          {Array.from({ length: 3 }, (_, i) => now.getFullYear() - i).map((y) => (
            <option key={y} value={y}>{y}</option>
          ))}
        </select>
      </div>

      {/* Summary Cards */}
      <div className="grid grid-cols-4 gap-4">
        <SummaryCard label="Total Orders" value={s.total_orders.value} change={s.total_orders.change} />
        <SummaryCard label="Average Order Value" value={s.avg_order_value.value} change={s.avg_order_value.change} />
        <SummaryCard label="Total Cancelled Orders" value={s.total_cancelled.value} change={s.total_cancelled.change} />
        <SummaryCard label="Total Customers" value={s.total_customers.value} change={s.total_customers.change} />
      </div>

      {/* Charts */}
      <div className="grid grid-cols-2 gap-4">
        <div className="bg-white p-4 shadow rounded">
          <h3 className="font-semibold mb-2">Total Sales</h3>
          <LineChart width={300} height={200} data={charts.total_sales}>
            <CartesianGrid strokeDasharray="3 3" />
            <XAxis dataKey="date" />
            <YAxis />
            <Tooltip />
            <Line type="monotone" dataKey="value" stroke="#8884d8" />
          </LineChart>
        </div>
        <div className="bg-white p-4 shadow rounded">
          <h3 className="font-semibold mb-2">Average Order Value</h3>
          <LineChart width={300} height={200} data={charts.avg_order_value}>
            <CartesianGrid strokeDasharray="3 3" />
            <XAxis dataKey="date" />
            <YAxis />
            <Tooltip />
            <Line type="monotone" dataKey="value" stroke="#82ca9d" />
          </LineChart>
        </div>
      </div>

      {/* Additional Summary Cards */}
      <div className="grid grid-cols-4 gap-4">
        <SummaryCard label="Total Price" value={s.total_price.value} change={s.total_price.change} />
        <SummaryCard label="Total Discounts" value={s.total_discounts.value} change={s.total_discounts.change} />
        <SummaryCard label="Total Net Profit" value={s.total_net_profit.value} change={s.total_net_profit.change} />
        <SummaryCard label="Outstanding Amount" value={s.outstanding_amount.value} change={s.outstanding_amount.change} />
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div className="bg-white p-4 shadow rounded">
          <h3 className="font-semibold mb-2">Number of Customers</h3>
          <LineChart width={300} height={200} data={charts.number_of_customers}>
            <CartesianGrid strokeDasharray="3 3" />
            <XAxis dataKey="date" />
            <YAxis />
            <Tooltip />
            <Line type="monotone" dataKey="value" stroke="#ff7300" />
          </LineChart>
        </div>
        <div className="bg-white p-4 shadow rounded">
          <h3 className="font-semibold mb-2">Number of Orders</h3>
          <LineChart width={300} height={200} data={charts.number_of_orders}>
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

