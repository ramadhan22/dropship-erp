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

  // metrics row should render even while loading, so do not early return

  // Card component for each metric. It receives backend values and the global
  // loading state so we can render skeletons while data is being fetched.
  const SummaryCard = ({
    label,
    value,
    change,
    loading: cardLoading,
  }: {
    label: string;
    value?: number;
    change?: number;
    loading: boolean;
  }) => (
    <div
      className="bg-white rounded-xl shadow p-6 flex flex-col min-w-[180px] flex-shrink-0"
      aria-label={label}
    >
      {/* metric label from backend */}
      <div className="text-xs font-semibold text-gray-400 uppercase mb-2">
        {label}
      </div>
      {cardLoading || value === undefined ? (
        // Loading skeleton when metrics are still fetching
        <>
          <div className="h-6 bg-gray-200 rounded animate-pulse mb-2" />
          <div className="h-4 bg-gray-200 rounded animate-pulse mt-2" />
        </>
      ) : (
        <>
          {/* main numeric value */}
          <div className="text-2xl font-bold text-gray-900">
            {value.toLocaleString()}
          </div>
          {/* percent change indicator */}
          <div
            className={`mt-2 flex flex-row items-center text-left ${
              change && change > 0
                ? "text-green-600"
                : change && change < 0
                ? "text-red-600"
                : "text-gray-400"
            }`}
          >
            {change && change > 0 && <span className="mr-1">▲</span>}
            {change && change < 0 && <span className="mr-1">▼</span>}
            <span>{Math.abs((change ?? 0) * 100).toFixed(1)}%</span>
          </div>
        </>
      )}
    </div>
  );

  // metrics values from backend; fallback to empty object when loading
  const charts = data?.charts || {};
  const s = (data?.summary as any) || {};
  // global flag for metrics loading state
  const metricsLoading = loading || !data;

  return (
    <div className="p-4">
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


      {/* Summary Cards - horizontal metrics row */}
      <div className="flex flex-row gap-x-4 max-w-screen-lg mx-auto w-full overflow-x-auto">
        <SummaryCard
          label="TOTAL ORDERS"
          value={s.total_orders?.value}
          change={s.total_orders?.change}
          loading={metricsLoading}
        />
        <SummaryCard
          label="AVERAGE ORDER VALUE"
          value={s.avg_order_value?.value}
          change={s.avg_order_value?.change}
          loading={metricsLoading}
        />
        <SummaryCard
          label="TOTAL CANCELLED ORDERS"
          value={s.total_cancelled?.value}
          change={s.total_cancelled?.change}
          loading={metricsLoading}
        />
        <SummaryCard
          label="TOTAL CUSTOMERS"
          value={s.total_customers?.value}
          change={s.total_customers?.change}
          loading={metricsLoading}
        />
      </div>
      {/* Charts */}
      <div className="grid grid-cols-2 gap-6 mt-8">
        <div className="bg-white rounded-xl shadow p-4 h-64">
          <h3 className="text-sm uppercase text-gray-400 mb-1">Total Sales</h3>
          <LineChart width={300} height={200} data={charts.total_sales}>
            <CartesianGrid strokeDasharray="3 3" />
            <XAxis dataKey="date" />
            <YAxis />
            <Tooltip />
            <Line type="monotone" dataKey="value" stroke="#8884d8" />
          </LineChart>
        </div>
        <div className="bg-white rounded-xl shadow p-4 h-64">
          <h3 className="text-sm uppercase text-gray-400 mb-1">Average Order Value</h3>
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
      <div className="grid grid-cols-4 gap-4 mt-8">
        <SummaryCard
          label="Total Price"
          value={s.total_price?.value}
          change={s.total_price?.change}
          loading={metricsLoading}
        />
        <SummaryCard
          label="Total Discounts"
          value={s.total_discounts?.value}
          change={s.total_discounts?.change}
          loading={metricsLoading}
        />
        <SummaryCard
          label="Total Net Profit"
          value={s.total_net_profit?.value}
          change={s.total_net_profit?.change}
          loading={metricsLoading}
        />
        <SummaryCard
          label="Outstanding Amount"
          value={s.outstanding_amount?.value}
          change={s.outstanding_amount?.change}
          loading={metricsLoading}
        />
      </div>

      <div className="grid grid-cols-2 gap-6 mt-8">
        <div className="bg-white rounded-xl shadow p-4 h-64">
          <h3 className="text-sm uppercase text-gray-400 mb-1">Number of Orders</h3>
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

