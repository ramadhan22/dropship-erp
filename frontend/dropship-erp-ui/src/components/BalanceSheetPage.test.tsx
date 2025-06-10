// File: src/components/BalanceSheetPage.test.tsx
import "@testing-library/jest-dom";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import * as api from "../api";
import BalanceSheetPage from "./BalanceSheetPage";

describe("BalanceSheetPage", () => {
  it("fetch & display", async () => {
    const mock = [{ category: "Assets", accounts: [], total: 100 }];
    jest
      .spyOn(api, "fetchBalanceSheet")
      // cast to any so TS accepts the mock return type
      .mockResolvedValue({ data: mock } as any);
    render(<BalanceSheetPage />);
    fireEvent.change(screen.getByLabelText(/Shop/i), {
      target: { value: "S" },
    });
    fireEvent.change(screen.getByLabelText(/Period/i), {
      target: { value: "2025-05" },
    });
    fireEvent.click(screen.getByRole("button", { name: /Fetch/i }));
    await waitFor(() => screen.getByText(/Assets/i));
  });
});
