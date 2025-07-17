import {
  Alert,
  Box,
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  FormControl,
  InputLabel,
  MenuItem,
  Select,
} from "@mui/material";
import { LocalizationProvider } from "@mui/x-date-pickers";
import { DatePicker } from "@mui/x-date-pickers/DatePicker";
import { AdapterDateFns } from "@mui/x-date-pickers/AdapterDateFns";
import { useEffect, useState } from "react";
import { listShopeeReturns } from "../api/shopeeReturns";
import { listAllStores } from "../api";
import type { ShopeeOrderReturn, ShopeeReturnUser, Store } from "../types";
import SortableTable from "./SortableTable";
import type { Column } from "./SortableTable";
import { getCurrentMonthDateRange } from "../utils/date";
import usePagination from "../usePagination";
import JsonTabs from "./JsonTabs";

export default function ShopeeOrderReturnPage() {
  const [list, setList] = useState<ShopeeOrderReturn[]>([]);
  const [stores, setStores] = useState<Store[]>([]);
  const [msg, setMsg] = useState<{ type: "success" | "error"; text: string } | null>(null);
  const [loading, setLoading] = useState(false);
  
  // Filter states
  const [selectedStore, setSelectedStore] = useState("all");
  const [firstOfMonth, lastOfMonth] = getCurrentMonthDateRange();
  const [createTimeFrom, setCreateTimeFrom] = useState<Date | null>(firstOfMonth);
  const [createTimeTo, setCreateTimeTo] = useState<Date | null>(lastOfMonth);
  const [updateTimeFrom, setUpdateTimeFrom] = useState<Date | null>(null);
  const [updateTimeTo, setUpdateTimeTo] = useState<Date | null>(null);
  const [status, setStatus] = useState("");
  const [negotiationStatus, setNegotiationStatus] = useState("");
  const [sellerProofStatus, setSellerProofStatus] = useState("");
  const [sellerCompensationStatus, setSellerCompensationStatus] = useState("");
  const [pageSize, setPageSize] = useState("20");
  
  // Detail modal
  const [detailOpen, setDetailOpen] = useState(false);
  const [selectedReturn, setSelectedReturn] = useState<ShopeeOrderReturn | null>(null);
  
  const { paginated, controls, setPage } = usePagination(list);

  const fetchStores = async () => {
    try {
      const storeList = await listAllStores();
      setStores(storeList);
    } catch (e: unknown) {
      const error = e as { response?: { data?: { error?: string } }; message?: string };
      setMsg({ type: "error", text: "Failed to load stores: " + (error.response?.data?.error || error.message || "Unknown error") });
    }
  };

  const fetchData = async () => {
    setLoading(true);
    try {
      const params = {
        store: selectedStore === "all" ? "" : selectedStore,
        page_size: pageSize,
        ...(createTimeFrom && { create_time_from: Math.floor(createTimeFrom.getTime() / 1000).toString() }),
        ...(createTimeTo && { create_time_to: Math.floor(createTimeTo.getTime() / 1000).toString() }),
        ...(updateTimeFrom && { update_time_from: Math.floor(updateTimeFrom.getTime() / 1000).toString() }),
        ...(updateTimeTo && { update_time_to: Math.floor(updateTimeTo.getTime() / 1000).toString() }),
        ...(status && { status }),
        ...(negotiationStatus && { negotiation_status: negotiationStatus }),
        ...(sellerProofStatus && { seller_proof_status: sellerProofStatus }),
        ...(sellerCompensationStatus && { seller_compensation_status: sellerCompensationStatus }),
      };

      const res = await listShopeeReturns(params);
      setList(res.data.data || []);
      setMsg(null);
      setPage(1);
    } catch (e: unknown) {
      const error = e as { response?: { data?: { error?: string } }; message?: string };
      setMsg({ type: "error", text: error.response?.data?.error || error.message || "Unknown error" });
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchStores();
  }, []);

  useEffect(() => {
    fetchData();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [selectedStore, createTimeFrom, createTimeTo, updateTimeFrom, updateTimeTo, status, negotiationStatus, sellerProofStatus, sellerCompensationStatus, pageSize]);

  const resetFilters = () => {
    setSelectedStore("all");
    setCreateTimeFrom(firstOfMonth);
    setCreateTimeTo(lastOfMonth);
    setUpdateTimeFrom(null);
    setUpdateTimeTo(null);
    setStatus("");
    setNegotiationStatus("");
    setSellerProofStatus("");
    setSellerCompensationStatus("");
    setPageSize("20");
  };

  const formatTimestamp = (timestamp: number) => {
    return new Date(timestamp * 1000).toLocaleString();
  };

  const getStatusColor = (status: string) => {
    switch (status.toLowerCase()) {
      case "pending":
        return "orange";
      case "completed":
        return "green";
      case "cancelled":
        return "red";
      default:
        return "gray";
    }
  };

  const columns: Column<ShopeeOrderReturn>[] = [
    { label: "Return SN", key: "return_sn" },
    { label: "Order SN", key: "order_sn" },
    {
      label: "Status",
      key: "status",
      render: (value) => (
        <span style={{ color: getStatusColor(String(value)), fontWeight: "bold" }}>
          {String(value)}
        </span>
      ),
    },
    { label: "Negotiation Status", key: "negotiation_status" },
    {
      label: "Refund Amount",
      key: "refund_amount",
      align: "right",
      render: (value) =>
        Number(value).toLocaleString("id-ID", {
          style: "currency",
          currency: "IDR",
        }),
    },
    { label: "Currency", key: "currency" },
    {
      label: "Created",
      key: "create_time",
      render: (value) => formatTimestamp(Number(value)),
    },
    {
      label: "Updated",
      key: "update_time",
      render: (value) => formatTimestamp(Number(value)),
    },
    { label: "User", key: "user", render: (value: ShopeeReturnUser) => value?.username || "" },
    {
      label: "",
      render: (_, row) => (
        <Button
          size="small"
          onClick={() => {
            setSelectedReturn(row);
            setDetailOpen(true);
          }}
        >
          Details
        </Button>
      ),
    },
  ];

  return (
    <LocalizationProvider dateAdapter={AdapterDateFns}>
      <Box sx={{ p: 2 }}>
        <h1>Shopee Order Returns</h1>
        
        {msg && (
          <Alert severity={msg.type} sx={{ mb: 2 }} onClose={() => setMsg(null)}>
            {msg.text}
          </Alert>
        )}

        {/* Filter Section */}
        <Box sx={{ mb: 3, p: 2, border: "1px solid #ddd", borderRadius: 1 }}>
          <h3>Filters</h3>
          <Box sx={{ display: "flex", gap: 2, flexWrap: "wrap", alignItems: "center" }}>
            {/* Store Filter */}
            <FormControl sx={{ minWidth: 200 }}>
              <InputLabel>Store</InputLabel>
              <Select
                value={selectedStore}
                label="Store"
                onChange={(e) => setSelectedStore(e.target.value)}
              >
                <MenuItem value="all">All Stores</MenuItem>
                {stores.map((store) => (
                  <MenuItem key={store.store_id} value={store.nama_toko}>
                    {store.nama_toko}
                  </MenuItem>
                ))}
              </Select>
            </FormControl>

            {/* Create Time Range */}
            <DatePicker
              label="Create From"
              value={createTimeFrom}
              onChange={setCreateTimeFrom}
              slotProps={{ textField: { size: "small" } }}
            />
            <DatePicker
              label="Create To"
              value={createTimeTo}
              onChange={setCreateTimeTo}
              slotProps={{ textField: { size: "small" } }}
            />

            {/* Update Time Range */}
            <DatePicker
              label="Update From"
              value={updateTimeFrom}
              onChange={setUpdateTimeFrom}
              slotProps={{ textField: { size: "small" } }}
            />
            <DatePicker
              label="Update To"
              value={updateTimeTo}
              onChange={setUpdateTimeTo}
              slotProps={{ textField: { size: "small" } }}
            />

            {/* Status Filters */}
            <FormControl sx={{ minWidth: 150 }}>
              <InputLabel>Status</InputLabel>
              <Select
                value={status}
                label="Status"
                onChange={(e) => setStatus(e.target.value)}
              >
                <MenuItem value="">All</MenuItem>
                <MenuItem value="pending">Pending</MenuItem>
                <MenuItem value="completed">Completed</MenuItem>
                <MenuItem value="cancelled">Cancelled</MenuItem>
              </Select>
            </FormControl>

            <FormControl sx={{ minWidth: 150 }}>
              <InputLabel>Negotiation Status</InputLabel>
              <Select
                value={negotiationStatus}
                label="Negotiation Status"
                onChange={(e) => setNegotiationStatus(e.target.value)}
              >
                <MenuItem value="">All</MenuItem>
                <MenuItem value="pending">Pending</MenuItem>
                <MenuItem value="accepted">Accepted</MenuItem>
                <MenuItem value="rejected">Rejected</MenuItem>
              </Select>
            </FormControl>

            {/* Page Size */}
            <FormControl sx={{ minWidth: 120 }}>
              <InputLabel>Page Size</InputLabel>
              <Select
                value={pageSize}
                label="Page Size"
                onChange={(e) => setPageSize(e.target.value)}
              >
                <MenuItem value="10">10</MenuItem>
                <MenuItem value="20">20</MenuItem>
                <MenuItem value="50">50</MenuItem>
                <MenuItem value="100">100</MenuItem>
              </Select>
            </FormControl>

            <Button variant="contained" onClick={fetchData} disabled={loading}>
              {loading ? "Loading..." : "Apply Filters"}
            </Button>
            <Button variant="outlined" onClick={resetFilters}>
              Reset
            </Button>
          </Box>
        </Box>

        {/* Table */}
        <SortableTable
          data={paginated}
          columns={columns}
        />
        {controls}

        {/* Detail Modal */}
        <Dialog
          open={detailOpen}
          onClose={() => setDetailOpen(false)}
          maxWidth="md"
          fullWidth
        >
          <DialogTitle>Return Details</DialogTitle>
          <DialogContent>
            {selectedReturn && (
              <Box sx={{ mt: 1 }}>
                <h4>Basic Information</h4>
                <p><strong>Return SN:</strong> {selectedReturn.return_sn}</p>
                <p><strong>Order SN:</strong> {selectedReturn.order_sn}</p>
                <p><strong>Status:</strong> {selectedReturn.status}</p>
                <p><strong>Reason:</strong> {selectedReturn.reason}</p>
                <p><strong>Refund Amount:</strong> {Number(selectedReturn.refund_amount).toLocaleString("id-ID", { style: "currency", currency: "IDR" })}</p>
                <p><strong>Created:</strong> {formatTimestamp(selectedReturn.create_time)}</p>
                <p><strong>Updated:</strong> {formatTimestamp(selectedReturn.update_time)}</p>
                
                <h4>Items</h4>
                {selectedReturn.item && selectedReturn.item.length > 0 ? (
                  <JsonTabs items={selectedReturn.item} />
                ) : (
                  <p>No items</p>
                )}

                <h4>Complete Data</h4>
                <JsonTabs items={[selectedReturn]} />
              </Box>
            )}
          </DialogContent>
          <DialogActions>
            <Button onClick={() => setDetailOpen(false)}>Close</Button>
          </DialogActions>
        </Dialog>
      </Box>
    </LocalizationProvider>
  );
}