import { useEffect, useState } from "react";
import {
  Table,
  TableHead,
  TableRow,
  TableCell,
  TableBody,
  Button,
} from "@mui/material";
import { listJournal, deleteJournal } from "../api/journal";
import type { JournalEntry } from "../types";

export default function JournalPage() {
  const [list, setList] = useState<JournalEntry[]>([]);
  const fetchData = () => listJournal().then((r) => setList(r.data));
  useEffect(() => {
    fetchData();
  }, []);
  return (
    <div>
      <h2>Journal Entries</h2>
      <Table size="small">
        <TableHead>
          <TableRow>
            <TableCell>ID</TableCell>
            <TableCell>Date</TableCell>
            <TableCell>Description</TableCell>
            <TableCell></TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {list.map((j) => (
            <TableRow key={j.journal_id}>
              <TableCell>{j.journal_id}</TableCell>
              <TableCell>
                {new Date(j.entry_date).toLocaleDateString()}
              </TableCell>
              <TableCell>{j.description}</TableCell>
              <TableCell>
                <Button
                  size="small"
                  onClick={() => {
                    deleteJournal(j.journal_id).then(fetchData);
                  }}
                >
                  Del
                </Button>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  );
}
