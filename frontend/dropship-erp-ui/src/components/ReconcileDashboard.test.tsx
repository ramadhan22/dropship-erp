import "@testing-library/jest-dom";
import { render } from "@testing-library/react";
import { waitFor, screen, fireEvent } from "@testing-library/dom";
import * as api from "../api/reconcile";
import ReconcileDashboard from "./ReconcileDashboard";

jest.mock("../api/reconcile", () => ({
  listCandidates: jest.fn().mockResolvedValue({ data: [] }),
  bulkReconcile: jest.fn(),
}));

test("load candidates", async () => {
  render(<ReconcileDashboard />);
  fireEvent.change(screen.getByLabelText(/Shop/i), { target: { value: "S" } });
  fireEvent.click(screen.getByRole("button", { name: /Refresh/i }));
  await waitFor(() => expect(api.listCandidates).toHaveBeenCalled());
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
  fireEvent.change(screen.getByLabelText(/Shop/i), { target: { value: "S" } });
  await waitFor(() => expect(api.listCandidates).toHaveBeenCalled());
  await screen.findByText("A");
  fireEvent.click(screen.getByLabelText(/Status mismatch only/i));
  expect(screen.queryByText("B")).toBeNull();
});
