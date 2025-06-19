import { api } from "./index";
import type { JournalEntry, JournalLine } from "../types";

export function listJournal(params?: {
  from?: string;
  to?: string;
  q?: string;
}) {
  const q = new URLSearchParams();
  if (params?.from) q.append("from", params.from);
  if (params?.to) q.append("to", params.to);
  if (params?.q) q.append("q", params.q);
  const qs = q.toString();
  const url = qs ? `/journal/?${qs}` : "/journal/";
  return api.get<JournalEntry[]>(url);
}

export function getJournal(id: number) {
  return api.get<JournalEntry>(`/journal/${id}`);
}

export function getJournalLines(id: number) {
  return api.get<JournalLine & { account_name: string }[]>(
    `/journal/${id}/lines`,
  );
}

export interface JournalEntryWithLines {
  entry: JournalEntry;
  lines: (JournalLine & { account_name: string })[];
}

export function getJournalLinesBySource(sourceId: string) {
  return api.get<JournalEntryWithLines[]>(`/journal/source/${sourceId}/lines`);
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
