import type { ActivistStat, AreaStats, BBox } from "../types/landing";

const MOCK_AREAS: AreaStats[] = [
  {
    // Пример области вокруг Москвы
    bbox: [37.3, 55.5, 37.9, 55.9],
    activists: [
      {
        id: "a1",
        name: "Иван Петров",
        createdCount: 12,
        resolvedCount: 5
      },
      {
        id: "a2",
        name: "Анна Смирнова",
        createdCount: 7,
        resolvedCount: 11
      }
    ]
  },
  {
    // Вторая область (для примера, чуть севернее)
    bbox: [37.3, 55.9, 37.9, 56.3],
    activists: [
      {
        id: "a3",
        name: "Сергей Иванов",
        createdCount: 4,
        resolvedCount: 9
      }
    ]
  }
];

function intersects(b1: BBox, b2: BBox): boolean {
  const [minLng1, minLat1, maxLng1, maxLat1] = b1;
  const [minLng2, minLat2, maxLng2, maxLat2] = b2;

  const lngOverlap = minLng1 <= maxLng2 && maxLng1 >= minLng2;
  const latOverlap = minLat1 <= maxLat2 && maxLat1 >= minLat2;

  return lngOverlap && latOverlap;
}

export function getActivistsByBboxMock(bbox: BBox): ActivistStat[] {
  const area = MOCK_AREAS.find((a) => intersects(a.bbox, bbox));
  return area?.activists ?? [];
}


