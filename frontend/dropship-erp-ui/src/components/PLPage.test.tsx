import "@testing-library/jest-dom";
import { render } from "@testing-library/react";
import { fireEvent, screen, waitFor } from "@testing-library/dom";
import * as api from "../api/pl";
import PLPage from "./PLPage";

jest.mock("../api", () => ({
  listAllStores: jest.fn().mockResolvedValue([]),
}));

jest.mock("../api/pl", () => ({
  fetchPL: jest.fn().mockResolvedValue({ data: { net_profit: 1 } }),
}));

test("fetch pl", async () => {
  render(<PLPage />);
  fireEvent.change(screen.getByLabelText(/Shop/i), { target: { value: "S" } });
  fireEvent.change(screen.getByLabelText(/Period/i, { selector: "input" }), {
    target: { value: "2025-05" },
  });
  fireEvent.click(screen.getByRole("button", { name: /Fetch/i }));
  await waitFor(() => expect(api.fetchPL).toHaveBeenCalled());
});
