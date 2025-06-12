import { useState } from "react";
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
import type { Account } from "../types";

export default function GLPage() {
  const [shop, setShop] = useState("");
  const [from, setFrom] = useState("");
  const [to, setTo] = useState("");
  const [data, setData] = useState<Account[]>([]);

  const handleFetch = async () => {
    const res = await fetchGeneralLedger(shop, from, to);
    setData(res.data);
  };

  return (
    <div>
      <h2>General Ledger</h2>
      <div style={{ display: "flex", gap: "0.5rem" }}>
        <TextField
          label="Shop"
          value={shop}
          onChange={(e) => setShop(e.target.value)}
        />
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
          {data.map((a) => (
            <TableRow key={a.account_id}>
              <TableCell>{a.account_name}</TableCell>
              <TableCell>{a.balance}</TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  );
}
