import "@testing-library/jest-dom";
import { render } from "@testing-library/react";
import { fireEvent, screen, waitFor } from "@testing-library/dom";
import * as api from "../api/pl";
import PLPage from "./PLPage";

jest.mock("../api", () => ({
  listAllStores: jest.fn().mockResolvedValue([
    { store_id: 1, nama_toko: "S", jenis_channel_id: 1 },
  ]),
}));

jest.mock("../api/pl", () => ({
  fetchPL: jest.fn().mockResolvedValue({ data: { net_profit: 1 } }),
}));

test("fetch pl", async () => {
  render(<PLPage />);
  await waitFor(() =>
    expect(api.fetchPL).toHaveBeenCalledWith(
      "",
      new Date().toISOString().slice(0, 7),
    ),
  );
  fireEvent.change(screen.getByLabelText(/Period/i, { selector: "input" }), {
    target: { value: "2025-05" },
  });
  fireEvent.click(screen.getByRole("button", { name: /Fetch/i }));
  await waitFor(() =>
    expect(api.fetchPL).toHaveBeenLastCalledWith("", "2025-05"),
  );
});
