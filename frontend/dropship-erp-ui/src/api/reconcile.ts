import { api } from "./index";
import type { ReconciledTransaction } from "../types";

export function listUnmatched(shop: string) {
  return api.get<ReconciledTransaction[]>(`/reconcile/unmatched?shop=${shop}`);
}

export function bulkReconcile(pairs: [string, string][], shop: string) {
  return api.post("/reconcile/bulk", { pairs, shop });
}
