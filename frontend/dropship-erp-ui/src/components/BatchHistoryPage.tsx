import { useEffect, useState } from "react";
import SortableTable from "./SortableTable";
import type { Column } from "./SortableTable";
import { listBatchHistory } from "../api";
import type { BatchHistory } from "../types";

export default function BatchHistoryPage() {
  const [data, setData] = useState<BatchHistory[]>([]);

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
  ];

  return (
    <div>
      <h2>Batch History</h2>
      <SortableTable columns={columns} data={data} />
    </div>
  );
}
