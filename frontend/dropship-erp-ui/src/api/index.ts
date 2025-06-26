import axios from "axios";
import { loadingEmitter } from "../loadingEmitter";
import type {
  BalanceCategory,
  Metric,
  JenisChannel,
  Store,
  StoreWithChannel,
  DropshipPurchase,
  DropshipPurchaseDetail,
  Account,
  ShopeeSettled,
  ShopeeSettledSummary,
  ShopeeAffiliateSale,
  ShopeeAffiliateSummary,
  SalesProfit,
} from "../types";

export interface ImportResponse {
  inserted: number;
}

// Base URL for API calls. In Jest/Node we read from process.env; in Vite builds
// you can still set VITE_API_URL, otherwise we fall back to localhost.
// Base URL for API calls – in Vite builds import.meta.env is available; otherwise we default to localhost
let BASE_URL = "http://localhost:8080/api";

if (typeof process !== "undefined" && process.env?.VITE_API_URL) {
  BASE_URL = process.env.VITE_API_URL;
}

try {
  // Access import.meta dynamically so tests running in CommonJS don't fail
  const meta = Function("return import.meta")();
  if (meta?.env?.VITE_API_URL) {
    BASE_URL = meta.env.VITE_API_URL;
  }
} catch {
  // ignore if import.meta is not available
}

export const api = axios.create({
  baseURL: BASE_URL,
});

// Global loading indicator hooks into axios requests
api.interceptors.request.use((config) => {
  loadingEmitter.start();
  return config;
});

api.interceptors.response.use(
  (res) => {
    loadingEmitter.end();
    return res;
  },
  (err) => {
    loadingEmitter.end();
    return Promise.reject(err);
  },
);

// Dropship import
export function importDropship(file: File, channel?: string) {
  const data = new FormData();
  data.append("file", file);
  if (channel) {
    data.append("channel", channel);
  }
  return api.post<ImportResponse>("/dropship/import", data, {
    headers: { "Content-Type": "multipart/form-data" },
  });
}

// Shopee import
export function importShopee(files: File[]) {
  const data = new FormData();
  for (const file of files) {
    data.append("file", file);
  }
  return api.post<ImportResponse>("/shopee/import", data, {
    headers: { "Content-Type": "multipart/form-data" },
  });
}

export function importShopeeAffiliate(files: File[]) {
  const data = new FormData();
  for (const file of files) {
    data.append("file", file);
  }
  return api.post<ImportResponse>("/shopee/affiliate", data, {
    headers: { "Content-Type": "multipart/form-data" },
  });
}

export const confirmShopeeSettle = (orderSN: string) =>
  api.post<{ success: boolean }>(`/shopee/settle/${orderSN}`);

export const getShopeeSettleDetail = (orderSN: string) =>
  api.get<{ data: ShopeeSettled; dropship_total: number }>(
    `/shopee/settled/${orderSN}`,
  );

// Reconcile
export function reconcile(purchaseId: string, orderId: string, shop: string) {
  return api.post("/reconcile", {
    purchase_id: purchaseId,
    order_id: orderId,
    shop,
  });
}

// Compute metrics
export function computeMetrics(shop: string, period: string) {
  return api.post("/metrics", { shop, period });
}

// Fetch cached metrics
export function fetchMetrics(shop: string, period: string) {
  return api.get<Metric>(`/metrics?shop=${shop}&period=${period}`);
}

// Fetch balance sheet
export function fetchBalanceSheet(shop: string, period: string) {
  return api.get<BalanceCategory[]>(
    `/balancesheet?shop=${shop}&period=${period}`,
  );
}

// === JenisChannel & Store Master Data ===
export function createJenisChannel(jenisChannel: string) {
  return api.post("/jenis-channels", { jenis_channel: jenisChannel });
}

export function createStore(jenisChannelId: number, namaToko: string) {
  return api.post("/stores", {
    jenis_channel_id: jenisChannelId,
    nama_toko: namaToko,
  });
}

export function updateStore(id: number, data: Partial<Store>) {
  return api.put(`/stores/${id}`, data);
}

export function deleteStore(id: number) {
  return api.delete(`/stores/${id}`);
}

export function fetchShopeeAuthURL() {
  return api.get<{ url: string }>("/config/shopee-auth-url");
}

export function listJenisChannels() {
  return api.get<JenisChannel[]>("/jenis-channels");
}

export function listStores(channelId: number) {
  return api.get<Store[]>(`/jenis-channels/${channelId}/stores`);
}

export function listStoresByChannelName(channel: string) {
  const q = new URLSearchParams({ channel });
  return api.get<Store[]>(`/stores?${q.toString()}`);
}

export function listAllStoresDirect() {
  return api.get<StoreWithChannel[]>("/stores/all");
}

export function getStore(id: number) {
  return api.get<Store>(`/stores/${id}`);
}

// Fetch stores across all channels by first listing channels then querying each
// channel's stores. Returns a flat array of Store objects.
export async function listAllStores() {
  const channels = await listJenisChannels().then((r) => r.data);
  const lists = await Promise.all(
    channels.map((c) =>
      listStores(c.jenis_channel_id).then((r) => r.data ?? []),
    ),
  );
  return lists.flat();
}

export function listShopeeSettled(params: {
  channel?: string;
  store?: string;
  from?: string;
  to?: string;
  order?: string;
  sort?: string;
  dir?: string;
  page?: number;
  page_size?: number;
}) {
  const q = new URLSearchParams();
  if (params.channel) q.append("channel", params.channel);
  if (params.store) q.append("store", params.store);
  if (params.from) q.append("from", params.from);
  if (params.to) q.append("to", params.to);
  if (params.order) q.append("order", params.order);
  if (params.sort) q.append("sort", params.sort);
  if (params.dir) q.append("dir", params.dir);
  if (params.page) q.append("page", String(params.page));
  if (params.page_size) q.append("page_size", String(params.page_size));
  return api.get<{ data: ShopeeSettled[]; total: number }>(
    `/shopee/settled?${q.toString()}`,
  );
}

