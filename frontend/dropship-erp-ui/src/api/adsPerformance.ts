import { api } from "./index";

// API functions for ads performance
export const fetchAdsCampaigns = async (params?: { 
  store_id?: number; 
  status?: string; 
  limit?: number; 
  offset?: number; 
}) => {
  const searchParams = new URLSearchParams();
  if (params?.store_id) searchParams.set("store_id", params.store_id.toString());
  if (params?.status) searchParams.set("status", params.status);
  if (params?.limit) searchParams.set("limit", params.limit.toString());
  if (params?.offset) searchParams.set("offset", params.offset.toString());

  const response = await api.get(`/ads/campaigns?${searchParams}`);
  return response.data;
};

export const fetchAdsPerformanceSummary = async (params?: {
  store_id?: number;
  start_date?: string;
  end_date?: string;
}) => {
  const searchParams = new URLSearchParams();
  if (params?.store_id) searchParams.set("store_id", params.store_id.toString());
  if (params?.start_date) searchParams.set("start_date", params.start_date);
  if (params?.end_date) searchParams.set("end_date", params.end_date);

  const response = await api.get(`/ads/summary?${searchParams}`);
  return response.data;
};

export const fetchAdsCampaignsFromShopee = async (data: {
  store_id: number;
}) => {
  const response = await api.post("/ads/campaigns/fetch", data);
  return response.data;
};

export const fetchAdsPerformanceFromShopee = async (data: {
  store_id: number;
  campaign_id: number;
  start_date: string;
  end_date: string;
}) => {
  const response = await api.post("/ads/performance/fetch", data);
  return response.data;
};

export const syncHistoricalAdsPerformance = async (data: {
  store_id: number;
}) => {
  const response = await api.post("/ads/sync/historical", data);
  return response.data;
};