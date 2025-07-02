import { api } from "./index";
import type { TaxPayment } from "../types";

export const fetchTaxPayment = (store: string, type: string, period: string) =>
  api.get<TaxPayment>(`/tax-payment?store=${store}&type=${type}&period=${period}`);

export const payTax = (tp: TaxPayment) =>
  api.post<{ success: boolean }>("/tax-payment/pay", tp);
