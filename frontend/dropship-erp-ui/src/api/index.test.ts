// File: src/api/index.test.ts

import "@testing-library/jest-dom";
import {
  api,
  computeMetrics,
  fetchBalanceSheet,
  fetchMetrics,
  importDropship,
  importShopee,
  reconcile,
} from "./index";

// Turn the axios‐style api.post and api.get into Jest mocks
(api.post as jest.Mock) = jest.fn();
(api.get as jest.Mock) = jest.fn();

describe("API layer", () => {
  beforeEach(() => {
    // Clear call history between tests
    (api.post as jest.Mock).mockClear();
    (api.get as jest.Mock).mockClear();
  });

  it("importDropship calls api.post correctly and resolves data", async () => {
    (api.post as jest.Mock).mockResolvedValue({ data: { inserted: 2 } });

    const file = new File(["data"], "file.csv", { type: "text/csv" });
    const result = await importDropship(file);
    expect(api.post).toHaveBeenCalledWith(
      "/dropship/import",
      expect.any(FormData),
      { headers: { "Content-Type": "multipart/form-data" } },
    );
    expect(result).toEqual({ data: { inserted: 2 } });
  });

  it("importShopee calls api.post correctly and resolves data", async () => {
    (api.post as jest.Mock).mockResolvedValue({ data: { inserted: 5 } });

    const file = new File(["data"], "orders.xlsx", {
      type: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
    });
    const result = await importShopee(file);
    expect(api.post).toHaveBeenCalledWith(
      "/shopee/import",
      expect.any(FormData),
      { headers: { "Content-Type": "multipart/form-data" } },
    );
    expect(result).toEqual({ data: { inserted: 5 } });
  });

  it("reconcile calls api.post correctly and resolves", async () => {
    (api.post as jest.Mock).mockResolvedValue({ data: { matched: true } });

    const result = await reconcile("P1", "O1", "ShopX");
    expect(api.post).toHaveBeenCalledWith("/reconcile", {
      purchase_id: "P1",
      order_id: "O1",
      shop: "ShopX",
    });
    expect(result).toEqual({ data: { matched: true } });
  });

  it("computeMetrics calls api.post correctly and resolves", async () => {
    (api.post as jest.Mock).mockResolvedValue({ data: {} });

    await expect(computeMetrics("ShopX", "2025-05")).resolves.toEqual({
      data: {},
    });
    expect(api.post).toHaveBeenCalledWith("/metrics", {
      shop: "ShopX",
      period: "2025-05",
    });
  });

  it("fetchMetrics calls api.get and returns typed data", async () => {
    const fakeMetric = {
      shop_username: "ShopX",
      period: "2025-05",
      sum_revenue: 0,
      sum_cogs: 0,
      sum_fees: 0,
      net_profit: 42,
      ending_cash_balance: 100,
    };
    (api.get as jest.Mock).mockResolvedValue({ data: fakeMetric });

    const res = await fetchMetrics("ShopX", "2025-05");
    expect(api.get).toHaveBeenCalledWith(`/metrics?shop=ShopX&period=2025-05`);
    expect(res.data).toEqual(fakeMetric);
  });

  it("fetchBalanceSheet calls api.get and returns typed array", async () => {
    const fakeSheet = [{ category: "Assets", accounts: [], total: 500 }];
    (api.get as jest.Mock).mockResolvedValue({ data: fakeSheet });

    const res = await fetchBalanceSheet("ShopX", "2025-05");
    expect(api.get).toHaveBeenCalledWith(
      `/balancesheet?shop=ShopX&period=2025-05`,
    );
    expect(res.data).toEqual(fakeSheet);
  });
});
