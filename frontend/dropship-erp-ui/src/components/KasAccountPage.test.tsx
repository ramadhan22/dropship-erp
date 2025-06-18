import "@testing-library/jest-dom";
import { render } from "@testing-library/react";
import { fireEvent, screen, waitFor } from "@testing-library/dom";
import KasAccountPage from "./KasAccountPage";
import * as kasApi from "../api/kasAccounts";

jest.mock("../api/kasAccounts", () => ({
  listKasAccounts: jest.fn().mockResolvedValue({ data: [] }),
  adjustKasBalance: jest.fn(),
}));

jest.mock("../api", () => ({
  listAccounts: jest.fn().mockResolvedValue({ data: [] }),
}));

test("fetch asset accounts", async () => {
  render(<KasAccountPage />);
  await waitFor(() => expect(kasApi.listKasAccounts).toHaveBeenCalled());
});

test("adjust balance", async () => {
  (kasApi.listKasAccounts as jest.Mock).mockResolvedValueOnce({
    data: [{ asset_id: 1, account_id: 10, balance: 100 }],
  });
  render(<KasAccountPage />);
  await waitFor(() => screen.getByText(/1/));
  fireEvent.click(screen.getByRole("button", { name: /Adjust/i }));
  await waitFor(() => screen.getByRole("button", { name: /^Save$/i }));
  fireEvent.change(screen.getByRole("spinbutton", { name: /Balance/i }), {
    target: { value: "150" },
  });
  fireEvent.click(screen.getByRole("button", { name: /^Save$/i }));
  await waitFor(() =>
    expect(kasApi.adjustKasBalance).toHaveBeenCalledWith(1, 150),
  );
});
