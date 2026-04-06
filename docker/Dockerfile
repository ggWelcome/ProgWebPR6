# Використовуємо офіційний образ MySQL
FROM mysql:8.0

# Задаємо змінні середовища
ENV MYSQL_ROOT_PASSWORD=rootpass
ENV MYSQL_DATABASE=powerdb
ENV MYSQL_USER=poweruser
ENV MYSQL_PASSWORD=powerpass

# Копіюємо SQL-скрипт для ініціалізації
COPY init.sql /docker-entrypoint-initdb.d/
