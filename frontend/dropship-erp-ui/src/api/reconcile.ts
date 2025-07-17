import { api } from "./index";
import type { AxiosRequestConfig } from "axios";
import type {
  ReconciledTransaction,
  ReconcileCandidate,
  ShopeeOrderDetail,
  ShopeeEscrowDetail,
} from "../types";

export function listUnmatched(shop: string) {
  return api.get<ReconciledTransaction[]>(`/reconcile/unmatched?shop=${shop}`);
}

export function listCandidates(
  shop: string,
  order: string,
  status: string,
  from: string,
  to: string,
  page: number,
  pageSize: number,
  config?: AxiosRequestConfig,
) {
  const q = new URLSearchParams();
  if (shop) q.append("shop", shop);
  if (order) q.append("order", order);
  if (status) q.append("status", status);
  if (from) q.append("from", from);
  if (to) q.append("to", to);
  q.append("page", String(page));
  q.append("page_size", String(pageSize));
  const qs = q.toString();
  const url = qs ? `/reconcile/candidates?${qs}` : "/reconcile/candidates";
  return api.get<{ data: ReconcileCandidate[]; total: number }>(url, config);
}

export function bulkReconcile(pairs: [string, string][], shop: string) {
  return api.post("/reconcile/bulk", { pairs, shop });
}

export function reconcileCheck(
  kodePesanan: string,
  config?: AxiosRequestConfig,
) {
  return api.post<{ message: string }>("/reconcile/check", {
    kode_pesanan: kodePesanan,
  }, config);
}

export function fetchShopeeDetail(invoice: string) {
  return api.get<ShopeeOrderDetail | { status: string; batch_id: number; message: string }>(`/reconcile/status?invoice=${invoice}`);
}

export function checkJobStatus(batchId: number) {
  return api.get<{ batch_id: number; status: string }>(`/reconcile/job-status/${batchId}`);
}

export function fetchEscrowDetail(invoice: string) {
  return api.get<ShopeeEscrowDetail>(`/reconcile/escrow?invoice=${invoice}`);
}

export function fetchShopeeToken(invoice: string) {
  return api.get<{ access_token: string }>(
    `/reconcile/token?invoice=${invoice}`,
  );
}

export function cancelPurchase(kodePesanan: string) {
  return api.post("/reconcile/cancel", { kode_pesanan: kodePesanan });
}

export function updateShopeeStatus(invoice: string) {
  return api.post("/reconcile/update_status", { invoice });
}

export function updateShopeeStatuses(
  invoices: string[],
  config?: AxiosRequestConfig,
) {
  return api.post("/reconcile/update_statuses", { invoices }, config);
}

export function createReconcileBatch(
  shop: string,
  order: string,
  status: string,
  from: string,
  to: string,
) {
  return api.post("/reconcile/batch", { shop, order, status, from, to });
}
