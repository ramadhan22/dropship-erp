import "@testing-library/jest-dom";
import { render } from "@testing-library/react";
import { fireEvent, screen, waitFor } from "@testing-library/dom";
import * as expApi from "../api/expenses";
import * as baseApi from "../api";
import ExpensePage from "./ExpensePage";

jest.mock("../api/expenses", () => ({
  listExpenses: jest.fn().mockResolvedValue({ data: { data: [] } }),
  createExpense: jest.fn(),
  deleteExpense: jest.fn(),
}));
jest.mock("../api", () => ({
  listAccounts: jest.fn().mockResolvedValue({
    data: [
      {
        account_id: 10,
        account_code: "A10",
        account_name: "Asset 10",
        account_type: "asset",
        parent_id: null,
        balance: 0,
      },
      {
        account_id: 5,
        account_code: "E5",
        account_name: "Expense 5",
        account_type: "expense",
        parent_id: null,
        balance: 0,
      },
    ],
  }),
}));

test("renders and creates expense", async () => {
  render(<ExpensePage />);
  await waitFor(() => expect(expApi.listExpenses).toHaveBeenCalled());
  fireEvent.click(screen.getByRole("button", { name: /Add Expense/i }));
  fireEvent.change(screen.getByLabelText(/Description/i), {
    target: { value: "x" },
  });
  const assetInput = screen.getByRole("combobox", { name: /Asset Account/i });
  fireEvent.change(assetInput, { target: { value: "A10" } });
  fireEvent.keyDown(assetInput, { key: "ArrowDown" });
  fireEvent.keyDown(assetInput, { key: "Enter" });
  const expInput = screen.getAllByRole("combobox", { name: /Expense Account/i })[0];
  fireEvent.change(expInput, { target: { value: "E5" } });
  fireEvent.keyDown(expInput, { key: "ArrowDown" });
  fireEvent.keyDown(expInput, { key: "Enter" });
  fireEvent.change(screen.getAllByLabelText(/Amount/i)[0], {
    target: { value: "5" },
  });
  fireEvent.click(screen.getByRole("button", { name: /Save/i }));
  await waitFor(() =>
    expect(expApi.createExpense).toHaveBeenCalledWith({
      description: "x",
      asset_account_id: 10,
      lines: [{ account_id: 5, amount: 5 }],
      date: expect.any(String),
    }),
  );
});
