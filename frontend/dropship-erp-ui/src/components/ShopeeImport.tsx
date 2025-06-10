import { Alert, Button, TextField } from "@mui/material";
import { useState } from "react";
import { importShopee } from "../api";

export default function ShopeeImport() {
  const [path, setPath] = useState("");
  const [msg, setMsg] = useState<{
    type: "success" | "error";
    text: string;
  } | null>(null);

  const handleSubmit = async () => {
    try {
      await importShopee(path);
      setMsg({ type: "success", text: "Shopee import successful!" });
    } catch (e: any) {
      setMsg({ type: "error", text: e.response?.data?.error || e.message });
    }
  };

  return (
    <div>
      <h2>Import Shopee CSV</h2>
      <TextField
        label="Local file path"
        fullWidth
        value={path}
        onChange={(e) => setPath(e.target.value)}
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
