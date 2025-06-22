import "@testing-library/jest-dom";
import { render } from "@testing-library/react";
import { fireEvent, screen, waitFor } from "@testing-library/dom";
import * as api from "../api";
import WithdrawPage from "./WithdrawPage";

jest.mock("../api", () => ({
  listAllStores: jest.fn().mockResolvedValue([]),
  withdrawShopeeBalance: jest.fn(),
}));

test("submits withdraw", async () => {
  (api.listAllStores as jest.Mock).mockResolvedValue([{ store_id: 1, nama_toko: "Shop" }]);
  (api.withdrawShopeeBalance as jest.Mock).mockResolvedValue({});
  render(<WithdrawPage />);
  await waitFor(() => expect(api.listAllStores).toHaveBeenCalled());
  fireEvent.mouseDown(screen.getByText(/Select Store/i));
  fireEvent.click(await screen.findByText("Shop"));
  fireEvent.change(screen.getByLabelText(/Amount/i), { target: { value: "100" } });
  fireEvent.click(screen.getByRole("button", { name: /Withdraw/i }));
  await waitFor(() => expect(api.withdrawShopeeBalance).toHaveBeenCalled());
});
