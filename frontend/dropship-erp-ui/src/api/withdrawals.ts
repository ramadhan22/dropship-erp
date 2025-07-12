import { api } from "./index";
import type { Withdrawal, WalletTransaction } from "../types";

export const createWithdrawal = (w: Partial<Withdrawal>) =>
  api.post<Withdrawal>("/withdrawals/", w);

export const listWithdrawals = () =>
  api.get<Withdrawal[]>("/withdrawals/");

export const importWithdrawals = (file: File) => {
  const data = new FormData();
  data.append("file", file);
  return api.post("/withdrawals/import", data, {
    headers: { "Content-Type": "multipart/form-data" },
  });
};

export function listWalletWithdrawals(params: {
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
    `/wallet-withdrawals?${q.toString()}`,
  );
}

export function createWalletWithdrawalJournal(payload: {
  store: string;
  transaction_id: number;
  create_time: number;
  amount: number;
}) {
  return api.post(`/wallet-withdrawals/journal`, payload);
}

export function createAllWalletWithdrawalJournal(payload: { store: string }) {
  return api.post(`/wallet-withdrawals/journal-all`, payload);
}
