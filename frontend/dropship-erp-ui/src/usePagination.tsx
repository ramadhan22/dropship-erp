import { useState } from "react";
import { Pagination } from "@mui/material";

export default function usePagination<T>(data: T[], defaultSize: number = 20) {
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(defaultSize);

  const paginated = data.slice((page - 1) * pageSize, page * pageSize);
  const total = data.length;
  const controls = (
    <div
      style={{
        marginTop: "1rem",
        display: "flex",
        justifyContent: "space-between",
        alignItems: "center",
      }}
    >
      <div>Total: {total}</div>
      <div style={{ display: "flex", alignItems: "center", gap: "0.5rem" }}>
        <select
          value={pageSize}
          onChange={(e) => {
            setPageSize(Number(e.target.value));
            setPage(1);
          }}
        >
          {[10, 20, 50, 100, 250, 500, 1000].map((n) => (
            <option key={n} value={n}>
              {n}
            </option>
          ))}
        </select>
        <Pagination
          page={page}
          count={Math.max(1, Math.ceil(total / pageSize))}
          onChange={(_, val) => setPage(val)}
        />
      </div>
    </div>
  );
  return { paginated, controls, page, setPage, pageSize, setPageSize };
}
