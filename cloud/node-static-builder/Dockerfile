FROM node:lts-alpine as builder

RUN mkdir -p /app

WORKDIR /app

COPY function/package.json ./
COPY function/package*.json ./
RUN npm install

FROM builder AS application

# COPY function files and folders
COPY function/ ./

RUN npm run build

FROM nginx:stable

COPY nginx.conf /etc/nginx/nginx.conf

RUN mkdir -p /app

WORKDIR /app

COPY --from=application /app/build /app/build

CMD ["nginx"]