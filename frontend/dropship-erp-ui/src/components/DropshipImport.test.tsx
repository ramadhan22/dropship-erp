// File: src/components/DropshipImport.test.tsx

import "@testing-library/jest-dom";
import {
  fireEvent,
  render,
  screen,
  waitFor,
  within,
} from "@testing-library/react";
import * as api from "../api";
import DropshipImport from "./DropshipImport";

// Mock the api module so importDropship calls do not hit the network
jest.mock("../api", () => ({
  importDropship: jest.fn(),
  listDropshipPurchases: jest
    .fn()
    .mockResolvedValue({ data: { data: [], total: 0 } }),
  listJenisChannels: jest.fn().mockResolvedValue({ data: [] }),
  listStores: jest.fn().mockResolvedValue({ data: [] }),
}));

describe("DropshipImport", () => {
  it("renders import button and opens dialog", async () => {
    render(<DropshipImport />);
    await waitFor(() => expect(api.listDropshipPurchases).toHaveBeenCalled());
    const btn = screen.getByRole("button", { name: /Import/i });
    expect(btn).toBeInTheDocument();
    fireEvent.click(btn);
    expect(screen.getByLabelText(/CSV file/i)).toBeInTheDocument();
  });

  it("shows success message on successful import", async () => {
    // Arrange: mock implementation of importDropship to resolve
    (api.importDropship as jest.Mock).mockResolvedValue({
      data: { inserted: 3 },
    });

    render(<DropshipImport />);
    await waitFor(() => expect(api.listDropshipPurchases).toHaveBeenCalled());
    fireEvent.click(screen.getByRole("button", { name: /Import/i }));
    const dialog = await screen.findByRole("dialog");
    const file = new File(["data"], "data.csv", { type: "text/csv" });
    fireEvent.change(within(dialog).getByLabelText(/CSV file/i), {
      target: { files: [file] },
    });
    fireEvent.click(within(dialog).getByRole("button", { name: /Import/i }));

    await waitFor(() =>
      expect(
        screen.getByText(/Imported 3 rows successfully!/i),
      ).toBeInTheDocument(),
    );
  });

  it("shows error message on failure", async () => {
    (api.importDropship as jest.Mock).mockRejectedValue(new Error("boom"));

    render(<DropshipImport />);
    await waitFor(() => expect(api.listDropshipPurchases).toHaveBeenCalled());
    fireEvent.click(screen.getByRole("button", { name: /Import/i }));
    const dialog = await screen.findByRole("dialog");
    const file = new File(["bad"], "bad.csv", { type: "text/csv" });
    fireEvent.change(within(dialog).getByLabelText(/CSV file/i), {
      target: { files: [file] },
    });
    fireEvent.click(within(dialog).getByRole("button", { name: /Import/i }));

    await waitFor(() => expect(screen.getByText(/boom/i)).toBeInTheDocument());
  });
});
