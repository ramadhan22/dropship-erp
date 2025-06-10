// File: src/components/BalanceSheetPage.tsx

import {
  Button,
  Card,
  CardContent,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableRow,
  TextField,
  Typography,
} from "@mui/material";
import { useState } from "react";
import { fetchBalanceSheet } from "../api";
import type { BalanceCategory } from "../types";

export default function BalanceSheetPage() {
  const [shop, setShop] = useState("");
  const [period, setPeriod] = useState("");
  const [data, setData] = useState<BalanceCategory[]>([]);

  const handleFetch = async () => {
    const res = await fetchBalanceSheet(shop, period);
    setData(res.data);
  };

  return (
    <div>
      <h2>Balance Sheet</h2>
      <TextField
        label="Shop"
        value={shop}
        onChange={(e) => setShop(e.target.value)}
        sx={{ mr: 2 }}
      />
      <TextField
        label="Period (YYYY-MM)"
        value={period}
        onChange={(e) => setPeriod(e.target.value)}
        sx={{ mr: 2 }}
      />
      <Button variant="contained" onClick={handleFetch}>
        Fetch
      </Button>

      <div style={{ marginTop: "1rem" }}>
        {data.map((cat) => (
          <Card key={cat.category} sx={{ mb: 2 }}>
            <CardContent>
              <Typography variant="h6">
                {cat.category} (Total: {cat.total})
              </Typography>
              <Table size="small">
                <TableHead>
                  <TableRow>
                    <TableCell>Code</TableCell>
                    <TableCell>Name</TableCell>
                    <TableCell>Balance</TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {cat.accounts.map((a) => (
                    <TableRow key={a.account_id}>
                      <TableCell>{a.account_code}</TableCell>
                      <TableCell>{a.account_name}</TableCell>
                      <TableCell>{a.balance}</TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  );
}
