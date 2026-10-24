import { App, Button, Checkbox, Input, Modal, Tabs } from "antd";
import { RefreshCw, Search } from "lucide-react";
import { useEffect, useMemo, useState } from "react";
import { useTranslation } from "react-i18next";

import {
  fetchChannelModels,
  type ModelCatalogEntry,
} from "@/services/api/image";
import type { ModelChannel } from "@/stores/use-config-store";

// Channel model selector: fetch upstream models or add them manually, then include checked models in the channel list.
export function ModelSelectModal({
  open,
  channel,
  selectedModels,
  onConfirm,
  onClose,
}: {
  open: boolean;
  channel: ModelChannel | null;
  selectedModels: ModelChannel["models"];
  onConfirm: (models: ModelCatalogEntry[]) => void;
  onClose: () => void;
}) {
  const { message } = App.useApp();
  const { t } = useTranslation();
  const [existing, setExisting] = useState<ModelCatalogEntry[]>([]);
  const [fetched, setFetched] = useState<ModelCatalogEntry[]>([]);
  const [selected, setSelected] = useState<Set<string>>(new Set());
  const [activeTab, setActiveTab] = useState("new");
  const [search, setSearch] = useState("");
  const [manual, setManual] = useState("");
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (!open) return;
    const models = selectedModels.map(({ name: id, displayName }) => ({
      id,
      displayName,
    }));
    setExisting(models);
    setFetched([]);
    setSelected(new Set(models.map((model) => model.id)));
    setActiveTab(models.length ? "existing" : "new");
    setSearch("");
    setManual("");
  }, [open, selectedModels]);

  const currentList = activeTab === "new" ? fetched : existing;
  const visibleList = useMemo(() => {
    const keyword = search.trim().toLowerCase();
    return keyword
      ? currentList.filter((model) =>
          `${model.displayName || ""} ${model.id}`
            .toLowerCase()
            .includes(keyword),
        )
      : currentList;
  }, [currentList, search]);
  const visibleSelectedCount = visibleList.filter((model) =>
    selected.has(model.id),
  ).length;

  const toggle = (id: string, checked: boolean) =>
    setSelected((current) => {
      const next = new Set(current);
      if (checked) next.add(id);
      else next.delete(id);
      return next;
    });

  const selectVisible = (checked: boolean) =>
    setSelected((current) => {
      const next = new Set(current);
      visibleList.forEach(({ id }) =>
        checked ? next.add(id) : next.delete(id),
      );
      return next;
    });

  const addManual = () => {
    const id = manual.trim();
    if (!id) return;
    if (![...fetched, ...existing].some((model) => model.id === id))
      setFetched((current) => [{ id }, ...current]);
    setSelected((current) => new Set(current).add(id));
    setManual("");
    setActiveTab("new");
  };

  const fetchModels = async () => {
    if (!channel) return;
    if (!channel.baseUrl.trim() || !channel.apiKey.trim()) {
      message.error(t("config.modelSelect.missingConfig"));
      return;
    }
    setLoading(true);
    try {
      const models = await fetchChannelModels(channel);
      setFetched(models);
      setActiveTab("new");
      message.success(
        t("config.modelSelect.fetched", { count: models.length }),
      );
    } catch (error) {
      message.error(
        error instanceof Error
          ? error.message
          : t("config.modelSelect.fetchFailed"),
      );
    } finally {
      setLoading(false);
    }
  };

  const confirm = () => {
    const ordered = [...existing, ...fetched]
      .filter(
        (model, index, list) =>
          list.findIndex((item) => item.id === model.id) === index,
      )
      .filter((model) => selected.has(model.id));
    onConfirm(ordered);
    onClose();
  };

  return (
    <Modal
      open={open}
      width={880}
      centered
      onCancel={onClose}
      title={
        <span>
          {t("config.modelSelect.title")}{" "}
          <span className="ml-2 text-xs font-normal text-stone-500">
            {t("config.modelSelect.selected", {
              selected: selected.size,
              total: new Set([...existing, ...fetched].map((model) => model.id))
                .size,
            })}
          </span>
        </span>
      }
      styles={{ body: { maxHeight: "62vh", overflowY: "auto" } }}
      footer={[
        <Button key="cancel" onClick={onClose}>
          {t("common.cancel")}
        </Button>,
        <Button key="confirm" type="primary" onClick={confirm}>
          {t("config.modelSelect.confirm")}
        </Button>,
      ]}
    >
      <div className="flex flex-wrap items-center gap-3">
        <Input
          className="min-w-[200px] flex-1"
          value={search}
          onChange={(event) => setSearch(event.target.value)}
          placeholder={t("config.modelSelect.search")}
          prefix={<Search className="size-4 text-stone-400" />}
          allowClear
        />
        <Input
          className="min-w-[180px] flex-1"
          value={manual}
          onChange={(event) => setManual(event.target.value)}
          onPressEnter={addManual}
          placeholder={t("config.modelSelect.modelName")}
        />
        <Button onClick={addManual}>{t("config.modelSelect.add")}</Button>
        <Button
          icon={<RefreshCw className="size-4" />}
          loading={loading}
          onClick={() => void fetchModels()}
        >
          {t("config.modelSelect.fetch")}
        </Button>
      </div>
      <div className="mt-2 text-xs text-stone-500">
        {t("config.modelSelect.description")}
      </div>

      <Tabs
        className="mt-3"
        activeKey={activeTab}
        onChange={setActiveTab}
        items={[
          {
            key: "new",
            label: t("config.modelSelect.fetchedTab", {
              count: fetched.length,
            }),
          },
          {
            key: "existing",
            label: t("config.modelSelect.existingTab", {
              count: existing.length,
            }),
          },
        ]}
      />

      <div className="mb-3 flex items-center justify-between gap-2">
        <span className="text-xs text-stone-500">
          {t("config.modelSelect.visibleSelected", {
            selected: visibleSelectedCount,
            total: visibleList.length,
          })}
        </span>
        <div className="flex gap-2">
          <Button
            size="small"
            disabled={!visibleList.length}
            onClick={() => selectVisible(true)}
          >
            {t("config.modelSelect.selectVisible")}
          </Button>
          <Button
            size="small"
            disabled={!visibleSelectedCount}
            onClick={() => selectVisible(false)}
          >
            {t("config.modelSelect.clearVisible")}
          </Button>
        </div>
      </div>

      {visibleList.length ? (
        <div className="grid grid-cols-1 gap-x-8 gap-y-3 md:grid-cols-2">
          {visibleList.map(({ id, displayName }) => (
            <Checkbox
              key={id}
              checked={selected.has(id)}
              onChange={(event) => toggle(id, event.target.checked)}
            >
              <span
                className="inline-flex min-w-0 items-baseline gap-2"
                title={id}
              >
                <span className="truncate">{displayName || id}</span>
                {displayName && displayName !== id ? (
                  <span className="shrink-0 text-xs text-stone-500">
                    ID: {id}
                  </span>
                ) : null}
              </span>
            </Checkbox>
          ))}
        </div>
      ) : (
        <div className="py-8 text-center text-sm text-stone-500">
          {t(
            activeTab === "new"
              ? "config.modelSelect.fetchedEmpty"
              : "config.modelSelect.existingEmpty",
          )}
        </div>
      )}
    </Modal>
  );
}
