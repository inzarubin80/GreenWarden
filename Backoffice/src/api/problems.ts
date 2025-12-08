import { apiClient } from "./client";
import type { Problem } from "../types/problem";

export async function fetchProblems(): Promise<Problem[]> {
  // Adjust the path according to your Go API routing, e.g. /problems
  return apiClient.get<Problem[]>("/problems");
}


