import { api } from "./index";
import type { Metric } from "../types";

export function fetchPL(shop: string, period: string) {
  return api.get<Metric>(`/pl?shop=${shop}&period=${period}`);
}

export interface ProfitLossParams {
  type: "Monthly" | "Yearly";
  month?: number;
  year: number;
  store?: string;
}

export function fetchProfitLoss(params: ProfitLossParams) {
  return api.get("/profitloss", { params });
}
