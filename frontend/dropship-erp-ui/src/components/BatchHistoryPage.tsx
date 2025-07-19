import { useEffect, useState } from "react";
import {
  Button,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
} from "@mui/material";
import SortableTable from "./SortableTable";
import type { Column } from "./SortableTable";
import { listBatchHistory, listBatchDetails } from "../api";
import type { BatchHistory, BatchHistoryDetail } from "../types";

// Helper function to format PostgreSQL interval strings into readable format
const formatDuration = (intervalStr: string): string => {
  if (!intervalStr) return "-";
  
  // PostgreSQL INTERVAL format examples: "00:05:30", "01:23:45.123456", "2 days 03:45:30"
  const matches = intervalStr.match(/(?:(\d+)\s+days?\s+)?(\d{2}):(\d{2}):(\d{2})(?:\.(\d+))?/);
  if (!matches) return intervalStr; // Return original if can't parse
  
  const [, days, hours, minutes, seconds] = matches;
  const parts = [];
  
  if (days && parseInt(days) > 0) parts.push(`${days}d`);
  if (hours && parseInt(hours) > 0) parts.push(`${hours}h`);
  if (minutes && parseInt(minutes) > 0) parts.push(`${minutes}m`);
  if (seconds && parseInt(seconds) > 0) parts.push(`${seconds}s`);
  
  return parts.length > 0 ? parts.join(" ") : "< 1s";
};

export default function BatchHistoryPage() {
  const [data, setData] = useState<BatchHistory[]>([]);
  const [details, setDetails] = useState<BatchHistoryDetail[]>([]);
  const [open, setOpen] = useState(false);
  const DEFAULT_STATUSES = ["pending", "processing", "completed", "failed"];
  const [status, setStatus] = useState<string[]>(DEFAULT_STATUSES.slice(0, 2));
  const [typ, setTyp] = useState("");

  useEffect(() => {
    listBatchHistory({ status, type: typ || undefined }).then((res) =>
      setData(res.data),
    );
  }, [status, typ]);

  const columns: Column<BatchHistory>[] = [
    { label: "ID", key: "id" },
    { label: "Type", key: "process_type" },
    {
      label: "Started",
      key: "started_at",
      render: (v) => new Date(v).toLocaleString(),
    },
    {
      label: "Ended",
      key: "ended_at",
      render: (v) => v ? new Date(v).toLocaleString() : "-",
    },
    {
      label: "Duration",
      key: "time_spent",
      render: (v) => v ? formatDuration(v) : "-",
    },
    { label: "Total", key: "total_data", align: "right" },
    { label: "Done", key: "done_data", align: "right" },
    { label: "Status", key: "status" },
    { label: "Error", key: "error_message" },
    { label: "File", key: "file_name" },
    { label: "Path", key: "file_path" },
    {
      label: "",
      render: (_, row) => (
        <Button
          size="small"
          onClick={() => {
            listBatchDetails(row.id).then((r) => {
              setDetails(r.data);
              setOpen(true);
            });
          }}
        >
          View
        </Button>
      ),
    },
  ];

  return (
    <div>
      <h2>Batch History</h2>
      <div style={{ display: "flex", gap: "0.5rem", marginBottom: "1rem" }}>
        <select
          multiple
          value={status}
          onChange={(e) =>
            setStatus(Array.from(e.target.selectedOptions).map((o) => o.value))
          }
        >
          <option value="pending">pending</option>
          <option value="processing">processing</option>
          <option value="completed">completed</option>
          <option value="failed">failed</option>
        </select>
        <input
          placeholder="Type"
          value={typ}
          onChange={(e) => setTyp(e.target.value)}
          style={{ height: "2rem" }}
        />
      </div>
      <SortableTable columns={columns} data={data} />
      <Dialog
        open={open}
        onClose={() => setOpen(false)}
        maxWidth="md"
        fullWidth
      >
        <DialogTitle>Batch Detail</DialogTitle>
        <DialogContent>
          <SortableTable
            columns={[
              { label: "ID", key: "id" },
              { label: "Batch", key: "batch_id" },
              { label: "Reference", key: "reference" },
              { label: "Store", key: "store" },
              { label: "Status", key: "status" },
              { label: "Error", key: "error_message" },
            ]}
            data={details}
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOpen(false)}>Close</Button>
        </DialogActions>
      </Dialog>
    </div>
  );
}
