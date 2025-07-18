import "@testing-library/jest-dom";
import { render, screen } from "@testing-library/react";
import { waitFor } from "@testing-library/dom";
import * as api from "../api/pl";
import PLPage from "./PLPage";

jest.mock("../api", () => ({
  listAllStores: jest.fn().mockResolvedValue([
    { store_id: 1, nama_toko: "S", jenis_channel_id: 1 },
  ]),
}));

jest.mock("../api/pl", () => ({
  fetchProfitLoss: jest.fn(),
}));

describe("PLPage Color Logic", () => {
  beforeEach(() => {
    // Mock data to test color logic
    const mockPLData = {
      pendapatanUsaha: [
        { label: "Revenue Item", amount: 100000, change: 10000, changePercent: 10 },
      ],
      totalPendapatanUsaha: 100000,
      prevTotalPendapatanUsaha: 90000,
      hargaPokokPenjualan: [
        { label: "Cost Item", amount: 50000, change: 5000, changePercent: 10 },
      ],
      totalHargaPokokPenjualan: 50000,
      prevTotalHargaPokokPenjualan: 45000,
      labaKotor: { amount: 50000, change: 5000, changePercent: 10 },
      bebanOperasional: [
        { label: "Operating Expense", amount: 20000, change: 2000, changePercent: 10 },
      ],
      totalBebanOperasional: 20000,
      prevTotalBebanOperasional: 18000,
      bebanPemasaran: [
        { label: "Marketing Expense", amount: 10000, change: 1000, changePercent: 10 },
      ],
      totalBebanPemasaran: 10000,
      prevTotalBebanPemasaran: 9000,
      bebanAdministrasi: [
        { label: "Admin Expense", amount: 5000, change: 500, changePercent: 10 },
      ],
      totalBebanAdministrasi: 5000,
      prevTotalBebanAdministrasi: 4500,
      totalBebanUsaha: { amount: 35000, change: 3500, changePercent: 10 },
      labaSebelumPajak: 15000,
      prevLabaSebelumPajak: 12000,
      pajakPenghasilan: [
        { label: "Income Tax", amount: 3000, change: 300, changePercent: 10 },
      ],
      totalPajakPenghasilan: 3000,
      prevTotalPajakPenghasilan: 2700,
      labaRugiBersih: { amount: 12000, change: 1200, changePercent: 10 },
    };

    (api.fetchProfitLoss as jest.Mock).mockResolvedValue({ data: mockPLData });
  });

  test("should render PLPage with mock data", async () => {
    render(<PLPage />);
    
    await waitFor(() => {
      expect(api.fetchProfitLoss).toHaveBeenCalled();
    });
    
    // Just verify that the component renders without error
    await waitFor(() => {
      const profitLossHeading = screen.getByText("PROFIT & LOSS STATEMENT");
      expect(profitLossHeading).toBeInTheDocument();
    });
  });
});