import "@testing-library/jest-dom";
import { render } from "@testing-library/react";
import { waitFor, screen, fireEvent } from "@testing-library/dom";
import * as api from "../api/reconcile";
import ReconcileDashboard from "./ReconcileDashboard";

jest.mock("../api", () => ({
  listAllStores: jest.fn().mockResolvedValue([{ store_id: 1, nama_toko: "S", jenis_channel_id: 1 }]),
}));

jest.mock("../api/reconcile", () => ({
  listCandidates: jest.fn().mockResolvedValue({ data: [] }),
  bulkReconcile: jest.fn(),
}));

beforeEach(() => {
  jest.clearAllMocks();
});

test("load all data on mount", async () => {
  render(<ReconcileDashboard />);
  await waitFor(() => expect(api.listCandidates).toHaveBeenCalledWith(""));
});

test("load candidates with filter", async () => {
  render(<ReconcileDashboard />);
  await waitFor(() => expect(api.listCandidates).toHaveBeenCalled());
  jest.clearAllMocks();
  await screen.findByText("S");
  fireEvent.change(screen.getByLabelText(/Shop/i), { target: { value: "S" } });
  fireEvent.click(screen.getByRole("button", { name: /Refresh/i }));
  await waitFor(() => expect(api.listCandidates).toHaveBeenCalledWith("S"));
});

test("filter status mismatch", async () => {
  (api.listCandidates as jest.Mock).mockResolvedValueOnce({
    data: [
      {
        kode_pesanan: "A",
        nama_toko: "X",
        status_pesanan_terakhir: "diproses",
        no_pesanan: "1",
      },
      {
        kode_pesanan: "B",
        nama_toko: "X",
        status_pesanan_terakhir: "diproses",
        no_pesanan: null,
      },
    ],
  });
  render(<ReconcileDashboard />);
  await screen.findByText("S");
  fireEvent.change(screen.getByLabelText(/Shop/i), { target: { value: "S" } });
  await waitFor(() => expect(screen.getByLabelText(/Shop/i)).toHaveValue("S"));
  await waitFor(() => expect(api.listCandidates).toHaveBeenCalled());
  fireEvent.click(screen.getByLabelText(/Status mismatch only/i));
  expect(screen.queryByText("B")).toBeNull();
});
