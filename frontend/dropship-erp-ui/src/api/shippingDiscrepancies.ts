import { api } from "./index";
import type { ShippingDiscrepancy } from "../types";

export const listShippingDiscrepancies = (params: {
  store_name?: string;
  type?: string;
  page?: number;
  page_size?: number;
}) => {
  const q = new URLSearchParams();
  if (params.store_name) q.append("store_name", params.store_name);
  if (params.type) q.append("type", params.type);
  if (params.page) q.append("page", params.page.toString());
  if (params.page_size) q.append("page_size", params.page_size.toString());
  return api.get<{
    data: ShippingDiscrepancy[];
    page: number;
    page_size: number;
    total_count: number;
    store_name?: string;
    type?: string;
  }>(`/shipping-discrepancies/?${q.toString()}`);
};

export const getShippingDiscrepancyStats = (params: {
  start_date?: string;
  end_date?: string;
  type?: string; // "amounts" or "counts"
}) => {
  const q = new URLSearchParams();
  if (params.start_date) q.append("start_date", params.start_date);
  if (params.end_date) q.append("end_date", params.end_date);
  if (params.type) q.append("type", params.type);
  return api.get<{
    start_date: string;
    end_date: string;
    type: string;
    stats: Record<string, number>;
  }>(`/shipping-discrepancies/stats?${q.toString()}`);
};

export const getShippingDiscrepancyByInvoice = (invoice: string) =>
  api.get<ShippingDiscrepancy>(`/shipping-discrepancies/invoice/${invoice}`);