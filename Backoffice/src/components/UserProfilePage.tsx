import React from "react";
import { useLocation, useNavigate, useParams } from "react-router-dom";

interface UserProfileState {
  name?: string;
  avatarUrl?: string;
  boostyUrl?: string;
}

export const UserProfilePage: React.FC = () => {
  const navigate = useNavigate();
  const { id } = useParams<{ id: string }>();
  const location = useLocation();
  const state = (location.state as UserProfileState | null) ?? {};

  const displayName =
    state.name && state.name.trim().length > 0
      ? state.name
      : id
        ? `Активист #${id}`
        : "Активист";

  const initials =
    displayName
      .replace(/^Активист #/, "")
      .trim()
      .split(" ")
      .map((p) => p[0])
      .join("")
      .toUpperCase() || "A";

  return (
    <div className="app-layout">
      <div className="map-column" />
      <div className="card-column">
        <div className="landing-section">
          <button
            type="button"
            className="status"
            onClick={() => navigate(-1)}
          >
            Назад
          </button>
        </div>
        <div className="landing-section user-profile">
          <div className="user-profile-header">
            <div className="user-profile-avatar">
              {state.avatarUrl ? (
                // eslint-disable-next-line jsx-a11y/img-redundant-alt
                <img src={state.avatarUrl} alt="Аватар активиста" />
              ) : (
                <span>{initials}</span>
              )}
            </div>
            <div className="user-profile-main">
              <div className="user-profile-name">{displayName}</div>
              {id && (
                <div className="user-profile-id">ID пользователя: {id}</div>
              )}
            </div>
          </div>
          <div className="user-profile-body">
            <p>
              Это страница активиста, который создаёт и закрывает экологические
              проблемы на карте.
            </p>
            {state.boostyUrl ? (
              <a
                className="button-primary"
                href={state.boostyUrl}
                target="_blank"
                rel="noreferrer"
              >
                Поддержать активиста
              </a>
            ) : (
              <div className="status">
                Ссылка на донат ещё не настроена для этого пользователя.
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
};


