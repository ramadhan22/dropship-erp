// File: src/components/ShopeeImport.test.tsx

import "@testing-library/jest-dom";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import * as api from "../api";
import ShopeeImport from "./ShopeeImport";

// Mock the api module so importShopee calls do not hit the network
jest.mock("../api", () => ({
  importShopee: jest.fn(),
}));

describe("ShopeeImport", () => {
  it("shows success message on successful import", async () => {
    // Arrange: mock implementation of importShopee to resolve
    (api.importShopee as jest.Mock).mockResolvedValue({
      data: { inserted: 2 },
    });

    render(<ShopeeImport />);
    const file = new File(["data"], "orders.xlsx", {
      type: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
    });
    fireEvent.change(screen.getByLabelText(/XLSX file/i), {
      target: { files: [file] },
    });
    fireEvent.click(screen.getByRole("button", { name: /Import/i }));

    await waitFor(() =>
      expect(
        screen.getByText(/Imported 2 rows successfully!/i),
      ).toBeInTheDocument(),
    );
  });

  it("shows error on failure", async () => {
    // Arrange: mock implementation to reject
    (api.importShopee as jest.Mock).mockRejectedValue(new Error("fail"));

    render(<ShopeeImport />);
    const file = new File(["bad"], "bad.xlsx", {
      type: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
    });
    fireEvent.change(screen.getByLabelText(/XLSX file/i), {
      target: { files: [file] },
    });
    fireEvent.click(screen.getByRole("button", { name: /Import/i }));

    await waitFor(() => expect(screen.getByText(/fail/i)).toBeInTheDocument());
  });
});
