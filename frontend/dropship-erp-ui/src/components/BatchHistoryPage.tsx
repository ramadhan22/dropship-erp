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

export default function BatchHistoryPage() {
  const [data, setData] = useState<BatchHistory[]>([]);
  const [details, setDetails] = useState<BatchHistoryDetail[]>([]);
  const [open, setOpen] = useState(false);

  useEffect(() => {
    listBatchHistory().then((res) => setData(res.data));
  }, []);

  const columns: Column<BatchHistory>[] = [
    { label: "Type", key: "process_type" },
    {
      label: "Started",
      key: "started_at",
      render: (v) => new Date(v).toLocaleString(),
    },
    { label: "Total", key: "total_data", align: "right" },
    { label: "Done", key: "done_data", align: "right" },
    { label: "Status", key: "status" },
    { label: "Error", key: "error_message" },
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
      <SortableTable columns={columns} data={data} />
      <Dialog open={open} onClose={() => setOpen(false)} maxWidth="md" fullWidth>
        <DialogTitle>Batch Detail</DialogTitle>
        <DialogContent>
          <SortableTable
            columns={[
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
