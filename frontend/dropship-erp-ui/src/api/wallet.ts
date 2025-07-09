import { api } from "./index";
import type { WalletTransaction } from "../types";

export function listWalletTransactions(params: {
  store: string;
  page_no?: number;
  page_size?: number;
  create_time_from?: number;
  create_time_to?: number;
  wallet_type?: string;
  transaction_type?: string;
  money_flow?: string;
  transaction_tab_type?: string;
}) {
  const q = new URLSearchParams();
  q.append("store", params.store);
  if (params.page_no != null) q.append("page_no", String(params.page_no));
  if (params.page_size != null) q.append("page_size", String(params.page_size));
  if (params.create_time_from != null)
    q.append("create_time_from", String(params.create_time_from));
  if (params.create_time_to != null)
    q.append("create_time_to", String(params.create_time_to));
  if (params.wallet_type) q.append("wallet_type", params.wallet_type);
  if (params.transaction_type)
    q.append("transaction_type", params.transaction_type);
  if (params.money_flow) q.append("money_flow", params.money_flow);
  if (params.transaction_tab_type)
    q.append("transaction_tab_type", params.transaction_tab_type);
  return api.get<{ data: WalletTransaction[]; has_next_page: boolean }>(
    `/wallet/transactions?${q.toString()}`,
  );
}
