// File: src/components/ReconcileForm.test.tsx
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import * as api from "../api";
import ReconcileForm from "./ReconcileForm";

describe("ReconcileForm", () => {
  it("success path", async () => {
    jest.spyOn(api, "reconcile").mockResolvedValue({} as any);
    render(<ReconcileForm />);
    fireEvent.change(screen.getByLabelText(/Shop/i), {
      target: { value: "Shop" },
    });
    fireEvent.change(screen.getByLabelText(/Purchase ID/i), {
      target: { value: "P1" },
    });
    fireEvent.change(screen.getByLabelText(/Order ID/i), {
      target: { value: "O1" },
    });
    fireEvent.click(screen.getByRole("button", { name: /Reconcile/i }));
    await waitFor(() => screen.getByText(/Reconciliation successful/i));
  });
});
