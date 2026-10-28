import { act } from "react";
import { createRoot, type Root } from "react-dom/client";
import { afterEach, beforeEach, expect, it, vi } from "vitest";
import i18n from "@/i18n";
import { CanvasNode } from "./canvas-node";
import { CanvasNodeType, type CanvasNodeImage, type CanvasNodeData } from "@/types/canvas";

let container: HTMLDivElement;
let root: Root;
const pending = (id: string): CanvasNodeImage => ({ id, status: "loading", content: "", naturalWidth: 0, naturalHeight: 0, bytes: 0, mimeType: "", generationProgress: { startedAt: Date.now(), stage: "generating" } });
const render = (images: CanvasNodeImage[], status: "loading" | "success" | "idle" = "loading") => {
    const primary = images.find((image) => image.content);
    const data: CanvasNodeData = { id: "node", type: CanvasNodeType.Image, title: "Test", position: { x: 0, y: 0 }, width: 340, height: 240, metadata: { status, images, content: primary?.content, primaryImageId: primary?.id, generationProgress: images[0]?.generationProgress } };
    const noop = vi.fn();
    act(() => root.render(<CanvasNode data={data} scale={1} isSelected={false} isRelated={false} isFocusRelated={false} isConnectionTarget={false} isConnecting={false} showPanel={false} showImageInfo={false} onMouseDown={noop} onHoverStart={noop} onHoverEnd={noop} onConnectStart={noop} onResizeStart={noop} onResize={noop} onResizeEnd={noop} onContentChange={noop} onTitleChange={noop} onContextMenu={noop} />));
};
beforeEach(async () => {
    (globalThis as typeof globalThis & { IS_REACT_ACT_ENVIRONMENT: boolean }).IS_REACT_ACT_ENVIRONMENT = true;
    await i18n.changeLanguage("zh-CN");
    container = document.createElement("div");
    document.body.append(container);
    root = createRoot(container);
});
afterEach(() => { act(() => root.unmount()); container.remove(); });

it("keeps batch progress visible after the first image succeeds", () => {
    const first = pending("one");
    const second = pending("two");
    render([first, second]);
    expect(container.textContent).toContain("已完成 0 / 2 张");
    const ready = { ...first, status: "success" as const, content: "data:image/png;base64,dGVzdA==" };
    render([ready, second]);
    expect(container.querySelector("img")).not.toBeNull();
    expect(container.textContent).toContain("已完成 1 / 2 张");
    expect(container.querySelectorAll('[role="progressbar"]')).toHaveLength(1);
    render([ready, { ...second, status: "error", errorDetails: "Failed" }], "success");
    expect(container.querySelector('[role="progressbar"]')).toBeNull();
});

it("shows progress for a single image and removes it on completion", () => {
    const image = pending("one");
    render([image]);
    expect(container.querySelector('[role="progressbar"]')).not.toBeNull();
    render([{ ...image, status: "success", content: "data:image/png;base64,dGVzdA==" }], "success");
    expect(container.querySelector("img")).not.toBeNull();
    expect(container.querySelector('[role="progressbar"]')).toBeNull();
});

it("does not leave a waiting bar after a batch is cancelled", () => {
    render([pending("one"), pending("two")]);
    render(["one", "two"].map((id) => ({ ...pending(id), status: "error", errorDetails: "Request cancelled" })), "idle");
    expect(container.querySelector('[role="progressbar"]')).toBeNull();
    expect(container.textContent).toContain("Request cancelled");
});