export function sumShopeeSettled(params: {
  channel?: string;
  store?: string;
  from?: string;
  to?: string;
}) {
  const q = new URLSearchParams();
  if (params.channel) q.append("channel", params.channel);
  if (params.store) q.append("store", params.store);
  if (params.from) q.append("from", params.from);
  if (params.to) q.append("to", params.to);
  return api.get<ShopeeSettledSummary>(
    `/shopee/settled/summary?${q.toString()}`,
  );
}

export function listShopeeAffiliate(params: {
  no_pesanan?: string;
  from?: string;
  to?: string;
  page?: number;
  page_size?: number;
}) {
  const q = new URLSearchParams();
  if (params.no_pesanan) q.append("no_pesanan", params.no_pesanan);
  if (params.from) q.append("from", params.from);
  if (params.to) q.append("to", params.to);
  if (params.page) q.append("page", String(params.page));
  if (params.page_size) q.append("page_size", String(params.page_size));
  return api.get<{ data: ShopeeAffiliateSale[]; total: number }>(
    `/shopee/affiliate?${q.toString()}`,
  );
}

export function sumShopeeAffiliate(params: {
  no_pesanan?: string;
  from?: string;
  to?: string;
}) {
  const q = new URLSearchParams();
  if (params.no_pesanan) q.append("no_pesanan", params.no_pesanan);
  if (params.from) q.append("from", params.from);
  if (params.to) q.append("to", params.to);
  return api.get<ShopeeAffiliateSummary>(
    `/shopee/affiliate/summary?${q.toString()}`,
  );
}

export function listSalesProfit(params: {
  channel?: string;
  store?: string;
  from?: string;
  to?: string;
  order?: string;
  sort?: string;
  dir?: string;
  page?: number;
  page_size?: number;
}) {
  const q = new URLSearchParams();
  if (params.channel) q.append("channel", params.channel);
  if (params.store) q.append("store", params.store);
  if (params.from) q.append("from", params.from);
  if (params.to) q.append("to", params.to);
  if (params.order) q.append("order", params.order);
  if (params.sort) q.append("sort", params.sort);
  if (params.dir) q.append("dir", params.dir);
  if (params.page) q.append("page", String(params.page));
  if (params.page_size) q.append("page_size", String(params.page_size));
  return api.get<{ data: SalesProfit[]; total: number }>(
    `/sales?${q.toString()}`,
  );
}

export interface DropshipPurchaseList {
  data: DropshipPurchase[];
  total: number;
}

export interface ProductSales {
  nama_produk: string;
  total_qty: number;
  total_value: number;
}

export function listDropshipPurchases(params: {
  channel?: string;
  store?: string;
  from?: string;
  to?: string;
  order?: string;
  sort?: string;
  dir?: string;
  page?: number;
  page_size?: number;
}) {
  const q = new URLSearchParams();
  if (params.channel) q.append("channel", params.channel);
  if (params.store) q.append("store", params.store);
  if (params.from) q.append("from", params.from);
  if (params.to) q.append("to", params.to);
  if (params.order) q.append("order", params.order);
  if (params.sort) q.append("sort", params.sort);
  if (params.dir) q.append("dir", params.dir);
  if (params.page) q.append("page", String(params.page));
  if (params.page_size) q.append("page_size", String(params.page_size));
  const qs = q.toString();
  const url = qs ? `/dropship/purchases?${qs}` : "/dropship/purchases";
  return api.get<DropshipPurchaseList>(url);
}

export function sumDropshipPurchases(params: {
  channel?: string;
  store?: string;
  from?: string;
  to?: string;
}) {
  const q = new URLSearchParams();
  if (params.channel) q.append("channel", params.channel);
  if (params.store) q.append("store", params.store);
  if (params.from) q.append("from", params.from);
  if (params.to) q.append("to", params.to);
  return api.get<{ total: number }>(
    `/dropship/purchases/summary?${q.toString()}`,
  );
}

export function fetchTopProducts(params: {
  channel?: string;
  store?: string;
  from?: string;
  to?: string;
  limit?: number;
}) {
  const q = new URLSearchParams();
  if (params.channel) q.append("channel", params.channel);
  if (params.store) q.append("store", params.store);
  if (params.from) q.append("from", params.from);
  if (params.to) q.append("to", params.to);
  if (params.limit) q.append("limit", String(params.limit));
  return api.get<ProductSales[]>(`/dropship/top-products?${q.toString()}`);
}
// === Accounts CRUD ===
export function createAccount(acc: Partial<Account>) {
  return api.post("/accounts", acc);
}

export function listAccounts() {
  return api.get<Account[]>("/accounts");
}

export function getAccount(id: number) {
  return api.get<Account>(`/accounts/${id}`);
}

export function updateAccount(id: number, acc: Partial<Account>) {
  return api.put(`/accounts/${id}`, acc);
}

export function deleteAccount(id: number) {
  return api.delete(`/accounts/${id}`);
}

export function getDropshipPurchaseDetails(id: string) {
  return api.get<DropshipPurchaseDetail[]>(`/dropship/purchases/${id}/details`);
}

export const withdrawShopeeBalance = (store: string, amount: number) =>
  api.post("/withdraw", { store, amount });

export const fetchPendingBalance = (store: string) =>
  api.get<{ pending_balance: number }>(`/pending-balance?store=${store}`);
