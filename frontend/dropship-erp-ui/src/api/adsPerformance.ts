import { api } from './index';

export interface AdsPerformance {
  id: number;
  store_id: number;
  campaign_id: string;
  campaign_name: string;
  campaign_type: string;
  campaign_status: string;
  date_from: string;
  date_to: string;
  ads_viewed: number;
  total_clicks: number;
  orders_count: number;
  products_sold: number;
  sales_from_ads: number;
  ad_costs: number;
  click_rate: number;
  roas: number;
  daily_budget: number;
  target_roas: number;
  performance_change_percentage: number;
  created_at: string;
  updated_at: string;
}

export interface AdsPerformanceSummary {
  total_ads_viewed: number;
  total_clicks: number;
  total_orders: number;
  total_products_sold: number;
  total_sales_from_ads: number;
  total_ad_costs: number;
  average_click_rate: number;
  average_roas: number;
  date_from: string;
  date_to: string;
}

export interface AdsPerformanceFilter {
  store_id?: number;
  campaign_status?: string;
  campaign_type?: string;
  date_from?: string;
  date_to?: string;
  limit?: number;
  offset?: number;
}

export interface AdsPerformanceResponse {
  ads: AdsPerformance[];
  limit: number;
  offset: number;
  count: number;
}

export interface RefreshAdsDataRequest {
  date_from: string;
  date_to: string;
  store_id?: number;
}

export interface RefreshAdsDataResponse {
  message: string;
  date_from: string;
  date_to: string;
}

export async function getAdsPerformance(filter: AdsPerformanceFilter = {}): Promise<AdsPerformanceResponse> {
  const params = new URLSearchParams();
  
  if (filter.store_id) params.append('store_id', filter.store_id.toString());
  if (filter.campaign_status) params.append('campaign_status', filter.campaign_status);
  if (filter.campaign_type) params.append('campaign_type', filter.campaign_type);
  if (filter.date_from) params.append('date_from', filter.date_from);
  if (filter.date_to) params.append('date_to', filter.date_to);
  if (filter.limit) params.append('limit', filter.limit.toString());
  if (filter.offset) params.append('offset', filter.offset.toString());

  const response = await api.get(`/ads-performance?${params.toString()}`);
  return response.data;
}

export async function getAdsPerformanceSummary(filter: Omit<AdsPerformanceFilter, 'limit' | 'offset'> = {}): Promise<AdsPerformanceSummary> {
  const params = new URLSearchParams();
  
  if (filter.store_id) params.append('store_id', filter.store_id.toString());
  if (filter.campaign_status) params.append('campaign_status', filter.campaign_status);
  if (filter.campaign_type) params.append('campaign_type', filter.campaign_type);
  if (filter.date_from) params.append('date_from', filter.date_from);
  if (filter.date_to) params.append('date_to', filter.date_to);

  const response = await api.get(`/ads-performance/summary?${params.toString()}`);
  return response.data;
}

export async function refreshAdsData(request: RefreshAdsDataRequest): Promise<RefreshAdsDataResponse> {
  const response = await api.post('/ads-performance/refresh', request);
  return response.data;
}