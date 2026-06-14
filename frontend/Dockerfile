# syntax=docker/dockerfile:1.7
FROM node:24-alpine AS builder

WORKDIR /src

COPY package.json package-lock.json ./
RUN --mount=type=cache,target=/root/.npm npm ci

COPY . .
RUN npm run build-only

FROM nginxinc/nginx-unprivileged:1.28-alpine

COPY --chown=101:101 nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=builder --chown=101:101 /src/dist /usr/share/nginx/html

EXPOSE 8080

USER 101:101

CMD ["nginx", "-g", "daemon off;"]
