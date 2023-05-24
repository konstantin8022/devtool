FROM node:12.13.1-alpine3.10
ENV PACKAGES nodejs-npm nginx
ENV UPDATED_PACKAGES sqlite openssl apk-tools

RUN apk --update add --no-cache $PACKAGES && \
    ln -sf /dev/stdout /var/log/nginx/access.log && \
    ln -sf /dev/stderr /var/log/nginx/error.log && \
    apk add --no-cache -u $UPDATED_PACKAGES && \
    rm -rf /var/cache/apk && \
    :

WORKDIR /app
ENV HOME /app
ENV LANG C.UTF-8

COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY . .

RUN set -ex && \
    npm install && \
    npm cache clean --force && \
    npm run build && \
    :

EXPOSE 8080

ENTRYPOINT ["sh", "./docker-entrypoint.sh"]
