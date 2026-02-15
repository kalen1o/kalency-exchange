import { describe, expect, it } from "vitest";
import { formatLocaleDateTime } from "./datetime";

describe("formatLocaleDateTime", () => {
  it("formats ISO timestamps via Intl.DateTimeFormat", () => {
    const input = "2026-02-15T00:01:00Z";
    const options: Intl.DateTimeFormatOptions = {
      timeZone: "UTC",
      year: "numeric",
      month: "2-digit",
      day: "2-digit",
      hour: "2-digit",
      minute: "2-digit",
      second: "2-digit",
      hour12: true
    };

    const expected = new Intl.DateTimeFormat("en-US", options).format(new Date(input));
    expect(formatLocaleDateTime(input, "en-US", options)).toBe(expected);
  });

  it("returns fallback symbol for invalid values", () => {
    expect(formatLocaleDateTime(null)).toBe("-");
    expect(formatLocaleDateTime(undefined)).toBe("-");
    expect(formatLocaleDateTime("not-a-date")).toBe("-");
  });
});

