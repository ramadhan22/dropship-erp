import "@testing-library/jest-dom";
import { render } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { fireEvent, screen, waitFor } from "@testing-library/dom";
import * as api from "../api";
import StoreDetailPage from "./StoreDetailPage";

jest.mock("../api", () => ({
  getStore: jest.fn(),
  updateStore: jest.fn(),
}));

test("save store auth params", async () => {
  (api.getStore as jest.Mock).mockResolvedValue({
    data: { store_id: 1, nama_toko: "Shop", jenis_channel_id: 2 },
  });
  (api.updateStore as jest.Mock).mockResolvedValue({});
  render(
    <MemoryRouter initialEntries={["/stores/1?code=abc&shop_id=123"]}>
      <Routes>
        <Route path="/stores/:id" element={<StoreDetailPage />} />
      </Routes>
    </MemoryRouter>,
  );
  await waitFor(() => expect(api.getStore).toHaveBeenCalledWith(1));
  await screen.findByRole("button", { name: /Save/i });
  fireEvent.click(screen.getByRole("button", { name: /Save/i }));
  await waitFor(() =>
    expect(api.updateStore).toHaveBeenCalledWith(1, {
      nama_toko: "Shop",
      jenis_channel_id: 2,
      code_id: "abc",
      shop_id: "123",
    }),
  );
});
