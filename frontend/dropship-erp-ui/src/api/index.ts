import axios from "axios";
import type { BalanceCategory, Metric } from "../types";

export interface ImportResponse {
  inserted: number;
}

// Base URL for API calls. In Jest/Node we read from process.env; in Vite builds
// you can still set VITE_API_URL, otherwise we fall back to localhost.
// Base URL for API calls – in Vite builds import.meta.env is available; otherwise we default to localhost
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

export const api = axios.create({
  baseURL: BASE_URL,
});

// Dropship import
export function importDropship(file: File) {
  const data = new FormData();
  data.append("file", file);
  return api.post<ImportResponse>("/dropship/import", data, {
    headers: { "Content-Type": "multipart/form-data" },
  });
}

// Shopee import
export function importShopee(file: File) {
  const data = new FormData();
  data.append("file", file);
  return api.post<ImportResponse>("/shopee/import", data, {
    headers: { "Content-Type": "multipart/form-data" },
  });
}

// Reconcile
export function reconcile(purchaseId: string, orderId: string, shop: string) {
  return api.post("/reconcile", {
    purchase_id: purchaseId,
    order_id: orderId,
    shop,
  });
}

// Compute metrics
export function computeMetrics(shop: string, period: string) {
  return api.post("/metrics", { shop, period });
}

// Fetch cached metrics
export function fetchMetrics(shop: string, period: string) {
  return api.get<Metric>(`/metrics?shop=${shop}&period=${period}`);
}

// Fetch balance sheet
export function fetchBalanceSheet(shop: string, period: string) {
  return api.get<BalanceCategory[]>(
    `/balancesheet?shop=${shop}&period=${period}`,
  );
}
