import { api } from "./index";
import type { Account } from "../types";

export function fetchGeneralLedger(shop: string, from: string, to: string) {
  return api.get<Account[]>(
    `/generalledger?shop=${shop}&from=${from}&to=${to}`,
  );
}
