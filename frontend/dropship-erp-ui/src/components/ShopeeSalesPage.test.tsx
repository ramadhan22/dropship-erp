import "@testing-library/jest-dom";
import { render } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { fireEvent, screen, waitFor } from "@testing-library/dom";
import * as api from "../api";
import ShopeeSalesPage from "./ShopeeSalesPage";

jest.mock("../api", () => ({
  listJenisChannels: jest.fn().mockResolvedValue({ data: [] }),
  listStoresByChannelName: jest.fn().mockResolvedValue({ data: [] }),
  listShopeeSettled: jest.fn().mockResolvedValue({
    data: {
      data: [
        {
          nama_toko: "TOKO",
          no_pesanan: "SN1",
          is_data_mismatch: false,
          is_settled_confirmed: false,
        },
      ],
      total: 1,
    },
  }),
  sumShopeeSettled: jest.fn().mockResolvedValue({
    data: {
      harga_asli_produk: 0,
      total_diskon_produk: 0,
      gmv: 0,
      diskon_voucher_ditanggung_penjual: 0,
      biaya_administrasi: 0,
      biaya_layanan_termasuk_ppn_11: 0,
      total_penghasilan: 0,
    },
  }),
  importShopee: jest.fn(),
  confirmShopeeSettle: jest.fn().mockResolvedValue({ data: { success: true } }),
}));

test("confirm settle button calls API", async () => {
  render(
    <MemoryRouter>
      <ShopeeSalesPage />
    </MemoryRouter>,
  );

  await waitFor(() => expect(api.listShopeeSettled).toHaveBeenCalled());

  const btn = await screen.findByRole("button", { name: /Confirm Settle/i });
  fireEvent.click(btn);

  await waitFor(() =>
    expect(api.confirmShopeeSettle).toHaveBeenCalledWith("SN1"),
  );
});
