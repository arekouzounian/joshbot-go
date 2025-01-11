FROM tiangolo/uwsgi-nginx-flask:python3.12

COPY ./app /app

ENV LISTEN_PORT 8080
EXPOSE 8080

RUN pip install --no-cache-dir --upgrade -r /app/requirements.txt