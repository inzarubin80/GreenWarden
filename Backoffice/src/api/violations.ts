import type { Paged, Violation, ViolationDetails } from "../types/violation";
import { apiClient } from "./client";

export async function getViolationsByBbox(
  bbox: [number, number, number, number]
): Promise<Paged<Violation>> {
  const qs = `bbox=${bbox.join(",")}`;
  return apiClient.get<Paged<Violation>>(`/violations?${qs}`);
}

export async function getViolationById(id: string): Promise<ViolationDetails> {
  return apiClient.get<ViolationDetails>(
    `/violations/${encodeURIComponent(id)}`
  );
}


