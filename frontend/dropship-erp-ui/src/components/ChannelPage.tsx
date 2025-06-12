import { Alert, Button, TextField } from "@mui/material";
import { useEffect, useState } from "react";
import { createJenisChannel, createStore, listJenisChannels, listStores } from "../api";
import type { JenisChannel, Store } from "../types";

export default function ChannelPage() {
  const [channels, setChannels] = useState<JenisChannel[]>([]);
  const [selected, setSelected] = useState<number | "">("");
  const [stores, setStores] = useState<Store[]>([]);
  const [channelName, setChannelName] = useState("");
  const [storeName, setStoreName] = useState("");
  const [msg, setMsg] = useState<{ type: "success" | "error"; text: string } | null>(null);

  useEffect(() => {
    listJenisChannels().then((res) => setChannels(res.data));
  }, []);

  const handleCreateChannel = async () => {
    try {
      const res = await createJenisChannel(channelName);
      setChannels([...channels, { jenis_channel_id: res.data.jenis_channel_id, jenis_channel: channelName }]);
      setChannelName("");
      setMsg({ type: "success", text: "Channel created" });
    } catch (e: any) {
      setMsg({ type: "error", text: e.response?.data?.error || e.message });
    }
  };

  const handleSelect = async (id: number) => {
    setSelected(id);
    try {
      const res = await listStores(id);
      setStores(res.data ?? []);
    } catch (e: any) {
      setStores([]);
      setMsg({ type: "error", text: e.response?.data?.error || e.message });
    }
  };

  const handleCreateStore = async () => {
    try {
      if (selected === "") return;
      await createStore(Number(selected), storeName);
      const res = await listStores(Number(selected));
      setStores(res.data);
      setStoreName("");
      setMsg({ type: "success", text: "Store created" });
    } catch (e: any) {
      setMsg({ type: "error", text: e.response?.data?.error || e.message });
    }
  };

  return (
    <div>
      <h2>Channel Master Data</h2>
      <div>
        <TextField
          label="New Channel"
          value={channelName}
          onChange={(e) => setChannelName(e.target.value)}
          size="small"
        />
        <Button variant="contained" onClick={handleCreateChannel} sx={{ ml: 1 }}>
          Add Channel
        </Button>
      </div>
      <div style={{ marginTop: "1rem" }}>
        <select
          aria-label="Channel Select"
          value={selected}
          onChange={(e) => handleSelect(Number(e.target.value))}
        >
          <option value="" key="placeholder">Select Channel</option>
          {channels.map((c) => (
            <option key={c.jenis_channel_id} value={c.jenis_channel_id}>
              {c.jenis_channel}
            </option>
          ))}
        </select>
        <TextField
          label="Store Name"
          value={storeName}
          onChange={(e) => setStoreName(e.target.value)}
          size="small"
          sx={{ ml: 1 }}
        />
        <Button variant="contained" onClick={handleCreateStore} sx={{ ml: 1 }}>
          Add Store
        </Button>
      </div>
      {msg && (
        <Alert severity={msg.type} sx={{ mt: 2 }}>
          {msg.text}
        </Alert>
      )}
      {stores.length > 0 && (
        <ul>
          {stores.map((s) => (
            <li key={s.store_id}>{s.nama_toko}</li>
          ))}
        </ul>
      )}
    </div>
  );
}
