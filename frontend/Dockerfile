FROM node:18-alpine AS base
WORKDIR /app
COPY package*.json ./

RUN npm install -g pnpm

WORKDIR /app

COPY package.json pnpm-lock.yaml ./

RUN pnpm install --frozen-lockfile

FROM base AS dev
COPY . .

EXPOSE 4321
CMD [ "pnpm", "run", "dev", "--", "--host" ]


FROM base AS prod-builder
COPY . .
RUN pnpm run build


FROM nginx:1.25-alpine
COPY --from=prod-builder /app/dist /usr/share/nginx/html
EXPOSE 80