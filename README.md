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

## Frontend

![image](https://github.com/eeboAvitoLovers/eal-backend/assets/145232152/515a89fc-d80e-4ece-830d-f13bbc852bb6)

![image](https://github.com/eeboAvitoLovers/eal-backend/assets/145232152/d5b45e1a-c8df-4777-aefd-cd62cc8ea398)

![image](https://github.com/eeboAvitoLovers/eal-backend/assets/145232152/f620e150-d7fe-4242-be0a-e9e491b2739a)

## Machine Learning 

Пример кластеризации похожих сообщений

![image](https://github.com/eeboAvitoLovers/eal-backend/assets/145232152/360f7d18-5b16-4cd9-ae7b-894f8ad3fd7d)

## Database scheme

База данных PostgreSQL

![image](https://github.com/eeboAvitoLovers/eal-backend/assets/145232152/5c36547e-81f6-42fd-89ff-f5c52a5e26b2)

## TODO:
1) Протестировать и исправить работу backend составляющей
2) Автоматизировать кластеризацию новых записей в базе данных
3) Связать Frontend и Backend
