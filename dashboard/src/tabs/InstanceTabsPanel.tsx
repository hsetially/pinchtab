import { useEffect, useMemo, useRef, useState } from "react";
import type { InstanceTab } from "../generated/types";
import { useAppStore } from "../stores/useAppStore";
import { TabsChart } from "../components/molecules";
import InstanceStats from "../components/molecules/InstanceStats";
import { ErrorBoundary } from "../components/atoms";
import TabBar from "./TabBar";
import SelectedTabPanel from "./SelectedTabPanel";

interface Props {
  tabs: InstanceTab[];
  emptyMessage?: string;
  instanceId?: string;
}

export default function InstanceTabsPanel({
  tabs,
  emptyMessage = "No tabs open",
  instanceId,
}: Props) {
  const [selectedTabId, setSelectedTabId] = useState<string | null>(null);
  const [selectionPinned, setSelectionPinned] = useState(false);
  const manualTabsRef = useRef(false);

  const {
    instances,
    tabsChartData,
    memoryChartData,
    serverChartData,
    currentMetrics,
    settings,
    monitoringShowTelemetry: showTelemetry,
    setMonitoringShowTelemetry: setShowTelemetry,
  } = useAppStore();

  const memoryEnabled = settings.monitoring?.memoryMetrics ?? false;

  const selectedInstance = instances.find((i) => i.id === instanceId);
  const chartInstances = useMemo(
    () =>
      selectedInstance
        ? [
            {
              id: selectedInstance.id,
              profileName: selectedInstance.profileName || "Unknown",
            },
          ]
        : [],
    [selectedInstance],
  );

  useEffect(() => {
    if (tabs.length === 0) {
      setSelectedTabId(null);
      setSelectionPinned(false);
      if (!manualTabsRef.current) {
        setShowTelemetry(true);
      }
      return;
    }

    if (showTelemetry && !manualTabsRef.current) {
      setShowTelemetry(false);
    }
    manualTabsRef.current = false;

    if (selectionPinned && tabs.some((tab) => tab.id === selectedTabId)) {
      return;
    }

    if (
      !tabs.some((tab) => tab.id === selectedTabId) ||
      selectedTabId !== tabs[0].id
    ) {
      if (selectedTabId !== tabs[0].id) {
        setSelectedTabId(tabs[0].id);
      }
      if (selectionPinned) {
        setSelectionPinned(false);
      }
    }
  }, [selectedTabId, selectionPinned, tabs, showTelemetry, setShowTelemetry]);

  const selectedTab = useMemo(
    () => tabs.find((tab) => tab.id === selectedTabId) ?? null,
    [selectedTabId, tabs],
  );

  return (
    <div className="flex min-h-0 flex-1 flex-col">
      <TabBar
        tabs={tabs}
        selectedTabId={selectedTabId}
        pinnedTabId={selectionPinned ? selectedTabId : null}
        telemetryActive={showTelemetry}
        onSelect={(id) => {
          setSelectedTabId(id);
          setSelectionPinned(true);
          setShowTelemetry(false);
        }}
        onTogglePinned={(id) => {
          if (selectionPinned && selectedTabId === id) {
            setSelectionPinned(false);
            setSelectedTabId(tabs[0]?.id ?? null);
            return;
          }
          setSelectedTabId(id);
          setSelectionPinned(true);
          setShowTelemetry(false);
        }}
        onSetTelemetry={(active) => {
          manualTabsRef.current = !active;
          setShowTelemetry(active);
        }}
      />

      {tabs.length === 0 && !showTelemetry ? (
        <div className="flex flex-1 items-center justify-center py-8 text-sm text-text-muted">
          {emptyMessage}
        </div>
      ) : showTelemetry ? (
        <div className="flex-1 overflow-auto">
          <ErrorBoundary
            fallback={
              <div className="flex h-50 items-center justify-center rounded-lg border border-destructive/50 bg-bg-surface text-sm text-destructive">
                Chart crashed - check console
              </div>
            }
          >
            <TabsChart
              data={tabsChartData || []}
              memoryData={memoryEnabled ? memoryChartData : undefined}
              serverData={serverChartData || []}
              instances={chartInstances}
              selectedInstanceId={instanceId || null}
              onSelectInstance={() => {}}
            />
          </ErrorBoundary>
          <InstanceStats
            instance={selectedInstance}
            metrics={instanceId ? currentMetrics[instanceId] : null}
            tabs={tabs}
          />
        </div>
      ) : (
        <SelectedTabPanel selectedTab={selectedTab} instanceId={instanceId} />
      )}
    </div>
  );
}
