import { act } from "react";
import { createRoot, type Root } from "react-dom/client";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import i18n from "@/i18n";
import { CanvasImageProgress } from "./canvas-image-progress";

let container: HTMLDivElement;
let root: Root;

beforeEach(async () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-10-29T00:00:00Z"));
    (globalThis as typeof globalThis & { IS_REACT_ACT_ENVIRONMENT: boolean }).IS_REACT_ACT_ENVIRONMENT = true;
    await i18n.changeLanguage("zh-CN");
    container = document.createElement("div");
    document.body.append(container);
    root = createRoot(container);
});
afterEach(() => {
    act(() => root.unmount());
    container.remove();
    vi.useRealTimers();
});

describe("canvas image progress", () => {
    it("shows elapsed time without inventing a percentage or ETA", () => {
        act(() => root.render(<CanvasImageProgress progress={{ startedAt: Date.now() - 12_000, stage: "generating" }} />));
        expect(container.textContent).toContain("12秒");
        expect(container.textContent).toContain("模型处理中");
        expect(container.textContent).toContain("模型未提供百分比或剩余时间");
        expect(container.querySelector('[role="progressbar"]')?.hasAttribute("aria-valuenow")).toBe(false);
        expect(container.textContent).not.toContain("%");
        act(() => vi.advanceTimersByTime(2_000));
        expect(container.textContent).toContain("14秒");
    });

    it("changes to saving only when the caller reports an image response", () => {
        const startedAt = Date.now() - 15_000;
        act(() => root.render(<CanvasImageProgress progress={{ startedAt, stage: "generating" }} />));
        act(() => root.render(<CanvasImageProgress progress={{ startedAt, stage: "saving" }} />));
        expect(container.textContent).toContain("图片已生成，正在保存");
        expect(container.textContent).toContain("15秒");
    });

    it("uses settled batch counts and distinguishes failed images", () => {
        act(() => root.render(<CanvasImageProgress compact completed={1} failed={1} total={3} />));
        expect(container.textContent).toContain("已完成 1 / 3 张");
        expect(container.textContent).toContain("1 张失败");
        expect(container.querySelector('[role="progressbar"]')?.getAttribute("aria-valuenow")).toBe("2");
        expect(container.querySelector('[role="progressbar"]')?.getAttribute("aria-valuemax")).toBe("3");
    });

    it("resets elapsed time for a retry and clears its timer when removed", () => {
        act(() => root.render(<CanvasImageProgress progress={{ startedAt: Date.now() - 60_000, stage: "generating" }} />));
        act(() => root.render(<CanvasImageProgress progress={{ startedAt: Date.now(), stage: "generating" }} />));
        expect(container.textContent).toContain("0秒");
        act(() => root.render(null));
        expect(vi.getTimerCount()).toBe(0);
    });
});
