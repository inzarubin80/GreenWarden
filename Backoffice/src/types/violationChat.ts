export interface ViolationChatMessage {
  id: string;
  violation_id: string;
  user_id: string;
  user_name?: string;
  user_avatar_url?: string;
  user_boosty_url?: string;
  text: string;
  is_system: boolean;
  created_at: string;
  updated_at?: string | null;
}

export interface PaginatedViolationChatMessages {
  items: ViolationChatMessage[];
  page: number;
  page_size: number;
  total: number;
}

