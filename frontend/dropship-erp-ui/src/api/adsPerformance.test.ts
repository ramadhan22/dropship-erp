import { api } from "./index";
import {
  fetchAdsCampaigns,
  fetchAdsPerformanceSummary,
  fetchAdsCampaignsFromShopee,
  fetchAdsPerformanceFromShopee,
  syncHistoricalAdsPerformance,
} from "./adsPerformance";

// Mock the api module
jest.mock("./index", () => ({
  api: {
    get: jest.fn(),
    post: jest.fn(),
  },
}));

const mockApi = api as jest.Mocked<typeof api>;

describe("adsPerformance API", () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  describe("fetchAdsCampaigns", () => {
    it("should call api.get with correct endpoint and return response.data", async () => {
      const mockResponse = { data: { campaigns: [] } };
      mockApi.get.mockResolvedValue(mockResponse);

      const result = await fetchAdsCampaigns({ limit: 100 });

      expect(mockApi.get).toHaveBeenCalledWith("/ads/campaigns?limit=100");
      expect(result).toEqual(mockResponse.data);
    });

    it("should handle params correctly", async () => {
      const mockResponse = { data: { campaigns: [] } };
      mockApi.get.mockResolvedValue(mockResponse);

      await fetchAdsCampaigns({
        store_id: 123,
        status: "active",
        limit: 50,
        offset: 10,
      });

      expect(mockApi.get).toHaveBeenCalledWith(
        "/ads/campaigns?store_id=123&status=active&limit=50&offset=10"
      );
    });
  });

  describe("fetchAdsPerformanceSummary", () => {
    it("should call api.get with correct endpoint and return response.data", async () => {
      const mockResponse = { data: { total_campaigns: 5 } };
      mockApi.get.mockResolvedValue(mockResponse);

      const result = await fetchAdsPerformanceSummary({
        start_date: "2023-01-01",
        end_date: "2023-12-31",
      });

      expect(mockApi.get).toHaveBeenCalledWith(
        "/ads/summary?start_date=2023-01-01&end_date=2023-12-31"
      );
      expect(result).toEqual(mockResponse.data);
    });
  });

  describe("fetchAdsCampaignsFromShopee", () => {
    it("should call api.post with correct endpoint and data", async () => {
      const mockResponse = { data: { success: true } };
      mockApi.post.mockResolvedValue(mockResponse);

      const data = { store_id: 123 };
      const result = await fetchAdsCampaignsFromShopee(data);

      expect(mockApi.post).toHaveBeenCalledWith("/ads/campaigns/fetch", data);
      expect(result).toEqual(mockResponse.data);
    });
  });

  describe("fetchAdsPerformanceFromShopee", () => {
    it("should call api.post with correct endpoint and data", async () => {
      const mockResponse = { data: { success: true } };
      mockApi.post.mockResolvedValue(mockResponse);

      const data = {
        store_id: 123,
        campaign_id: 456,
        start_date: "2023-01-01",
        end_date: "2023-12-31",
      };
      const result = await fetchAdsPerformanceFromShopee(data);

      expect(mockApi.post).toHaveBeenCalledWith("/ads/performance/fetch", data);
      expect(result).toEqual(mockResponse.data);
    });
  });

  describe("syncHistoricalAdsPerformance", () => {
    it("should call api.post with correct endpoint and data", async () => {
      const mockResponse = { data: { batch_id: "abc123" } };
      mockApi.post.mockResolvedValue(mockResponse);

      const data = { store_id: 123 };
      const result = await syncHistoricalAdsPerformance(data);

      expect(mockApi.post).toHaveBeenCalledWith("/ads/sync/historical", data);
      expect(result).toEqual(mockResponse.data);
    });
  });
});