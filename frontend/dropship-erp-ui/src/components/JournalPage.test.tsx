import "@testing-library/jest-dom";
import { render } from "@testing-library/react";
import { fireEvent, screen, waitFor } from "@testing-library/dom";
import * as api from "../api/journal";
import JournalPage from "./JournalPage";

jest.mock("../api/journal", () => ({
  listJournal: jest.fn().mockResolvedValue({ data: [] }),
  deleteJournal: jest.fn(),
  createJournal: jest.fn().mockResolvedValue({}),
}));
jest.mock("../api", () => ({
  listAccounts: jest.fn().mockResolvedValue({ data: [] }),
}));

test("fetch list", async () => {
  render(<JournalPage />);
  await waitFor(() => expect(api.listJournal).toHaveBeenCalled());
  expect(screen.getByText(/Journal Entries/i)).toBeInTheDocument();
});

test("create journal", async () => {
  render(<JournalPage />);
  await waitFor(() => expect(api.listJournal).toHaveBeenCalled());
  fireEvent.click(screen.getByRole("button", { name: /Add Journal/i }));
  fireEvent.change(screen.getByLabelText(/Account/i), {
    target: { value: "1" },
  });
  fireEvent.change(screen.getByLabelText(/Debit/i), {
    target: { value: "10" },
  });
  fireEvent.click(screen.getByRole("button", { name: /Save/i }));
  await waitFor(() =>
    expect(api.createJournal).toHaveBeenCalledWith(
      expect.objectContaining({
        entry: expect.any(Object),
      }),
    ),
  );
});
