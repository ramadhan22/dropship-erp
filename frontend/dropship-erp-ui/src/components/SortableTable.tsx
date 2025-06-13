import { Table, TableHead, TableRow, TableCell, TableBody, TableSortLabel } from "@mui/material";
import { useState } from "react";

export interface Column<T> {
  label: string;
  key?: keyof T;
  render?: (value: any, row: T) => React.ReactNode;
}

export default function SortableTable<T extends Record<string, any>>({ columns, data }: { columns: Column<T>[]; data: T[] }) {
  const [sortKey, setSortKey] = useState<keyof T | null>(null);
  const [direction, setDirection] = useState<"asc" | "desc">("asc");

  const sorted = [...data];
  if (sortKey) {
    sorted.sort((a, b) => {
      const aVal = a[sortKey];
      const bVal = b[sortKey];
      if (aVal === bVal) return 0;
      if (aVal == null) return -1;
      if (bVal == null) return 1;
      return (aVal > bVal ? 1 : -1) * (direction === "asc" ? 1 : -1);
    });
  }

  const handleSort = (key: keyof T) => {
    if (sortKey === key) {
      setDirection(direction === "asc" ? "desc" : "asc");
    } else {
      setSortKey(key);
      setDirection("asc");
    }
  };

  return (
    <Table size="small">
      <TableHead>
        <TableRow>
          {columns.map((col) => (
            <TableCell key={String(col.label)}>
              {col.key ? (
                <TableSortLabel
                  active={sortKey === col.key}
                  direction={sortKey === col.key ? direction : "asc"}
                  onClick={() => handleSort(col.key!)}
                >
                  {col.label}
                </TableSortLabel>
              ) : (
                col.label
              )}
            </TableCell>
          ))}
        </TableRow>
      </TableHead>
      <TableBody>
        {sorted.map((row, idx) => (
          <TableRow key={idx}>
            {columns.map((col) => (
              <TableCell key={String(col.label)}>
                {col.render ? col.render(col.key ? (row as any)[col.key] : undefined, row) : col.key ? String((row as any)[col.key]) : null}
              </TableCell>
            ))}
          </TableRow>
        ))}
      </TableBody>
    </Table>
  );
}
