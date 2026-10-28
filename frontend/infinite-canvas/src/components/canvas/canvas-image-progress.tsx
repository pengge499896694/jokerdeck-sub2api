import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";

import { canvasThemes } from "@/lib/canvas-theme";
import { formatDuration } from "@/lib/image-utils";
import { useThemeStore } from "@/stores/use-theme-store";
import type { CanvasImageProgress as CanvasImageProgressState } from "@/types/canvas";

export function CanvasImageProgress({ progress, completed = 0, failed = 0, total = 1, compact = false }: {
    progress?: CanvasImageProgressState;
    completed?: number;
    failed?: number;
    total?: number;
    compact?: boolean;
}) {
    const { t } = useTranslation();
    const theme = canvasThemes[useThemeStore((state) => state.theme)];
    const [mountedAt] = useState(Date.now);
    const [now, setNow] = useState(Date.now);
    const startedAt = progress?.startedAt ?? mountedAt;
    useEffect(() => {
        const timer = window.setInterval(() => setNow(Date.now()), 1000);
        return () => window.clearInterval(timer);
    }, []);
    const settled = Math.min(total, completed + failed);
    const stage = t(`generation.canvas.${progress?.stage ?? "generating"}`);

    return (
        <div className={`flex w-full flex-col justify-center gap-2 ${compact ? "rounded-xl border px-3 py-2" : "h-full px-6 text-center"}`}
            style={{ background: compact ? theme.toolbar.panel : theme.node.fill, color: theme.node.text, borderColor: theme.node.stroke }}>
            <div className="text-xs font-medium" role="status">{stage}</div>
            <div className="flex justify-between gap-2 text-[11px]" style={{ color: theme.node.muted }}>
                <span>{t("generation.canvas.elapsed", { duration: formatDuration(Math.max(0, now - startedAt)) })}</span>
                {total > 1 ? <span>{t("generation.canvas.completed", { completed, total })}</span> : null}
            </div>
            <div className="relative h-1.5 overflow-hidden rounded-full" style={{ background: theme.node.stroke }}
                role="progressbar" aria-label={stage} aria-valuemin={0} aria-valuemax={total}
                aria-valuenow={total > 1 ? settled : undefined}>
                {total > 1 && settled > 0 ? <div className="h-full rounded-full transition-[width] duration-300" style={{ width: `${settled / total * 100}%`, background: theme.node.activeStroke }} /> : null}
                <div className="canvas-generation-wait absolute inset-y-0 w-1/3 rounded-full opacity-60" style={{ background: theme.node.activeStroke }} />
            </div>
            {failed > 0 ? <div className="text-[11px] text-red-500">{t("generation.canvas.failed", { count: failed })}</div> : null}
            {!compact ? <p className="text-[10px] leading-4" style={{ color: theme.node.muted }}>{t("generation.canvas.waitHint")}</p> : null}
        </div>
    );
}
