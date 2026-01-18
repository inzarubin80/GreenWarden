import React, { useCallback, useEffect, useMemo, useState } from "react";
import { Routes, Route, useParams, useNavigate, Link } from "react-router-dom";
import { MapView } from "./components/MapView";
import { ProblemCard } from "./components/ProblemCard";
import { LandingPage } from "./components/LandingPage";
import { UserProfilePage } from "./components/UserProfilePage";
import { ViolationSharePage } from "./components/ViolationSharePage";
import type { Problem } from "./types/problem";
import { getViolationsByBbox, getViolationById } from "./api/violations";
import type { Violation, ViolationDetails } from "./types/violation";
import { getStatusLabel, getRequestStatusLabel } from "./types/status";
import { GOOGLE_PLAY_URL } from "./config/mobile";

const violationToProblem = (v: Violation): Problem => ({
  id: String(v.id),
  title: v.description || `Нарушение #${v.id}`,
  description: v.description,
  status: v.status,
  latitude: v.lat,
  longitude: v.lng,
  address: v.address,
  createdAt: v.created_at
});

interface MapScreenProps {
  violationId?: string;
}

const MapScreen: React.FC<MapScreenProps> = ({ violationId }) => {
  const [problems, setProblems] = useState<Problem[]>([]);
  const [selected, setSelected] = useState<Problem | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [currentBbox, setCurrentBbox] = useState<
    [number, number, number, number] | null
  >(null);
  const [statusFilter, setStatusFilter] = useState<string>("all");
  const [details, setDetails] = useState<ViolationDetails | null>(null);
  const [detailsLoading, setDetailsLoading] = useState(false);
  const [detailsError, setDetailsError] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState<"overview" | "requests" | "photos">(
    "overview"
  );

  const navigate = useNavigate();
  const isDetailMode = useMemo(() => !!violationId, [violationId]);

  const availableStatuses = useMemo(() => {
    const set = new Set<string>();
    for (const p of problems) {
      if (p.status) {
        set.add(p.status);
      }
    }
    return Array.from(set).sort();
  }, [problems]);

  const filteredProblems = useMemo(() => {
    if (statusFilter === "all") return problems;
    return problems.filter((p) => p.status === statusFilter);
  }, [problems, statusFilter]);

  const ensureDetailsLoaded = useCallback(
    async (id: string) => {
      if (!id) return;
      try {
        setDetailsLoading(true);
        setDetailsError(null);
        const d = await getViolationById(id);
        setDetails(d);
      } catch (e) {
        setDetailsError(
          e instanceof Error ? e.message : "Не удалось загрузить детали"
        );
      } finally {
        setDetailsLoading(false);
      }
    },
    []
  );

  const loadViolations = useCallback(
    async (bbox: [number, number, number, number]) => {
      try {
        setLoading(true);
        setError(null);
        const resp = await getViolationsByBbox(bbox);
        const items = resp.items ?? [];
        const mapped = items.map(violationToProblem);
        setProblems(mapped);
      } catch (e) {
        setError(
          e instanceof Error ? e.message : "Не удалось загрузить нарушения"
        );
      } finally {
        setLoading(false);
      }
    },
    []
  );

  const handleBoundsChange = useCallback(
    (bbox: [number, number, number, number]) => {
      setCurrentBbox(bbox);
      void loadViolations(bbox);
    },
    [loadViolations]
  );

  const handleSelect = (problem: Problem) => {
    setSelected(problem);
    setActiveTab("overview");
    setDetails(null);
    setDetailsError(null);
  };

  useEffect(() => {
    if (filteredProblems.length === 0) {
      setSelected(null);
      return;
    }

    setSelected((prev) => {
      if (!prev) {
        return filteredProblems[0];
      }
      const found = filteredProblems.find((p) => p.id === prev.id);
      return found ?? filteredProblems[0];
    });
  }, [filteredProblems]);

  useEffect(() => {
    if (!violationId) {
      return;
    }

    const loadById = async () => {
      try {
        setLoading(true);
        setError(null);
        const d = await getViolationById(violationId);
        setDetails(d);
        const p: Problem = {
          id: violationId,
          title: d.description || `Нарушение #${violationId}`,
          description: d.description,
          status: undefined,
          latitude: d.lat,
          longitude: d.lng,
          address: undefined,
          createdAt: undefined
        };
        setProblems([p]);
        setSelected(p);
      } catch (e) {
        const msg =
          e instanceof Error ? e.message : "Не удалось загрузить нарушение";
        if (msg.toLowerCase().includes("not found")) {
          setError("Нарушение не найдено");
        } else {
          setError(msg);
        }
        setProblems([]);
        setSelected(null);
      } finally {
        setLoading(false);
      }
    };

    void loadById();
  }, [violationId]);

  return (
    <div className="app-layout">
      <div className="map-column">
        <MapView
          problems={filteredProblems}
          selectedId={selected?.id ?? null}
          onSelect={handleSelect}
          onBoundsChange={handleBoundsChange}
        />
      </div>
      <div className="card-column">
        <div className="card-main">
          <div className="filters-row">
            <Link className="button-secondary" to="/about">
              О проекте и приложении
            </Link>
          </div>
          {!isDetailMode && (
            <div className="filters-row">
              <label className="filter-label">
                Статус:
                <select
                  className="filter-select"
                  value={statusFilter}
                  onChange={(e) => setStatusFilter(e.target.value)}
                >
                  <option value="all">Все</option>
                  {availableStatuses.map((st) => (
                    <option key={st} value={st}>
                      {getStatusLabel(st)}
                    </option>
                  ))}
                </select>
              </label>
            </div>
          )}
          {isDetailMode && (
            <button
              type="button"
              className="status"
              onClick={() => navigate("/")}
            >
              На главную карту
            </button>
          )}
          {loading && <div className="status">Загрузка нарушений...</div>}
          {error && <div className="status error">{error}</div>}
          {!loading && !error && selected && (
            <>
              <div className="tabs">
                <button
                  type="button"
                  className={`tab-button${
                    activeTab === "overview" ? " tab-button--active" : ""
                  }`}
                  onClick={() => setActiveTab("overview")}
                >
                  Обзор
                </button>
                <button
                  type="button"
                  className={`tab-button${
                    activeTab === "requests" ? " tab-button--active" : ""
                  }`}
                  onClick={() => {
                    setActiveTab("requests");
                    if (!details) {
                      void ensureDetailsLoaded(selected.id);
                    }
                  }}
                >
                  Действия
                  {details && details.requests.length > 0
                    ? ` (${details.requests.length})`
                    : ""}
                </button>
                <button
                  type="button"
                  className={`tab-button${
                    activeTab === "photos" ? " tab-button--active" : ""
                  }`}
                  onClick={() => {
                    setActiveTab("photos");
                    if (!details) {
                      void ensureDetailsLoaded(selected.id);
                    }
                  }}
                >
                  Фото
                  {details && details.photos.length > 0
                    ? ` (${details.photos.length})`
                    : ""}
                </button>
              </div>
              {activeTab === "overview" && <ProblemCard problem={selected} />}
              {activeTab === "requests" && (
                <div className="tab-content">
                  {detailsLoading && (
                    <div className="status">Загрузка действий...</div>
                  )}
                  {detailsError && (
                    <div className="status error">{detailsError}</div>
                  )}
                  {!detailsLoading &&
                    !detailsError &&
                    details &&
                    details.requests.length === 0 && (
                      <div className="status">Действий пока нет</div>
                    )}
                  {!detailsLoading &&
                    !detailsError &&
                    details &&
                    details.requests.length > 0 && (
                      <div className="requests-list">
                        {details.requests.map((r) => (
                          <div key={r.id} className="request-card">
                            <div className="request-user">
                              <div className="request-user-avatar">
                                {r.author_avatar_url ? (
                                  // eslint-disable-next-line jsx-a11y/img-redundant-alt
                                  <img
                                    src={r.author_avatar_url}
                                    alt="Аватар пользователя"
                                  />
                                ) : (
                                  <span>
                                    {String(r.created_by_user_id)
                                      .split(" ")
                                      .map((p) => p[0])
                                      .join("")
                                      .toUpperCase() || "A"}
                                  </span>
                                )}
                              </div>
                              {/*
                                Пока сервер не отдаёт имя, показываем техническое имя
                                «Пользователь #ID». Когда появится author_name,
                                можно будет подставить его вместо этого текста.
                              */}
                              <Link
                                className="request-user-name"
                                to={`/user/${r.created_by_user_id}`}
                                state={{
                                  name: `Пользователь #${r.created_by_user_id}`,
                                  avatarUrl: r.author_avatar_url,
                                  boostyUrl: r.author_boosty_url
                                }}
                              >
                                {`Пользователь #${r.created_by_user_id}`}
                              </Link>
                            </div>
                            <div className="request-header">
                              <span className="request-status">
                                {getRequestStatusLabel(r.status)}
                              </span>
                              <span className="request-date">
                                {new Date(
                                  r.created_at
                                ).toLocaleString("ru-RU", {
                                  day: "2-digit",
                                  month: "2-digit",
                                  year: "numeric",
                                  hour: "2-digit",
                                  minute: "2-digit"
                                })}
                              </span>
                            </div>
                            {r.comment && (
                              <div className="request-comment">
                                {r.comment}
                              </div>
                            )}
                            <div className="request-footer">
                              <span>
                                Фото: {r.photos?.length ?? 0} · Лайки:{" "}
                                {r.likes} / Дизлайки: {r.dislikes}
                              </span>
                            </div>
                          </div>
                        ))}
                      </div>
                    )}
                </div>
              )}
              {activeTab === "photos" && (
                <div className="tab-content">
                  {detailsLoading && (
                    <div className="status">Загрузка фото...</div>
                  )}
                  {detailsError && (
                    <div className="status error">{detailsError}</div>
                  )}
                  {!detailsLoading &&
                    !detailsError &&
                    details &&
                    details.photos.length === 0 && (
                      <div className="status">Фото пока нет</div>
                    )}
                  {!detailsLoading &&
                    !detailsError &&
                    details &&
                    details.photos.length > 0 && (
                      <div className="photos-grid">
                        {details.photos.map((p) => (
                          <a
                            key={p.id || p.url}
                            href={p.url}
                            target="_blank"
                            rel="noreferrer"
                            className="photo-thumb"
                          >
                            {/* eslint-disable-next-line jsx-a11y/img-redundant-alt */}
                            <img
                              src={p.thumb_url || p.url}
                              alt="Фото нарушения"
                            />
                          </a>
                        ))}
                      </div>
                    )}
                </div>
              )}
            </>
          )}
          {!loading && !error && !selected && (
            <div className="status">
              {isDetailMode ? "Нарушение не найдено" : "Проблемы не найдены"}
            </div>
          )}
        </div>
        <div className="landing-section">
          <h2 className="landing-subtitle">Мобильное приложение</h2>
          <p className="landing-text">
            С помощью мобильного приложения жители и активисты могут фиксировать
            проблемы на местности, а также отмечать их решение.
          </p>
          <div className="landing-buttons">
            <a
              className="button-secondary"
              href={GOOGLE_PLAY_URL}
              target="_blank"
              rel="noreferrer"
            >
              Скачать в Google Play
            </a>
          </div>
        </div>
      </div>
    </div>
  );
};

const ViolationRoute: React.FC = () => {
  const { id } = useParams<{ id: string }>();

  return <MapScreen violationId={id} />;
};

const IndexRoute: React.FC = () => {
  return <MapScreen />;
};

const NotFoundRoute: React.FC = () => {
  const navigate = useNavigate();

  return (
    <div className="app-layout">
      <div className="map-column" />
      <div className="card-column">
        <div className="status error">Страница не найдена</div>
        <button
          type="button"
          className="status"
          onClick={() => navigate("/")}
        >
          На карту
        </button>
      </div>
    </div>
  );
};

export default function App() {
  return (
    <Routes>
      <Route path="/" element={<IndexRoute />} />
      {/* Backward compatibility + share links */}
      <Route path="/violation/:id" element={<ViolationSharePage />} />
      <Route path="/violations/:id" element={<ViolationSharePage />} />
      <Route path="/about" element={<LandingPage />} />
      <Route path="/user/:id" element={<UserProfilePage />} />
      <Route path="*" element={<NotFoundRoute />} />
    </Routes>
  );
}

