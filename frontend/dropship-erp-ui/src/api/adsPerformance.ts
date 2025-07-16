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

  const response = await fetch(`/api/ads/campaigns?${searchParams}`);
  if (!response.ok) throw new Error("Failed to fetch campaigns");
  return response.json();
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

  const response = await fetch(`/api/ads/summary?${searchParams}`);
  if (!response.ok) throw new Error("Failed to fetch summary");
  return response.json();
};

export const fetchAdsCampaignsFromShopee = async (data: {
  store_id: number;
}) => {
  const response = await fetch("/api/ads/campaigns/fetch", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(data),
  });
  if (!response.ok) throw new Error("Failed to fetch campaigns from Shopee");
  return response.json();
};

export const fetchAdsPerformanceFromShopee = async (data: {
  store_id: number;
  campaign_id: number;
  start_date: string;
  end_date: string;
}) => {
  const response = await fetch("/api/ads/performance/fetch", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(data),
  });
  if (!response.ok) throw new Error("Failed to fetch performance data from Shopee");
  return response.json();
};

export const syncHistoricalAdsPerformance = async (data: {
  store_id: number;
}) => {
  const response = await fetch("/api/ads/sync/historical", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(data),
  });
  if (!response.ok) throw new Error("Failed to start historical sync");
  return response.json();
};