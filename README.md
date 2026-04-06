# ProgWebPR6
Горбачов Ярослав ТВ-23

# Smart Grid CRUD API (MySQL + Go)

## Запуск MySQL у Docker
```bash
cd docker
docker build -t smartgrid-mysql .
docker run -d -p 3306:3306 --name smartgrid-db smartgrid-mysql
