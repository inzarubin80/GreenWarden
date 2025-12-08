import React, { useCallback, useState } from "react";
import { MapView } from "./components/MapView";
import { ProblemCard } from "./components/ProblemCard";
import type { Problem } from "./types/problem";
import { getViolationsByBbox } from "./api/violations";
import type { Violation } from "./types/violation";

export default function App() {
  const [problems, setProblems] = useState<Problem[]>([]);
  const [selected, setSelected] = useState<Problem | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [currentBbox, setCurrentBbox] = useState<
    [number, number, number, number] | null
  >(null);

  const violationToProblem = (v: Violation): Problem => ({
    id: String(v.id),
    title: v.description || `Нарушение #${v.id}`,
    description: v.description,
    status: v.status,
    latitude: v.lat,
    longitude: v.lng,
    address: v.address,
    createdAt: v.created_at
  });

  const loadViolations = useCallback(
    async (bbox: [number, number, number, number]) => {
      try {
        setLoading(true);
        setError(null);
        const resp = await getViolationsByBbox(bbox);
        const items = resp.items ?? [];
        const mapped = items.map(violationToProblem);
        setProblems(mapped);
        if (mapped.length > 0) {
          setSelected((prev) =>
            prev ? mapped.find((p) => p.id === prev.id) ?? mapped[0] : mapped[0]
          );
        } else {
          setSelected(null);
        }
      } catch (e) {
        setError(
          e instanceof Error ? e.message : "Не удалось загрузить нарушения"
        );
      } finally {
        setLoading(false);
      }
    },
    []
  );

  const handleBoundsChange = useCallback(
    (bbox: [number, number, number, number]) => {
      setCurrentBbox(bbox);
      void loadViolations(bbox);
    },
    [loadViolations]
  );

  const handleSelect = (problem: Problem) => {
    setSelected(problem);
  };

  return (
    <div className="app-layout">
      <div className="map-column">
        <MapView
          problems={problems}
          selectedId={selected?.id ?? null}
          onSelect={handleSelect}
          onBoundsChange={handleBoundsChange}
        />
      </div>
      <div className="card-column">
        {loading && <div className="status">Загрузка нарушений...</div>}
        {error && <div className="status error">Ошибка: {error}</div>}
        {!loading && !error && selected && (
          <ProblemCard problem={selected} />
        )}
        {!loading && !error && !selected && (
          <div className="status">Проблемы не найдены</div>
        )}
      </div>
    </div>
  );
}


