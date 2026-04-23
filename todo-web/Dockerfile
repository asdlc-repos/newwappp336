# ---- Build stage ----
FROM node:20-alpine AS build
WORKDIR /app

# Install dependencies first to leverage layer caching.
COPY package.json package-lock.json* ./
RUN if [ -f package-lock.json ]; then npm ci; else npm install; fi

# Copy the rest of the source and build.
COPY . .
RUN npm run build

# ---- Runtime stage ----
FROM nginx:alpine
WORKDIR /usr/share/nginx/html

# Replace nginx's default config with our custom one (port 8080, SPA fallback, /health).
COPY nginx.conf /etc/nginx/nginx.conf

# Copy built static assets from the build stage.
RUN rm -rf ./*
COPY --from=build /app/dist/ ./

# Entrypoint substitutes runtime VITE_API_URL into env.js before starting nginx.
COPY docker-entrypoint.sh /docker-entrypoint.sh
RUN chmod +x /docker-entrypoint.sh

EXPOSE 8080

ENTRYPOINT ["/docker-entrypoint.sh"]
