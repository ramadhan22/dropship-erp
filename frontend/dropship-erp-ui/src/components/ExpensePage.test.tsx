import "@testing-library/jest-dom";
import { render } from "@testing-library/react";
import { fireEvent, screen, waitFor } from "@testing-library/dom";
import * as expApi from "../api/expenses";
import ExpensePage from "./ExpensePage";

jest.mock("../api/expenses", () => ({
  listExpenses: jest.fn().mockResolvedValue({ data: [] }),
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
  fireEvent.change(screen.getByLabelText(/Amount/i), {
    target: { value: "5" },
  });
  fireEvent.change(screen.getByLabelText(/Account/i), {
    target: { value: "1" },
  });
  fireEvent.click(screen.getByRole("button", { name: /Save/i }));
  await waitFor(() =>
    expect(expApi.createExpense).toHaveBeenCalledWith({
      description: "x",
      amount: 5,
      account_id: 1,
      date: expect.any(String),
    }),
  );
});
