// File: src/components/DropshipImport.test.tsx

import "@testing-library/jest-dom";
import { render } from "@testing-library/react";
import { fireEvent, screen, waitFor, within } from "@testing-library/dom";
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
    expect(screen.getByLabelText(/CSV files/i)).toBeInTheDocument();
  });

  it("shows success message on successful import", async () => {
    (api.importDropship as jest.Mock)
      .mockResolvedValueOnce({ data: { inserted: 2 } })
      .mockResolvedValueOnce({ data: { inserted: 1 } });

    render(<DropshipImport />);
    await waitFor(() => expect(api.listDropshipPurchases).toHaveBeenCalled());
    fireEvent.click(screen.getByRole("button", { name: /Import/i }));
    const dialog = await screen.findByRole("dialog");
    const file1 = new File(["data1"], "1.csv", { type: "text/csv" });
    const file2 = new File(["data2"], "2.csv", { type: "text/csv" });
    fireEvent.change(within(dialog).getByLabelText(/CSV files/i), {
      target: { files: [file1, file2] },
    });
    fireEvent.click(within(dialog).getByRole("button", { name: /Import/i }));

    await waitFor(() =>
      expect(
        screen.getByText(/Imported 3 rows from 2 files successfully!/i),
      ).toBeInTheDocument(),
    );
    expect(api.importDropship).toHaveBeenCalledTimes(2);
  });

  it("shows error message on failure", async () => {
    (api.importDropship as jest.Mock).mockRejectedValue(new Error("boom"));

    render(<DropshipImport />);
    await waitFor(() => expect(api.listDropshipPurchases).toHaveBeenCalled());
    fireEvent.click(screen.getByRole("button", { name: /Import/i }));
    const dialog = await screen.findByRole("dialog");
    const file = new File(["bad"], "bad.csv", { type: "text/csv" });
    fireEvent.change(within(dialog).getByLabelText(/CSV files/i), {
      target: { files: [file] },
    });
    fireEvent.click(within(dialog).getByRole("button", { name: /Import/i }));

    await waitFor(() => expect(screen.getByText(/boom/i)).toBeInTheDocument());
  });
});
