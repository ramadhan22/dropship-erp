import { api } from "./index";
import type { ReconciledTransaction, ReconcileCandidate } from "../types";

export function listUnmatched(shop: string) {
  return api.get<ReconciledTransaction[]>(`/reconcile/unmatched?shop=${shop}`);
}

export function listCandidates(shop: string) {
  return api.get<ReconcileCandidate[]>(`/reconcile/candidates?shop=${shop}`);
}

export function bulkReconcile(pairs: [string, string][], shop: string) {
  return api.post("/reconcile/bulk", { pairs, shop });
}

export function reconcileCheck(kodePesanan: string) {
  return api.post<{ message: string }>("/reconcile/check", {
    kode_pesanan: kodePesanan,
  });
}
