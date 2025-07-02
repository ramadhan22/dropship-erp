import "@testing-library/jest-dom";
import { render } from "@testing-library/react";
import { waitFor } from "@testing-library/dom";
import * as api from "../api/shopeeAdjustments";
import ShopeeAdjustmentPage from "./ShopeeAdjustmentPage";

jest.mock("../api/shopeeAdjustments", () => ({
  listShopeeAdjustments: jest.fn().mockResolvedValue({ data: [] }),
  updateShopeeAdjustment: jest.fn().mockResolvedValue({}),
  deleteShopeeAdjustment: jest.fn().mockResolvedValue({}),
}));

test("fetches list on load", async () => {
  render(<ShopeeAdjustmentPage />);
  await waitFor(() => expect(api.listShopeeAdjustments).toHaveBeenCalled());
});
