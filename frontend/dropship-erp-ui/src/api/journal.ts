import { api } from "./index";
import type { JournalEntry, JournalLine } from "../types";

export function listJournal() {
  return api.get<JournalEntry[]>("/journal/");
}

export function getJournal(id: number) {
  return api.get<JournalEntry>(`/journal/${id}`);
}

export function getJournalLines(id: number) {
  return api.get<JournalLine & { account_name: string }[]>(
    `/journal/${id}/lines`
  );
}

export function deleteJournal(id: number) {
  return api.delete(`/journal/${id}`);
}

export function createJournal(data: {
  entry: Partial<JournalEntry>;
  lines: Partial<JournalLine>[];
}) {
  return api.post("/journal/", data);
}
