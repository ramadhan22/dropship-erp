import { api } from "./index";
import type { KasAccountBalance } from "../types";

export function listKasAccounts() {
  return api.get<KasAccountBalance[]>("/asset-accounts/");
}

export function adjustKasBalance(id: number, balance: number) {
  return api.put(`/asset-accounts/${id}/balance`, { balance });
}
