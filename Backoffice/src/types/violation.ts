export interface ViolationPhoto {
  url?: string;
  thumb_url?: string;
}

export interface Violation {
  id: string;
  description?: string;
  lat: number;
  lng: number;
  status?: string;
  address?: string;
  created_at?: string;
  photos?: ViolationPhoto[];
}

export interface Paged<T> {
  items: T[];
  page?: number;
  page_size?: number;
  total?: number;
}


