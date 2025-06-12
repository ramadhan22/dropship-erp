import "@testing-library/jest-dom";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import * as api from "../api";
import AccountPage from "./AccountPage";

jest.mock("../api", () => ({
  listAccounts: jest.fn(),
  createAccount: jest.fn(),
}));

test("creates account", async () => {
  (api.listAccounts as jest.Mock).mockResolvedValue({ data: [] });

  render(<AccountPage />);

  await waitFor(() => expect(api.listAccounts).toHaveBeenCalled());

  fireEvent.change(screen.getByLabelText(/Code/i), {
    target: { value: "100" },
  });
  fireEvent.change(screen.getByLabelText(/Name/i), {
    target: { value: "Cash" },
  });
  fireEvent.change(screen.getByLabelText(/Type/i), {
    target: { value: "Asset" },
  });

  fireEvent.click(screen.getByRole("button", { name: /Add Account/i }));

  await waitFor(() =>
    expect(api.createAccount).toHaveBeenCalledWith({
      account_code: "100",
      account_name: "Cash",
      account_type: "Asset",
      parent_id: null,
    }),
  );
});
