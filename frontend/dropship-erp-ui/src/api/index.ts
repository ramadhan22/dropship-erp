import axios from "axios";
import type { BalanceCategory, Metric } from "../types";

// Base URL for API calls. In Jest/Node we read from process.env; in Vite builds
// you can still set VITE_API_URL, otherwise we fall back to localhost.
// Base URL for API calls – in Vite builds import.meta.env is available; otherwise we default to localhost
const BASE_URL = import.meta.env.VITE_API_URL ?? "http://localhost:8080/api";

export const api = axios.create({
  baseURL: BASE_URL,
});

// Dropship import
export function importDropship(filePath: string) {
  return api.post("/dropship/import", { file_path: filePath });
}

// Shopee import
export function importShopee(filePath: string) {
  return api.post("/shopee/import", { file_path: filePath });
}

// Reconcile
export function reconcile(
  purchaseId: string,
  orderId: string,
  shop: string
) {
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
    `/balancesheet?shop=${shop}&period=${period}`
  );
}