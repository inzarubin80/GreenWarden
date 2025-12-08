export type BBox = [number, number, number, number]; // [minLng, minLat, maxLng, maxLat]

export interface ActivistStat {
  id: string;
  name: string;
  avatarUrl?: string;
  createdCount: number;
  resolvedCount: number;
}

export interface AreaStats {
  bbox: BBox;
  activists: ActivistStat[];
}


