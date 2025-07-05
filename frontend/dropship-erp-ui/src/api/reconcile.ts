import { api } from "./index";
import type {
  ReconciledTransaction,
  ReconcileCandidate,
  ShopeeOrderDetail,
} from "../types";

export function listUnmatched(shop: string) {
  return api.get<ReconciledTransaction[]>(`/reconcile/unmatched?shop=${shop}`);
}

export function listCandidates(
  shop: string,
  order?: string,
  from?: string,
  to?: string,
) {
  const q = new URLSearchParams();
  if (shop) q.append("shop", shop);
  if (order) q.append("order", order);
  if (from) q.append("from", from);
  if (to) q.append("to", to);
  const qs = q.toString();
  const url = qs ? `/reconcile/candidates?${qs}` : "/reconcile/candidates";
  return api.get<ReconcileCandidate[]>(url);
}

export function bulkReconcile(pairs: [string, string][], shop: string) {
  return api.post("/reconcile/bulk", { pairs, shop });
}

export function reconcileCheck(kodePesanan: string) {
  return api.post<{ message: string }>("/reconcile/check", {
    kode_pesanan: kodePesanan,
  });
}

export function fetchShopeeDetail(invoice: string) {
  return api.get<ShopeeOrderDetail>(`/reconcile/status?invoice=${invoice}`);
}

export function fetchShopeeToken(invoice: string) {
  return api.get<{ access_token: string }>(`/reconcile/token?invoice=${invoice}`);
}

export function cancelPurchase(kodePesanan: string) {
  return api.post("/reconcile/cancel", { kode_pesanan: kodePesanan });
}

export function updateShopeeStatus(invoice: string) {
  return api.post("/reconcile/update_status", { invoice });
}
