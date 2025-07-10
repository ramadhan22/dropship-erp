import { api } from "./index";
import type { WalletTransaction } from "../types";

export function listAdsTopups(params: {
  store: string;
  page_no?: number;
  page_size?: number;
  create_time_from?: number;
  create_time_to?: number;
}) {
  const q = new URLSearchParams();
  q.append("store", params.store);
  if (params.page_no != null) q.append("page_no", String(params.page_no));
  if (params.page_size != null) q.append("page_size", String(params.page_size));
  if (params.create_time_from != null)
    q.append("create_time_from", String(params.create_time_from));
  if (params.create_time_to != null)
    q.append("create_time_to", String(params.create_time_to));
  return api.get<{ data: WalletTransaction[]; has_next_page: boolean }>(
    `/ads-topups?${q.toString()}`,
  );
}

export function createAdsTopupJournal(payload: {
  store: string;
  transaction_id: number;
  create_time: number;
  amount: number;
}) {
  return api.post(`/ads-topups/journal`, payload);
}
