import { useState } from "react";
import { Pagination } from "@mui/material";

export default function usePagination<T>(data: T[], pageSize: number = 10) {
  const [page, setPage] = useState(1);

  const paginated = data.slice((page - 1) * pageSize, page * pageSize);
  const controls = (
    <Pagination
      sx={{ mt: 2 }}
      page={page}
      count={Math.max(1, Math.ceil(data.length / pageSize))}
      onChange={(_, val) => setPage(val)}
    />
  );
  return { paginated, controls, page, setPage };
}
