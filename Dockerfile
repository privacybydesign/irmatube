FROM node:14 AS builder

RUN apt-get update && apt-get install -y \
    php \
    php-cli \
    php-zip \
    php-xml \
    php-mbstring \
    php-curl \
    php-sqlite3 \
    php-ldap \
    unzip \
    cron

RUN curl -sS https://getcomposer.org/installer | php -- --install-dir=/usr/local/bin --filename=composer
WORKDIR /app

COPY . .

RUN cd /app/www && npm install
RUN cd /app/www && composer install

FROM php:8.0-apache

COPY --from=builder /app/www /var/www/html
COPY --from=builder /app/data /app/data

RUN chown -R www-data:www-data /var/www/html \
    && chmod -R 755 /var/www/html


RUN echo "Listen 8080" >> /etc/apache2/ports.conf

EXPOSE 8080

