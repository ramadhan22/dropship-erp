import { useEffect, useState } from "react";
import { Pagination } from "@mui/material";

export interface FetchParams {
  page: number;
  pageSize: number;
}

export type Fetcher<T> = (
  params: FetchParams,
) => Promise<{ data: T[]; total: number }>;

export default function useServerPagination<T>(
  fetcher: Fetcher<T>,
  defaultSize = 20,
) {
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(defaultSize);
  const [data, setData] = useState<T[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const load = async () => {
    setLoading(true);
    try {
      const res = await fetcher({ page, pageSize });
      setData(res.data);
      setTotal(res.total);
      const pages = Math.max(1, Math.ceil(res.total / pageSize));
      if (page > pages) {
        setPage(pages);
      }
      setError(null);
    } catch (e: any) {
      setError(e.message);
    }
    setLoading(false);
  };

  useEffect(() => {
    load();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [page, pageSize]);

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

  return { data, loading, error, controls, page, setPage, pageSize, setPageSize, reload: load };
}
