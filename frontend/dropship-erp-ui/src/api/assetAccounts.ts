import { api } from "./index";
import type { AssetAccountBalance } from "../types";

export function listAssetAccounts() {
  return api.get<AssetAccountBalance[]>("/asset-accounts/");
}

export function adjustAssetBalance(id: number, balance: number) {
  return api.put(`/asset-accounts/${id}/balance`, { balance });
}
