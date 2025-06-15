import {
  Table,
  TableHead,
  TableRow,
  TableCell,
  TableBody,
  TableSortLabel,
} from "@mui/material";
import { useState } from "react";

export interface Column<T> {
  label: string;
  key?: keyof T;
  render?: (value: any, row: T) => React.ReactNode;
  align?: "left" | "right" | "center";
}

export default function SortableTable<T extends Record<string, any>>({
  columns,
  data,
  defaultSort,
  onSortChange,
}: {
  columns: Column<T>[];
  data: T[];
  defaultSort?: { key: keyof T; direction?: "asc" | "desc" };
  onSortChange?: (key: keyof T, direction: "asc" | "desc") => void;
}) {
  const [sortKey, setSortKey] = useState<keyof T | null>(
    defaultSort?.key ?? null,
  );
  const [direction, setDirection] = useState<"asc" | "desc">(
    defaultSort?.direction ?? "asc",
  );

  const sorted = onSortChange
    ? data
    : (() => {
        const s = [...data];
        if (sortKey) {
          s.sort((a, b) => {
            const aVal = a[sortKey];
            const bVal = b[sortKey];
            if (aVal === bVal) return 0;
            if (aVal == null) return -1;
            if (bVal == null) return 1;
            return (aVal > bVal ? 1 : -1) * (direction === "asc" ? 1 : -1);
          });
        }
        return s;
      })();

  const handleSort = (key: keyof T) => {
    let dir: "asc" | "desc" = "asc";
    if (sortKey === key) {
      dir = direction === "asc" ? "desc" : "asc";
      setDirection(dir);
    } else {
      setSortKey(key);
      dir = "asc";
      setDirection(dir);
    }
    if (onSortChange) {
      onSortChange(key, dir);
    }
  };

  return (
    <Table size="small">
      <TableHead>
        <TableRow>
          {columns.map((col) => (
            <TableCell key={String(col.label)} align={col.align}>
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
              <TableCell key={String(col.label)} align={col.align}>
                {col.render
                  ? col.render(col.key ? (row as any)[col.key] : undefined, row)
                  : col.key
                    ? String((row as any)[col.key])
                    : null}
              </TableCell>
            ))}
          </TableRow>
        ))}
      </TableBody>
    </Table>
  );
}
