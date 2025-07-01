import "@testing-library/jest-dom";
import { render } from "@testing-library/react";
import { waitFor } from "@testing-library/dom";
import WithdrawalPage from "./WithdrawalPage";
import * as api from "../api/withdrawals";
import * as common from "../api";

jest.mock("../api/withdrawals", () => ({
  listWithdrawals: jest.fn().mockResolvedValue({ data: [] }),
  createWithdrawal: jest.fn(),
  importWithdrawals: jest.fn(),
}));

jest.mock("../api", () => ({
  listAllStores: jest.fn().mockResolvedValue([]),
}));

test("fetch withdrawals", async () => {
  render(<WithdrawalPage />);
  await waitFor(() => expect(api.listWithdrawals).toHaveBeenCalled());
  await waitFor(() => expect(common.listAllStores).toHaveBeenCalled());
});
