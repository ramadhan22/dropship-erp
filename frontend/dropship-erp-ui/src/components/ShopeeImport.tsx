import { Alert, Button } from "@mui/material";
import { useState } from "react";
import { importShopee } from "../api";

export default function ShopeeImport() {
  const [files, setFiles] = useState<File[]>([]);
  const [msg, setMsg] = useState<{
    type: "success" | "error";
    text: string;
  } | null>(null);

  const handleSubmit = async () => {
    try {
      if (files.length === 0) return;
      const res = await importShopee(files);
      setMsg({
        type: "success",
        text: `Imported ${res.data.inserted} rows successfully!`,
      });
    } catch (e: any) {
      setMsg({ type: "error", text: e.response?.data?.error || e.message });
    }
  };

  return (
    <div>
      <h2>Import Shopee XLSX</h2>
      <input
        type="file"
        multiple
        aria-label="XLSX file"
        onChange={(e) => setFiles(Array.from(e.target.files || []))}
      />
      <Button variant="contained" onClick={handleSubmit} sx={{ mt: 2 }}>
        Import
      </Button>
      {msg && (
        <Alert severity={msg.type} sx={{ mt: 2 }}>
          {msg.text}
        </Alert>
      )}
    </div>
  );
}
