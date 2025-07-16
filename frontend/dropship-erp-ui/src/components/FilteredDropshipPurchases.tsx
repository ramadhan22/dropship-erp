import { useState } from "react";
import { Box, Typography, Paper, Alert } from "@mui/material";
import FilterPanel from "./FilterPanel";
import SortableTable from "./SortableTable";
import useServerPagination from "../useServerPagination";
import { listDropshipPurchasesFiltered } from "../api";
import type { FilterField } from "./FilterPanel";
import type { Column } from "./SortableTable";
import type { FilterParams, SortCondition, FetchParams } from "../useServerPagination";
import type { DropshipPurchase } from "../types";

// Define the fields available for filtering
const filterFields: FilterField[] = [
  { key: "kode_pesanan", label: "Order Code", type: "text" },
  { key: "kode_transaksi", label: "Transaction Code", type: "text" },
  { key: "waktu_pesanan_terbuat", label: "Order Date", type: "date" },
  { key: "status_pesanan_terakhir", label: "Order Status", type: "select", options: [
    { value: "Selesai", label: "Completed" },
    { value: "Dibatalkan", label: "Cancelled" },
    { value: "Dikemas", label: "Packed" },
    { value: "Dikirim", label: "Shipped" },
  ]},
  { key: "jenis_channel", label: "Channel", type: "select", options: [
    { value: "Shopee", label: "Shopee" },
    { value: "Tokopedia", label: "Tokopedia" },
    { value: "Lazada", label: "Lazada" },
  ]},
  { key: "nama_toko", label: "Store Name", type: "text" },
  { key: "total_transaksi", label: "Total Transaction", type: "number" },
  { key: "biaya_mitra_jakmall", label: "Partner Fee", type: "number" },
  { key: "provinsi", label: "Province", type: "text" },
  { key: "kota", label: "City", type: "text" },
];

// Define table columns
const columns: Column<DropshipPurchase>[] = [
  { label: "Order Code", key: "kode_pesanan" },
  { label: "Transaction Code", key: "kode_transaksi" },
  {
    label: "Order Date",
    key: "waktu_pesanan_terbuat",
    render: (value) => value ? new Date(value).toLocaleDateString("id-ID") : "",
  },
  { label: "Status", key: "status_pesanan_terakhir" },
  { label: "Channel", key: "jenis_channel" },
  { label: "Store", key: "nama_toko" },
  {
    label: "Total Transaction",
    key: "total_transaksi",
    align: "right",
    render: (value) =>
      Number(value).toLocaleString("id-ID", {
        style: "currency",
        currency: "IDR",
      }),
  },
  {
    label: "Partner Fee",
    key: "biaya_mitra_jakmall",
    align: "right",
    render: (value) =>
      Number(value).toLocaleString("id-ID", {
        style: "currency",
        currency: "IDR",
      }),
  },
  { label: "Province", key: "provinsi" },
  { label: "City", key: "kota" },
];

export default function FilteredDropshipPurchases() {
  const [error, setError] = useState<string | null>(null);

  const fetcher = async (params: FetchParams) => {
    try {
      setError(null);
      
      const apiParams: any = {
        page: params.page,
        page_size: params.pageSize,
      };

      // Convert filters to JSON string
      if (params.filters?.filters) {
        apiParams.filters = JSON.stringify(params.filters.filters);
      }

      // Convert sort to JSON string
      if (params.filters?.sort && params.filters.sort.length > 0) {
        apiParams.sort = JSON.stringify(params.filters.sort);
      }

      const response = await listDropshipPurchasesFiltered(apiParams);
      return response.data;
    } catch (err: any) {
      setError(err.response?.data?.error || err.message);
      // Return empty result on error
      return {
        data: [],
        total: 0,
        page: params.page,
        page_size: params.pageSize,
        total_pages: 0,
      };
    }
  };

  const {
    data,
    loading,
    controls,
    applyFilters,
    applySort,
    filters,
    sortConditions,
  } = useServerPagination<DropshipPurchase>(fetcher, 20);

  const handleFiltersChange = (newFilters: FilterParams | undefined) => {
    applyFilters(newFilters);
  };

  const handleSortChange = (sort: SortCondition[]) => {
    applySort(sort);
  };

  return (
    <Box sx={{ p: 3 }}>
      <Typography variant="h4" sx={{ mb: 3 }}>
        Advanced Filtering Demo - Dropship Purchases
      </Typography>
      
      <Typography variant="body1" sx={{ mb: 3, color: "text.secondary" }}>
        This page demonstrates the new advanced filtering and sorting capabilities. 
        You can create complex filter conditions, combine them with AND/OR logic, 
        and sort by multiple columns.
      </Typography>

      {error && (
        <Alert severity="error" sx={{ mb: 2 }}>
          {error}
        </Alert>
      )}

      <FilterPanel
        fields={filterFields}
        onFiltersChange={handleFiltersChange}
        onSortChange={handleSortChange}
        currentFilters={filters}
        currentSort={sortConditions}
        loading={loading}
      />

      <Paper sx={{ mt: 2, p: 2 }}>
        <Typography variant="h6" sx={{ mb: 2 }}>
          Results
        </Typography>
        
        {loading ? (
          <Typography>Loading...</Typography>
        ) : (
          <>
            <SortableTable 
              columns={columns} 
              data={data}
              onSortChange={(key, direction) => {
                const newSort = [{ field: key as string, direction }];
                handleSortChange(newSort);
              }}
            />
            {controls}
          </>
        )}
      </Paper>
    </Box>
  );
}