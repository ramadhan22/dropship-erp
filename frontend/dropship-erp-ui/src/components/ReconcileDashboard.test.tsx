import "@testing-library/jest-dom";
import { render } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { waitFor, screen, fireEvent } from "@testing-library/dom";
import * as api from "../api/reconcile";
import ReconcileDashboard from "./ReconcileDashboard";

jest.mock("../api", () => ({
  listAllStores: jest
    .fn()
    .mockResolvedValue([{ store_id: 1, nama_toko: "S", jenis_channel_id: 1 }]),
}));

jest.mock("../api/reconcile", () => ({
  listCandidates: jest.fn().mockResolvedValue({ data: { data: [], total: 0 } }),
  reconcileCheck: jest.fn().mockResolvedValue({ data: { message: "ok" } }),
  updateShopeeStatuses: jest.fn().mockResolvedValue({}),
  fetchShopeeDetail: jest.fn().mockResolvedValue({
    data: { order_sn: "INV", order_status: "PROCESSED" },
  }),
  fetchEscrowDetail: jest.fn().mockResolvedValue({
    data: { order_sn: "INV", escrow_amount: 1000 },
  }),
}));

beforeEach(() => {
  jest.clearAllMocks();
});

test("load all data on mount", async () => {
  render(
    <MemoryRouter>
      <ReconcileDashboard />
    </MemoryRouter>,
  );
  await waitFor(() => expect(api.listCandidates).toHaveBeenCalled());
});

test("load candidates with filter", async () => {
  render(
    <MemoryRouter>
      <ReconcileDashboard />
    </MemoryRouter>,
  );
  await waitFor(() => expect(api.listCandidates).toHaveBeenCalled());
  jest.clearAllMocks();
  await screen.findByText("S");
  fireEvent.change(screen.getByLabelText(/Shop/i), { target: { value: "S" } });
  fireEvent.click(screen.getByRole("button", { name: /Refresh/i }));
  await waitFor(() => expect(api.listCandidates).toHaveBeenCalled());
});

test("filter by invoice", async () => {
  render(
    <MemoryRouter>
      <ReconcileDashboard />
    </MemoryRouter>,
  );
  await waitFor(() => expect(api.listCandidates).toHaveBeenCalled());
  jest.clearAllMocks();
  fireEvent.change(screen.getByLabelText(/Search Invoice/i), {
    target: { value: "INV" },
  });
  fireEvent.click(screen.getByRole("button", { name: /Refresh/i }));
  await waitFor(() => expect(api.listCandidates).toHaveBeenCalled());
});

test("click reconcile button", async () => {
  (api.listCandidates as jest.Mock).mockResolvedValueOnce({
    data: {
      data: [
        {
          kode_pesanan: "A",
          kode_invoice_channel: "INV",
          nama_toko: "X",
          status_pesanan_terakhir: "diproses",
          no_pesanan: "INV",
          shopee_order_status: "PROCESSED",
        },
      ],
      total: 1,
    },
  });
  render(
    <MemoryRouter>
      <ReconcileDashboard />
    </MemoryRouter>,
  );
  await screen.findByText("Reconcile");
  fireEvent.click(screen.getByText("Reconcile"));
  await waitFor(() => expect(api.reconcileCheck).toHaveBeenCalledWith("A"));
});

test("reconcile all button", async () => {
  const rows = [
    {
      kode_pesanan: "A",
      kode_invoice_channel: "INV",
      nama_toko: "X",
      status_pesanan_terakhir: "diproses",
      no_pesanan: "INV",
      shopee_order_status: "PROCESSED",
    },
    {
      kode_pesanan: "B",
      kode_invoice_channel: "INV2",
      nama_toko: "X",
      status_pesanan_terakhir: "diproses",
      no_pesanan: "INV2",
      shopee_order_status: "PROCESSED",
    },
  ];
  (api.listCandidates as jest.Mock)
    .mockResolvedValueOnce({ data: { data: rows, total: 2 } })
    .mockResolvedValueOnce({ data: { data: rows, total: 2 } });
  render(
    <MemoryRouter>
      <ReconcileDashboard />
    </MemoryRouter>,
  );
  await screen.findByRole("button", { name: /Reconcile All/i });
  fireEvent.click(screen.getByRole("button", { name: /Reconcile All/i }));
  await screen.findByRole("dialog");
  await waitFor(() =>
    expect(api.updateShopeeStatuses).toHaveBeenCalledWith(["INV", "INV2"]),
  );
  await waitFor(() => expect(api.reconcileCheck).toHaveBeenCalledTimes(2));
});

test("check status button", async () => {
  (api.listCandidates as jest.Mock).mockResolvedValueOnce({
    data: {
      data: [
        {
          kode_pesanan: "A",
          kode_invoice_channel: "INV",
          nama_toko: "X",
          status_pesanan_terakhir: "diproses",
          no_pesanan: "INV",
          shopee_order_status: "PROCESSED",
        },
      ],
      total: 1,
    },
  });
  render(
    <MemoryRouter>
      <ReconcileDashboard />
    </MemoryRouter>,
  );
  await screen.findByRole("button", { name: /Check Status/i });
  fireEvent.click(screen.getByRole("button", { name: /Check Status/i }));
  await waitFor(() =>
    expect(api.fetchShopeeDetail).toHaveBeenCalledWith("INV"),
  );
});

test("check status completed uses escrow endpoint", async () => {
  (api.listCandidates as jest.Mock).mockResolvedValueOnce({
    data: {
      data: [
        {
          kode_pesanan: "A",
          kode_invoice_channel: "INV",
          nama_toko: "X",
          status_pesanan_terakhir: "selesai",
          no_pesanan: "INV",
          shopee_order_status: "COMPLETED",
        },
      ],
      total: 1,
    },
  });
  render(
    <MemoryRouter>
      <ReconcileDashboard />
    </MemoryRouter>,
  );
  await screen.findByRole("button", { name: /Check Status/i });
  fireEvent.click(screen.getByRole("button", { name: /Check Status/i }));
  await waitFor(() =>
    expect(api.fetchEscrowDetail).toHaveBeenCalledWith("INV"),
  );
});

test("show shopee status column", async () => {
  (api.listCandidates as jest.Mock).mockResolvedValueOnce({
    data: {
      data: [
        {
          kode_pesanan: "A",
          kode_invoice_channel: "INV",
          nama_toko: "X",
          status_pesanan_terakhir: "diproses",
          no_pesanan: "INV",
          shopee_order_status: "PROCESSED",
        },
      ],
      total: 1,
    },
  });
  render(
    <MemoryRouter>
      <ReconcileDashboard />
    </MemoryRouter>,
  );
  await screen.findByText("PROCESSED");
  expect(api.fetchShopeeDetail).not.toHaveBeenCalled();
});
