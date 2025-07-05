import { api } from "./index";
import type { Expense } from "../types";

export function createExpense(exp: Partial<Expense>) {
  return api.post<Expense>("/expenses/", exp);
}

export function listExpenses(params?: { page?: number; page_size?: number }) {
  return api.get<{ data: Expense[]; total: number }>("/expenses/", { params });
}

export function deleteExpense(id: string) {
  return api.delete(`/expenses/${id}`);
}

export function getExpense(id: string) {
  return api.get<Expense>(`/expenses/${id}`);
}

export function updateExpense(id: string, exp: Partial<Expense>) {
  return api.put(`/expenses/${id}`, exp);
}
