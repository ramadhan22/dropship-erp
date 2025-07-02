import { api } from "./index";
import type { ShopeeAdjustment } from "../types";

export const importShopeeAdjustments = (file: File) => {
  const data = new FormData();
  data.append("file", file);
  return api.post("/shopee/adjustments/import", data, {
    headers: { "Content-Type": "multipart/form-data" },
  });
};

export const listShopeeAdjustments = (params: {
  from?: string;
  to?: string;
}) => {
  const q = new URLSearchParams();
  if (params.from) q.append("from", params.from);
  if (params.to) q.append("to", params.to);
  return api.get<ShopeeAdjustment[]>(`/shopee/adjustments/?${q.toString()}`);
};

export const updateShopeeAdjustment = (
  id: number,
  data: Partial<ShopeeAdjustment>
) => api.put(`/shopee/adjustments/${id}`, data);

export const deleteShopeeAdjustment = (id: number) =>
  api.delete(`/shopee/adjustments/${id}`);
