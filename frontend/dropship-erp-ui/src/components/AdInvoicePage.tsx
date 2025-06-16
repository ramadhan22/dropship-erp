import { useEffect, useState } from "react";
import { Button, Alert } from "@mui/material";
import SortableTable from "./SortableTable";
import type { Column } from "./SortableTable";
import { importAdInvoice, listAdInvoices } from "../api/adInvoices";
import type { AdInvoice } from "../types";

export default function AdInvoicePage() {
  const [file, setFile] = useState<File | null>(null);
  const [list, setList] = useState<AdInvoice[]>([]);
  const [sortKey, setSortKey] = useState<keyof AdInvoice>("invoice_date");
  const [sortDir, setSortDir] = useState<"asc" | "desc">("desc");
  const [msg, setMsg] = useState<{
    type: "success" | "error";
    text: string;
  } | null>(null);

  const fetchData = () =>
    listAdInvoices({ sort: sortKey as string, dir: sortDir }).then((r) =>
      setList(r.data),
    );

  useEffect(() => {
    fetchData();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [sortKey, sortDir]);

  const columns: Column<AdInvoice>[] = [
    { label: "Invoice", key: "invoice_no" },
    { label: "Store", key: "store" },
    {
      label: "Date",
      key: "invoice_date",
      render: (v) => new Date(v).toLocaleDateString(),
    },
    {
      label: "Total",
      key: "total",
      align: "right",
      render: (v) =>
        Number(v).toLocaleString("id-ID", {
          style: "currency",
          currency: "IDR",
        }),
    },
  ];

  return (
    <div>
      <h2>Ads Invoices</h2>
      <div style={{ display: "flex", gap: "0.5rem", marginBottom: "1rem" }}>
        <input
          type="file"
          accept="application/pdf"
          onChange={(e) => setFile(e.target.files?.[0] ?? null)}
        />
        <Button
          variant="contained"
          onClick={async () => {
            if (!file) return;
            try {
              await importAdInvoice(file);
              setFile(null);
              fetchData();
              setMsg({ type: "success", text: "uploaded" });
            } catch (e: any) {
              setMsg({ type: "error", text: e.message });
            }
          }}
        >
          Upload
        </Button>
      </div>
      {msg && <Alert severity={msg.type}>{msg.text}</Alert>}
      <SortableTable
        columns={columns}
        data={list}
        defaultSort={{ key: sortKey, direction: sortDir }}
        onSortChange={(k, d) => {
          setSortKey(k as keyof AdInvoice);
          setSortDir(d);
        }}
      />
    </div>
  );
}
