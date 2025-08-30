# Multi-stage: build web and then Nginx static
FROM node:24-alpine AS webbuild
WORKDIR /app
COPY --from=web /package*.json  ./
RUN npm install
COPY --from=web . .
RUN npm run build

FROM nginx:alpine
COPY nginx.conf /etc/nginx/nginx.conf
COPY --from=webbuild /app/dist /usr/share/nginx/html/app
