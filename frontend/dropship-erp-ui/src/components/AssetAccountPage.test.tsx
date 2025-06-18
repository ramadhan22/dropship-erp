import "@testing-library/jest-dom";
import { render } from "@testing-library/react";
import { fireEvent, screen, waitFor } from "@testing-library/dom";
import AssetAccountPage from "./AssetAccountPage";
import * as assetApi from "../api/assetAccounts";

jest.mock("../api/assetAccounts", () => ({
  listAssetAccounts: jest.fn().mockResolvedValue({ data: [] }),
  adjustAssetBalance: jest.fn(),
}));

jest.mock("../api", () => ({
  listAccounts: jest.fn().mockResolvedValue({ data: [] }),
}));

test("fetch asset accounts", async () => {
  render(<AssetAccountPage />);
  await waitFor(() => expect(assetApi.listAssetAccounts).toHaveBeenCalled());
});

test("adjust balance", async () => {
  (assetApi.listAssetAccounts as jest.Mock).mockResolvedValueOnce({
    data: [{ asset_id: 1, account_id: 10, balance: 100 }],
  });
  render(<AssetAccountPage />);
  await waitFor(() => screen.getByText(/1/));
  fireEvent.click(screen.getByRole("button", { name: /Adjust/i }));
  await waitFor(() => screen.getByRole("button", { name: /^Save$/i }));
  fireEvent.change(screen.getByRole("spinbutton", { name: /Balance/i }), {
    target: { value: "150" },
  });
  fireEvent.click(screen.getByRole("button", { name: /^Save$/i }));
  await waitFor(() =>
    expect(assetApi.adjustAssetBalance).toHaveBeenCalledWith(1, 150),
  );
});
