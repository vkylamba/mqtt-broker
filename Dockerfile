FROM python:3.8
ENV PYTHONUNBUFFERED 1
RUN apt-get -y update && apt-get -y install mosquitto
RUN mkdir /app
WORKDIR /app
COPY ./src /app/
COPY ./requirements.txt /app/requirements.txt
EXPOSE 9024
RUN pip install -r requirements.txt
