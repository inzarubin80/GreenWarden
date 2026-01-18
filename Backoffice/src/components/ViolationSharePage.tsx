import React, { useEffect, useMemo, useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import { getViolationById } from "../api/violations";
import { getViolationChat } from "../api/violationChat";
import type { ViolationDetails, ViolationRequest } from "../types/violation";
import type { ViolationChatMessage } from "../types/violationChat";
import { getRequestStatusLabel } from "../types/status";

type Participant = {
  userId: string;
  name: string;
  avatarUrl?: string;
  boostyUrl?: string;
};

type ParticipantBadges = {
  reported?: boolean;
  resolved?: boolean;
};

function formatDateTimeRu(value?: string) {
  if (!value) return "";
  const d = new Date(value);
  if (Number.isNaN(d.getTime())) return value;
  return d.toLocaleString("ru-RU", {
    day: "2-digit",
    month: "2-digit",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit"
  });
}

function initialsFromName(name: string) {
  const cleaned = name.trim();
  if (!cleaned) return "A";
  return cleaned
    .split(" ")
    .filter(Boolean)
    .slice(0, 2)
    .map((p) => p[0])
    .join("")
    .toUpperCase();
}

function requestAuthorName(r: ViolationRequest): string {
  return r.author_name && r.author_name.trim().length > 0
    ? r.author_name
    : `Пользователь #${r.created_by_user_id}`;
}

function chatAuthorName(m: ViolationChatMessage): string {
  return m.user_name && m.user_name.trim().length > 0
    ? m.user_name
    : `Пользователь #${m.user_id}`;
}

export const ViolationSharePage: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();

  const [details, setDetails] = useState<ViolationDetails | null>(null);
  const [chat, setChat] = useState<ViolationChatMessage[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!id) {
      setError("Неверная ссылка");
      setLoading(false);
      return;
    }
    let cancelled = false;
    const load = async () => {
      try {
        setLoading(true);
        setError(null);
        const [v, c] = await Promise.all([
          getViolationById(id),
          getViolationChat(id, 1, 200)
        ]);
        if (cancelled) return;
        setDetails(v);
        setChat(c.items ?? []);
      } catch (e) {
        if (cancelled) return;
        const msg = e instanceof Error ? e.message : "Не удалось загрузить";
        setError(msg.toLowerCase().includes("404") ? "Нарушение не найдено" : msg);
      } finally {
        if (!cancelled) setLoading(false);
      }
    };
    void load();
    return () => {
      cancelled = true;
    };
  }, [id]);

  const openRequest = useMemo(() => {
    if (!details) return null;
    return details.requests.find((r) => r.status === "open") ?? null;
  }, [details]);

  const resolutionRequest = useMemo(() => {
    if (!details) return null;
    const candidates = details.requests.filter(
      (r) => r.status === "closed" || r.status === "partially_closed"
    );
    if (candidates.length === 0) return null;
    return candidates
      .slice()
      .sort(
        (a, b) => new Date(a.created_at).getTime() - new Date(b.created_at).getTime()
      )
      .at(-1)!;
  }, [details]);

  const participantBadges = useMemo<Map<string, ParticipantBadges>>(() => {
    const m = new Map<string, ParticipantBadges>();
    if (openRequest) {
      m.set(openRequest.created_by_user_id, { ...(m.get(openRequest.created_by_user_id) ?? {}), reported: true });
    }
    if (resolutionRequest) {
      m.set(resolutionRequest.created_by_user_id, { ...(m.get(resolutionRequest.created_by_user_id) ?? {}), resolved: true });
    }
    return m;
  }, [openRequest, resolutionRequest]);

  const participants = useMemo<Participant[]>(() => {
    const byId = new Map<string, Participant>();

    const upsert = (p: Participant) => {
      const prev = byId.get(p.userId);
      if (!prev) {
        byId.set(p.userId, p);
        return;
      }
      byId.set(p.userId, {
        userId: p.userId,
        name: prev.name && !prev.name.startsWith("Пользователь #") ? prev.name : p.name,
        avatarUrl: prev.avatarUrl || p.avatarUrl,
        boostyUrl: prev.boostyUrl || p.boostyUrl
      });
    };

    if (details) {
      for (const r of details.requests) {
        upsert({
          userId: r.created_by_user_id,
          name: requestAuthorName(r),
          avatarUrl: r.author_avatar_url,
          boostyUrl: r.author_boosty_url
        });
      }
    }

    for (const m of chat) {
      upsert({
        userId: m.user_id,
        name: chatAuthorName(m),
        avatarUrl: m.user_avatar_url,
        boostyUrl: m.user_boosty_url
      });
    }

    return Array.from(byId.values()).sort((a, b) => {
      // donators first, then name
      const ad = a.boostyUrl ? 0 : 1;
      const bd = b.boostyUrl ? 0 : 1;
      if (ad !== bd) return ad - bd;
      return a.name.localeCompare(b.name, "ru");
    });
  }, [details, chat]);

  const title = useMemo(() => {
    if (!details) return "Нарушение";
    return details.description && details.description.trim().length > 0
      ? details.description
      : `Нарушение #${id}`;
  }, [details, id]);

  const openPhotos = openRequest?.photos ?? [];
  const openHeroPhoto = openPhotos[0];
  const openMorePhotos = openPhotos.slice(1);

  return (
    <div className="app-layout share-layout">
      <div className="map-column share-hero">
        <div className="share-hero-inner">
          <div className="share-topbar">
            <button type="button" className="button-ghost" onClick={() => navigate("/")}>
              На карту
            </button>
            <Link className="button-secondary" to="/about">
              О проекте
            </Link>
          </div>

          <h1 className="share-title">{title}</h1>
          {details && (
            <div className="share-meta">
              <div>
                Координаты: {details.lat.toFixed(5)}, {details.lng.toFixed(5)}
              </div>
              {openRequest && (
                <div>Зафиксировано: {formatDateTimeRu(openRequest.created_at)}</div>
              )}
            </div>
          )}

          <div className="share-hero-block">
            <div className="share-hero-block-title">Фото проблемы</div>
            {openHeroPhoto ? (
              <>
                <a
                  href={openHeroPhoto.url}
                  target="_blank"
                  rel="noreferrer"
                  className="share-photo share-photo--hero"
                >
                  {/* eslint-disable-next-line jsx-a11y/img-redundant-alt */}
                  <img
                    src={openHeroPhoto.thumb_url || openHeroPhoto.url}
                    alt="Фото проблемы"
                  />
                </a>
                {openMorePhotos.length > 0 && (
                  <div className="share-photos share-photos--strip">
                    {openMorePhotos.map((p) => (
                      <a
                        key={p.id || p.url}
                        href={p.url}
                        target="_blank"
                        rel="noreferrer"
                        className="share-photo"
                      >
                        {/* eslint-disable-next-line jsx-a11y/img-redundant-alt */}
                        <img src={p.thumb_url || p.url} alt="Фото проблемы" />
                      </a>
                    ))}
                  </div>
                )}
              </>
            ) : (
              <div className="status">Фото проблемы пока нет</div>
            )}
          </div>
        </div>
      </div>

      <div className="card-column share-column">
        {loading && <div className="status">Загрузка...</div>}
        {error && <div className="status error">{error}</div>}

        {!loading && !error && details && (
          <>
            <div className="share-section">
              <div className="share-section-title">Участники</div>
              <div className="participants-grid">
                {participants.map((p) => (
                  <div key={p.userId} className="participant-card">
                    <Link
                      className="participant-main"
                      to={`/user/${p.userId}`}
                      state={{
                        name: p.name,
                        avatarUrl: p.avatarUrl,
                        boostyUrl: p.boostyUrl
                      }}
                    >
                      <div className="participant-avatar">
                        {p.avatarUrl ? (
                          // eslint-disable-next-line jsx-a11y/img-redundant-alt
                          <img src={p.avatarUrl} alt="Аватар" />
                        ) : (
                          <span>{initialsFromName(p.name)}</span>
                        )}
                      </div>
                      <div className="participant-info">
                        <div className="participant-name">{p.name}</div>
                        <div className="participant-sub-row">
                          <span className="participant-sub">ID: {p.userId}</span>
                          {(() => {
                            const b = participantBadges.get(p.userId);
                            if (!b) return null;
                            return (
                              <span className="participant-badges">
                                {b.reported && <span className="badge badge--report">Зафиксировал</span>}
                                {b.resolved && <span className="badge badge--resolve">Решил</span>}
                              </span>
                            );
                          })()}
                        </div>
                      </div>
                    </Link>
                    {p.boostyUrl ? (
                      <a
                        className="button-primary participant-donate"
                        href={p.boostyUrl}
                        target="_blank"
                        rel="noreferrer"
                      >
                        Поддержать
                      </a>
                    ) : (
                      <div className="participant-donate muted">Донат не указан</div>
                    )}
                  </div>
                ))}
              </div>
            </div>

            <div className="share-section">
              <div className="share-section-title">Таймлайн</div>

              <div className="timeline-card">
                <div className="timeline-header">
                  <div className="timeline-badge">Зафиксировал</div>
                  <div className="timeline-date">
                    {openRequest ? formatDateTimeRu(openRequest.created_at) : "—"}
                  </div>
                </div>
                {openRequest ? (
                  <>
                    <div className="timeline-user">
                      <Link
                        className="timeline-user-link"
                        to={`/user/${openRequest.created_by_user_id}`}
                        state={{
                          name: requestAuthorName(openRequest),
                          avatarUrl: openRequest.author_avatar_url,
                          boostyUrl: openRequest.author_boosty_url
                        }}
                      >
                        {requestAuthorName(openRequest)}
                      </Link>
                    </div>
                    {openRequest.comment && (
                      <div className="timeline-text">{openRequest.comment}</div>
                    )}
                  </>
                ) : (
                  <div className="status">Данных о фиксации нет</div>
                )}
              </div>

              <div className="timeline-card">
                <div className="timeline-header">
                  {(() => {
                    const label = resolutionRequest
                      ? getRequestStatusLabel(resolutionRequest.status)
                      : "Не решено";
                    const cls = resolutionRequest
                      ? resolutionRequest.status === "partially_closed"
                        ? "timeline-badge timeline-badge--partial"
                        : "timeline-badge"
                      : "timeline-badge timeline-badge--neutral";
                    return <div className={cls}>{label}</div>;
                  })()}
                  <div className="timeline-date">
                    {resolutionRequest
                      ? formatDateTimeRu(resolutionRequest.created_at)
                      : "—"}
                  </div>
                </div>
                {resolutionRequest ? (
                  <>
                    <div className="timeline-user">
                      <Link
                        className="timeline-user-link"
                        to={`/user/${resolutionRequest.created_by_user_id}`}
                        state={{
                          name: requestAuthorName(resolutionRequest),
                          avatarUrl: resolutionRequest.author_avatar_url,
                          boostyUrl: resolutionRequest.author_boosty_url
                        }}
                      >
                        {requestAuthorName(resolutionRequest)}
                      </Link>
                    </div>
                    {resolutionRequest.comment && (
                      <div className="timeline-text">{resolutionRequest.comment}</div>
                    )}
                    {resolutionRequest.photos?.length ? (
                      <div className="share-photos share-photos--small">
                        {resolutionRequest.photos.map((p) => (
                          <a
                            key={p.id || p.url}
                            href={p.url}
                            target="_blank"
                            rel="noreferrer"
                            className="share-photo"
                          >
                            {/* eslint-disable-next-line jsx-a11y/img-redundant-alt */}
                            <img src={p.thumb_url || p.url} alt="Фото решения" />
                          </a>
                        ))}
                      </div>
                    ) : (
                      <div className="status">Фото решения не приложено</div>
                    )}
                  </>
                ) : (
                  <div className="status">Пока никто не отметил решение</div>
                )}
              </div>
            </div>

            <div className="share-section">
              <div className="share-section-title">Чат</div>
              {chat.length === 0 ? (
                <div className="status">Сообщений пока нет</div>
              ) : (
                <div className="chat-list">
                  {chat.map((m) => (
                    <div
                      key={m.id}
                      className={`chat-message${m.is_system ? " chat-message--system" : ""}`}
                    >
                      <div className="chat-avatar">
                        {m.user_avatar_url ? (
                          // eslint-disable-next-line jsx-a11y/img-redundant-alt
                          <img src={m.user_avatar_url} alt="Аватар" />
                        ) : (
                          <span>{initialsFromName(chatAuthorName(m))}</span>
                        )}
                      </div>
                      <div className="chat-body">
                        <div className="chat-head">
                          <Link
                            className="chat-user"
                            to={`/user/${m.user_id}`}
                            state={{
                              name: chatAuthorName(m),
                              avatarUrl: m.user_avatar_url,
                              boostyUrl: m.user_boosty_url
                            }}
                          >
                            {chatAuthorName(m)}
                          </Link>
                          <span className="chat-date">{formatDateTimeRu(m.created_at)}</span>
                        </div>
                        <div className="chat-text">{m.text}</div>
                        {!m.is_system && m.user_boosty_url && (
                          <a
                            className="chat-donate"
                            href={m.user_boosty_url}
                            target="_blank"
                            rel="noreferrer"
                          >
                            Поддержать автора
                          </a>
                        )}
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </>
        )}
      </div>
    </div>
  );
};

