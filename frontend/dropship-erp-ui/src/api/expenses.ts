import { api } from "./index";
import type { Expense } from "../types";

export function createExpense(exp: Partial<Expense>) {
  return api.post<Expense>("/expenses/", exp);
}

export function listExpenses() {
  return api.get<Expense[]>("/expenses/");
}

export function deleteExpense(id: string) {
  return api.delete(`/expenses/${id}`);
}
