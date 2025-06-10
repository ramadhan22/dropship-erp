import { Alert, Button, TextField } from "@mui/material";
import { useState } from "react";
import { importDropship } from "../api";

export default function DropshipImport() {
  const [path, setPath] = useState("");
  const [msg, setMsg] = useState<{
    type: "success" | "error";
    text: string;
  } | null>(null);

  const handleSubmit = async () => {
    try {
      await importDropship(path);
      setMsg({ type: "success", text: "Imported successfully!" });
    } catch (e: any) {
      setMsg({ type: "error", text: e.response?.data?.error || e.message });
    }
  };

  return (
    <div>
      <h2>Import Dropship CSV</h2>
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
