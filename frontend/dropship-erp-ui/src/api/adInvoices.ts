import { api } from "./index";
import type { AdInvoice } from "../types";

export function importAdInvoice(file: File) {
  const data = new FormData();
  data.append("file", file);
  return api.post("/ad-invoices/", data, {
    headers: { "Content-Type": "multipart/form-data" },
  });
}

export function listAdInvoices(params?: { sort?: string; dir?: string }) {
  const q = new URLSearchParams();
  if (params?.sort) q.append("sort", params.sort);
  if (params?.dir) q.append("dir", params.dir);
  return api.get<AdInvoice[]>(`/ad-invoices/?${q.toString()}`);
}
