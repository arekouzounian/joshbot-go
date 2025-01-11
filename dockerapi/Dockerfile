FROM tiangolo/uwsgi-nginx-flask:python3.12

COPY ./app /app

ENV LISTEN_PORT 6969
EXPOSE 6969

RUN pip install --no-cache-dir --upgrade -r /app/requirements.txt