FROM node:14 AS build

WORKDIR /app

COPY . .

RUN cd /app/www && npm install

FROM php:8.0-apache

COPY --from=build /app/www /var/www/html
COPY --from=build /app/data /app/data

RUN chown -R www-data:www-data /var/www/html \
    && chmod -R 755 /var/www/html


RUN echo "Listen 8080" >> /etc/apache2/ports.conf

EXPOSE 8080

