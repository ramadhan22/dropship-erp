import "@testing-library/jest-dom";
import { render } from "@testing-library/react";
import { screen, waitFor } from "@testing-library/dom";
import * as api from "../api";
import SalesSummaryPage from "./SalesSummaryPage";

jest.mock("../api", () => ({
  listJenisChannels: jest.fn().mockResolvedValue({ data: [] }),
  listStoresByChannelName: jest.fn(),
  fetchDailyPurchaseTotals: jest.fn(),
  fetchTopProducts: jest.fn(),
}));

test("displays totals after fetching data", async () => {
  (api.fetchDailyPurchaseTotals as jest.Mock).mockResolvedValue({
    data: [
      { date: "2025-05-01", total: 150, count: 2 },
      { date: "2025-05-02", total: 200, count: 1 },
    ],
  });

  render(<SalesSummaryPage />);

  await waitFor(() => expect(api.fetchDailyPurchaseTotals).toHaveBeenCalled());

  await waitFor(() =>
    expect(screen.getByText(/Total Revenue:/i)).toBeInTheDocument(),
  );

  expect(screen.getByText(/350/)).toBeInTheDocument();
  expect(screen.getByText(/Total Orders:/i)).toBeInTheDocument();
});
