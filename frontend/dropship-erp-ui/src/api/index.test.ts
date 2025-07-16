// File: src/api/index.test.ts

import "@testing-library/jest-dom";
import {
  api,
  fetchBalanceSheet,
  importDropship,
  createJenisChannel,
  createStore,
  listJenisChannels,
  listStores,
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
    (api.post as jest.Mock).mockResolvedValue({ data: { queued: 2 } });

    const fileA = new File(["data"], "a.csv", { type: "text/csv" });
    const fileB = new File(["data"], "b.csv", { type: "text/csv" });
    const result = await importDropship([fileA, fileB], "Shopee");
    expect(api.post).toHaveBeenCalledWith(
      "/dropship/import",
      expect.any(FormData),
      { headers: { "Content-Type": "multipart/form-data" } },
    );
    expect(result).toEqual({ data: { queued: 2 } });
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
  it("createJenisChannel posts correctly", async () => {
    (api.post as jest.Mock).mockResolvedValue({
      data: { jenis_channel_id: 1 },
    });

    const res = await createJenisChannel("Tokopedia");
    expect(api.post).toHaveBeenCalledWith("/jenis-channels", {
      jenis_channel: "Tokopedia",
    });
    expect(res).toEqual({ data: { jenis_channel_id: 1 } });
  });

  it("createStore posts correctly", async () => {
    (api.post as jest.Mock).mockResolvedValue({ data: { store_id: 2 } });

    const res = await createStore(1, "ShopA");
    expect(api.post).toHaveBeenCalledWith("/stores", {
      jenis_channel_id: 1,
      nama_toko: "ShopA",
    });
    expect(res).toEqual({ data: { store_id: 2 } });
  });

  it("listJenisChannels fetches list", async () => {
    (api.get as jest.Mock).mockResolvedValue({ data: [] });

    const res = await listJenisChannels();
    expect(api.get).toHaveBeenCalledWith("/jenis-channels");
    expect(res.data).toEqual([]);
  });

  it("listStores fetches stores for channel", async () => {
    (api.get as jest.Mock).mockResolvedValue({ data: [] });

    const res = await listStores(3);
    expect(api.get).toHaveBeenCalledWith("/jenis-channels/3/stores");
    expect(res.data).toEqual([]);
  });
});
