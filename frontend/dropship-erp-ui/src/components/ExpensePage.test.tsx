import "@testing-library/jest-dom";
import { render } from "@testing-library/react";
import { fireEvent, screen, waitFor } from "@testing-library/dom";
import * as expApi from "../api/expenses";
import ExpensePage from "./ExpensePage";

jest.mock("../api/expenses", () => ({
  listExpenses: jest.fn().mockResolvedValue({ data: { data: [] } }),
  createExpense: jest.fn(),
  deleteExpense: jest.fn(),
}));

test("renders and creates expense", async () => {
  render(<ExpensePage />);
  await waitFor(() => expect(expApi.listExpenses).toHaveBeenCalled());
  fireEvent.click(screen.getByRole("button", { name: /Add Expense/i }));
  fireEvent.change(screen.getByLabelText(/Description/i), {
    target: { value: "x" },
  });
  fireEvent.change(screen.getByLabelText(/Asset Account/i), {
    target: { value: "10" },
  });
  fireEvent.change(screen.getAllByLabelText(/Expense Account/i)[0], {
    target: { value: "5" },
  });
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
