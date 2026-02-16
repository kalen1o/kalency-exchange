import React from "react";
import { render } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { useBeforeUnload } from "./useBeforeUnload";

function HookHost() {
  useBeforeUnload(true);
  return null;
}

describe("useBeforeUnload", () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("registers beforeunload handler that blocks tab close", () => {
    const addEventListenerSpy = vi.spyOn(window, "addEventListener");
    render(<HookHost />);

    const beforeUnloadCall = addEventListenerSpy.mock.calls.find((call) => call[0] === "beforeunload");
    expect(beforeUnloadCall).toBeDefined();

    const beforeUnloadHandler = beforeUnloadCall?.[1] as (event: BeforeUnloadEvent) => void;
    const event = new Event("beforeunload", { cancelable: true }) as BeforeUnloadEvent;
    beforeUnloadHandler(event);

    expect(event.defaultPrevented).toBe(true);
  });

  it("removes beforeunload handler on unmount", () => {
    const removeEventListenerSpy = vi.spyOn(window, "removeEventListener");
    const { unmount } = render(<HookHost />);
    unmount();

    expect(removeEventListenerSpy).toHaveBeenCalledWith("beforeunload", expect.any(Function));
  });
});
