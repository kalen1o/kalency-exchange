import { describe, expect, it } from "vitest";
import { formatFloatPrice } from "./price";

describe("formatFloatPrice", () => {
  it("formats prices with two decimal places and quote currency", () => {
    expect(formatFloatPrice(102.2, "usd")).toBe("102.20 USD");
  });

  it("returns fallback symbol for missing values", () => {
    expect(formatFloatPrice(null, "USD")).toBe("-");
    expect(formatFloatPrice(undefined, "USD")).toBe("-");
    expect(formatFloatPrice(Number.NaN, "USD")).toBe("-");
  });
});
