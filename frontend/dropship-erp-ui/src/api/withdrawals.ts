import { api } from "./index";
import type { Withdrawal } from "../types";

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
