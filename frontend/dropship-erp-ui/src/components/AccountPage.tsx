import { Alert, Button, TextField } from "@mui/material";
import { useEffect, useState } from "react";
import { createAccount, listAccounts } from "../api";
import type { Account } from "../types";

export default function AccountPage() {
  const [accounts, setAccounts] = useState<Account[]>([]);
  const [code, setCode] = useState("");
  const [name, setName] = useState("");
  const [type, setType] = useState("");
  const [parent, setParent] = useState("");
  const [msg, setMsg] = useState<{
    type: "success" | "error";
    text: string;
  } | null>(null);

  useEffect(() => {
    listAccounts().then((res) => setAccounts(res.data));
  }, []);

  const handleCreate = async () => {
    try {
      const res = await createAccount({
        account_code: code,
        account_name: name,
        account_type: type,
        parent_id: parent ? Number(parent) : null,
      });
      setAccounts([
        ...accounts,
        {
          account_id: res.data.account_id,
          account_code: code,
          account_name: name,
          account_type: type,
          parent_id: parent ? Number(parent) : null,
          balance: 0,
        },
      ]);
      setCode("");
      setName("");
      setType("");
      setParent("");
      setMsg({ type: "success", text: "Account created" });
    } catch (e: any) {
      setMsg({ type: "error", text: e.response?.data?.error || e.message });
    }
  };

  return (
    <div>
      <h2>Accounts</h2>
      <div>
        <TextField
          label="Code"
          value={code}
          onChange={(e) => setCode(e.target.value)}
          size="small"
        />
        <TextField
          label="Name"
          value={name}
          onChange={(e) => setName(e.target.value)}
          size="small"
          sx={{ ml: 1 }}
        />
        <TextField
          label="Type"
          value={type}
          onChange={(e) => setType(e.target.value)}
          size="small"
          sx={{ ml: 1 }}
        />
        <select
          aria-label="Parent ID"
          value={parent}
          onChange={(e) => setParent(e.target.value)}
          style={{ marginLeft: "0.5rem" }}
        >
          <option value="">No Parent</option>
          {accounts.map((a) => (
            <option key={a.account_id} value={String(a.account_id)}>
              {a.account_code} - {a.account_name}
            </option>
          ))}
        </select>
        <Button variant="contained" onClick={handleCreate} sx={{ ml: 1 }}>
          Add Account
        </Button>
      </div>
      {msg && (
        <Alert severity={msg.type} sx={{ mt: 2 }}>
          {msg.text}
        </Alert>
      )}
      {accounts.length > 0 && (
        <ul>
          {accounts.map((a) => (
            <li key={a.account_id}>
              {a.account_code} - {a.account_name}
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
