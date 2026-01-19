import React, { useCallback, useMemo, useRef } from "react";
import {
  YMaps,
  Map,
  Placemark,
  ZoomControl,
  FullscreenControl
} from "@pbe/react-yandex-maps";
import type { Problem } from "../types/problem";

const DEFAULT_CENTER: [number, number] = [55.751244, 37.618423]; // Moscow

interface MapViewProps {
  problems: Problem[];
  selectedId: string | null;
  onSelect(problem: Problem): void;
  onBoundsChange?(bbox: [number, number, number, number]): void;
}

export const MapView: React.FC<MapViewProps> = ({
  problems,
  selectedId,
  onSelect,
  onBoundsChange
}) => {
  const didEmitInitialBounds = useRef(false);

  const center = useMemo<[number, number]>(() => {
    if (problems.length === 0) {
      return DEFAULT_CENTER;
    }
    const first = problems[0];
    return [first.latitude, first.longitude];
  }, [problems]);

  const emitBounds = useCallback(
    (target: any): boolean => {
      if (!onBoundsChange) return;
      try {
        const b = target?.getBounds?.();
        if (!b || !b[0] || !b[1]) return;
        const sw = b[0]; // [lat1, lon1]
        const ne = b[1]; // [lat2, lon2]
        const minLat = Math.min(sw[0], ne[0]);
        const maxLat = Math.max(sw[0], ne[0]);
        const minLng = Math.min(sw[1], ne[1]);
        const maxLng = Math.max(sw[1], ne[1]);
        const bbox: [number, number, number, number] = [minLng, minLat, maxLng, maxLat];
        onBoundsChange(bbox);
        return true;
      } catch {
        // ignore map errors
      }
      return false;
    },
    [onBoundsChange]
  );

  const mapInstanceRef = useCallback(
    (map: any) => {
      // Emit initial bounds once to load markers without requiring user movement
      if (!map || !onBoundsChange || didEmitInitialBounds.current) return;

      let tries = 0;
      const tryEmit = () => {
        tries += 1;
        const ok = emitBounds(map);
        if (ok) {
          didEmitInitialBounds.current = true;
          return;
        }
        // Bounds are often null right after map init; retry for a bit.
        if (tries < 20) setTimeout(tryEmit, 150);
      };
      setTimeout(tryEmit, 0);
    },
    [emitBounds, onBoundsChange]
  );

  return (
    <YMaps
      query={{
        // TODO: заменить на реальный ключ, если нужен
        apikey: import.meta.env.VITE_YANDEX_MAPS_API_KEY
      }}
    >
      <Map
        state={{
          center,
          zoom: problems.length === 1 ? 14 : 10
        }}
        instanceRef={mapInstanceRef}
        width="100%"
        height="100%"
        onBoundsChange={(e: any) => {
          try {
            const target = e.get("target");
            emitBounds(target);
          } catch {
            // ignore map errors
          }
        }}
      >
        <ZoomControl />
        <FullscreenControl />
        {problems.map((p) => (
          <Placemark
            key={p.id}
            geometry={[p.latitude, p.longitude]}
            options={{
              preset:
                p.id === selectedId
                  ? "islands#redIcon"
                  : "islands#blueCircleDotIcon"
            }}
            properties={{
              balloonContentHeader: p.title,
              balloonContentBody: p.description ?? "",
              hintContent: p.title
            }}
            onClick={() => onSelect(p)}
          />
        ))}
      </Map>
    </YMaps>
  );
};


