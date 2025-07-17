import { api } from "./index";
import type { ShopeeReturnResponse } from "../types";

export interface ShopeeReturnListParams {
  store?: string;
  page_no?: string;
  page_size?: string;
  create_time_from?: string;
  create_time_to?: string;
  update_time_from?: string;
  update_time_to?: string;
  status?: string;
  negotiation_status?: string;
  seller_proof_status?: string;
  seller_compensation_status?: string;
}

export const listShopeeReturns = (params: ShopeeReturnListParams = {}) => {
  const q = new URLSearchParams();
  
  Object.entries(params).forEach(([key, value]) => {
    if (value) {
      q.append(key, value);
    }
  });

  return api.get<ShopeeReturnResponse>(`/shopee/returns?${q.toString()}`);
};