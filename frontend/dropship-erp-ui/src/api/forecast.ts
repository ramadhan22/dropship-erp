import axios from "axios";

let BASE_URL = "http://localhost:8080/api";

if (typeof process !== "undefined" && process.env?.VITE_API_URL) {
  BASE_URL = process.env.VITE_API_URL;
}

try {
  // Access import.meta dynamically so tests running in CommonJS don't fail
  const meta = Function("return import.meta")();
  if (meta?.env?.VITE_API_URL) {
    BASE_URL = meta.env.VITE_API_URL;
  }
} catch {
  // ignore if import.meta is not available
}

export interface ForecastDataPoint {
  date: string;
  value: number;
  source: "historical" | "forecast";
}

export interface ForecastResult {
  metric: string;
  historicalData: ForecastDataPoint[];
  forecastData: ForecastDataPoint[];
  totalForecast: number;
  totalHistorical: number;
  growthRate: number;
  confidence: number;
  method: string;
}

export interface ForecastResponse {
  sales: ForecastResult;
  expenses: ForecastResult;
  profit: ForecastResult;
  period: string;
  generated: string;
}

export interface ForecastRequest {
  shop: string;
  period: "monthly" | "yearly";
  startDate: string;
  endDate: string;
  forecastTo: string;
}

export interface ForecastParams {
  shop: string;
  period: string;
  suggestedStartDate: string;
  suggestedEndDate: string;
  suggestedForecastTo: string;
  currentDate: string;
}

export interface ForecastSummary {
  shop: string;
  period: string;
  days: number;
  forecastSales: number;
  forecastExpenses: number;
  forecastProfit: number;
  historicalSales: number;
  historicalExpenses: number;
  historicalProfit: number;
  salesGrowthRate: number;
  expensesGrowthRate: number;
  profitGrowthRate: number;
  salesConfidence: number;
  expensesConfidence: number;
  profitConfidence: number;
  generated: string;
}

export async function generateForecast(request: ForecastRequest): Promise<{ data: ForecastResponse }> {
  const response = await axios.post(`${BASE_URL}/forecast/generate`, request);
  return response.data;
}

export async function getForecastParams(shop?: string, period?: string): Promise<{ data: ForecastParams }> {
  const params = new URLSearchParams();
  if (shop) params.set("shop", shop);
  if (period) params.set("period", period);
  
  const response = await axios.get(`${BASE_URL}/forecast/params?${params}`);
  return { data: response.data };
}

export async function getForecastSummary(shop: string, period?: string, days?: number): Promise<{ data: ForecastSummary }> {
  const params = new URLSearchParams();
  params.set("shop", shop);
  if (period) params.set("period", period);
  if (days) params.set("days", days.toString());
  
  const response = await axios.get(`${BASE_URL}/forecast/summary?${params}`);
  return { data: response.data };
}