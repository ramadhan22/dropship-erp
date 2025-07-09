export function formatCurrency(value: number | null | undefined): string {
  if (value == null) return "";
  return new Intl.NumberFormat("id-ID", {
    style: "currency",
    currency: "IDR",
  }).format(value);
}

export function formatDate(value: string | number | Date): string {
  return new Date(value).toLocaleDateString("id-ID");
}

export function formatDateTime(value: string | number | Date): string {
  return new Date(value).toLocaleString("id-ID");
}
