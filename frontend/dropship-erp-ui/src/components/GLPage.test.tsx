import "@testing-library/jest-dom";
import { render } from "@testing-library/react";
import { fireEvent, screen, waitFor } from "@testing-library/dom";
import * as api from "../api/gl";
import GLPage from "./GLPage";

jest.mock("../api", () => ({
  listAllStores: jest.fn().mockResolvedValue([]),
}));

jest.mock("../api/gl", () => ({
  fetchGeneralLedger: jest.fn().mockResolvedValue({ data: [] }),
}));

test("fetch gl", async () => {
  render(<GLPage />);
  fireEvent.change(screen.getByLabelText(/Shop/i), { target: { value: "S" } });
  fireEvent.change(screen.getByLabelText(/^From$/i), {
    target: { value: "2025-01-01" },
  });
  fireEvent.change(screen.getByLabelText(/^To$/i), {
    target: { value: "2025-01-31" },
  });
  fireEvent.click(screen.getByRole("button", { name: /Fetch/i }));
  await waitFor(() => expect(api.fetchGeneralLedger).toHaveBeenCalled());
});
