import "@testing-library/jest-dom";
import { render } from "@testing-library/react";
import { screen, waitFor } from "@testing-library/dom";
import * as api from "../api";
import SalesSummaryPage from "./SalesSummaryPage";

jest.mock("../api", () => ({
  listJenisChannels: jest.fn().mockResolvedValue({ data: [] }),
  listStoresByChannelName: jest.fn(),
  listShopeeSettled: jest.fn(),
}));

test("displays totals after fetching data", async () => {
  (api.listShopeeSettled as jest.Mock).mockResolvedValue({
    data: {
      data: [
        { waktu_pesanan_dibuat: "2025-05-01T00:00:00Z", total_penerimaan: 100 },
        { waktu_pesanan_dibuat: "2025-05-01T01:00:00Z", total_penerimaan: 50 },
        { waktu_pesanan_dibuat: "2025-05-02T00:00:00Z", total_penerimaan: 200 },
      ],
      total: 3,
    },
  });

  render(<SalesSummaryPage />);

  await waitFor(() => expect(api.listShopeeSettled).toHaveBeenCalled());

  await waitFor(() =>
    expect(screen.getByText(/Total Revenue:/i)).toBeInTheDocument(),
  );

  expect(screen.getByText(/350/)).toBeInTheDocument();
  expect(screen.getByText(/Total Orders:/i)).toBeInTheDocument();
});
