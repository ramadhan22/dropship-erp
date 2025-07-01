import { useEffect, useState } from "react";
import {
  Alert,
  Button,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  FormControl,
  InputLabel,
  MenuItem,
  Select,
  TextField,
} from "@mui/material";
import SortableTable from "./SortableTable";
import { Link } from "react-router-dom";
import {
  createJenisChannel,
  createStore,
  updateStore,
  deleteStore,
  listJenisChannels,
  listAllStoresDirect,
  fetchShopeeAuthURL,
} from "../api";
import type { JenisChannel, StoreWithChannel } from "../types";

export default function ChannelPage() {
  const [channels, setChannels] = useState<JenisChannel[]>([]);
  const [stores, setStores] = useState<StoreWithChannel[]>([]);
  const [channelName, setChannelName] = useState("");
  const [storeName, setStoreName] = useState("");
  const [storeChannel, setStoreChannel] = useState<number | "">("");
  const [msg, setMsg] = useState<{ type: "success" | "error"; text: string } | null>(null);
  const [openChannel, setOpenChannel] = useState(false);
  const [openStore, setOpenStore] = useState(false);
  const [editing, setEditing] = useState<StoreWithChannel | null>(null);

  useEffect(() => {
    listJenisChannels().then((res) => setChannels(res.data));
    refreshStores();
  }, []);

  async function refreshStores() {
    const res = await listAllStoresDirect();
    setStores(res.data);
  }

  const handleCreateChannel = async () => {
    try {
      const res = await createJenisChannel(channelName);
      setChannels([...channels, { jenis_channel_id: res.data.jenis_channel_id, jenis_channel: channelName }]);
      setChannelName("");
      setOpenChannel(false);
      setMsg({ type: "success", text: "Channel created" });
    } catch (e: any) {
      setMsg({ type: "error", text: e.response?.data?.error || e.message });
    }
  };

  const handleSaveStore = async () => {
    try {
      if (editing) {
        await updateStore(editing.store_id, {
          nama_toko: storeName,
          jenis_channel_id: Number(storeChannel),
        });
      } else {
        await createStore(Number(storeChannel), storeName);
      }
      setStoreName("");
      setStoreChannel("");
      setOpenStore(false);
      setEditing(null);
      await refreshStores();
      setMsg({ type: "success", text: editing ? "Store updated" : "Store created" });
    } catch (e: any) {
      setMsg({ type: "error", text: e.response?.data?.error || e.message });
    }
  };

  const handleEdit = (st: StoreWithChannel) => {
    setEditing(st);
    setStoreName(st.nama_toko);
    setStoreChannel(st.jenis_channel_id);
    setOpenStore(true);
  };

  const handleDelete = async (id: number) => {
    if (!window.confirm("Delete store?")) return;
    try {
      await deleteStore(id);
      await refreshStores();
    } catch (e: any) {
      setMsg({ type: "error", text: e.response?.data?.error || e.message });
    }
  };

  const columns = [
    { label: "Channel", key: "jenis_channel" as const },
    { label: "Store", key: "nama_toko" as const },
    {
      label: "Authorize",
      render: (_: any, row: StoreWithChannel) =>
        row.jenis_channel === "Shopee" && !row.code_id && !row.shop_id ? (
          <Button
            size="small"
            onClick={() =>
              fetchShopeeAuthURL(row.store_id).then((r) =>
                window.open(r.data.url, "_blank"),
              )
            }
          >
            Authorize
          </Button>
        ) : null,
    },
    {
      label: "Actions",
      render: (_: any, row: StoreWithChannel) => (
        <>
          <Button component={Link} to={`/stores/${row.store_id}`} size="small">
            Detail
          </Button>
          <Button size="small" onClick={() => handleEdit(row)}>
            Edit
          </Button>
          <Button size="small" color="error" onClick={() => handleDelete(row.store_id)}>
            Delete
          </Button>
        </>
      ),
    },
  ];

  return (
    <div>
      <h2>Channel Master Data</h2>
      <div style={{ marginBottom: "0.5rem" }}>
        <Button variant="contained" onClick={() => setOpenChannel(true)} sx={{ mr: 1 }}>
          New Channel
        </Button>
        <Button variant="contained" onClick={() => setOpenStore(true)}>
          New Store
        </Button>
      </div>
      {msg && (
        <Alert severity={msg.type} sx={{ mb: 1 }}>
          {msg.text}
        </Alert>
      )}
      <SortableTable columns={columns} data={stores} />

      <Dialog open={openChannel} onClose={() => setOpenChannel(false)}>
        <DialogTitle>Add Channel</DialogTitle>
        <DialogContent>
          <TextField
            label="Channel Name"
            value={channelName}
            onChange={(e) => setChannelName(e.target.value)}
            autoFocus
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOpenChannel(false)}>Cancel</Button>
          <Button onClick={handleCreateChannel} variant="contained">
            Save
          </Button>
        </DialogActions>
      </Dialog>

      <Dialog
        open={openStore}
        onClose={() => {
          setOpenStore(false);
          setEditing(null);
        }}
      >
        <DialogTitle>{editing ? "Edit Store" : "Add Store"}</DialogTitle>
        <DialogContent sx={{ display: "flex", flexDirection: "column", gap: 1 }}>
          <FormControl size="small" sx={{ mt: 1 }}>
            <InputLabel id="channel-label">Channel</InputLabel>
            <Select
              labelId="channel-label"
              label="Channel"
              value={storeChannel}
              onChange={(e) => setStoreChannel(e.target.value as number)}
            >
              <MenuItem value="">
                <em>Select Channel</em>
              </MenuItem>
              {channels.map((c) => (
                <MenuItem key={c.jenis_channel_id} value={c.jenis_channel_id}>
                  {c.jenis_channel}
                </MenuItem>
              ))}
            </Select>
          </FormControl>
          <TextField
            label="Store Name"
            value={storeName}
            onChange={(e) => setStoreName(e.target.value)}
            size="small"
          />
        </DialogContent>
        <DialogActions>
          <Button
            onClick={() => {
              setOpenStore(false);
              setEditing(null);
            }}
          >
            Cancel
          </Button>
          <Button variant="contained" onClick={handleSaveStore} disabled={storeChannel === "" || storeName === ""}>
            Save
          </Button>
        </DialogActions>
      </Dialog>
    </div>
  );
}
