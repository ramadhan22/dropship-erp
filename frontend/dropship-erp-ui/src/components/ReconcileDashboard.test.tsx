import "@testing-library/jest-dom";
import { render } from "@testing-library/react";
import { waitFor, screen, fireEvent } from "@testing-library/dom";
import * as api from "../api/reconcile";
import ReconcileDashboard from "./ReconcileDashboard";

jest.mock("../api/reconcile", () => ({
  listUnmatched: jest.fn().mockResolvedValue({ data: [] }),
  bulkReconcile: jest.fn(),
}));

test("load unmatched", async () => {
  render(<ReconcileDashboard />);
  screen.getByLabelText(/Shop/i).focus();
  fireEvent.change(screen.getByLabelText(/Shop/i), { target: { value: "S" } });
  fireEvent.click(screen.getByRole("button", { name: /Refresh/i }));
  await waitFor(() => expect(api.listUnmatched).toHaveBeenCalled());
});
