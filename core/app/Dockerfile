FROM node:24-alpine AS builder

WORKDIR /app

COPY package*.json ./

RUN npm cache clean --force
RUN rm -f package-lock.json
RUN npm install

COPY . .

RUN npm run build

FROM nginx:alpine

COPY --from=builder /app/dist /usr/share/nginx/html

COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 3000

CMD ["nginx", "-g", "daemon off;"] 