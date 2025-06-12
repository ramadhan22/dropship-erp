import { api } from "./index";
import type { Metric } from "../types";

export function fetchPL(shop: string, period: string) {
  return api.get<Metric>(`/pl?shop=${shop}&period=${period}`);
}
