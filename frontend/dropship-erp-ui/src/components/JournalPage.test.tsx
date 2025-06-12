import "@testing-library/jest-dom";
import { render } from "@testing-library/react";
import { waitFor, screen } from "@testing-library/dom";
import * as api from "../api/journal";
import JournalPage from "./JournalPage";

jest.mock("../api/journal", () => ({
  listJournal: jest.fn().mockResolvedValue({ data: [] }),
  deleteJournal: jest.fn(),
}));

test("fetch list", async () => {
  render(<JournalPage />);
  await waitFor(() => expect(api.listJournal).toHaveBeenCalled());
  expect(screen.getByText(/Journal Entries/i)).toBeInTheDocument();
});
