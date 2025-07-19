import { useCallback, useEffect, useState } from "react";
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
  BarChart,
  Bar,
} from "recharts";
import {
  generateForecast,
  getForecastParams,
  getForecastSummary,
  type ForecastResponse,
  type ForecastSummary,
  type ForecastRequest,
} from "../api/forecast";
import SummaryCard from "./SummaryCard";
import LoadingOverlay from "./LoadingOverlay";

export default function ForecastPage() {
  const [shop, setShop] = useState("");
  const [period, setPeriod] = useState<"monthly" | "yearly">("monthly");
  const [startDate, setStartDate] = useState("");
  const [endDate, setEndDate] = useState("");
  const [forecastTo, setForecastTo] = useState("");
  const [forecastData, setForecastData] = useState<ForecastResponse | null>(null);
  const [summary, setSummary] = useState<ForecastSummary | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Fetch suggested parameters when shop or period changes
  useEffect(() => {
    const fetchParams = async () => {
      if (!shop) return;
      
      try {
        const { data } = await getForecastParams(shop, period);
        setStartDate(data.suggestedStartDate);
        setEndDate(data.suggestedEndDate);
        setForecastTo(data.suggestedForecastTo);
      } catch (err) {
        console.error("Failed to fetch forecast params:", err);
      }
    };

    fetchParams();
  }, [shop, period]);

  // Fetch summary data when parameters are available
  useEffect(() => {
    const fetchSummary = async () => {
      if (!shop) return;
      
      try {
        const { data } = await getForecastSummary(shop, period);
        setSummary(data);
      } catch (err) {
        console.error("Failed to fetch forecast summary:", err);
      }
    };

    fetchSummary();
  }, [shop, period]);

  const handleGenerateForecast = useCallback(async () => {
    if (!shop || !startDate || !endDate || !forecastTo) {
      setError("Please fill in all required fields");
      return;
    }

    setLoading(true);
    setError(null);

    try {
      const request: ForecastRequest = {
        shop,
        period,
        startDate,
        endDate,
        forecastTo,
      };

      const response = await generateForecast(request);
      setForecastData(response.data);
    } catch (err: any) {
      setError(err.response?.data?.error || "Failed to generate forecast");
      console.error("Forecast generation failed:", err);
    } finally {
      setLoading(false);
    }
  }, [shop, period, startDate, endDate, forecastTo]);

  // Prepare chart data by combining historical and forecast data
  const prepareChartData = useCallback((result: any) => {
    const combined = [
      ...result.historicalData.map((d: any) => ({
        date: new Date(d.date).toISOString().split('T')[0],
        value: d.value,
        type: 'historical',
      })),
      ...result.forecastData.map((d: any) => ({
        date: new Date(d.date).toISOString().split('T')[0],
        value: d.value,
        type: 'forecast',
      })),
    ];

    return combined.sort((a, b) => a.date.localeCompare(b.date));
  }, []);

  const formatCurrency = (value: number) => {
    return new Intl.NumberFormat('id-ID', {
      style: 'currency',
      currency: 'IDR',
      minimumFractionDigits: 0,
    }).format(value);
  };

  const formatPercentage = (value: number) => {
    return `${(value * 100).toFixed(1)}%`;
  };

  return (
    <div className="container mx-auto p-4">
      <h1 className="text-3xl font-bold mb-6">Sales, Expenses & Profit Forecast</h1>

      {/* Input Form */}
      <div className="bg-white rounded-lg shadow-md p-6 mb-6">
        <h2 className="text-xl font-semibold mb-4">Forecast Parameters</h2>
        
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-5 gap-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Shop
            </label>
            <input
              type="text"
              value={shop}
              onChange={(e) => setShop(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              placeholder="Enter shop name"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Period
            </label>
            <select
              value={period}
              onChange={(e) => setPeriod(e.target.value as "monthly" | "yearly")}
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
            >
              <option value="monthly">Monthly</option>
              <option value="yearly">Yearly</option>
            </select>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Start Date
            </label>
            <input
              type="date"
              value={startDate}
              onChange={(e) => setStartDate(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              End Date
            </label>
            <input
              type="date"
              value={endDate}
              onChange={(e) => setEndDate(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Forecast To
            </label>
            <input
              type="date"
              value={forecastTo}
              onChange={(e) => setForecastTo(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>
        </div>

        <div className="mt-4">
          <button
            onClick={handleGenerateForecast}
            disabled={loading}
            className="bg-blue-600 text-white px-6 py-2 rounded-md hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {loading ? "Generating..." : "Generate Forecast"}
          </button>
        </div>

        {error && (
          <div className="mt-4 p-3 bg-red-100 border border-red-300 text-red-700 rounded-md">
            {error}
          </div>
        )}
      </div>

      {/* Summary Cards */}
      {summary && (
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-6">
          <SummaryCard
            label="Forecast Sales"
            value={summary.forecastSales}
            change={summary.salesGrowthRate}
            loading={false}
          />
          <SummaryCard
            label="Forecast Expenses"
            value={summary.forecastExpenses}
            change={summary.expensesGrowthRate}
            loading={false}
          />
          <SummaryCard
            label="Forecast Profit"
            value={summary.forecastProfit}
            change={summary.profitGrowthRate}
            loading={false}
          />
        </div>
      )}

      {/* Forecast Charts */}
      {forecastData && (
        <div className="space-y-6">
          {/* Sales Forecast Chart */}
          <div className="bg-white rounded-lg shadow-md p-6">
            <h3 className="text-lg font-semibold mb-4">Sales Forecast</h3>
            <div className="mb-2 text-sm text-gray-600">
              Method: {forecastData.sales.method} | 
              Confidence: {formatPercentage(forecastData.sales.confidence)} | 
              Growth Rate: {formatPercentage(forecastData.sales.growthRate)}
            </div>
            <ResponsiveContainer width="100%" height={300}>
              <LineChart data={prepareChartData(forecastData.sales)}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="date" />
                <YAxis tickFormatter={(value) => `${(value / 1000000).toFixed(1)}M`} />
                <Tooltip formatter={(value: number) => [formatCurrency(value), "Sales"]} />
                <Legend />
                <Line
                  type="monotone"
                  dataKey="value"
                  stroke="#2563eb"
                  strokeWidth={2}
                  dot={(props) => {
                    const { payload } = props;
                    return (
                      <circle
                        cx={props.cx}
                        cy={props.cy}
                        r={3}
                        fill={payload?.type === 'forecast' ? '#ef4444' : '#2563eb'}
                      />
                    );
                  }}
                  name="Sales"
                />
              </LineChart>
            </ResponsiveContainer>
          </div>

          {/* Expenses Forecast Chart */}
          <div className="bg-white rounded-lg shadow-md p-6">
            <h3 className="text-lg font-semibold mb-4">Expenses Forecast</h3>
            <div className="mb-2 text-sm text-gray-600">
              Method: {forecastData.expenses.method} | 
              Confidence: {formatPercentage(forecastData.expenses.confidence)} | 
              Growth Rate: {formatPercentage(forecastData.expenses.growthRate)}
            </div>
            <ResponsiveContainer width="100%" height={300}>
              <LineChart data={prepareChartData(forecastData.expenses)}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="date" />
                <YAxis tickFormatter={(value) => `${(value / 1000000).toFixed(1)}M`} />
                <Tooltip formatter={(value: number) => [formatCurrency(value), "Expenses"]} />
                <Legend />
                <Line
                  type="monotone"
                  dataKey="value"
                  stroke="#dc2626"
                  strokeWidth={2}
                  dot={(props) => {
                    const { payload } = props;
                    return (
                      <circle
                        cx={props.cx}
                        cy={props.cy}
                        r={3}
                        fill={payload?.type === 'forecast' ? '#ef4444' : '#dc2626'}
                      />
                    );
                  }}
                  name="Expenses"
                />
              </LineChart>
            </ResponsiveContainer>
          </div>

          {/* Profit Forecast Chart */}
          <div className="bg-white rounded-lg shadow-md p-6">
            <h3 className="text-lg font-semibold mb-4">Profit Forecast</h3>
            <div className="mb-2 text-sm text-gray-600">
              Method: {forecastData.profit.method} | 
              Confidence: {formatPercentage(forecastData.profit.confidence)} | 
              Growth Rate: {formatPercentage(forecastData.profit.growthRate)}
            </div>
            <ResponsiveContainer width="100%" height={300}>
              <LineChart data={prepareChartData(forecastData.profit)}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="date" />
                <YAxis tickFormatter={(value) => `${(value / 1000000).toFixed(1)}M`} />
                <Tooltip formatter={(value: number) => [formatCurrency(value), "Profit"]} />
                <Legend />
                <Line
                  type="monotone"
                  dataKey="value"
                  stroke="#16a34a"
                  strokeWidth={2}
                  dot={(props) => {
                    const { payload } = props;
                    return (
                      <circle
                        cx={props.cx}
                        cy={props.cy}
                        r={3}
                        fill={payload?.type === 'forecast' ? '#ef4444' : '#16a34a'}
                      />
                    );
                  }}
                  name="Profit"
                />
              </LineChart>
            </ResponsiveContainer>
          </div>

          {/* Summary Bar Chart */}
          <div className="bg-white rounded-lg shadow-md p-6">
            <h3 className="text-lg font-semibold mb-4">Forecast vs Historical Summary</h3>
            <ResponsiveContainer width="100%" height={300}>
              <BarChart
                data={[
                  {
                    metric: 'Sales',
                    Historical: forecastData.sales.totalHistorical,
                    Forecast: forecastData.sales.totalForecast,
                  },
                  {
                    metric: 'Expenses',
                    Historical: forecastData.expenses.totalHistorical,
                    Forecast: forecastData.expenses.totalForecast,
                  },
                  {
                    metric: 'Profit',
                    Historical: forecastData.profit.totalHistorical,
                    Forecast: forecastData.profit.totalForecast,
                  },
                ]}
              >
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="metric" />
                <YAxis tickFormatter={(value) => `${(value / 1000000).toFixed(1)}M`} />
                <Tooltip formatter={(value: number) => formatCurrency(value)} />
                <Legend />
                <Bar dataKey="Historical" fill="#64748b" name="Historical" />
                <Bar dataKey="Forecast" fill="#3b82f6" name="Forecast" />
              </BarChart>
            </ResponsiveContainer>
          </div>
        </div>
      )}

      {loading && <LoadingOverlay />}
    </div>
  );
}