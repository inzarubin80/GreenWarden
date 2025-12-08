import React from "react";
import type { Problem } from "../types/problem";

interface ProblemCardProps {
  problem: Problem;
}

export const ProblemCard: React.FC<ProblemCardProps> = ({ problem }) => {
  return (
    <div className="problem-card">
      <h2 className="problem-title">{problem.title}</h2>
      {problem.status && (
        <div className="problem-status">
          Статус: <span>{problem.status}</span>
        </div>
      )}
      {problem.address && (
        <div className="problem-address">Адрес: {problem.address}</div>
      )}
      {problem.description && (
        <p className="problem-description">{problem.description}</p>
      )}
      <div className="problem-meta">
        <span>
          Координаты: {problem.latitude.toFixed(5)},{" "}
          {problem.longitude.toFixed(5)}
        </span>
        {problem.createdAt && (
          <span>
            Создано:{" "}
            {new Date(problem.createdAt).toLocaleString("ru-RU", {
              day: "2-digit",
              month: "2-digit",
              year: "numeric",
              hour: "2-digit",
              minute: "2-digit"
            })}
          </span>
        )}
      </div>
    </div>
  );
};


