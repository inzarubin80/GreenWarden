export interface ViolationPhoto {
  id?: string;
  violation_id?: string;
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

export interface ViolationRequest {
  id: string;
  status: string;
  created_by_user_id: string;
  author_name?: string;
  comment?: string;
  created_at: string;
  photos: ViolationPhoto[];
  likes: number;
  dislikes: number;
  user_vote: string;
  author_boosty_url?: string;
  author_avatar_url?: string;
}

export interface ViolationDetails {
  user_id: string;
  description?: string;
  lat: number;
  lng: number;
  photos: ViolationPhoto[];
  requests: ViolationRequest[];
}

export interface Paged<T> {
  items: T[];
  page?: number;
  page_size?: number;
  total?: number;
}

