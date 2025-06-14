import "@testing-library/jest-dom";
import { render } from "@testing-library/react";
import { fireEvent, screen, waitFor } from "@testing-library/dom";
import * as api from "../api/gl";
import GLPage from "./GLPage";

jest.mock("../api", () => ({
  listAllStores: jest.fn().mockResolvedValue([]),
}));

jest.mock("../api/gl", () => ({
  fetchGeneralLedger: jest.fn().mockResolvedValue({
    data: [
      {
        account_id: 1,
        account_code: "1001",
        account_name: "Cash",
        account_type: "Asset",
        parent_id: null,
        balance: 100,
      },
    ],
  }),
}));

test("fetch gl grouped", async () => {
  render(<GLPage />);
  fireEvent.change(screen.getByLabelText(/Shop/i), { target: { value: "S" } });
  fireEvent.change(screen.getByLabelText(/^From$/i), {
    target: { value: "2025-01-01" },
  });
  fireEvent.change(screen.getByLabelText(/^To$/i), {
    target: { value: "2025-01-31" },
  });
  fireEvent.click(screen.getByRole("button", { name: /Fetch/i }));
  await waitFor(() => screen.getByText(/Asset/i));
  expect(api.fetchGeneralLedger).toHaveBeenCalled();
});
