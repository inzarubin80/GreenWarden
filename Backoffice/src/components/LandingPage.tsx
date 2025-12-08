import React, { useCallback, useState } from "react";
import {
  YMaps,
  Map,
  ZoomControl,
  FullscreenControl,
  Rectangle
} from "@pbe/react-yandex-maps";
import { Link } from "react-router-dom";
import type { BBox, ActivistStat } from "../types/landing";
import { getActivistsByBboxMock } from "../mock/landingData";

const DEFAULT_CENTER: [number, number] = [55.751244, 37.618423]; // Moscow

const initialBBox: BBox = [37.3, 55.5, 37.9, 55.9];

export const LandingPage: React.FC = () => {
  const [currentBBox, setCurrentBBox] = useState<BBox | null>(initialBBox);
  const [activists, setActivists] = useState<ActivistStat[]>(
    getActivistsByBboxMock(initialBBox)
  );

  const handleBoundsChange = useCallback((bbox: BBox) => {
    setCurrentBBox(bbox);
    const data = getActivistsByBboxMock(bbox);
    setActivists(data);
  }, []);

  return (
    <div className="app-layout">
      <div className="map-column">
        <YMaps
          query={{
            apikey: import.meta.env.VITE_YANDEX_MAPS_API_KEY
          }}
        >
          <Map
            defaultState={{
              center: DEFAULT_CENTER,
              zoom: 10
            }}
            width="100%"
            height="100%"
            onBoundsChange={(e: any) => {
              try {
                const target = e.get("target");
                const b = target?.getBounds?.();
                if (!b || !b[0] || !b[1]) return;
                const sw = b[0]; // [lat1, lon1]
                const ne = b[1]; // [lat2, lon2]
                const minLat = Math.min(sw[0], ne[0]);
                const maxLat = Math.max(sw[0], ne[0]);
                const minLng = Math.min(sw[1], ne[1]);
                const maxLng = Math.max(sw[1], ne[1]);
                const bbox: BBox = [minLng, minLat, maxLng, maxLat];
                handleBoundsChange(bbox);
              } catch {
                // ignore map errors
              }
            }}
          >
            <ZoomControl />
            <FullscreenControl />
            {currentBBox && (
              <Rectangle
                geometry={[
                  [currentBBox[1], currentBBox[0]],
                  [currentBBox[3], currentBBox[2]]
                ]}
                options={{
                  fillColor: "rgba(34, 197, 94, 0.1)",
                  strokeColor: "rgba(34, 197, 94, 0.9)",
                  strokeWidth: 2
                }}
              />
            )}
          </Map>
        </YMaps>
      </div>
      <div className="card-column">
        <div className="landing-section">
          <Link className="button-secondary" to="/">
            На карту
          </Link>
        </div>
        <div className="landing-section">
          <h2 className="landing-subtitle">Как это работает</h2>
          <ol className="landing-steps">
            <li>Откройте карту и найдите проблемный участок.</li>
            <li>Создайте проблему через мобильное приложение.</li>
            <li>Координаторы распределяют задачи между активистами.</li>
            <li>После решения проблемы отмечайте её статус как решённую.</li>
          </ol>
        </div>

        <div className="landing-section">
          <h2 className="landing-subtitle">Активисты в выбранной области</h2>
          {activists.length === 0 ? (
            <div className="status">
              В текущей области пока нет активностей активистов.
            </div>
          ) : (
            <div className="activists-list">
              {activists.map((a) => (
                <div key={a.id} className="activist-card">
                  <div className="activist-main">
                    <div className="activist-avatar">
                      {a.name.charAt(0).toUpperCase()}
                    </div>
                    <div>
                      <div className="activist-name">{a.name}</div>
                      <div className="activist-meta">
                        Создал проблем: {a.createdCount} · Решил проблем:{" "}
                        {a.resolvedCount}
                      </div>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
};


