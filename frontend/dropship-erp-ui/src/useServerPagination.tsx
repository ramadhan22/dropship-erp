import { useEffect, useState, useCallback } from "react";
import { Pagination } from "@mui/material";

export interface FilterCondition {
  field: string;
  operator: string;
  value: any;
  values?: any[];
}

export interface FilterGroup {
  logic: string; // "AND" or "OR"
  conditions: FilterCondition[];
  groups?: FilterGroup[];
}

export interface SortCondition {
  field: string;
  direction: "asc" | "desc";
}

export interface FilterParams {
  filters?: FilterGroup;
  sort?: SortCondition[];
  pagination?: {
    page: number;
    page_size: number;
  };
}

export interface FetchParams {
  page: number;
  pageSize: number;
  filters?: FilterParams;
}

export interface QueryResult<T> {
  data: T[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

export type Fetcher<T> = (
  params: FetchParams,
) => Promise<QueryResult<T> | { data: T[]; total: number }>;

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
  const [filters, setFilters] = useState<FilterParams | undefined>();
  const [sortConditions, setSortConditions] = useState<SortCondition[]>([]);

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const params: FetchParams = {
        page,
        pageSize,
        filters: filters ? {
          ...filters,
          sort: sortConditions.length > 0 ? sortConditions : filters.sort,
          pagination: { page, page_size: pageSize }
        } : {
          sort: sortConditions.length > 0 ? sortConditions : undefined,
          pagination: { page, page_size: pageSize }
        }
      };
      
      const res = await fetcher(params);
      
      // Handle both old and new response formats
      if ('page' in res && 'page_size' in res) {
        // New format with QueryResult
        setData(res.data);
        setTotal(res.total);
        const pages = Math.max(1, Math.ceil(res.total / pageSize));
        if (page > pages) {
          setPage(pages);
        }
      } else {
        // Legacy format
        setData(res.data);
        setTotal(res.total);
        const pages = Math.max(1, Math.ceil(res.total / pageSize));
        if (page > pages) {
          setPage(pages);
        }
      }
      setError(null);
    } catch (e: any) {
      setError(e.message);
    }
    setLoading(false);
  }, [page, pageSize, filters, sortConditions, fetcher]);

  useEffect(() => {
    load();
  }, [load]);

  const applyFilters = useCallback((newFilters: FilterParams | undefined) => {
    setFilters(newFilters);
    setPage(1); // Reset to first page when filters change
  }, []);

  const applySort = useCallback((sort: SortCondition[]) => {
    setSortConditions(sort);
    setPage(1); // Reset to first page when sort changes
  }, []);

  const clearFilters = useCallback(() => {
    setFilters(undefined);
    setSortConditions([]);
    setPage(1);
  }, []);

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

  return { 
    data, 
    loading, 
    error, 
    controls, 
    page, 
    setPage, 
    pageSize, 
    setPageSize, 
    reload: load,
    filters,
    applyFilters,
    applySort,
    clearFilters,
    sortConditions
  };
}
