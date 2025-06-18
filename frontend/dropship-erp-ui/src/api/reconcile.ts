import { api } from "./index";
import type { ReconciledTransaction, ReconcileCandidate } from "../types";

export function listUnmatched(shop: string) {
  return api.get<ReconciledTransaction[]>(`/reconcile/unmatched?shop=${shop}`);
}

export function listCandidates(shop: string, order?: string) {
  const q = new URLSearchParams();
  if (shop) q.append("shop", shop);
  if (order) q.append("order", order);
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
