import { useEffect, useState } from "react";
import {
  Button,
  TextField,
  Table,
  TableHead,
  TableRow,
  TableCell,
  TableBody,
} from "@mui/material";
import { fetchGeneralLedger } from "../api/gl";
import { listAllStores } from "../api";
import type { Account, Store } from "../types";
import usePagination from "../usePagination";

export default function GLPage() {
  const [shop, setShop] = useState("");
  const [stores, setStores] = useState<Store[]>([]);
  const [from, setFrom] = useState("");
  const [to, setTo] = useState("");
  const [data, setData] = useState<Account[]>([]);
  const { paginated, controls } = usePagination(data);

  useEffect(() => {
    listAllStores().then((s) => setStores(s));
  }, []);

  const handleFetch = async () => {
    const res = await fetchGeneralLedger(shop, from, to);
    setData(res.data);
  };

  return (
    <div>
      <h2>General Ledger</h2>
      <div style={{ display: "flex", gap: "0.5rem" }}>
        <select
          aria-label="Shop"
          value={shop}
          onChange={(e) => setShop(e.target.value)}
        >
          <option value="">Select Store</option>
          {stores.map((s) => (
            <option key={s.store_id} value={s.nama_toko}>
              {s.nama_toko}
            </option>
          ))}
        </select>
        <TextField
          label="From"
          value={from}
          onChange={(e) => setFrom(e.target.value)}
        />
        <TextField
          label="To"
          value={to}
          onChange={(e) => setTo(e.target.value)}
        />
        <Button onClick={handleFetch}>Fetch</Button>
      </div>
      <Table size="small">
        <TableHead>
          <TableRow>
            <TableCell>Account</TableCell>
            <TableCell>Balance</TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {paginated.map((a) => (
            <TableRow key={a.account_id}>
              <TableCell>{a.account_name}</TableCell>
              <TableCell>{a.balance}</TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
      {controls}
    </div>
  );
}
