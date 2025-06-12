import { api } from "./index";
import type { JournalEntry } from "../types";

export function listJournal() {
  return api.get<JournalEntry[]>("/journal/");
}

export function getJournal(id: number) {
  return api.get<JournalEntry>(`/journal/${id}`);
}

export function deleteJournal(id: number) {
  return api.delete(`/journal/${id}`);
}
