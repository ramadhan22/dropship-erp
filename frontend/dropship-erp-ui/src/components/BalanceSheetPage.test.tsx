// File: src/components/BalanceSheetPage.test.tsx
import "@testing-library/jest-dom";
import { render } from "@testing-library/react";
import { fireEvent, screen, waitFor } from "@testing-library/dom";
import * as api from "../api";
import * as pl from "../api/pl";
import BalanceSheetPage from "./BalanceSheetPage";

jest.mock("../api", () => ({
  fetchBalanceSheet: jest.fn().mockResolvedValue({ data: [] }),
  listAllStores: jest.fn().mockResolvedValue([]),
}));
jest.mock("../api/pl", () => ({
  fetchProfitLoss: jest.fn().mockResolvedValue({
    data: { labaRugiBersih: { amount: 0 } },
  }),
}));

describe("BalanceSheetPage", () => {
  it("auto fetch on mount", async () => {
    render(<BalanceSheetPage />);
    await waitFor(() => expect(api.fetchBalanceSheet).toHaveBeenCalled());
    await waitFor(() => expect(pl.fetchProfitLoss).toHaveBeenCalled());
  });

  it("fetch & display", async () => {
    const mock = [{ category: "Assets", accounts: [], total: 100 }];
    jest
      .spyOn(api, "fetchBalanceSheet")
      // cast to any so TS accepts the mock return type
      .mockResolvedValue({ data: mock } as any);
    jest.spyOn(pl, "fetchProfitLoss").mockResolvedValue({
      data: { labaRugiBersih: { amount: 50 } },
    } as any);
    render(<BalanceSheetPage />);
    fireEvent.change(screen.getByLabelText(/Shop/i), {
      target: { value: "S" },
    });
    fireEvent.click(screen.getByRole("button", { name: /Fetch/i }));
    await waitFor(() =>
      expect(screen.queryAllByText(/Assets/i).length).toBeGreaterThan(0),
    );
    await waitFor(() => expect(pl.fetchProfitLoss).toHaveBeenCalled());
  });
});
