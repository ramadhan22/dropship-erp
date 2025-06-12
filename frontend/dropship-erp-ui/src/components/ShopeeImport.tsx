import { Alert, Button } from "@mui/material";
import { useState } from "react";
import { importShopee } from "../api";

export default function ShopeeImport() {
  const [file, setFile] = useState<File | null>(null);
  const [msg, setMsg] = useState<{
    type: "success" | "error";
    text: string;
  } | null>(null);

  const handleSubmit = async () => {
    try {
      if (!file) return;
      await importShopee(file);
      setMsg({ type: "success", text: "Shopee import successful!" });
    } catch (e: any) {
      setMsg({ type: "error", text: e.response?.data?.error || e.message });
    }
  };

  return (
    <div>
      <h2>Import Shopee XLSX</h2>
      <input
        type="file"
        aria-label="XLSX file"
        onChange={(e) => setFile(e.target.files?.[0] || null)}
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
