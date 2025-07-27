FROM python:3.11-slim

WORKDIR /app

RUN apt-get update && apt-get install -y gcc libpq-dev netcat-openbsd && rm -rf /var/lib/apt/lists/*

COPY . /app/

RUN pip install --upgrade pip
RUN pip install django psycopg2-binary

ENV DJANGO_SETTINGS_MODULE=myproject.settings

RUN chmod +x /app/wait-for-it.sh

CMD ["/app/wait-for-it.sh", "db", "5432", "sh", "-c", "python manage.py migrate && python django_example/benchmark_scd.py"] 