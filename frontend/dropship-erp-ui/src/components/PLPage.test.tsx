import "@testing-library/jest-dom";
import { render } from "@testing-library/react";
import { waitFor } from "@testing-library/dom";
import * as api from "../api/pl";
import PLPage from "./PLPage";

jest.mock("../api", () => ({
  listAllStores: jest.fn().mockResolvedValue([
    { store_id: 1, nama_toko: "S", jenis_channel_id: 1 },
  ]),
}));

jest.mock("../api/pl", () => ({
  fetchProfitLoss: jest.fn().mockResolvedValue({ data: {} }),
}));

test("fetch profit loss", async () => {
  render(<PLPage />);
  await waitFor(() =>
    expect(api.fetchProfitLoss).toHaveBeenCalledWith({
      type: "Monthly",
      month: new Date().getMonth() + 1,
      year: new Date().getFullYear(),
      store: "",
      comparison: true,
    }),
  );
});
