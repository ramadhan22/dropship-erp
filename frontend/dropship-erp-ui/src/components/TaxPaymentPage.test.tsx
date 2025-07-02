import "@testing-library/jest-dom";
import { render } from "@testing-library/react";
import { fireEvent, screen, waitFor } from "@testing-library/dom";
import * as api from "../api/tax";
import TaxPaymentPage from "./TaxPaymentPage";

jest.mock("../api", () => ({
  listAllStores: jest.fn().mockResolvedValue([]),
}));

jest.mock("../api/tax", () => ({
  fetchTaxPayment: jest.fn().mockResolvedValue({ data: { revenue: 1000, tax_amount: 5, is_paid: false } }),
  payTax: jest.fn().mockResolvedValue({ data: { success: true } }),
}));

test("fetch and pay", async () => {
  render(<TaxPaymentPage />);
  fireEvent.change(screen.getByLabelText(/Period/i), { target: { value: "2025-06" } });
  fireEvent.click(screen.getByRole("button", { name: /Fetch/i }));
  await waitFor(() => expect(api.fetchTaxPayment).toHaveBeenCalled());
  await waitFor(() => screen.getByText(/Tax Amount/i));
  fireEvent.click(screen.getByRole("button", { name: /Bayar Pajak/i }));
  await waitFor(() => expect(api.payTax).toHaveBeenCalled());
});
