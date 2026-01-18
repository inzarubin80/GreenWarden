import { apiClient } from "./client";
import type { PaginatedViolationChatMessages } from "../types/violationChat";

export async function getViolationChat(
  id: string,
  page = 1,
  pageSize = 200
): Promise<PaginatedViolationChatMessages> {
  const qs = `page=${encodeURIComponent(String(page))}&page_size=${encodeURIComponent(
    String(pageSize)
  )}`;
  return apiClient.get<PaginatedViolationChatMessages>(
    `/violations/${encodeURIComponent(id)}/chat?${qs}`
  );
}

