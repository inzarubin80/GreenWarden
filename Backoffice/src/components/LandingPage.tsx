import React from "react";
import { Link } from "react-router-dom";

import shot1 from "../assets/onboarding-shot-1.svg";
import shot2 from "../assets/onboarding-shot-2.svg";
import shot3 from "../assets/onboarding-shot-3.svg";
import { GOOGLE_PLAY_URL } from "../config/mobile";

export const LandingPage: React.FC = () => {
  return (
    <div className="about-page">
      <header className="about-hero">
        <div className="about-hero-top">
          <Link className="button-ghost" to="/">
            На карту
          </Link>
          <span className="about-pill">О проекте и приложении</span>
        </div>

        <div className="about-hero-main">
          <div className="about-hero-copy">
            <h1 className="about-title">GreenWarden</h1>
            <p className="about-subtitle">
              Фиксируйте экологические проблемы на местности, следите за их
              решением и поддерживайте активистов.
            </p>
            <div className="about-cta">
              <a className="button-primary" href={GOOGLE_PLAY_URL} target="_blank" rel="noreferrer">
                Скачать Android
              </a>
              <a className="button-secondary" href="#" onClick={(e) => e.preventDefault()}>
                Скачать iOS (скоро)
              </a>
            </div>
            <div className="about-note">
              Эта страница — онбординг. Скриншоты ниже временные и будут заменены
              на реальные из мобильного приложения.
            </div>
          </div>

          <div className="about-hero-shots">
            <img className="about-shot about-shot--big" src={shot1} alt="Скриншот: карта" />
            <img className="about-shot" src={shot2} alt="Скриншот: создание проблемы" />
            <img className="about-shot" src={shot3} alt="Скриншот: чат" />
          </div>
        </div>
      </header>

      <main className="about-content">
        <section className="about-section">
          <h2 className="about-section-title">Как это работает</h2>
          <ol className="about-steps">
            <li>Установите мобильное приложение GreenWarden.</li>
            <li>Откройте карту и найдите проблемный участок.</li>
            <li>Создайте проблему с описанием и фотографиями.</li>
            <li>Когда проблема решена — отметьте решение и приложите фото.</li>
            <li>Поддержите активиста и поделитесь проблемой.</li>
          </ol>
        </section>

        <section className="about-section">
          <h2 className="about-section-title">Почему это важно</h2>
          <div className="about-cards">
            <div className="about-card">
              <div className="about-card-title">Прозрачность</div>
              <div className="about-card-text">
                У каждой проблемы есть история: кто зафиксировал, что сделали,
                и как дошли до решения.
              </div>
            </div>
            <div className="about-card">
              <div className="about-card-title">Сообщество</div>
              <div className="about-card-text">
                Можно делиться ссылкой на проблему и подключать друзей, соседей
                и волонтеров.
              </div>
            </div>
            <div className="about-card">
              <div className="about-card-title">Поддержка</div>
              <div className="about-card-text">
                Акцент на участниках: у активистов есть донат‑ссылки — можно
                поддержать тех, кто помогает городу.
              </div>
            </div>
          </div>
        </section>

        <section className="about-section about-section--apps">
          <div className="about-app">
            <div className="about-app-head">
              <h3 className="about-app-title">Android</h3>
              <a className="button-primary" href={GOOGLE_PLAY_URL} target="_blank" rel="noreferrer">
                Скачать
              </a>
            </div>
            <div className="about-screens">
              <img className="about-shot" src={shot1} alt="Android: экран 1" />
              <img className="about-shot" src={shot2} alt="Android: экран 2" />
              <img className="about-shot" src={shot3} alt="Android: экран 3" />
            </div>
          </div>

          <div className="about-app">
            <div className="about-app-head">
              <h3 className="about-app-title">iOS</h3>
              <a className="button-secondary" href="#" onClick={(e) => e.preventDefault()}>
                Скоро
              </a>
            </div>
            <div className="about-screens">
              <img className="about-shot" src={shot1} alt="iOS: экран 1" />
              <img className="about-shot" src={shot2} alt="iOS: экран 2" />
              <img className="about-shot" src={shot3} alt="iOS: экран 3" />
            </div>
          </div>
        </section>
      </main>
    </div>
  );
};


