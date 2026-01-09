# Build Stage
FROM node:lts-alpine AS build

WORKDIR /app

COPY package*.json ./

RUN npm install

COPY . .

RUN chmod -R +x node_modules/.bin

RUN npm run build

# Production Stage
FROM nginx:latest AS production

# Копируем кастомную конфигурацию Nginx
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Копируем собранные файлы (Vite собирает в dist, а не build)
COPY --from=build /app/dist /usr/share/nginx/html

# Открываем порт 3000
EXPOSE 3000

# Запускаем Nginx
CMD ["nginx", "-g", "daemon off;"]

