# EAL Backend

Бэкенд-приложение команды ЭЭБО для хакатона BEST-HACK

## Конфигурация

Файл конфигурации `config.yaml` находится в `./internal/config/config.yaml`

## Запуск

1) Клонирование репозитория
    ```
    git clone git clone https://github.com/eeboAvitoLovers/eal-backend
    ```
2) Сборка Docker контейнера
    ```
    sudo docker build -t eal-backend .
    ```
3) Запуск контейнера
   ```
   sudo docker run -it -p 8080:8080 eal-backend
   ```
